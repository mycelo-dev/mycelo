package outbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mycelo-dev/mycelo/backend/internal/retrypolicy"
	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

const (
	idlePollInterval     = 500 * time.Millisecond
	consumerSyncInterval = 2 * time.Second

	defaultOutboundEventBatchLimit = 100
	defaultLeaseTTLMs              = int64(90_000)

	failureCategoryEventPayload   = "event_payload"
	failureCategoryEndpoint4xx    = "endpoint_response_4xx"
	failureCategoryEndpoint5xx    = "endpoint_response_5xx"
	failureCategoryEndpointOther  = "endpoint_response_other"
	failureCategoryEndpointClient = "endpoint_transport"
)

// EventReader provides cursor-based event reads for outbound consumers.
type EventReader interface {
	GetEventsAfterCursor(ctx context.Context, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error)
}

// MappingStore provides mapping state operations needed by outbound consumers.
type MappingStore interface {
	GetDestinationTopicMappings(ctx context.Context) ([]DestinationTopicMapping, error)
	GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (OutboundMappingState, error)
	UpdateOutboundMappingDeliveryState(ctx context.Context, destinationID string, topicID string, leaseHolder string, update DeliveryStateUpdate) error
	ClaimOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, leaseExpiresAtMillis int64) (bool, error)
	ReleaseOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string) error
	ApplyDeadLetterSkipInTx(ctx context.Context, leaseHolder string, insert DeadLetterEventInsert, update DeliveryStateUpdate) error
}

// ConsumerService orchestrates outbound consumer lifecycles and delivery retries.
type ConsumerService struct {
	store                MappingStore
	reader               EventReader
	deliveryClient       DeliveryClient
	instanceID           string
	leaseTTLMs           int64
	eventBatchLimit      int
	idlePollInterval     time.Duration
	consumerSyncInterval time.Duration
	retryDelay           func(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration
}

type poolStreamEventReader struct {
	limit int
}

func (r poolStreamEventReader) GetEventsAfterCursor(ctx context.Context, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error) {
	if limit <= 0 {
		limit = r.limit
	}

	return get_events.GetEventsAfterCursor(ctx, topic, after, offset, limit)
}

// NewConsumerService builds a consumer service with injected dependencies.
func NewConsumerService(store MappingStore, reader EventReader, deliveryClient DeliveryClient) *ConsumerService {
	return &ConsumerService{
		store:                store,
		reader:               reader,
		deliveryClient:       deliveryClient,
		instanceID:           outboundInstanceID(),
		leaseTTLMs:           outboundLeaseTTLMillis(),
		eventBatchLimit:      outboundEventBatchLimit(),
		idlePollInterval:     idlePollInterval,
		consumerSyncInterval: consumerSyncInterval,
		retryDelay: func(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
			return retrypolicy.ComputeDelayWithFullJitter(failureCount, baseDelay, maxDelay)
		},
	}
}

// NewDefaultConsumerService wires the default repository, reader, and HTTP delivery client.
func NewDefaultConsumerService() *ConsumerService {
	lim := outboundEventBatchLimit()

	return NewConsumerService(
		NewOutboundRepository(),
		poolStreamEventReader{limit: lim},
		NewHTTPDeliveryClient(httpClient),
	)
}

// StartConsumers starts outbound consumers using the default service wiring.
func StartConsumers(ctx context.Context) error {
	return NewDefaultConsumerService().Start(ctx)
}

// Start syncs destination-topic mappings and starts one consumer loop per mapping.
func (s *ConsumerService) Start(ctx context.Context) error {
	var consumersMu sync.Mutex
	consumers := make(map[string]context.CancelFunc)

	syncConsumers := func() error {
		mappings, err := s.store.GetDestinationTopicMappings(ctx)
		if err != nil {
			return err
		}

		activeKeys := make(map[string]struct{}, len(mappings))

		for _, mapping := range mappings {
			mapping := mapping
			key := mapping.DestinationID + ":" + mapping.TopicID
			activeKeys[key] = struct{}{}

			consumersMu.Lock()
			if _, exists := consumers[key]; exists {
				consumersMu.Unlock()
				continue
			}

			consumerCtx, cancel := context.WithCancel(ctx)
			consumers[key] = cancel
			consumersMu.Unlock()

			go func(m DestinationTopicMapping, k string) {
				defer func() {
					consumersMu.Lock()
					delete(consumers, k)
					consumersMu.Unlock()
				}()

				if err := s.consumeEvents(consumerCtx, m.DestinationID, m.TopicID, m.LastDeliveredEventID); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("outbound consumer stopped for destination %s and topic %s: %v", m.DestinationID, m.TopicID, err)
				}
			}(mapping, key)
		}

		consumersMu.Lock()
		defer consumersMu.Unlock()

		for key, cancel := range consumers {
			if _, exists := activeKeys[key]; exists {
				continue
			}

			cancel()
			delete(consumers, key)
		}

		return nil
	}

	if err := syncConsumers(); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(s.consumerSyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				consumersMu.Lock()
				for _, cancel := range consumers {
					cancel()
				}
				consumersMu.Unlock()

				return
			case <-ticker.C:
				if err := syncConsumers(); err != nil {
					log.Printf("failed to sync outbound consumers: %v", err)
				}
			}
		}
	}()

	return nil
}

// consumeEvents drains one mapping with at-least-once delivery semantics: HTTP success can be observed before the durable cursor commits.
func (s *ConsumerService) consumeEvents(ctx context.Context, destinationID string, topicID string, startOffset int64) error {
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer releaseCancel()

		if err := s.store.ReleaseOutboundDeliveryLease(releaseCtx, destinationID, topicID, s.instanceID); err != nil {
			log.Printf("outbound lease release destination %s topic %s: %v", destinationID, topicID, err)
		}
	}()

	cursor := startOffset

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		nowMillis := time.Now().UnixMilli()
		held, err := s.store.ClaimOutboundDeliveryLease(ctx, destinationID, topicID, s.instanceID, nowMillis, nowMillis+s.leaseTTLMs)
		if err != nil {
			return err
		}
		if !held {
			sleepWithContext(ctx, s.idlePollInterval)

			continue
		}

		state, err := s.store.GetOutboundMappingState(ctx, destinationID, topicID)
		if err != nil {
			return err
		}

		reconcileCursorWithDB(&cursor, state.LastDeliveredEventID)

		if !state.MappingExists || !state.DeliveryFlag || state.TopicName == "" || state.Endpoint == "" {
			sleepWithContext(ctx, s.idlePollInterval)

			continue
		}

		nowMillis = time.Now().UnixMilli()
		if state.NextAttemptAt > nowMillis {
			sleepWithContext(ctx, time.Duration(state.NextAttemptAt-nowMillis)*time.Millisecond)

			continue
		}

		resp, err := s.reader.GetEventsAfterCursor(ctx, state.TopicName, 0, cursor, s.eventBatchLimit)
		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			state, err = s.store.GetOutboundMappingState(ctx, destinationID, topicID)
			if err != nil {
				return err
			}

			reconcileCursorWithDB(&cursor, state.LastDeliveredEventID)

			if !state.MappingExists || !state.DeliveryFlag || state.Endpoint == "" {
				break
			}

			reclaimMillis := time.Now().UnixMilli()
			heldEv, err := s.store.ClaimOutboundDeliveryLease(ctx, destinationID, topicID, s.instanceID, reclaimMillis, reclaimMillis+s.leaseTTLMs)
			if err != nil {
				return err
			}
			if !heldEv {
				break
			}

			data, err := jsonMarshalEvent(event)
			if err != nil {
				log.Printf("failed to marshal event %d: %v", event.ID, err)
				failure := fmt.Errorf("marshal event %d: %w", event.ID, err)
				if s.shouldSkipFailure(state, failureCategoryEventPayload, state.ConsecutiveFailureCount+1) {
					if err := s.skipEvent(ctx, destinationID, topicID, state, event, event.EventData, failureCategoryEventPayload, failure); err != nil {
						if errors.Is(err, ErrOutboundLeaseLost) {
							break
						}
						return err
					}
					cursor = event.ID

					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategoryEventPayload, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
					if errors.Is(err, ErrOutboundLeaseLost) {
						break
					}

					return err
				}
				break
			}

			deliveryID := newDeliveryRequestID()

			meta := &WebhookDeliveryMeta{
				EventID:       event.ID,
				DeliveryID:    deliveryID,
				Attempt:       state.ConsecutiveFailureCount + 1,
				SentAtUnixMs:  time.Now().UnixMilli(),
				SigningSecret: state.WebhookSigningSecret,
				TopicName:     state.TopicName,
			}

			deliveryResult, err := s.deliveryClient.Deliver(ctx, state.Endpoint, data, meta)
			if err != nil {
				log.Printf("failed to deliver event %d to %s: %v", event.ID, state.Endpoint, err)
				failure := fmt.Errorf("deliver event %d: %w", event.ID, err)
				if s.shouldSkipFailure(state, failureCategoryEndpointClient, state.ConsecutiveFailureCount+1) {
					if err := s.skipEvent(ctx, destinationID, topicID, state, event, data, failureCategoryEndpointClient, failure); err != nil {
						if errors.Is(err, ErrOutboundLeaseLost) {
							break
						}
						return err
					}
					cursor = event.ID

					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategoryEndpointClient, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
					if errors.Is(err, ErrOutboundLeaseLost) {
						break
					}

					return err
				}
				break
			}

			if deliveryResult.StatusCode < 200 || deliveryResult.StatusCode >= 300 {
				log.Printf("failed to deliver event %d to %s: received status %d", event.ID, state.Endpoint, deliveryResult.StatusCode)
				failureCategory := classifyFailureCategoryFromStatus(deliveryResult.StatusCode)
				failure := fmt.Errorf("deliver event %d: received status %d", event.ID, deliveryResult.StatusCode)
				if s.shouldSkipFailure(state, failureCategory, state.ConsecutiveFailureCount+1) {
					if err := s.skipEvent(ctx, destinationID, topicID, state, event, data, failureCategory, failure); err != nil {
						if errors.Is(err, ErrOutboundLeaseLost) {
							break
						}
						return err
					}
					cursor = event.ID

					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategory, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
					if errors.Is(err, ErrOutboundLeaseLost) {
						break
					}

					return err
				}
				break
			}

			if err := s.recordDeliverySuccess(ctx, destinationID, topicID, event.ID, state.LastFailedAt, state.LastSkippedEventID, state.LastSkippedAt); err != nil {
				if errors.Is(err, ErrOutboundLeaseLost) {
					break
				}

				return err
			}

			cursor = event.ID
		}

		if resp.Count == 0 {
			sleepWithContext(ctx, s.idlePollInterval)
		}
	}
}

func jsonMarshalEvent(event get_events.Event) ([]byte, error) {
	return json.Marshal(event)
}

func reconcileCursorWithDB(cursor *int64, dbLastDeliveredEventID int64) {
	if dbLastDeliveredEventID > *cursor {
		*cursor = dbLastDeliveredEventID
	}
}

func (s *ConsumerService) recordDeliverySuccess(ctx context.Context, destinationID string, topicID string, deliveredEventID int64, lastFailedAt int64, lastSkippedEventID int64, lastSkippedAt int64) error {
	nowMillis := time.Now().UnixMilli()

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, s.instanceID, DeliveryStateUpdate{
		LastDeliveredEventID:    deliveredEventID,
		LastAttemptedEventID:    deliveredEventID,
		LastFailedEventID:       0,
		LastSkippedEventID:      lastSkippedEventID,
		ConsecutiveFailureCount: 0,
		LastAttemptedAt:         nowMillis,
		LastSucceededAt:         nowMillis,
		LastFailedAt:            lastFailedAt,
		LastSkippedAt:           lastSkippedAt,
		NextAttemptAt:           nowMillis,
		LastErrorCategory:       "",
		LastError:               "",
	})
}

func (s *ConsumerService) recordDeliveryFailure(ctx context.Context, destinationID string, topicID string, cursor int64, eventID int64, previousFailureCount int, lastSucceededAt int64, lastSkippedEventID int64, lastSkippedAt int64, failureCategory string, failure error, retryBaseDelayMs int64, retryMaxDelayMs int64) error {
	nowMillis := time.Now().UnixMilli()
	nextFailureCount := previousFailureCount + 1
	nextAttemptAt := nowMillis + s.retryDelay(nextFailureCount, time.Duration(retryBaseDelayMs)*time.Millisecond, time.Duration(retryMaxDelayMs)*time.Millisecond).Milliseconds()

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, s.instanceID, DeliveryStateUpdate{
		LastDeliveredEventID:    cursor,
		LastAttemptedEventID:    eventID,
		LastFailedEventID:       eventID,
		LastSkippedEventID:      lastSkippedEventID,
		ConsecutiveFailureCount: nextFailureCount,
		LastAttemptedAt:         nowMillis,
		LastSucceededAt:         lastSucceededAt,
		LastFailedAt:            nowMillis,
		LastSkippedAt:           lastSkippedAt,
		NextAttemptAt:           nextAttemptAt,
		LastErrorCategory:       failureCategory,
		LastError:               failure.Error(),
	})
}

func classifyFailureCategoryFromStatus(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		return failureCategoryEndpoint4xx
	case statusCode >= 500 && statusCode < 600:
		return failureCategoryEndpoint5xx
	default:
		return failureCategoryEndpointOther
	}
}

func (s *ConsumerService) shouldSkipFailure(state OutboundMappingState, failureCategory string, nextFailureCount int) bool {
	if state.MaxConsecutiveFailuresBeforeSkip <= 0 {
		return false
	}
	if nextFailureCount < state.MaxConsecutiveFailuresBeforeSkip {
		return false
	}

	switch failureCategory {
	case failureCategoryEventPayload:
		return state.SkipOnEventPayloadError
	case failureCategoryEndpoint4xx:
		return state.SkipOnEndpoint4xx
	case failureCategoryEndpoint5xx:
		return state.SkipOnEndpoint5xx
	case failureCategoryEndpointClient:
		return state.SkipOnEndpointTransportError
	default:
		return false
	}
}

func (s *ConsumerService) skipEvent(ctx context.Context, destinationID string, topicID string, state OutboundMappingState, event get_events.Event, eventPayload []byte, failureCategory string, failure error) error {
	nowMillis := time.Now().UnixMilli()
	nextFailureCount := state.ConsecutiveFailureCount + 1

	payload := eventPayload
	if len(payload) == 0 {
		payload = buildFallbackDeadLetterPayload(event)
	}

	update := DeliveryStateUpdate{
		LastDeliveredEventID:    event.ID,
		LastAttemptedEventID:    event.ID,
		LastFailedEventID:       0,
		LastSkippedEventID:      event.ID,
		ConsecutiveFailureCount: 0,
		LastAttemptedAt:         nowMillis,
		LastSucceededAt:         state.LastSucceededAt,
		LastFailedAt:            nowMillis,
		LastSkippedAt:           nowMillis,
		NextAttemptAt:           nowMillis,
		LastErrorCategory:       failureCategory,
		LastError:               failure.Error(),
	}

	if !state.DeadLetterQueueEnabled {
		return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, s.instanceID, update)
	}

	insert := DeadLetterEventInsert{
		DestinationID:   destinationID,
		TopicID:         topicID,
		SourceEventID:   event.ID,
		Endpoint:        state.Endpoint,
		FailureCategory: failureCategory,
		FailureReason:   failure.Error(),
		FailureCount:    nextFailureCount,
		EventPayload:    payload,
		DeadLetteredAt:  nowMillis,
	}

	return s.store.ApplyDeadLetterSkipInTx(ctx, s.instanceID, insert, update)
}

func buildFallbackDeadLetterPayload(event get_events.Event) []byte {
	if len(event.EventData) > 0 {
		return event.EventData
	}

	return []byte(`null`)
}

func sleepWithContext(ctx context.Context, duration time.Duration) {
	if duration <= 0 {
		return
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func newDeliveryRequestID() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	return hex.EncodeToString(raw)
}

var outboundInstanceOnce sync.Once
var outboundInstanceValue string

func outboundInstanceID() string {
	outboundInstanceOnce.Do(func() {
		if v := strings.TrimSpace(os.Getenv("OUTBOUND_INSTANCE_ID")); v != "" {
			outboundInstanceValue = v

			return
		}

		raw := make([]byte, 6)
		if _, err := rand.Read(raw); err != nil {
			outboundInstanceValue = "anon-" + strconv.FormatInt(time.Now().UnixNano(), 10)

			return
		}

		outboundInstanceValue = "anon-" + hex.EncodeToString(raw) + "-" + strconv.Itoa(os.Getpid())
	})

	return outboundInstanceValue
}

func deliveryLeaseMinimumTTLMillis() int64 {
	return int64(DefaultHTTPDeliveryTimeout.Milliseconds()) + 7000
}

func outboundLeaseTTLMillis() int64 {
	floor := deliveryLeaseMinimumTTLMillis()

	if v := strings.TrimSpace(os.Getenv("OUTBOUND_LEASE_TTL_MS")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			if n < floor {
				return floor
			}

			return n
		}
	}

	if defaultLeaseTTLMs < floor {
		return floor
	}

	return defaultLeaseTTLMs
}

func outboundEventBatchLimit() int {
	if v := strings.TrimSpace(os.Getenv("OUTBOUND_EVENT_BATCH_LIMIT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 2000 {
			return n
		}
	}

	return defaultOutboundEventBatchLimit
}

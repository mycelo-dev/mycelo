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
	failureCategoryCircuitOpen    = "endpoint_circuit_open"

	deliveryModeOrdered   = "ordered"
	deliveryModeUnordered = "unordered"
)

// EventReader provides cursor-based event reads for outbound consumers.
type EventReader interface {
	GetEventsAfterCursor(ctx context.Context, tenantPublicID string, teamPublicID string, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error)
}

// MappingStore provides mapping state operations needed by outbound consumers.
type MappingStore interface {
	GetDestinationTopicMappings(ctx context.Context) ([]DestinationTopicMapping, error)
	GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (OutboundMappingState, error)
	UpdateOutboundMappingDeliveryState(ctx context.Context, destinationID string, topicID string, leaseHolder string, update DeliveryStateUpdate) error
	RecordDeliveryFailureEvent(ctx context.Context, event DeliveryFailureEventInsert) error
	InsertDeadLetterEvent(ctx context.Context, event DeadLetterEventInsert) error
	EnqueueUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, events []UnorderedDeliveryEventInsert) error
	UpdateUnorderedEnqueueCursor(ctx context.Context, destinationID string, topicID string, leaseHolder string, cursor int64) error
	ClaimUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, lockExpiresAtMillis int64, limit int) ([]UnorderedDeliveryClaim, error)
	MarkUnorderedDeliveryDelivered(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, deliveredAt int64) error
	MarkUnorderedDeliveryFailed(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, nextAttemptAt int64, failureCategory string, failureReason string, failedAt int64) error
	MarkUnorderedDeliverySkipped(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, failureCategory string, failureReason string, skippedAt int64) error
	AdvanceUnorderedContiguousCursor(ctx context.Context, destinationID string, topicID string) error
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
	circuitBreaker       *EndpointCircuitBreaker
	adaptiveMu           sync.Mutex
	adaptiveUnordered    map[string]*adaptiveUnorderedState
}

type poolStreamEventReader struct {
	limit int
}

type adaptiveUnorderedState struct {
	targetInFlight int
}

type unorderedEnqueueResult struct {
	discovered int
	hasMore    bool
	cursor     int64
}

func (r poolStreamEventReader) GetEventsAfterCursor(ctx context.Context, tenantPublicID string, teamPublicID string, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error) {
	if limit <= 0 {
		limit = r.limit
	}

	return get_events.GetEventsAfterCursorForTenant(ctx, tenantPublicID, teamPublicID, topic, after, offset, limit)
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
		circuitBreaker:    newEndpointCircuitBreaker(outboundCircuitBreakerFailureThreshold(), outboundCircuitBreakerCooldown()),
		adaptiveUnordered: make(map[string]*adaptiveUnorderedState),
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

				if err := s.consumeEvents(consumerCtx, m.DestinationID, m.TopicID, m.TenantPublicID, m.TeamPublicID, m.LastDeliveredEventID); err != nil && !errors.Is(err, context.Canceled) {
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
func (s *ConsumerService) consumeEvents(ctx context.Context, destinationID string, topicID string, tenantPublicID string, teamPublicID string, startOffset int64) error {
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer releaseCancel()

		if err := s.store.ReleaseOutboundDeliveryLease(releaseCtx, destinationID, topicID, s.instanceID); err != nil {
			log.Printf("outbound lease release destination %s topic %s: %v", destinationID, topicID, err)
		}
		s.forgetAdaptiveUnorderedState(destinationID, topicID)
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
			if state.MappingExists && !state.DeliveryFlag {
				s.releaseMappingLease(ctx, destinationID, topicID)
			}
			sleepWithContext(ctx, s.idlePollInterval)

			continue
		}

		if state.DeliveryMode == deliveryModeUnordered {
			if err := s.consumeUnorderedEvents(ctx, destinationID, topicID, tenantPublicID, teamPublicID, state); err != nil {
				return err
			}

			continue
		}

		nowMillis = time.Now().UnixMilli()
		if state.NextAttemptAt > nowMillis {
			sleepWithContext(ctx, time.Duration(state.NextAttemptAt-nowMillis)*time.Millisecond)

			continue
		}

		resp, err := s.reader.GetEventsAfterCursor(ctx, tenantPublicID, teamPublicID, state.TopicName, 0, cursor, s.eventBatchLimit)
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
				s.captureDeliveryFailure(ctx, destinationID, topicID, state, event.ID, failureCategoryEventPayload, failure)
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

			if !s.circuitBreaker.Allow(state.Endpoint) {
				incrementOutboundMetricFor("circuit_blocked_total", "endpoint", state.Endpoint)
				if err := s.recordCircuitOpen(ctx, destinationID, topicID, cursor, event.ID, state); err != nil {
					if errors.Is(err, ErrOutboundLeaseLost) {
						break
					}

					return err
				}
				break
			}

			attemptStartedAt := time.Now()
			deliveryResult, err := s.deliveryClient.Deliver(ctx, state.Endpoint, data, meta)
			observeOutboundAttemptDuration(attemptStartedAt)
			if err != nil {
				s.circuitBreaker.RecordFailure(state.Endpoint)
				incrementOutboundMetricFor("delivery_failure_total", "category", failureCategoryEndpointClient)
				log.Printf("failed to deliver event %d to %s: %v", event.ID, state.Endpoint, err)
				failure := fmt.Errorf("deliver event %d: %w", event.ID, err)
				s.captureDeliveryFailure(ctx, destinationID, topicID, state, event.ID, failureCategoryEndpointClient, failure)
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
				failureCategory := classifyFailureCategoryFromStatus(deliveryResult.StatusCode)
				if shouldOpenCircuitForFailureCategory(failureCategory) {
					s.circuitBreaker.RecordFailure(state.Endpoint)
				}
				incrementOutboundMetricFor("delivery_failure_total", "category", failureCategory)
				log.Printf("failed to deliver event %d to %s: received status %d", event.ID, state.Endpoint, deliveryResult.StatusCode)
				failure := fmt.Errorf("deliver event %d: received status %d", event.ID, deliveryResult.StatusCode)
				s.captureDeliveryFailure(ctx, destinationID, topicID, state, event.ID, failureCategory, failure)
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

			s.circuitBreaker.RecordSuccess(state.Endpoint)
			incrementOutboundMetric("delivery_success_total")
			observeOutboundDeliverySuccess()
			observeOutboundDeliveryLag(event.CreatedAt)
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

func (s *ConsumerService) consumeUnorderedEvents(ctx context.Context, destinationID string, topicID string, tenantPublicID string, teamPublicID string, state OutboundMappingState) error {
	enqueued, err := s.enqueueUnorderedEvents(ctx, destinationID, topicID, tenantPublicID, teamPublicID, state)
	if err != nil {
		return err
	}

	limit := s.currentAdaptiveUnorderedLimit(destinationID, topicID, state.UnorderedMaxInFlight)

	nowMillis := time.Now().UnixMilli()
	claims, err := s.store.ClaimUnorderedDeliveryEvents(ctx, destinationID, topicID, s.instanceID, nowMillis, nowMillis+s.leaseTTLMs, limit)
	if err != nil {
		return err
	}

	if len(claims) == 0 {
		s.observeAdaptiveUnorderedPressure(destinationID, topicID, state.UnorderedMaxInFlight, false, 0, limit)
		sleepWithContext(ctx, s.idlePollInterval)
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(claims))
	for _, claim := range claims {
		claim := claim
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.deliverUnorderedClaim(ctx, destinationID, topicID, state, claim); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil && !errors.Is(err, ErrOutboundLeaseLost) {
			return err
		}
	}

	pressure := enqueued.hasMore || enqueued.discovered >= limit || len(claims) >= limit
	s.observeAdaptiveUnorderedPressure(destinationID, topicID, state.UnorderedMaxInFlight, pressure, len(claims), limit)

	return s.store.AdvanceUnorderedContiguousCursor(ctx, destinationID, topicID)
}

func (s *ConsumerService) enqueueUnorderedEvents(ctx context.Context, destinationID string, topicID string, tenantPublicID string, teamPublicID string, state OutboundMappingState) (unorderedEnqueueResult, error) {
	cursor := state.UnorderedLastEnqueuedEventID
	if cursor < state.LastDeliveredEventID {
		cursor = state.LastDeliveredEventID
	}

	resp, err := s.reader.GetEventsAfterCursor(ctx, tenantPublicID, teamPublicID, state.TopicName, 0, cursor, unorderedDiscoveryBatchLimit(s.eventBatchLimit, state.UnorderedMaxInFlight))
	if err != nil {
		return unorderedEnqueueResult{}, err
	}
	if resp.Count == 0 {
		return unorderedEnqueueResult{cursor: cursor}, nil
	}

	events := make([]UnorderedDeliveryEventInsert, 0, len(resp.Events))
	for _, event := range resp.Events {
		events = append(events, UnorderedDeliveryEventInsert{SourceEventID: event.ID})
	}

	if err := s.store.EnqueueUnorderedDeliveryEvents(ctx, destinationID, topicID, events); err != nil {
		return unorderedEnqueueResult{}, err
	}

	if err := s.store.UpdateUnorderedEnqueueCursor(ctx, destinationID, topicID, s.instanceID, resp.Cursor); err != nil {
		return unorderedEnqueueResult{}, err
	}

	return unorderedEnqueueResult{discovered: resp.Count, hasMore: resp.HasMore, cursor: resp.Cursor}, nil
}

func (s *ConsumerService) deliverUnorderedClaim(ctx context.Context, destinationID string, topicID string, state OutboundMappingState, claim UnorderedDeliveryClaim) error {
	event := get_events.Event{
		ID:        claim.SourceEventID,
		Topic:     claim.Topic,
		EventData: claim.EventPayload,
		CreatedAt: claim.CreatedAt,
	}

	data, err := jsonMarshalEvent(event)
	if err != nil {
		failure := fmt.Errorf("marshal event %d: %w", event.ID, err)
		s.captureUnorderedFailure(ctx, destinationID, topicID, state, claim, failureCategoryEventPayload, failure)
		return nil
	}

	deliveryID := newDeliveryRequestID()
	meta := &WebhookDeliveryMeta{
		EventID:       event.ID,
		DeliveryID:    deliveryID,
		Attempt:       claim.AttemptCount,
		SentAtUnixMs:  time.Now().UnixMilli(),
		SigningSecret: state.WebhookSigningSecret,
		TopicName:     state.TopicName,
	}

	if !s.circuitBreaker.Allow(state.Endpoint) {
		incrementOutboundMetricFor("circuit_blocked_total", "endpoint", state.Endpoint)
		failure := fmt.Errorf("endpoint circuit open for %s", state.Endpoint)
		s.captureUnorderedFailure(ctx, destinationID, topicID, state, claim, failureCategoryCircuitOpen, failure)
		return nil
	}

	attemptStartedAt := time.Now()
	deliveryResult, err := s.deliveryClient.Deliver(ctx, state.Endpoint, data, meta)
	observeOutboundAttemptDuration(attemptStartedAt)
	if err != nil {
		s.circuitBreaker.RecordFailure(state.Endpoint)
		incrementOutboundMetricFor("delivery_failure_total", "category", failureCategoryEndpointClient)
		failure := fmt.Errorf("deliver event %d: %w", event.ID, err)
		s.captureUnorderedFailure(ctx, destinationID, topicID, state, claim, failureCategoryEndpointClient, failure)
		return nil
	}

	if deliveryResult.StatusCode < 200 || deliveryResult.StatusCode >= 300 {
		failureCategory := classifyFailureCategoryFromStatus(deliveryResult.StatusCode)
		if shouldOpenCircuitForFailureCategory(failureCategory) {
			s.circuitBreaker.RecordFailure(state.Endpoint)
		}
		incrementOutboundMetricFor("delivery_failure_total", "category", failureCategory)
		failure := fmt.Errorf("deliver event %d: received status %d", event.ID, deliveryResult.StatusCode)
		s.captureUnorderedFailure(ctx, destinationID, topicID, state, claim, failureCategory, failure)
		return nil
	}

	s.circuitBreaker.RecordSuccess(state.Endpoint)
	incrementOutboundMetric("delivery_success_total")
	observeOutboundDeliverySuccess()
	observeOutboundDeliveryLag(event.CreatedAt)

	return s.store.MarkUnorderedDeliveryDelivered(ctx, claim.DeliveryID, destinationID, topicID, s.instanceID, claim.SourceEventID, time.Now().UnixMilli())
}

func (s *ConsumerService) captureUnorderedFailure(ctx context.Context, destinationID string, topicID string, state OutboundMappingState, claim UnorderedDeliveryClaim, failureCategory string, failure error) {
	nowMillis := time.Now().UnixMilli()
	failureCount := claim.AttemptCount
	if failureCount <= 0 {
		failureCount = 1
	}

	insert := DeliveryFailureEventInsert{
		DestinationID:   destinationID,
		TopicID:         topicID,
		SourceEventID:   claim.SourceEventID,
		Endpoint:        state.Endpoint,
		FailureCategory: failureCategory,
		FailureReason:   failure.Error(),
		FailureCount:    failureCount,
		FailedAt:        nowMillis,
	}
	if err := s.store.RecordDeliveryFailureEvent(ctx, insert); err != nil {
		log.Printf("failed to record unordered delivery failure for event %d to %s: %v", claim.SourceEventID, state.Endpoint, err)
	}

	if s.shouldSkipFailure(state, failureCategory, failureCount) {
		if err := s.skipUnorderedClaim(ctx, destinationID, topicID, state, claim, failureCategory, failure, nowMillis); err != nil && !errors.Is(err, ErrOutboundLeaseLost) {
			log.Printf("failed to skip unordered delivery for event %d to %s: %v", claim.SourceEventID, state.Endpoint, err)
		}
		return
	}

	nextAttemptAt := nowMillis + s.retryDelay(failureCount, time.Duration(state.RetryBaseDelayMs)*time.Millisecond, time.Duration(state.RetryMaxDelayMs)*time.Millisecond).Milliseconds()
	if err := s.store.MarkUnorderedDeliveryFailed(ctx, claim.DeliveryID, destinationID, topicID, s.instanceID, claim.SourceEventID, nextAttemptAt, failureCategory, failure.Error(), nowMillis); err != nil && !errors.Is(err, ErrOutboundLeaseLost) {
		log.Printf("failed to record unordered delivery retry for event %d to %s: %v", claim.SourceEventID, state.Endpoint, err)
	}
}

func (s *ConsumerService) skipUnorderedClaim(ctx context.Context, destinationID string, topicID string, state OutboundMappingState, claim UnorderedDeliveryClaim, failureCategory string, failure error, nowMillis int64) error {
	if state.DeadLetterQueueEnabled {
		incrementOutboundMetric("dead_letter_write_total")
		insert := DeadLetterEventInsert{
			DestinationID:   destinationID,
			TopicID:         topicID,
			SourceEventID:   claim.SourceEventID,
			Endpoint:        state.Endpoint,
			FailureCategory: failureCategory,
			FailureReason:   failure.Error(),
			FailureCount:    claim.AttemptCount,
			EventPayload:    buildFallbackDeadLetterPayload(get_events.Event{EventData: claim.EventPayload}),
			DeadLetteredAt:  nowMillis,
		}
		if err := s.store.InsertDeadLetterEvent(ctx, insert); err != nil {
			return err
		}
	}

	return s.store.MarkUnorderedDeliverySkipped(ctx, claim.DeliveryID, destinationID, topicID, s.instanceID, claim.SourceEventID, failureCategory, failure.Error(), nowMillis)
}

func (s *ConsumerService) releaseMappingLease(ctx context.Context, destinationID string, topicID string) {
	if err := s.store.ReleaseOutboundDeliveryLease(ctx, destinationID, topicID, s.instanceID); err != nil {
		log.Printf("outbound lease release destination %s topic %s: %v", destinationID, topicID, err)
	}
}

func (s *ConsumerService) captureDeliveryFailure(ctx context.Context, destinationID string, topicID string, state OutboundMappingState, eventID int64, failureCategory string, failure error) {
	nowMillis := time.Now().UnixMilli()
	insert := DeliveryFailureEventInsert{
		DestinationID:   destinationID,
		TopicID:         topicID,
		SourceEventID:   eventID,
		Endpoint:        state.Endpoint,
		FailureCategory: failureCategory,
		FailureReason:   failure.Error(),
		FailureCount:    state.ConsecutiveFailureCount + 1,
		FailedAt:        nowMillis,
	}

	if err := s.store.RecordDeliveryFailureEvent(ctx, insert); err != nil {
		log.Printf("failed to record delivery failure for event %d to %s: %v", eventID, state.Endpoint, err)
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

func (s *ConsumerService) recordCircuitOpen(ctx context.Context, destinationID string, topicID string, cursor int64, eventID int64, state OutboundMappingState) error {
	nowMillis := time.Now().UnixMilli()
	remaining := s.circuitBreaker.RemainingCooldown(state.Endpoint).Milliseconds()
	if remaining < int64(s.idlePollInterval.Milliseconds()) {
		remaining = int64(s.idlePollInterval.Milliseconds())
	}

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, s.instanceID, DeliveryStateUpdate{
		LastDeliveredEventID:    cursor,
		LastAttemptedEventID:    eventID,
		LastFailedEventID:       eventID,
		LastSkippedEventID:      state.LastSkippedEventID,
		ConsecutiveFailureCount: state.ConsecutiveFailureCount,
		LastAttemptedAt:         nowMillis,
		LastSucceededAt:         state.LastSucceededAt,
		LastFailedAt:            nowMillis,
		LastSkippedAt:           state.LastSkippedAt,
		NextAttemptAt:           nowMillis + remaining,
		LastErrorCategory:       failureCategoryCircuitOpen,
		LastError:               fmt.Sprintf("endpoint circuit open for %s", state.Endpoint),
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

func shouldOpenCircuitForFailureCategory(failureCategory string) bool {
	switch failureCategory {
	case failureCategoryEndpoint5xx, failureCategoryEndpointOther, failureCategoryEndpointClient:
		return true
	default:
		return false
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

	incrementOutboundMetric("dead_letter_write_total")

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

func (s *ConsumerService) currentAdaptiveUnorderedLimit(destinationID string, topicID string, configuredCap int) int {
	capacity := normalizeUnorderedInFlightCap(configuredCap)
	key := adaptiveUnorderedKey(destinationID, topicID)

	s.adaptiveMu.Lock()
	defer s.adaptiveMu.Unlock()

	state := s.adaptiveUnordered[key]
	if state == nil {
		state = &adaptiveUnorderedState{targetInFlight: 1}
		s.adaptiveUnordered[key] = state
	}
	if state.targetInFlight < 1 {
		state.targetInFlight = 1
	}
	if state.targetInFlight > capacity {
		state.targetInFlight = capacity
	}

	return state.targetInFlight
}

func (s *ConsumerService) observeAdaptiveUnorderedPressure(destinationID string, topicID string, configuredCap int, pressure bool, claimed int, previousLimit int) {
	capacity := normalizeUnorderedInFlightCap(configuredCap)
	key := adaptiveUnorderedKey(destinationID, topicID)

	s.adaptiveMu.Lock()
	defer s.adaptiveMu.Unlock()

	state := s.adaptiveUnordered[key]
	if state == nil {
		state = &adaptiveUnorderedState{targetInFlight: 1}
		s.adaptiveUnordered[key] = state
	}

	target := state.targetInFlight
	if target < 1 {
		target = 1
	}
	if target > capacity {
		target = capacity
	}

	switch {
	case pressure && claimed >= previousLimit:
		if target < 4 {
			target *= 2
		} else {
			target += maxInt(1, target/2)
		}
	case !pressure && claimed == 0:
		target = maxInt(1, target/2)
	case !pressure && claimed < previousLimit/2:
		target--
	}

	state.targetInFlight = clampInt(target, 1, capacity)
}

func (s *ConsumerService) forgetAdaptiveUnorderedState(destinationID string, topicID string) {
	s.adaptiveMu.Lock()
	defer s.adaptiveMu.Unlock()

	delete(s.adaptiveUnordered, adaptiveUnorderedKey(destinationID, topicID))
}

func adaptiveUnorderedKey(destinationID string, topicID string) string {
	return destinationID + ":" + topicID
}

func normalizeUnorderedInFlightCap(value int) int {
	if value <= 0 {
		return 1
	}

	return value
}

func unorderedDiscoveryBatchLimit(baseLimit int, maxInFlight int) int {
	limit := baseLimit
	if limit <= 0 {
		limit = defaultOutboundEventBatchLimit
	}

	concurrencySizedLimit := normalizeUnorderedInFlightCap(maxInFlight) * 8
	if concurrencySizedLimit > limit {
		limit = concurrencySizedLimit
	}
	if limit > get_events.MaxEventsFetchUpperBound {
		return get_events.MaxEventsFetchUpperBound
	}

	return limit
}

func clampInt(value int, floor int, ceiling int) int {
	if value < floor {
		return floor
	}
	if value > ceiling {
		return ceiling
	}

	return value
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}

	return right
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

		if isTruthyEnv("OUTBOUND_REQUIRE_INSTANCE_ID") {
			log.Fatal("OUTBOUND_REQUIRE_INSTANCE_ID is enabled but OUTBOUND_INSTANCE_ID is empty")
		}

		log.Print("OUTBOUND_INSTANCE_ID is empty; using an ephemeral outbound lease holder id. Set OUTBOUND_INSTANCE_ID in multi-instance deployments.")

		raw := make([]byte, 6)
		if _, err := rand.Read(raw); err != nil {
			outboundInstanceValue = "anon-" + strconv.FormatInt(time.Now().UnixNano(), 10)

			return
		}

		outboundInstanceValue = "anon-" + hex.EncodeToString(raw) + "-" + strconv.Itoa(os.Getpid())
	})

	return outboundInstanceValue
}

func isTruthyEnv(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
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

package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mycelo-dev/mycelo/backend/internal/retrypolicy"
	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

const (
	idlePollInterval     = 500 * time.Millisecond
	consumerSyncInterval = 2 * time.Second

	failureCategoryEventPayload   = "event_payload"
	failureCategoryEndpoint4xx    = "endpoint_response_4xx"
	failureCategoryEndpoint5xx    = "endpoint_response_5xx"
	failureCategoryEndpointOther  = "endpoint_response_other"
	failureCategoryEndpointClient = "endpoint_transport"
)

// EventReader provides cursor-based event reads for outbound consumers.
type EventReader interface {
	GetEventsAfterCursor(ctx context.Context, topic string, after int64, offset int64) (get_events.EventsResponse, error)
}

// MappingStore provides the mapping state operations needed by outbound consumers.
type MappingStore interface {
	GetDestinationTopicMappings(ctx context.Context) ([]DestinationTopicMapping, error)
	GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (OutboundMappingState, error)
	UpdateOutboundMappingDeliveryState(ctx context.Context, destinationID string, topicID string, update DeliveryStateUpdate) error
	InsertDeadLetterEvent(ctx context.Context, event DeadLetterEventInsert) error
}

// DeliveryClient abstracts the transport used to deliver outbound payloads.
type DeliveryClient interface {
	Deliver(ctx context.Context, endpoint string, data []byte) (DeliveryResult, error)
}

// ConsumerService orchestrates outbound consumer lifecycles and delivery retries.
type ConsumerService struct {
	store                MappingStore
	reader               EventReader
	deliveryClient       DeliveryClient
	idlePollInterval     time.Duration
	consumerSyncInterval time.Duration
	retryDelay           func(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration
}

type streamEventReader struct{}

func (streamEventReader) GetEventsAfterCursor(ctx context.Context, topic string, after int64, offset int64) (get_events.EventsResponse, error) {
	return get_events.GetEventsAfterCursor(ctx, topic, after, offset)
}

// NewConsumerService builds a consumer service with injected dependencies.
func NewConsumerService(store MappingStore, reader EventReader, deliveryClient DeliveryClient) *ConsumerService {
	return &ConsumerService{
		store:                store,
		reader:               reader,
		deliveryClient:       deliveryClient,
		idlePollInterval:     idlePollInterval,
		consumerSyncInterval: consumerSyncInterval,
		retryDelay: func(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
			return retrypolicy.ComputeDelayWithFullJitter(failureCount, baseDelay, maxDelay)
		},
	}
}

// NewDefaultConsumerService wires the default repository, reader, and HTTP delivery client.
func NewDefaultConsumerService() *ConsumerService {
	return NewConsumerService(
		NewOutboundRepository(),
		streamEventReader{},
		NewHTTPDeliveryClient(httpClient),
	)
}

// StartConsumers starts outbound consumers using the default service wiring.
func StartConsumers(ctx context.Context) error {
	return NewDefaultConsumerService().Start(ctx)
}

// Start syncs destination-topic mappings and starts one consumer loop per mapping.
func (s *ConsumerService) Start(ctx context.Context) error {
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

			if _, exists := consumers[key]; exists {
				continue
			}

			consumerCtx, cancel := context.WithCancel(ctx)
			consumers[key] = cancel

			go func() {
				if err := s.consumeEvents(consumerCtx, mapping.DestinationID, mapping.TopicID, mapping.LastDeliveredEventID); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("outbound consumer stopped for destination %s and topic %s: %v", mapping.DestinationID, mapping.TopicID, err)
				}
			}()
		}

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
				for _, cancel := range consumers {
					cancel()
				}
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

// Each mapping gets its own consumer so the cursor stays isolated per topic/destination pair.
func (s *ConsumerService) consumeEvents(ctx context.Context, destinationID string, topicID string, startOffset int64) error {
	cursor := startOffset

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		state, err := s.store.GetOutboundMappingState(ctx, destinationID, topicID)
		if err != nil {
			return err
		}

		if !state.MappingExists || !state.DeliveryFlag || state.TopicName == "" || state.Endpoint == "" {
			sleepWithContext(ctx, s.idlePollInterval)
			continue
		}

		nowMillis := time.Now().UnixMilli()
		if state.NextAttemptAt > nowMillis {
			sleepWithContext(ctx, time.Duration(state.NextAttemptAt-nowMillis)*time.Millisecond)
			continue
		}

		resp, err := s.reader.GetEventsAfterCursor(ctx, state.TopicName, 0, cursor)
		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			state, err = s.store.GetOutboundMappingState(ctx, destinationID, topicID)
			if err != nil {
				return err
			}

			if !state.MappingExists || !state.DeliveryFlag || state.Endpoint == "" {
				break
			}

			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal event %d: %v", event.ID, err)
				failure := fmt.Errorf("marshal event %d: %w", event.ID, err)
				if s.shouldSkipFailure(state, failureCategoryEventPayload, state.ConsecutiveFailureCount+1) {
					if err := s.skipEvent(ctx, destinationID, topicID, state, event, event.EventData, failureCategoryEventPayload, failure); err != nil {
						return err
					}
					cursor = event.ID
					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategoryEventPayload, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
					return err
				}
				break
			}

			deliveryResult, err := s.deliveryClient.Deliver(ctx, state.Endpoint, data)
			if err != nil {
				log.Printf("failed to deliver event %d to %s: %v", event.ID, state.Endpoint, err)
				failure := fmt.Errorf("deliver event %d: %w", event.ID, err)
				if s.shouldSkipFailure(state, failureCategoryEndpointClient, state.ConsecutiveFailureCount+1) {
					if err := s.skipEvent(ctx, destinationID, topicID, state, event, data, failureCategoryEndpointClient, failure); err != nil {
						return err
					}
					cursor = event.ID
					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategoryEndpointClient, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
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
						return err
					}
					cursor = event.ID
					continue
				}
				if err := s.recordDeliveryFailure(ctx, destinationID, topicID, cursor, event.ID, state.ConsecutiveFailureCount, state.LastSucceededAt, state.LastSkippedEventID, state.LastSkippedAt, failureCategory, failure, state.RetryBaseDelayMs, state.RetryMaxDelayMs); err != nil {
					return err
				}
				break
			}

			cursor = event.ID

			if err := s.recordDeliverySuccess(ctx, destinationID, topicID, cursor, event.ID, state.LastFailedAt, state.LastSkippedEventID, state.LastSkippedAt); err != nil {
				return err
			}
		}

		if resp.Count == 0 {
			sleepWithContext(ctx, s.idlePollInterval)
		}
	}
}

func (s *ConsumerService) recordDeliverySuccess(ctx context.Context, destinationID string, topicID string, cursor int64, eventID int64, lastFailedAt int64, lastSkippedEventID int64, lastSkippedAt int64) error {
	nowMillis := time.Now().UnixMilli()

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, DeliveryStateUpdate{
		LastDeliveredEventID:    cursor,
		LastAttemptedEventID:    eventID,
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

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, DeliveryStateUpdate{
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

	if state.DeadLetterQueueEnabled {
		payload := eventPayload
		if len(payload) == 0 {
			payload = buildFallbackDeadLetterPayload(event)
		}

		if err := s.store.InsertDeadLetterEvent(ctx, DeadLetterEventInsert{
			DestinationID:   destinationID,
			TopicID:         topicID,
			SourceEventID:   event.ID,
			Endpoint:        state.Endpoint,
			FailureCategory: failureCategory,
			FailureReason:   failure.Error(),
			FailureCount:    nextFailureCount,
			EventPayload:    payload,
			DeadLetteredAt:  nowMillis,
		}); err != nil {
			return err
		}
	}

	return s.store.UpdateOutboundMappingDeliveryState(ctx, destinationID, topicID, DeliveryStateUpdate{
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
	})
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

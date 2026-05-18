package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelo-dev/mycelo/backend/internal/retrypolicy"
	"github.com/mycelo-dev/mycelo/backend/outbound"
	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

type consumerTestStore struct {
	mu                  sync.Mutex
	mapping             outbound.DestinationTopicMapping
	state               outbound.OutboundMappingState
	lastUpdate          outbound.DeliveryStateUpdate
	lastDeadLetterEvent outbound.DeadLetterEventInsert
	lastFailureEvent    outbound.DeliveryFailureEventInsert
	unorderedClaims     []outbound.UnorderedDeliveryClaim
	deliveredUnordered  []int64
	claimLimits         []int
	enqueuedCursor      int64
	updateCh            chan struct{}
	cancel              context.CancelFunc
}

func (s *consumerTestStore) GetDestinationTopicMappings(ctx context.Context) ([]outbound.DestinationTopicMapping, error) {
	return []outbound.DestinationTopicMapping{s.mapping}, nil
}

func (s *consumerTestStore) GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (outbound.OutboundMappingState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state, nil
}

func (s *consumerTestStore) UpdateOutboundMappingDeliveryState(ctx context.Context, destinationID string, topicID string, leaseHolder string, update outbound.DeliveryStateUpdate) error {
	s.mu.Lock()
	s.lastUpdate = update
	s.state.LastDeliveredEventID = update.LastDeliveredEventID
	s.state.LastAttemptedEventID = update.LastAttemptedEventID
	s.state.LastFailedEventID = update.LastFailedEventID
	s.state.LastSkippedEventID = update.LastSkippedEventID
	s.state.ConsecutiveFailureCount = update.ConsecutiveFailureCount
	s.state.LastSucceededAt = update.LastSucceededAt
	s.state.LastFailedAt = update.LastFailedAt
	s.state.LastSkippedAt = update.LastSkippedAt
	s.state.NextAttemptAt = update.NextAttemptAt
	s.state.LastErrorCategory = update.LastErrorCategory
	s.state.LastError = update.LastError
	s.mu.Unlock()

	select {
	case s.updateCh <- struct{}{}:
	default:
	}

	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *consumerTestStore) InsertDeadLetterEvent(ctx context.Context, event outbound.DeadLetterEventInsert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastDeadLetterEvent = event
	return nil
}

func (s *consumerTestStore) RecordDeliveryFailureEvent(ctx context.Context, event outbound.DeliveryFailureEventInsert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastFailureEvent = event
	return nil
}

func (s *consumerTestStore) EnqueueUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, events []outbound.UnorderedDeliveryEventInsert) error {
	return nil
}

func (s *consumerTestStore) UpdateUnorderedEnqueueCursor(ctx context.Context, destinationID string, topicID string, leaseHolder string, cursor int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enqueuedCursor = cursor
	return nil
}

func (s *consumerTestStore) ClaimUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, lockExpiresAtMillis int64, limit int) ([]outbound.UnorderedDeliveryClaim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.claimLimits = append(s.claimLimits, limit)
	if len(s.unorderedClaims) == 0 {
		return nil, nil
	}
	if limit > len(s.unorderedClaims) {
		limit = len(s.unorderedClaims)
	}

	claims := append([]outbound.UnorderedDeliveryClaim(nil), s.unorderedClaims[:limit]...)
	s.unorderedClaims = s.unorderedClaims[limit:]
	return claims, nil
}

func (s *consumerTestStore) MarkUnorderedDeliveryDelivered(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, deliveredAt int64) error {
	s.mu.Lock()
	s.deliveredUnordered = append(s.deliveredUnordered, sourceEventID)
	s.mu.Unlock()

	select {
	case s.updateCh <- struct{}{}:
	default:
	}

	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *consumerTestStore) MarkUnorderedDeliveryFailed(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, nextAttemptAt int64, failureCategory string, failureReason string, failedAt int64) error {
	return nil
}

func (s *consumerTestStore) MarkUnorderedDeliverySkipped(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, failureCategory string, failureReason string, skippedAt int64) error {
	return nil
}

func (s *consumerTestStore) AdvanceUnorderedContiguousCursor(ctx context.Context, destinationID string, topicID string) error {
	return nil
}

func (s *consumerTestStore) ClaimOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, leaseExpiresAtMillis int64) (bool, error) {
	return true, nil
}

func (s *consumerTestStore) ReleaseOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string) error {
	return nil
}

func (s *consumerTestStore) ApplyDeadLetterSkipInTx(ctx context.Context, leaseHolder string, insert outbound.DeadLetterEventInsert, update outbound.DeliveryStateUpdate) error {
	if err := s.InsertDeadLetterEvent(ctx, insert); err != nil {
		return err
	}

	return s.UpdateOutboundMappingDeliveryState(ctx, insert.DestinationID, insert.TopicID, leaseHolder, update)
}

type staticEventReader struct {
	response get_events.EventsResponse
}

func (r staticEventReader) GetEventsAfterCursor(ctx context.Context, tenantPublicID string, teamPublicID string, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error) {
	return r.response, nil
}

type staticDeliveryClient struct {
	result outbound.DeliveryResult
	err    error
}

func (c staticDeliveryClient) Deliver(ctx context.Context, endpoint string, data []byte, meta *outbound.WebhookDeliveryMeta) (outbound.DeliveryResult, error) {
	if c.err != nil {
		return outbound.DeliveryResult{}, c.err
	}

	return c.result, nil
}

func TestConsumerRetriesAndRecordsBlockingEvent(t *testing.T) {
	originalRandomDurationUpTo := retrypolicy.RandomDurationUpTo
	t.Cleanup(func() {
		retrypolicy.RandomDurationUpTo = originalRandomDurationUpTo
	})

	var gotMaxDelay time.Duration
	retrypolicy.RandomDurationUpTo = func(maxDelay time.Duration) time.Duration {
		gotMaxDelay = maxDelay
		return 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			LastDeliveredEventID:             41,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 0,
			LastSucceededAt:                  1234,
			LastSkippedEventID:               77,
			LastSkippedAt:                    88,
			MappingExists:                    true,
		},
		updateCh: make(chan struct{}, 1),
		cancel:   cancel,
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{
			response: get_events.EventsResponse{
				Events: []get_events.Event{
					{ID: 42, Topic: "orders.created", EventData: []byte(`{"order_id":"ord_123"}`), CreatedAt: 123456},
				},
				Count:  1,
				Cursor: 42,
			},
		},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 400}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	select {
	case <-store.updateCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery state update")
	}

	update := store.lastUpdate
	if update.LastDeliveredEventID != 41 {
		t.Fatalf("LastDeliveredEventID = %d, want %d", update.LastDeliveredEventID, 41)
	}
	if update.LastAttemptedEventID != 42 {
		t.Fatalf("LastAttemptedEventID = %d, want %d", update.LastAttemptedEventID, 42)
	}
	if update.LastFailedEventID != 42 {
		t.Fatalf("LastFailedEventID = %d, want %d", update.LastFailedEventID, 42)
	}
	if update.LastSkippedEventID != 77 {
		t.Fatalf("LastSkippedEventID = %d, want %d", update.LastSkippedEventID, 77)
	}
	if update.ConsecutiveFailureCount != 1 {
		t.Fatalf("ConsecutiveFailureCount = %d, want %d", update.ConsecutiveFailureCount, 1)
	}
	if update.LastSucceededAt != 1234 {
		t.Fatalf("LastSucceededAt = %d, want %d", update.LastSucceededAt, 1234)
	}
	if update.LastSkippedAt != 88 {
		t.Fatalf("LastSkippedAt = %d, want %d", update.LastSkippedAt, 88)
	}
	if update.LastErrorCategory != "endpoint_response_4xx" {
		t.Fatalf("LastErrorCategory = %q, want %q", update.LastErrorCategory, "endpoint_response_4xx")
	}
	if update.LastError != "deliver event 42: received status 400" {
		t.Fatalf("LastError = %q, want %q", update.LastError, "deliver event 42: received status 400")
	}

	failureEvent := store.lastFailureEvent
	if failureEvent.SourceEventID != 42 {
		t.Fatalf("failure SourceEventID = %d, want %d", failureEvent.SourceEventID, 42)
	}
	if failureEvent.FailureCategory != "endpoint_response_4xx" {
		t.Fatalf("failure FailureCategory = %q, want %q", failureEvent.FailureCategory, "endpoint_response_4xx")
	}
	if failureEvent.FailureReason != "deliver event 42: received status 400" {
		t.Fatalf("failure FailureReason = %q, want %q", failureEvent.FailureReason, "deliver event 42: received status 400")
	}
	if failureEvent.FailureCount != 1 {
		t.Fatalf("failure FailureCount = %d, want %d", failureEvent.FailureCount, 1)
	}
	if gotMaxDelay != 2*time.Second {
		t.Fatalf("RandomDurationUpTo received %s, want %s", gotMaxDelay, 2*time.Second)
	}
	if update.NextAttemptAt-update.LastAttemptedAt != int64((5 * time.Second).Milliseconds()) {
		t.Fatalf("retry delay = %dms, want %dms", update.NextAttemptAt-update.LastAttemptedAt, (5 * time.Second).Milliseconds())
	}
}

func TestConsumerRecordsSuccessAndClearsFailureState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			LastDeliveredEventID:             41,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 0,
			LastFailedAt:                     1111,
			LastSkippedEventID:               12,
			LastSkippedAt:                    2222,
			MappingExists:                    true,
		},
		updateCh: make(chan struct{}, 1),
		cancel:   cancel,
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{
			response: get_events.EventsResponse{
				Events: []get_events.Event{
					{ID: 42, Topic: "orders.created", EventData: []byte(`{"order_id":"ord_123"}`), CreatedAt: 123456},
				},
				Count:  1,
				Cursor: 42,
			},
		},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 202}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	select {
	case <-store.updateCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery state update")
	}

	update := store.lastUpdate
	if update.LastDeliveredEventID != 42 {
		t.Fatalf("LastDeliveredEventID = %d, want %d", update.LastDeliveredEventID, 42)
	}
	if update.LastAttemptedEventID != 42 {
		t.Fatalf("LastAttemptedEventID = %d, want %d", update.LastAttemptedEventID, 42)
	}
	if update.LastFailedEventID != 0 {
		t.Fatalf("LastFailedEventID = %d, want %d", update.LastFailedEventID, 0)
	}
	if update.LastSkippedEventID != 12 {
		t.Fatalf("LastSkippedEventID = %d, want %d", update.LastSkippedEventID, 12)
	}
	if update.ConsecutiveFailureCount != 0 {
		t.Fatalf("ConsecutiveFailureCount = %d, want %d", update.ConsecutiveFailureCount, 0)
	}
	if update.LastFailedAt != 1111 {
		t.Fatalf("LastFailedAt = %d, want %d", update.LastFailedAt, 1111)
	}
	if update.LastSkippedAt != 2222 {
		t.Fatalf("LastSkippedAt = %d, want %d", update.LastSkippedAt, 2222)
	}
	if update.LastErrorCategory != "" {
		t.Fatalf("LastErrorCategory = %q, want empty string", update.LastErrorCategory)
	}
	if update.LastError != "" {
		t.Fatalf("LastError = %q, want empty string", update.LastError)
	}
}

func TestConsumerUnorderedModeClaimsAndDeliversEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			DeliveryMode:                     "unordered",
			LastDeliveredEventID:             41,
			UnorderedLastEnqueuedEventID:     43,
			UnorderedMaxInFlight:             2,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 0,
			MappingExists:                    true,
		},
		unorderedClaims: []outbound.UnorderedDeliveryClaim{
			{DeliveryID: 1, SourceEventID: 42, AttemptCount: 1, Topic: "orders.created", EventPayload: []byte(`{"order_id":"ord_42"}`), CreatedAt: 123456},
			{DeliveryID: 2, SourceEventID: 43, AttemptCount: 1, Topic: "orders.created", EventPayload: []byte(`{"order_id":"ord_43"}`), CreatedAt: 123457},
		},
		updateCh: make(chan struct{}, 2),
		cancel:   cancel,
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{response: get_events.EventsResponse{Count: 0, Cursor: 43}},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 202}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	select {
	case <-store.updateCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for unordered delivery")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.deliveredUnordered) == 0 {
		t.Fatal("expected at least one unordered event to be marked delivered")
	}
	if store.deliveredUnordered[0] != 42 && store.deliveredUnordered[0] != 43 {
		t.Fatalf("delivered event = %d, want 42 or 43", store.deliveredUnordered[0])
	}
}

func TestConsumerUnorderedModeRampsClaimLimitUnderBacklog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	claims := make([]outbound.UnorderedDeliveryClaim, 0, 8)
	for i := int64(42); i < 50; i++ {
		claims = append(claims, outbound.UnorderedDeliveryClaim{
			DeliveryID:    i,
			SourceEventID: i,
			AttemptCount:  1,
			Topic:         "orders.created",
			EventPayload:  []byte(`{"order_id":"ord"}`),
			CreatedAt:     123456,
		})
	}

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                    "orders.created",
			Endpoint:                     "https://example.com/webhook",
			DeliveryFlag:                 true,
			DeliveryMode:                 "unordered",
			LastDeliveredEventID:         41,
			UnorderedLastEnqueuedEventID: 49,
			UnorderedMaxInFlight:         4,
			RetryBaseDelayMs:             2000,
			RetryMaxDelayMs:              60000,
			MappingExists:                true,
		},
		unorderedClaims: claims,
		updateCh:        make(chan struct{}, 8),
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{response: get_events.EventsResponse{Count: 0, Cursor: 49}},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 202}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	for i := 0; i < 3; i++ {
		select {
		case <-store.updateCh:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for unordered delivery ramp")
		}
	}
	cancel()

	store.mu.Lock()
	limits := append([]int(nil), store.claimLimits...)
	store.mu.Unlock()

	if len(limits) < 2 {
		t.Fatalf("claim limits = %v, want at least two claim cycles", limits)
	}
	if limits[0] != 1 {
		t.Fatalf("first claim limit = %d, want 1", limits[0])
	}
	if limits[1] <= limits[0] {
		t.Fatalf("claim limits did not ramp: %v", limits)
	}
	if limits[1] > 4 {
		t.Fatalf("claim limit exceeded configured cap: %v", limits)
	}
}

func TestConsumerSkipsConfiguredFailuresAndWritesDeadLetterEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			LastDeliveredEventID:             41,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 3,
			SkipOnEndpoint4xx:                true,
			DeadLetterQueueEnabled:           true,
			ConsecutiveFailureCount:          2,
			LastSucceededAt:                  999,
			MappingExists:                    true,
		},
		updateCh: make(chan struct{}, 1),
		cancel:   cancel,
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{
			response: get_events.EventsResponse{
				Events: []get_events.Event{
					{ID: 42, Topic: "orders.created", EventData: []byte(`{"order_id":"ord_123"}`), CreatedAt: 123456},
				},
				Count:  1,
				Cursor: 42,
			},
		},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 400}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	select {
	case <-store.updateCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery state update")
	}

	if store.lastDeadLetterEvent.SourceEventID != 42 {
		t.Fatalf("SourceEventID = %d, want %d", store.lastDeadLetterEvent.SourceEventID, 42)
	}
	if store.lastDeadLetterEvent.FailureCategory != "endpoint_response_4xx" {
		t.Fatalf("FailureCategory = %q, want %q", store.lastDeadLetterEvent.FailureCategory, "endpoint_response_4xx")
	}
	if store.lastDeadLetterEvent.FailureCount != 3 {
		t.Fatalf("FailureCount = %d, want %d", store.lastDeadLetterEvent.FailureCount, 3)
	}
	if store.lastFailureEvent.FailureCount != 3 {
		t.Fatalf("delivery failure FailureCount = %d, want %d", store.lastFailureEvent.FailureCount, 3)
	}

	update := store.lastUpdate
	if update.LastDeliveredEventID != 42 {
		t.Fatalf("LastDeliveredEventID = %d, want %d", update.LastDeliveredEventID, 42)
	}
	if update.LastAttemptedEventID != 42 {
		t.Fatalf("LastAttemptedEventID = %d, want %d", update.LastAttemptedEventID, 42)
	}
	if update.LastSkippedEventID != 42 {
		t.Fatalf("LastSkippedEventID = %d, want %d", update.LastSkippedEventID, 42)
	}
	if update.ConsecutiveFailureCount != 0 {
		t.Fatalf("ConsecutiveFailureCount = %d, want %d", update.ConsecutiveFailureCount, 0)
	}
	if update.LastSucceededAt != 999 {
		t.Fatalf("LastSucceededAt = %d, want %d", update.LastSucceededAt, 999)
	}
	if update.LastErrorCategory != "endpoint_response_4xx" {
		t.Fatalf("LastErrorCategory = %q, want %q", update.LastErrorCategory, "endpoint_response_4xx")
	}
}

func TestConsumerSkipsConfiguredTransportFailures(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 41,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			LastDeliveredEventID:             41,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 1,
			SkipOnEndpointTransportError:     true,
			DeadLetterQueueEnabled:           true,
			MappingExists:                    true,
		},
		updateCh: make(chan struct{}, 1),
		cancel:   cancel,
	}

	service := outbound.NewConsumerService(
		store,
		staticEventReader{
			response: get_events.EventsResponse{
				Events: []get_events.Event{
					{ID: 42, Topic: "orders.created", EventData: []byte(`{"order_id":"ord_123"}`), CreatedAt: 123456},
				},
				Count:  1,
				Cursor: 42,
			},
		},
		staticDeliveryClient{err: errors.New("dial tcp timeout")},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	select {
	case <-store.updateCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery state update")
	}

	if store.lastDeadLetterEvent.FailureCategory != "endpoint_transport" {
		t.Fatalf("FailureCategory = %q, want %q", store.lastDeadLetterEvent.FailureCategory, "endpoint_transport")
	}
	if store.lastUpdate.LastSkippedEventID != 42 {
		t.Fatalf("LastSkippedEventID = %d, want %d", store.lastUpdate.LastSkippedEventID, 42)
	}
}

package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelo-dev/mycelo/backend/outbound"
	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

type countingDeliveryClient struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (c *countingDeliveryClient) Deliver(ctx context.Context, endpoint string, data []byte, meta *outbound.WebhookDeliveryMeta) (outbound.DeliveryResult, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()

	if c.err != nil {
		return outbound.DeliveryResult{}, c.err
	}

	return outbound.DeliveryResult{StatusCode: 204}, nil
}

func (c *countingDeliveryClient) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.calls
}

func TestConsumerCircuitBreakerBlocksEndpointAfterThreshold(t *testing.T) {
	t.Setenv("OUTBOUND_CIRCUIT_BREAKER_FAILURE_THRESHOLD", "1")
	t.Setenv("OUTBOUND_CIRCUIT_BREAKER_COOLDOWN_MS", "60000")

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
			RetryBaseDelayMs:                 1,
			RetryMaxDelayMs:                  1,
			MaxConsecutiveFailuresBeforeSkip: 0,
			MappingExists:                    true,
		},
		updateCh: make(chan struct{}, 1),
	}

	client := &countingDeliveryClient{err: errors.New("dial tcp timeout")}
	service := outbound.NewConsumerService(
		store,
		staticEventReader{
			response: get_events.EventsResponse{
				Events: []get_events.Event{
					{ID: 42, Topic: "orders.created", EventData: []byte(`{"order_id":"ord_123"}`), CreatedAt: time.Now().UnixMilli()},
				},
				Count:  1,
				Cursor: 42,
			},
		},
		client,
	)

	ctx, cancel := context.WithCancel(context.Background())
	store.cancel = cancel
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	waitForConsumerUpdate(t, store.updateCh)

	if client.Calls() != 1 {
		t.Fatalf("delivery calls after first failure = %d, want 1", client.Calls())
	}

	store.mu.Lock()
	store.state.NextAttemptAt = 0
	store.updateCh = make(chan struct{}, 1)
	store.mu.Unlock()

	ctx, cancel = context.WithCancel(context.Background())
	store.cancel = cancel
	if err := service.Start(ctx); err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	waitForConsumerUpdate(t, store.updateCh)

	if client.Calls() != 1 {
		t.Fatalf("delivery calls after open circuit = %d, want 1", client.Calls())
	}
	if store.lastUpdate.LastErrorCategory != "endpoint_circuit_open" {
		t.Fatalf("LastErrorCategory = %q, want %q", store.lastUpdate.LastErrorCategory, "endpoint_circuit_open")
	}
}

func waitForConsumerUpdate(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery state update")
	}
}

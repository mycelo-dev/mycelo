package tests

import (
	"context"
	"testing"
	"time"

	"github.com/mycelo-dev/mycelo/backend/outbound"
	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

type offsetSpyReader struct {
	ch chan int64
}

func (r offsetSpyReader) GetEventsAfterCursor(ctx context.Context, tenantPublicID string, teamPublicID string, topic string, after int64, offset int64, limit int) (get_events.EventsResponse, error) {
	select {
	case r.ch <- offset:
	default:
	}

	return get_events.EventsResponse{Count: 0}, nil
}

func TestConsumerReconcilesEventReadOffsetWithDBCursor(t *testing.T) {
	offsetCh := make(chan int64, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	store := &consumerTestStore{
		mapping: outbound.DestinationTopicMapping{
			DestinationID:        "destination-1",
			TopicID:              "topic-1",
			LastDeliveredEventID: 10,
		},
		state: outbound.OutboundMappingState{
			TopicName:                        "orders.created",
			Endpoint:                         "https://example.com/webhook",
			DeliveryFlag:                     true,
			LastDeliveredEventID:             99,
			RetryBaseDelayMs:                 2000,
			RetryMaxDelayMs:                  60000,
			MaxConsecutiveFailuresBeforeSkip: 0,
			MappingExists:                    true,
		},
	}

	service := outbound.NewConsumerService(
		store,
		offsetSpyReader{ch: offsetCh},
		staticDeliveryClient{result: outbound.DeliveryResult{StatusCode: 204}},
	)

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	var first int64
	select {
	case first = <-offsetCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first event read")
	}

	if first != 99 {
		t.Fatalf("first read offset = %d, want 99 (sync mapping cursor was 10)", first)
	}
}

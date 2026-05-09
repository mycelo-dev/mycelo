package outbound

import (
	"context"
	"encoding/json"
	"log"
	"time"

	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

func StartConsumers(ctx context.Context) error {
	consumers := make(map[string]context.CancelFunc)

	syncConsumers := func() error {
		mappings, err := GetDestinationTopicMappings(ctx)
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
				if err := ConsumeEvents(consumerCtx, mapping.DestinationID, mapping.TopicID, mapping.LastDeliveredEventID); err != nil {
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
		ticker := time.NewTicker(2 * time.Second)
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
func ConsumeEvents(ctx context.Context, destinationID string, topicID string, startOffset int64) error {
	cursor := startOffset

	for {
		state, err := GetOutboundMappingState(ctx, destinationID, topicID)
		if err != nil {
			return err
		}

		if !state.MappingExists || !state.DeliveryFlag || state.TopicName == "" || state.Endpoint == "" {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		resp, err := get_events.GetEventsAfterCursor(ctx, state.TopicName, 0, cursor)
		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			state, err = GetOutboundMappingState(ctx, destinationID, topicID)
			if err != nil {
				return err
			}

			if !state.MappingExists || !state.DeliveryFlag || state.Endpoint == "" {
				break
			}

			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal event %d: %v", event.ID, err)
				continue
			}

			httpResp, err := DeliverToHttp(state.Endpoint, data)
			if err != nil {
				log.Printf("failed to deliver event %d to %s: %v", event.ID, state.Endpoint, err)
				continue
			}

			if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
				log.Printf("failed to deliver event %d to %s: received status %d", event.ID, state.Endpoint, httpResp.StatusCode)
				httpResp.Body.Close()
				continue
			}
			httpResp.Body.Close()

			cursor = event.ID

			if err := UpdateOutboundMappingCursor(ctx, destinationID, topicID, cursor); err != nil {
				return err
			}
		}

		if resp.Count == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

package outbound

import (
	"context"
	"encoding/json"
	"log"
	"time"

	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

// now we need to keep on delivering the events to an outbound endpoint.
// we need to store the outbound endpoint address, the topic for which we need to send
// the event and a flag. if the flag is off then we no need to send events to that event for that topic

func ConsumeEvents(ctx context.Context, topic string, startOffset int64) error {
	cursor := startOffset
	endpoint := GetUrl()

	for {
		resp, err := get_events.GetEventsAfterCursor(ctx, topic, 0, cursor)
		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal event %d: %v", event.ID, err)
				continue
			}
			DeliverToHttp(endpoint, data) // HTTP POST, ignore failure for now
			cursor = event.ID
		}

		if resp.Count == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

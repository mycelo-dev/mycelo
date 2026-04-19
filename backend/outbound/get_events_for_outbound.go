package outbound

// here we read the events using the getEvents function

import (
	"context"
	"fmt"

	get_events "github.com/mycelo-dev/mycelo/backend/stream"
)

func GetEventsForOutbound(ctx context.Context, topic string, after int64, offset int64) get_events.EventsResponse {
	var resp get_events.EventsResponse

	resp, err := get_events.GetEventsAfterCursor(ctx, topic, after, offset)

	if err != nil {
		fmt.Println("error getting the events for the topic", err)
		return resp
	}

	return resp
}

package stream

import (
	"context"
	"encoding/json"

	db "github.com/mycelo-dev/mycelo/backend/core"
	queries "github.com/mycelo-dev/mycelo/backend/queries"
)

type Event struct {
	ID        int64           `json:"id"` // hidden from API
	Topic     string          `json:"topic"`
	EventData json.RawMessage `json:"event_data"`
	CreatedAt int64           `json:"created_at"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
	Count  int     `json:"count"`
	Cursor int64   `json:"cursor"`
	// HasMore bool `json:"has_more"`
}

func GetEventsAfterCursor(
	ctx context.Context,
	topic string,
	after int64, // timestamp
	offset int64, // id
) (EventsResponse, error) {

	sql := queries.GetEventsAfterCursorQuery()

	rows, err := db.Get().Query(ctx, sql, topic, after, offset)
	if err != nil {
		return EventsResponse{}, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	count := 0

	var lastID int64
	var lastCreatedAt int64

	for rows.Next() {
		var e Event
		var id int64

		if err := rows.Scan(&e.Topic, &e.EventData, &e.CreatedAt, &e.ID); err != nil {
			return EventsResponse{}, err
		}
		events = append(events, e)
		count++

		lastID = id
		lastCreatedAt = e.CreatedAt
	}

	// cursor derived from last row
	cursor := offset
	if count > 0 {
		cursor = lastID
		after = lastCreatedAt
	}

	return EventsResponse{
		Events: events,
		Count:  count,
		Cursor: cursor,
		// HasMore: false,
	}, nil
}

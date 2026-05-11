package stream

import (
	"context"
	"encoding/json"

	db "github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
)

// MaxEventsFetchUpperBound caps any single read to protect memory under hot topics.
const MaxEventsFetchUpperBound = 2000

const defaultEventsFetchBatch = 500

// Event is the persisted event shape used by the stream APIs.
type Event struct {
	ID        int64           `json:"id"` // hidden from API
	Topic     string          `json:"topic"`
	EventData json.RawMessage `json:"event_data"`
	CreatedAt int64           `json:"created_at"`
}

// EventsResponse returns events plus the cursor clients should continue from.
type EventsResponse struct {
	Events  []Event `json:"events"`
	Count   int     `json:"count"`
	Cursor  int64   `json:"cursor"`
	HasMore bool    `json:"has_more"`
}

// GetEventsAfterCursor returns topic events after the supplied cursor batching with a LIMIT.
func GetEventsAfterCursor(
	ctx context.Context,
	topic string,
	after int64, // timestamp
	offset int64, // id
	limit int,
) (EventsResponse, error) {
	if limit <= 0 {
		limit = defaultEventsFetchBatch
	}
	if limit > MaxEventsFetchUpperBound {
		limit = MaxEventsFetchUpperBound
	}

	sql := select_queries.GetEventsAfterCursorQuery()

	rows, err := db.Get().Query(ctx, sql, topic, after, offset, limit)
	if err != nil {
		return EventsResponse{}, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	count := 0

	var lastID int64

	for rows.Next() {
		var e Event

		if err := rows.Scan(&e.Topic, &e.EventData, &e.CreatedAt, &e.ID); err != nil {
			return EventsResponse{}, err
		}
		events = append(events, e)
		count++

		lastID = e.ID
	}

	// cursor derived from last row
	cursor := offset
	if count > 0 {
		cursor = lastID
	}

	hasMore := count > 0 && count == limit && limit > 0

	return EventsResponse{
		Events:  events,
		Count:   count,
		Cursor:  cursor,
		HasMore: hasMore,
	}, nil
}

// GetEventsBeforeCursor returns topic events in newest-first order before the supplied id cursor.
func GetEventsBeforeCursor(
	ctx context.Context,
	topic string,
	offset int64,
	limit int,
) (EventsResponse, error) {
	if limit <= 0 {
		limit = defaultEventsFetchBatch
	}
	if limit > MaxEventsFetchUpperBound {
		limit = MaxEventsFetchUpperBound
	}

	rows, err := db.Get().Query(ctx, select_queries.GetEventsBeforeCursorQuery(), topic, offset, limit)
	if err != nil {
		return EventsResponse{}, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	count := 0
	cursor := offset

	for rows.Next() {
		var e Event

		if err := rows.Scan(&e.Topic, &e.EventData, &e.CreatedAt, &e.ID); err != nil {
			return EventsResponse{}, err
		}

		events = append(events, e)
		count++
		cursor = e.ID
	}

	if err := rows.Err(); err != nil {
		return EventsResponse{}, err
	}

	return EventsResponse{
		Events:  events,
		Count:   count,
		Cursor:  cursor,
		HasMore: count > 0 && count == limit && limit > 0,
	}, nil
}

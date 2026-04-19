package stream

import (
	"context"
	"encoding/json"

	db "github.com/mycelo-dev/mycelo/backend/core"
	queries "github.com/mycelo-dev/mycelo/backend/queries"
)

type Event struct {
	Topic     string      `json:"topic"`
	EventData interface{} `json:"event_data"`
	CreatedAt int64       `json:"created_at"`
}

func GetEventsAfterCursor(
	ctx context.Context,
	topic string,
	after int64,
	offset int,
) ([]Event, error) {

	sql := queries.GetEventsAfterCursorQuery(after, offset)

	rows, err := db.Get().Query(ctx, sql, topic, after)

	if offset == 0 && after != 0 {

		rows, err = db.Get().Query(ctx, sql, topic, after)

	} else if offset != 0 && after == 0 {

		rows, err = db.Get().Query(ctx, sql, topic, offset)

	} else if offset != 0 && after != 0 {

		rows, err = db.Get().Query(ctx, sql, topic, offset, after)

	} else {

		rows, err = db.Get().Query(ctx, sql, topic, after)

	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event

	for rows.Next() {
		var e Event
		var eventBytes []byte

		if err := rows.Scan(&e.Topic, &eventBytes, &e.CreatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(eventBytes, &e.EventData); err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, nil
}

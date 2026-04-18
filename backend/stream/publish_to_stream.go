package stream

import (
	"context"
	"encoding/json"
	"time"

	db "github.com/mycelo-dev/mycelo/backend/core"
	insert_events_queries "github.com/mycelo-dev/mycelo/backend/queries"
)

func PublishToStream(ctx context.Context, topic string, event_data interface{}) error {

	sql := insert_events_queries.GetInsertEventsQueries()
	created_at := time.Now().UnixMilli()

	eventBytes, err := json.Marshal(event_data)
	if err != nil {
		return err
	}

	_, err = db.Get().Exec(ctx, sql, topic, eventBytes, created_at)
	return err
}

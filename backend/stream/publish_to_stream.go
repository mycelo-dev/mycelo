package P2S

import (
	"context"
	"fmt"
	"time"

	db "github.com/mycelo-dev/mycelo/backend/core"
	insert_events_queries "github.com/mycelo-dev/mycelo/backend/queries"
)

func PublishToStream(ctx context.Context, topic string, event_data string) {

	sql := insert_events_queries.GetInsertEventsQueries()

	created_at := time.Now().UnixMicro()

	_, err := db.Get().Exec(ctx, sql, topic, event_data, created_at)

	if err != nil {
		fmt.Println("error inserting data into events table: ", err)
	}
}

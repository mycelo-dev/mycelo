package topics_management

import (
	"context"
	"fmt"
	"time"

	db "github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries"
)

func CreateTopicRepository(ctx context.Context, topic_name string) error {

	query := queries.GetTopicsInsertQuery()

	created_at := time.Now().UnixMilli()
	updated_at := time.Now().UnixMilli()

	_, err := db.Get().Exec(
		ctx,
		query,
		1,
		1,
		topic_name,
		created_at,
		updated_at,
	)

	if err != nil {
		fmt.Println("error inserting topics: ", err)
	}

	return err
}

func UpdateTopicRepository(ctx context.Context, old_topic_name string, new_topic_name string) error {
	query := queries.GetQueryToUpdateTopic()

	updated_at := time.Now().UnixMilli()

	_, err := db.Get().Exec(ctx, query, new_topic_name, old_topic_name, 1, 1, updated_at)

	if err != nil {
		fmt.Println("failed to update the topic: ", err)
	}

	return err
}

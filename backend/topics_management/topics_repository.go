package topics_management

import (
	"context"
	"fmt"
	"time"

	db "github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

func CreateTopicRepository(ctx context.Context, topic_name string) error {

	query := insert_queries.GetTopicsInsertQuery()

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
	query := update_queries.GetQueryToUpdateTopic()

	updated_at := time.Now().UnixMilli()

	_, err := db.Get().Exec(ctx, query, new_topic_name, old_topic_name, 1, 1, updated_at)

	if err != nil {
		fmt.Println("failed to update the topic: ", err)
	}

	return err
}

func ListTopicsRepository(ctx context.Context) ([]TopicRecord, error) {
	query := select_queries.GetTopicsByTenantAndTeamQuery()

	rows, err := db.Get().Query(ctx, query, 1, 1)
	if err != nil {
		fmt.Println("failed to read topics: ", err)
		return nil, err
	}
	defer rows.Close()

	topics := make([]TopicRecord, 0)

	for rows.Next() {
		var topic TopicRecord

		if err := rows.Scan(&topic.Topic_id, &topic.Topic_name); err != nil {
			fmt.Println("failed to scan topic: ", err)
			return nil, err
		}

		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("failed while reading topic rows: ", err)
		return nil, err
	}

	return topics, nil
}

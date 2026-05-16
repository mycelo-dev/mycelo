package topics_management

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/auth"
	db "github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

// CreateTopicRepository inserts a topic record for the current tenant-team scope.
func CreateTopicRepository(ctx context.Context, topic_name string) error {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}

	query := insert_queries.GetTopicsInsertQuery()

	created_at := time.Now().UnixMilli()
	updated_at := time.Now().UnixMilli()

	commandTag, err := db.Get().Exec(
		ctx,
		query,
		authContext.TenantPublicId,
		authContext.TeamPublicId,
		topic_name,
		created_at,
		updated_at,
	)

	if err != nil {
		fmt.Println("error inserting topics: ", err)
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("tenant/team scope not found for topic creation")
	}

	return nil
}

// UpdateTopicRepository renames a topic for the current tenant-team scope.
func UpdateTopicRepository(ctx context.Context, old_topic_name string, new_topic_name string) error {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}

	query := update_queries.GetQueryToUpdateTopic()

	updated_at := time.Now().UnixMilli()

	_, err = db.Get().Exec(ctx, query, new_topic_name, old_topic_name, authContext.TenantPublicId, authContext.TeamPublicId, updated_at)

	if err != nil {
		fmt.Println("failed to update the topic: ", err)
	}

	return err
}

// ListTopicsRepository lists topics for the current tenant-team scope.
func ListTopicsRepository(ctx context.Context) ([]TopicRecord, error) {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := select_queries.GetTopicsByTenantAndTeamQuery()

	rows, err := db.Get().Query(ctx, query, authContext.TenantPublicId, authContext.TeamPublicId)
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

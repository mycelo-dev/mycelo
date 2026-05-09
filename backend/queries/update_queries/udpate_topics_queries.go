package update_queries

func GetQueryToUpdateTopic() string {
	return `
			UPDATE TOPICS
			SET topic_name = $1, updated_at = $5
			WHERE topic_name = $2 AND tenant_id = $3 AND team_id = $4
	`
}

package select_queries

func GetTopicsByTenantAndTeamQuery() string {
	return `
		SELECT
			topic_public_id,
			topic_name
		FROM topics
		WHERE tenant_id = $1
		AND team_id = $2
		ORDER BY topic_name ASC
	`
}

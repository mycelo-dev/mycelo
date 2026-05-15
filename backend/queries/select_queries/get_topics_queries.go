package select_queries

// GetTopicsByTenantAndTeamQuery lists topics for the current tenant-team scope.
func GetTopicsByTenantAndTeamQuery() string {
	return `
		SELECT
			topic_public_id,
			topic_name
		FROM topics
		WHERE tenant_public_id = $1
		AND team_public_id = $2
		ORDER BY topic_name ASC
	`
}

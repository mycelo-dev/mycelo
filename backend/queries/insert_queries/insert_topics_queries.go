package insert_queries

// GetTopicsInsertQuery inserts a topic row.
func GetTopicsInsertQuery() string {
	return `
			INSERT INTO topics
			(tenant_id, team_id, tenant_public_id, team_public_id, topic_name, created_at, updated_at)
			SELECT
				tenant.tenant_id,
				team.team_id,
				tenant.tenant_public_id,
				team.team_public_id,
				$3,
				$4,
				$5
			FROM tenants tenant
			INNER JOIN teams team
				ON team.tenant_id = tenant.tenant_id
			WHERE tenant.tenant_public_id = $1
			AND team.team_public_id = $2
	`
}

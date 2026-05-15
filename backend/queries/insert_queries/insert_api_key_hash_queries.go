package insert_queries

// GetInsertApiKeyHashQuery inserts a stored API key hash row.
func GetInsertApiKeyHashQuery() string {
	return `
			INSERT INTO api_keys
			(tenant_public_id, team_public_id, key_hash, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5);
	`
}

// GetInsertApiKeyHashForTeamQuery inserts or replaces a team API key for a tenant user.
func GetInsertApiKeyHashForTeamQuery() string {
	return `
			INSERT INTO api_keys
			(tenant_public_id, team_public_id, key_hash, created_at, updated_at)
			SELECT
				tenant.tenant_public_id,
				team.team_public_id,
				$4,
				$5,
				$6
			FROM tenants tenant
			INNER JOIN users app_user
				ON app_user.tenant_id = tenant.tenant_id
			INNER JOIN teams team
				ON team.tenant_id = tenant.tenant_id
			WHERE tenant.tenant_public_id = $1
			AND app_user.user_public_id = $2
			AND team.team_public_id = $3
			ON CONFLICT (tenant_public_id, team_public_id)
			DO UPDATE SET
				key_hash = EXCLUDED.key_hash,
				updated_at = EXCLUDED.updated_at
			RETURNING tenant_public_id::text, team_public_id::text
	`
}

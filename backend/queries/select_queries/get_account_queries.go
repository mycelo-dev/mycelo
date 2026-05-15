package select_queries

// GetTeamsForTenantUserQuery lists teams owned by a signed-up tenant user.
func GetTeamsForTenantUserQuery() string {
	return `
			SELECT
				team.team_public_id::text,
				team.team_name
			FROM teams team
			INNER JOIN tenants tenant
				ON tenant.tenant_id = team.tenant_id
			INNER JOIN users app_user
				ON app_user.tenant_id = tenant.tenant_id
			WHERE tenant.tenant_public_id = $1
			AND app_user.user_public_id = $2
			ORDER BY team.team_name ASC
	`
}

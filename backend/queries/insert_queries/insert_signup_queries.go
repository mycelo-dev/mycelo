package insert_queries

// GetInsertTenantQuery inserts a tenant and returns its internal and public identifiers.
func GetInsertTenantQuery() string {
	return `
			INSERT INTO tenants
			(tenant_name, created_at, updated_at)
			VALUES($1, $2, $3)
			RETURNING tenant_id, tenant_public_id::text
	`
}

// GetInsertTeamQuery inserts a team under a tenant and returns its public identifier.
func GetInsertTeamQuery() string {
	return `
			INSERT INTO teams
			(tenant_id, team_name, created_at, updated_at)
			VALUES($1, $2, $3, $4)
			RETURNING team_public_id::text
	`
}

// GetInsertUserQuery inserts a user under a tenant and returns its public identifier.
func GetInsertUserQuery() string {
	return `
			INSERT INTO users
			(tenant_id, user_name, email, password_hash, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5, $6)
			RETURNING user_public_id::text
	`
}

// GetInsertTeamForTenantUserQuery inserts a team for a signed-up tenant user.
func GetInsertTeamForTenantUserQuery() string {
	return `
			INSERT INTO teams
			(tenant_id, team_name, created_at, updated_at)
			SELECT
				tenant.tenant_id,
				$3,
				$4,
				$5
			FROM tenants tenant
			INNER JOIN users app_user
				ON app_user.tenant_id = tenant.tenant_id
			WHERE tenant.tenant_public_id = $1
			AND app_user.user_public_id = $2
			RETURNING team_public_id::text, team_name
	`
}

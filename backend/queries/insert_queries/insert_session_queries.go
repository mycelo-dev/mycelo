package insert_queries

// GetInsertSessionQuery stores a hashed operator-console session token.
func GetInsertSessionQuery() string {
	return `
			INSERT INTO account_sessions
			(tenant_public_id, user_public_id, session_hash, created_at, expires_at)
			VALUES($1, $2, $3, $4, $5)
	`
}

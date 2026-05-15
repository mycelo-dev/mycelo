package select_queries

// GetSessionContextQuery returns account scope for a valid operator-console session.
func GetSessionContextQuery() string {
	return `
			SELECT
				tenant_public_id::text,
				user_public_id::text
			FROM account_sessions
			WHERE session_hash = $1
			AND expires_at > $2
	`
}

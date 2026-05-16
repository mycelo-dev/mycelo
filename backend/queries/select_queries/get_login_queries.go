package select_queries

// GetAccountByEmailQuery loads account context by signup email.
func GetAccountByEmailQuery() string {
	return `
			SELECT
				tenant.tenant_public_id::text,
				app_user.user_public_id::text,
				tenant.tenant_name,
				app_user.user_name,
				app_user.email,
				app_user.password_hash
			FROM users app_user
			INNER JOIN tenants tenant
				ON tenant.tenant_id = app_user.tenant_id
			WHERE lower(app_user.email) = lower($1)
	`
}

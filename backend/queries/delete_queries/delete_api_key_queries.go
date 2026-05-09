package delete_queries

func GetRevokeApiKeyQuery() string {
	return `
			DELETE FROM api_keys
			WHERE
			tenant_public_id = $1 AND team_public_id = $2
	`
}

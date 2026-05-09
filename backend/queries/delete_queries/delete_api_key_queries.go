package delete_queries

// GetRevokeApiKeyQuery deletes the stored API key row for a tenant-team pair.
func GetRevokeApiKeyQuery() string {
	return `
			DELETE FROM api_keys
			WHERE
			tenant_public_id = $1 AND team_public_id = $2
	`
}

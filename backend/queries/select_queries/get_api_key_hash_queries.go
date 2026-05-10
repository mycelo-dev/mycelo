package select_queries

// GetApiKeyHashFromDbQuery reads a stored API key hash by tenant and team.
func GetApiKeyHashFromDbQuery() string {

	return `
			SELECT key_hash
			FROM 
			api_keys
			WHERE 
			tenant_public_id = $1 AND team_public_id = $2
	`
}

package select_queries

func GetApiKeyHashFromDbQuery() string {

	return `
			SELECT key_hash
			FROM 
			api_keys
			WHERE 
			tenant_public_id = $1 AND team_public_id = $2
	`
}

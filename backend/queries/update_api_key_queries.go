package queries

func GetRotateApiKeyQuery() string {

	return `
			UPDATE api_keys 
			SET key_hash = $1, updated_at = $4
			WHERE tenant_public_id = $2 AND team_public_id = $3
	`
}

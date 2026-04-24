package queries

func GetInsertApiKeyHashQuery() string {
	return `
			INSERT INTO api_keys
			(tenant_public_id, team_public_id, key_hash, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5);
	`
}

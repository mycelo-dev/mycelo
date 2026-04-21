package queries

func GetInsertDestinationQuery() string {
	return `
			INSERT INTO destinations
			(tenant_id, team_id, destination_name, destination_address, created_at, updated_at)
			VALUES($1, $2, $3, $4, $5, $6);
	`
}

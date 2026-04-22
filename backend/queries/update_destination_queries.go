package queries

func GetUpdateDestinationQuery() string {
	return `
			UPDATE destinations 
			SET destination_name = $1, destination_address = $2, updated_at = $3
			WHERE destination_public_id = $4
	`
}

package update_queries

func GetUpdateDestinationQuery() string {
	return `
			UPDATE destinations 
			SET destination_name = $1, destination_address = $2, updated_at = $3
			WHERE destination_public_id = $4
	`
}

func GetUpdateDeliveryFlagQuery() string {
	return `
			UPDATE destinations
			SET delivery_flag = $1, updated_at = $2
			WHERE destination_public_id = $3
	`
}

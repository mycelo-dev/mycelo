package queries

func GetDeleteDestinationQuery() string {
	return `
			DELETE FROM destinations
			WHERE destination_public_id = $1
			AND delivery_flag = false
	`
}

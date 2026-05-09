package select_queries

func GetReadDeliveryFlagByPublicIdQuery() string {
	return `
			SELECT delivery_flag 
			FROM destinations
			WHERE destination_public_id = $1
	`
}

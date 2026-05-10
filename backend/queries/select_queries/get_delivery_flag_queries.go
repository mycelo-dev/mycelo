package select_queries

// GetReadDeliveryFlagByPublicIdQuery reads the delivery flag for a destination.
func GetReadDeliveryFlagByPublicIdQuery() string {
	return `
			SELECT delivery_flag 
			FROM destinations
			WHERE destination_public_id = $1
	`
}

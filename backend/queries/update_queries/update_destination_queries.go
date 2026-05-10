package update_queries

// GetUpdateDestinationQuery updates a destination's editable fields.
func GetUpdateDestinationQuery() string {
	return `
			UPDATE destinations 
			SET
				destination_name = $1,
				destination_address = $2,
				updated_at = $3,
				webhook_signing_secret = CASE WHEN $5 THEN $6 ELSE webhook_signing_secret END
			WHERE destination_public_id = $4
	`
}

// GetUpdateDeliveryFlagQuery updates the delivery flag for a mapped destination-topic pair.
func GetUpdateDeliveryFlagQuery() string {
	return `
			UPDATE destinations
			SET delivery_flag = $1, updated_at = $2
			WHERE destination_public_id = $3
			AND EXISTS (
				SELECT 1
				FROM destination_topic_mapping
				WHERE destination_public_id = $3
				AND topic_public_id = $4
			)
	`
}

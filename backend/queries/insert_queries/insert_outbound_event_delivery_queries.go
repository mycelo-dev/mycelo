package insert_queries

// GetInsertOutboundEventDeliveryQuery creates a pending unordered delivery record.
func GetInsertOutboundEventDeliveryQuery() string {
	return `
		INSERT INTO outbound_event_deliveries
		(
			destination_public_id,
			topic_public_id,
			source_event_id,
			status,
			created_at,
			updated_at
		)
		SELECT
			$1,
			$2,
			source_event_id,
			'pending',
			$4,
			$4
		FROM unnest($3::bigint[]) AS source_event_id
		ON CONFLICT (destination_public_id, topic_public_id, source_event_id)
		DO NOTHING
	`
}

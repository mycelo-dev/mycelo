package insert_queries

// GetInsertOutboundDeliveryFailureQuery upserts the latest failed delivery attempt for one mapping event.
func GetInsertOutboundDeliveryFailureQuery() string {
	return `
		INSERT INTO outbound_delivery_failures
		(
			destination_public_id,
			topic_public_id,
			source_event_id,
			endpoint,
			failure_category,
			failure_reason,
			failure_count,
			first_failed_at,
			last_failed_at
		)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (destination_public_id, topic_public_id, source_event_id)
		DO UPDATE SET
			endpoint = EXCLUDED.endpoint,
			failure_category = EXCLUDED.failure_category,
			failure_reason = EXCLUDED.failure_reason,
			failure_count = GREATEST(outbound_delivery_failures.failure_count, EXCLUDED.failure_count),
			last_failed_at = EXCLUDED.last_failed_at
	`
}

package insert_queries

// GetInsertDeadLetterEventQuery inserts a dead-letter record for a skipped event.
func GetInsertDeadLetterEventQuery() string {
	return `
		INSERT INTO dead_letter_events
		(
			destination_public_id,
			topic_public_id,
			source_event_id,
			endpoint,
			failure_category,
			failure_reason,
			failure_count,
			event_payload,
			dead_lettered_at
		)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
}

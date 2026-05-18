package update_queries

// GetMarkOutboundEventDeliveryDeliveredQuery records one unordered delivery success.
func GetMarkOutboundEventDeliveryDeliveredQuery() string {
	return `
		UPDATE outbound_event_deliveries
		SET
			status = 'delivered',
			delivered_at = $3,
			locked_by = '',
			lock_expires_at = 0,
			next_attempt_at = 0,
			last_error_category = '',
			last_error = '',
			updated_at = $3
		WHERE delivery_id = $1
		AND locked_by = $2
	`
}

// GetMarkOutboundEventDeliveryFailedQuery records one unordered delivery failure and retry schedule.
func GetMarkOutboundEventDeliveryFailedQuery() string {
	return `
		UPDATE outbound_event_deliveries
		SET
			status = 'failed',
			next_attempt_at = $3,
			locked_by = '',
			lock_expires_at = 0,
			last_error_category = $4,
			last_error = $5,
			updated_at = $6
		WHERE delivery_id = $1
		AND locked_by = $2
	`
}

// GetMarkOutboundEventDeliverySkippedQuery records one unordered delivery skip.
func GetMarkOutboundEventDeliverySkippedQuery() string {
	return `
		UPDATE outbound_event_deliveries
		SET
			status = 'skipped',
			skipped_at = $3,
			next_attempt_at = 0,
			locked_by = '',
			lock_expires_at = 0,
			last_error_category = $4,
			last_error = $5,
			updated_at = $3
		WHERE delivery_id = $1
		AND locked_by = $2
	`
}

package select_queries

// GetClaimOutboundEventDeliveriesQuery atomically claims pending unordered deliveries.
func GetClaimOutboundEventDeliveriesQuery() string {
	return `
		WITH candidates AS (
			SELECT delivery_id
			FROM outbound_event_deliveries
			WHERE destination_public_id = $1
			AND topic_public_id = $2
			AND status IN ('pending', 'failed')
			AND next_attempt_at <= $3
			AND (locked_by = '' OR lock_expires_at <= $3)
			ORDER BY source_event_id ASC
			LIMIT $6
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbound_event_deliveries oed
		SET
			status = 'in_flight',
			locked_by = $4,
			lock_expires_at = $5,
			attempt_count = attempt_count + 1,
			last_attempted_at = $3,
			updated_at = $3
		FROM candidates c, events e
		WHERE oed.delivery_id = c.delivery_id
		AND e.id = oed.source_event_id
		RETURNING
			oed.delivery_id,
			oed.source_event_id,
			oed.attempt_count,
			e.topic,
			e.event_data,
			e.created_at
	`
}

// GetUnorderedInFlightDeliveryCountQuery counts active unordered deliveries for mode-change safety.
func GetUnorderedInFlightDeliveryCountQuery() string {
	return `
		SELECT COUNT(*)
		FROM outbound_event_deliveries
		WHERE destination_public_id = $1
		AND topic_public_id = $2
		AND status = 'in_flight'
	`
}

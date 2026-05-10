package select_queries

// GetDeadLetterEventsQuery lists dead-lettered events with optional destination/topic filters.
func GetDeadLetterEventsQuery() string {
	return `
		SELECT
			dle.dead_letter_event_id,
			dle.destination_public_id,
			d.destination_name,
			dle.topic_public_id,
			t.topic_name,
			dle.source_event_id,
			dle.endpoint,
			dle.failure_category,
			dle.failure_reason,
			dle.failure_count,
			dle.event_payload,
			dle.dead_lettered_at
		FROM dead_letter_events dle
		INNER JOIN destinations d
			ON d.destination_public_id = dle.destination_public_id
		INNER JOIN topics t
			ON t.topic_public_id = dle.topic_public_id
		WHERE ($1 = '' OR dle.destination_public_id::text = $1)
		AND ($2 = '' OR dle.topic_public_id::text = $2)
		AND d.tenant_id = $4
		AND d.team_id = $5
		AND t.tenant_id = $4
		AND t.team_id = $5
		ORDER BY dle.dead_lettered_at DESC, dle.dead_letter_event_id DESC
		LIMIT $3
	`
}

package select_queries

// GetOutboundMappingsQuery lists all outbound mapping cursors.
func GetOutboundMappingsQuery() string {
	return `
		SELECT
			dtm.destination_public_id,
			dtm.topic_public_id,
			d.tenant_public_id,
			d.team_public_id,
			dtm.last_delivered_event_id
		FROM destination_topic_mapping dtm
		INNER JOIN destinations d
			ON d.destination_public_id = dtm.destination_public_id
	`
}

// GetOutboundMappingStateQuery reads the delivery configuration and retry state for one mapping.
func GetOutboundMappingStateQuery() string {
	return `
		SELECT
			t.topic_name,
			d.destination_address,
			d.webhook_signing_secret,
			d.delivery_flag,
			dtm.last_delivered_event_id,
			dtm.retry_base_delay_ms,
			dtm.retry_max_delay_ms,
			dtm.max_consecutive_failures_before_skip,
			dtm.dead_letter_queue_enabled,
			dtm.skip_on_endpoint_4xx,
			dtm.skip_on_endpoint_5xx,
			dtm.skip_on_endpoint_transport_error,
			dtm.skip_on_event_payload_error,
			dtm.last_attempted_event_id,
			dtm.last_failed_event_id,
			dtm.last_skipped_event_id,
			dtm.consecutive_failure_count,
			dtm.last_succeeded_at,
			dtm.last_failed_at,
			dtm.last_skipped_at,
			dtm.next_attempt_at,
			dtm.last_error_category,
			dtm.last_error
		FROM destination_topic_mapping dtm
		INNER JOIN topics t
			ON t.topic_public_id = dtm.topic_public_id
		INNER JOIN destinations d
			ON d.destination_public_id = dtm.destination_public_id
		WHERE dtm.destination_public_id = $1
		AND dtm.topic_public_id = $2
	`
}

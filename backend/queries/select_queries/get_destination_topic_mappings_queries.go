package select_queries

// GetDestinationTopicMappingsByTenantAndTeamQuery lists mappings and their delivery state.
func GetDestinationTopicMappingsByTenantAndTeamQuery() string {
	return `
		SELECT
			d.destination_public_id,
			d.destination_name,
			d.destination_address,
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
			dtm.last_attempted_at,
			dtm.last_succeeded_at,
			dtm.last_failed_at,
			dtm.last_skipped_at,
			dtm.next_attempt_at,
			dtm.last_error_category,
			dtm.last_error,
			t.topic_public_id,
			t.topic_name
		FROM destination_topic_mapping dtm
		INNER JOIN destinations d
			ON d.destination_public_id = dtm.destination_public_id
		INNER JOIN topics t
			ON t.topic_public_id = dtm.topic_public_id
		WHERE d.tenant_id = $1
		AND d.team_id = $2
		AND t.tenant_id = $1
		AND t.team_id = $2
		ORDER BY d.destination_name ASC, t.topic_name ASC
	`
}

// GetDestinationTopicMappingPolicyQuery reads delivery policy for one destination-topic mapping.
func GetDestinationTopicMappingPolicyQuery() string {
	return `
		SELECT
			retry_base_delay_ms,
			retry_max_delay_ms,
			max_consecutive_failures_before_skip,
			dead_letter_queue_enabled,
			skip_on_endpoint_4xx,
			skip_on_endpoint_5xx,
			skip_on_endpoint_transport_error,
			skip_on_event_payload_error
		FROM destination_topic_mapping
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

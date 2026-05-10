package update_queries

// GetUpdateDestinationTopicMappingCursorQuery updates only the delivered event cursor.
func GetUpdateDestinationTopicMappingCursorQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET last_delivered_event_id = $3
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

// GetUpdateDestinationTopicMappingDeliveryStateQuery updates full delivery attempt metadata.
func GetUpdateDestinationTopicMappingDeliveryStateQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			last_delivered_event_id = $3,
			last_attempted_event_id = $4,
			last_failed_event_id = $5,
			last_skipped_event_id = $6,
			consecutive_failure_count = $7,
			last_attempted_at = $8,
			last_succeeded_at = $9,
			last_failed_at = $10,
			last_skipped_at = $11,
			next_attempt_at = $12,
			last_error_category = $13,
			last_error = $14
		WHERE destination_public_id = $1
		AND topic_public_id = $2
		AND delivery_lease_holder = $15
	`
}

// GetUpdateDestinationTopicMappingPolicyQuery updates configurable delivery controls for a mapping.
func GetUpdateDestinationTopicMappingPolicyQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			retry_base_delay_ms = $3,
			retry_max_delay_ms = $4,
			max_consecutive_failures_before_skip = $5,
			dead_letter_queue_enabled = $6,
			skip_on_endpoint_4xx = $7,
			skip_on_endpoint_5xx = $8,
			skip_on_endpoint_transport_error = $9,
			skip_on_event_payload_error = $10
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

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
			skip_on_event_payload_error = $10,
			delivery_mode = $11,
			unordered_max_in_flight = $12,
			unordered_last_enqueued_event_id = CASE
				WHEN delivery_mode <> $11 AND $11 = 'unordered' THEN last_delivered_event_id
				ELSE unordered_last_enqueued_event_id
			END,
			last_delivered_event_id = CASE
				WHEN delivery_mode <> $11 AND $11 = 'ordered' THEN GREATEST(last_delivered_event_id, unordered_last_enqueued_event_id)
				ELSE last_delivered_event_id
			END
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

// GetUpdateUnorderedEnqueueCursorQuery advances the unordered discovery cursor.
func GetUpdateUnorderedEnqueueCursorQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET unordered_last_enqueued_event_id = GREATEST(unordered_last_enqueued_event_id, $3)
		WHERE destination_public_id = $1
		AND topic_public_id = $2
		AND delivery_lease_holder = $4
	`
}

// GetUpdateUnorderedDeliverySuccessStateQuery records aggregate state after an unordered success.
func GetUpdateUnorderedDeliverySuccessStateQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			last_attempted_event_id = GREATEST(last_attempted_event_id, $3),
			consecutive_failure_count = 0,
			last_attempted_at = $4,
			last_succeeded_at = $4,
			next_attempt_at = $4,
			last_error_category = '',
			last_error = ''
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

// GetUpdateUnorderedDeliveryFailureStateQuery records aggregate state after an unordered failure.
func GetUpdateUnorderedDeliveryFailureStateQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			last_attempted_event_id = GREATEST(last_attempted_event_id, $3),
			last_failed_event_id = $3,
			consecutive_failure_count = consecutive_failure_count + 1,
			last_attempted_at = $4,
			last_failed_at = $4,
			next_attempt_at = $5,
			last_error_category = $6,
			last_error = $7
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

// GetAdvanceUnorderedContiguousCursorQuery advances the ordered-compatible cursor over contiguous delivered rows.
func GetAdvanceUnorderedContiguousCursorQuery() string {
	return `
		WITH current_cursor AS (
			SELECT last_delivered_event_id
			FROM destination_topic_mapping
			WHERE destination_public_id = $1
			AND topic_public_id = $2
		),
		next_cursor AS (
			SELECT COALESCE(MAX(oed.source_event_id), (SELECT last_delivered_event_id FROM current_cursor)) AS cursor
			FROM outbound_event_deliveries oed, current_cursor cc
			WHERE oed.destination_public_id = $1
			AND oed.topic_public_id = $2
			AND oed.source_event_id > cc.last_delivered_event_id
			AND oed.status IN ('delivered', 'skipped')
			AND NOT EXISTS (
				SELECT 1
				FROM outbound_event_deliveries gap
				WHERE gap.destination_public_id = $1
				AND gap.topic_public_id = $2
				AND gap.source_event_id > cc.last_delivered_event_id
				AND gap.source_event_id <= oed.source_event_id
				AND gap.status NOT IN ('delivered', 'skipped')
			)
		)
		UPDATE destination_topic_mapping
		SET last_delivered_event_id = GREATEST(last_delivered_event_id, (SELECT cursor FROM next_cursor))
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}

package insert_queries

// GetAssignTopicToDestinationQuery inserts a destination-topic mapping with its initial cursor.
func GetAssignTopicToDestinationQuery() string {

	return `
			INSERT INTO destination_topic_mapping
			(
				destination_public_id,
				topic_public_id,
				last_delivered_event_id,
				retry_base_delay_ms,
				retry_max_delay_ms,
				max_consecutive_failures_before_skip,
				dead_letter_queue_enabled,
				skip_on_endpoint_4xx,
				skip_on_endpoint_5xx,
				skip_on_endpoint_transport_error,
				skip_on_event_payload_error,
				delivery_mode,
				unordered_max_in_flight,
				unordered_last_enqueued_event_id
			)
			VALUES
			(
				$1,
				$2,
				COALESCE((
					SELECT MAX(e.id)
					FROM events e
					INNER JOIN topics t
						ON t.topic_name = e.topic
					WHERE t.topic_public_id = $2
				), 0),
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11,
				$12,
				COALESCE((
					SELECT MAX(e.id)
					FROM events e
					INNER JOIN topics t
						ON t.topic_name = e.topic
					WHERE t.topic_public_id = $2
				), 0)
			)
	`
}

package select_queries

// GetOutboundDeliveryFailuresQuery lists recent failed delivery attempts for a tenant-team scope.
func GetOutboundDeliveryFailuresQuery() string {
	return `
		SELECT
			odf.delivery_failure_id,
			odf.destination_public_id::text,
			d.destination_name,
			odf.topic_public_id::text,
			t.topic_name,
			odf.source_event_id,
			odf.endpoint,
			odf.failure_category,
			odf.failure_reason,
			odf.failure_count,
			odf.first_failed_at,
			odf.last_failed_at
		FROM outbound_delivery_failures odf
		INNER JOIN destinations d
			ON d.destination_public_id = odf.destination_public_id
		INNER JOIN topics t
			ON t.topic_public_id = odf.topic_public_id
		WHERE ($1 = '' OR odf.destination_public_id::text = $1)
		AND ($2 = '' OR odf.topic_public_id::text = $2)
		AND d.tenant_public_id = $4
		AND d.team_public_id = $5
		AND t.tenant_public_id = $4
		AND t.team_public_id = $5
		ORDER BY odf.last_failed_at DESC, odf.delivery_failure_id DESC
		LIMIT $3
	`
}

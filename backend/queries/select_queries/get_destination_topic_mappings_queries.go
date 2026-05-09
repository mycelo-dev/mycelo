package select_queries

func GetDestinationTopicMappingsByTenantAndTeamQuery() string {
	return `
		SELECT
			d.destination_public_id,
			d.destination_name,
			d.destination_address,
			d.delivery_flag,
			dtm.last_delivered_event_id,
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

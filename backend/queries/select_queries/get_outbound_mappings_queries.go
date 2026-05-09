package select_queries

func GetOutboundMappingsQuery() string {
	return `
		SELECT dtm.destination_public_id, dtm.topic_public_id, dtm.last_delivered_event_id
		FROM destination_topic_mapping dtm
	`
}

func GetOutboundMappingStateQuery() string {
	return `
		SELECT t.topic_name, d.destination_address, d.delivery_flag, dtm.last_delivered_event_id
		FROM destination_topic_mapping dtm
		INNER JOIN topics t
			ON t.topic_public_id = dtm.topic_public_id
		INNER JOIN destinations d
			ON d.destination_public_id = dtm.destination_public_id
		WHERE dtm.destination_public_id = $1
		AND dtm.topic_public_id = $2
	`
}

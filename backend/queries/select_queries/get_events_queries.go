package select_queries

// GetEventsAfterCursorQuery reads ordered events after a cursor.
func GetEventsAfterCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE tenant_public_id = $1
		AND team_public_id = $2
		AND topic = $3
		AND ($4 = 0 OR created_at > $4)
		AND id > $5
		ORDER BY id ASC
		LIMIT $6
	`
}

// GetEventsBeforeCursorQuery reads newest events before an id cursor.
func GetEventsBeforeCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE tenant_public_id = $1
		AND team_public_id = $2
		AND topic = $3
		AND ($4 = 0 OR id < $4)
		ORDER BY created_at DESC, id DESC
		LIMIT $5
	`
}

// GetEventTopicsByTenantAndTeamQuery lists topic names that have stored events.
func GetEventTopicsByTenantAndTeamQuery() string {
	return `
		SELECT DISTINCT topic
		FROM events
		WHERE tenant_public_id = $1
		AND team_public_id = $2
		ORDER BY topic ASC
	`
}

// GetTopicHeadsByTenantAndTeamQuery lists each configured topic with its latest event id.
func GetTopicHeadsByTenantAndTeamQuery() string {
	return `
		SELECT
			t.topic_public_id,
			t.topic_name,
			COALESCE(MAX(e.id), 0)
		FROM topics t
		LEFT JOIN events e
			ON e.tenant_public_id = t.tenant_public_id
			AND e.team_public_id = t.team_public_id
			AND e.topic = t.topic_name
		WHERE t.tenant_public_id = $1
		AND t.team_public_id = $2
		GROUP BY t.topic_public_id, t.topic_name
		ORDER BY t.topic_name ASC
	`
}

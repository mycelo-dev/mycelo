package select_queries

// GetEventsAfterCursorQuery reads ordered events after a cursor.
func GetEventsAfterCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE tenant_public_id = $1
		AND team_public_id = $2
		AND topic = $3
		AND created_at > $4
		AND id > $5
		ORDER BY created_at ASC, id ASC
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

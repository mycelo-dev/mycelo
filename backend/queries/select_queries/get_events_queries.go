package select_queries

// GetEventsAfterCursorQuery reads ordered events after a cursor.
func GetEventsAfterCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE topic = $1
		AND created_at > $2 
		AND id > $3
		ORDER BY created_at ASC, id ASC
		LIMIT $4
	`
}

// GetEventsBeforeCursorQuery reads newest events before an id cursor.
func GetEventsBeforeCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE topic = $1
		AND ($2 = 0 OR id < $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`
}

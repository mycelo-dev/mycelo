package queries

func GetEventsAfterCursorQuery(created_at int64, offset int) string {

	if offset == 0 && created_at != 0 {
		return `
			SELECT topic, event_data, created_at
			FROM events
			WHERE topic = $1
			AND created_at > $2
			ORDER BY created_at ASC
		`
	} else if offset != 0 && created_at == 0 {
		return `
			SELECT topic, event_data, created_at
			FROM events
			WHERE topic = $1
			AND id > $2
			ORDER BY created_at ASC
		`
	} else if offset != 0 && created_at != 0 {
		return `
			SELECT topic, event_data, created_at
			FROM events
			WHERE topic = $1
			AND id > $2
			AND created_at > $3
			ORDER BY created_at ASC
		`
	} else {
		return `
			SELECT topic, event_data, created_at
			FROM events
			WHERE topic = $1
			AND created_at > $2
			ORDER BY created_at ASC
		`
	}
}

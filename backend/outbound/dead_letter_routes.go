package outbound

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultDeadLetterEventsLimit = 100

// ListDeadLetterEventsRoute returns recent dead-letter records with optional mapping filters.
func ListDeadLetterEventsRoute(w http.ResponseWriter, r *http.Request) {
	destinationID := r.URL.Query().Get("destination_id")
	topicID := r.URL.Query().Get("topic_id")

	limit := defaultDeadLetterEventsLimit
	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	records, err := NewOutboundRepository().ListDeadLetterEvents(r.Context(), destinationID, topicID, limit)
	if err != nil {
		http.Error(w, "failed to read dead letter events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(records)
}

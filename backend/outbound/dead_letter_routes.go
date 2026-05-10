package outbound

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultDeadLetterEventsLimit = 100

// ListDeadLetterEventsRoute returns recent dead-letter records with optional mapping filters.
func ListDeadLetterEventsRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ReplayDeadLetterEventsRoute(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

// ReplayDeadLetterEventsRoute re-enqueues DLQ payloads as new stream events.
func ReplayDeadLetterEventsRoute(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DeadLetterEventID int64  `json:"dead_letter_event_id"`
		DestinationID     string `json:"destination_id"`
		TopicID           string `json:"topic_id"`
		Limit             int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.DeadLetterEventID < 0 {
		http.Error(w, "dead_letter_event_id must be non-negative", http.StatusBadRequest)
		return
	}

	limit := payload.Limit
	if payload.DeadLetterEventID > 0 {
		limit = 1
	}
	if limit <= 0 {
		limit = defaultDeadLetterEventsLimit
	}

	result, err := NewOutboundRepository().ReplayDeadLetterEvents(r.Context(), payload.DeadLetterEventID, payload.DestinationID, payload.TopicID, limit)
	if err != nil {
		http.Error(w, "failed to replay dead letter events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}

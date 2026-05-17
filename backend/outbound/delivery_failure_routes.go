package outbound

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultDeliveryFailuresLimit = 100

// ListDeliveryFailuresRoute returns recent failed delivery attempts with optional mapping filters.
func ListDeliveryFailuresRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	destinationID := r.URL.Query().Get("destination_id")
	topicID := r.URL.Query().Get("topic_id")

	limit := defaultDeliveryFailuresLimit
	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	records, err := NewOutboundRepository().ListDeliveryFailureEvents(r.Context(), destinationID, topicID, limit)
	if err != nil {
		http.Error(w, "failed to read delivery failures", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(records)
}

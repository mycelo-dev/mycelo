package outbound

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mycelo-dev/mycelo/backend/auth"
)

const (
	defaultDeadLetterEventsLimit       = 100
	maxBulkDeadLetterReplayLimit       = 25
	bulkDeadLetterReplayConfirmation   = "REPLAY_FILTERED_DLQ"
	bulkDeadLetterReplayCooldownPeriod = 30 * time.Second
)

var bulkDeadLetterReplayLimiter = struct {
	sync.Mutex
	lastByScope map[string]time.Time
}{
	lastByScope: make(map[string]time.Time),
}

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
		Confirmation      string `json:"confirmation"`
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
	if payload.DeadLetterEventID == 0 {
		if payload.Limit <= 0 {
			limit = maxBulkDeadLetterReplayLimit
		}
		if payload.Confirmation != bulkDeadLetterReplayConfirmation {
			http.Error(w, "bulk DLQ replay requires confirmation REPLAY_FILTERED_DLQ", http.StatusBadRequest)
			return
		}
		if limit > maxBulkDeadLetterReplayLimit {
			http.Error(w, "bulk DLQ replay limit exceeds maximum", http.StatusBadRequest)
			return
		}
		if retryAfter, ok := allowBulkDeadLetterReplay(r); !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			http.Error(w, "bulk DLQ replay rate limit exceeded", http.StatusTooManyRequests)
			return
		}
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

func allowBulkDeadLetterReplay(r *http.Request) (time.Duration, bool) {
	scope := r.RemoteAddr
	if authContext, err := auth.FromContext(r.Context()); err == nil {
		scope = authContext.TenantPublicId + ":" + authContext.TeamPublicId
	}

	now := time.Now()
	bulkDeadLetterReplayLimiter.Lock()
	defer bulkDeadLetterReplayLimiter.Unlock()

	lastReplay, exists := bulkDeadLetterReplayLimiter.lastByScope[scope]
	if exists {
		elapsed := now.Sub(lastReplay)
		if elapsed < bulkDeadLetterReplayCooldownPeriod {
			return bulkDeadLetterReplayCooldownPeriod - elapsed, false
		}
	}

	bulkDeadLetterReplayLimiter.lastByScope[scope] = now
	return 0, true
}

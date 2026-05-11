package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Publish validates and stores an incoming event payload.
func Publish(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", 400)
		return
	}

	var payload struct {
		Topic     string      `json:"topic"`
		EventData interface{} `json:"event_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}

	err := PublishToStream(r.Context(), payload.Topic, payload.EventData)
	if err != nil {
		http.Error(w, "failed to publish event", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("event stored"))
}

// GetEvents reads topic events using optional cursor parameters.
func GetEvents(w http.ResponseWriter, r *http.Request) {

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "topic is required", 400)
		return
	}

	after := int64(0)

	if a := r.URL.Query().Get("after"); a != "" {
		parsed, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			http.Error(w, "invalid after timestamp", 400)
			return
		}
		after = parsed
	}

	offset := int64(0)

	if o := r.URL.Query().Get("offset"); o != "" {
		parsed, err := strconv.Atoi(o)
		if err != nil {
			http.Error(w, "invalid offset value", 400)
			return
		}
		offset = int64(parsed)
	}

	limit := defaultEventsFetchBatch
	if ql := strings.TrimSpace(r.URL.Query().Get("limit")); ql != "" {
		parsed, err := strconv.Atoi(ql)
		if err != nil || parsed < 1 {
			http.Error(w, "invalid limit value", http.StatusBadRequest)
			return
		}

		limit = parsed

		if limit > MaxEventsFetchUpperBound {
			limit = MaxEventsFetchUpperBound
		}
	}

	var events EventsResponse
	var err error

	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("order")), "desc") {
		events, err = GetEventsBeforeCursor(
			r.Context(),
			topic,
			offset,
			limit,
		)
	} else {
		events, err = GetEventsAfterCursor(
			r.Context(),
			topic,
			after,
			offset,
			limit,
		)
	}
	if err != nil {
		http.Error(w, "failed to fetch events", 500)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

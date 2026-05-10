package topics_management

import (
	"encoding/json"
	"net/http"
)

// CreateTopicRoute decodes and creates a new topic.
func CreateTopicRoute(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		TopicName string `json:"topic_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := CreateTopicServices(r.Context(), payload.TopicName)

	if err != nil {
		http.Error(w, "failed to insert the topic", 500)
		return
	}
}

// UpdateTopicRoute decodes and renames an existing topic.
func UpdateTopicRoute(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OldTopicName string `json:"old_topic_name"`
		NewTopicName string `json:"new_topic_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := UpdateTopicServices(r.Context(), payload.OldTopicName, payload.NewTopicName)

	if err != nil {
		http.Error(w, "failed to update the topic", 500)
		return
	}

}

// ListTopicsRoute returns all topics as JSON.
func ListTopicsRoute(w http.ResponseWriter, r *http.Request) {
	topics, err := ListTopicsServices(r.Context())
	if err != nil {
		http.Error(w, "failed to read topics", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topics)
}

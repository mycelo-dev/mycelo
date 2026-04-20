package topics_management

import (
	"encoding/json"
	"net/http"
)

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

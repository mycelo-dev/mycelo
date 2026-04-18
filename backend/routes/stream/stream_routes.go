package stream

import (
	"encoding/json"
	"log"
	"net/http"

	stream "github.com/mycelo-dev/mycelo/backend/stream"
)

func publish(w http.ResponseWriter, r *http.Request) {

	var publish_payload struct {
		Topic     string      `json:"topic"`
		EventData interface{} `json:"event_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&publish_payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stream.PublishToStream(r.Context(), publish_payload.Topic, publish_payload.EventData)
}

func HandleRequests() {
	http.HandleFunc("/publish", publish)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

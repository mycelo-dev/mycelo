package destination_management

import (
	"encoding/json"
	"net/http"
)

func CreateDestinationRoute(w http.ResponseWriter, r *http.Request) {

	if err := json.NewDecoder(r.Body).Decode(&CreateDestination); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := CreateDestinationServices(r.Context(), CreateDestination.Destination_name, CreateDestination.Destination_address)

	if err != nil {
		http.Error(w, "failed to create the destination", 500)
		return
	}
}

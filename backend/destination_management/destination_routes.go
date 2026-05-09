package destination_management

import (
	"encoding/json"
	"net/http"
)

func CreateDestinationRoute(w http.ResponseWriter, r *http.Request) {

	var cd CreateDestination

	if err := json.NewDecoder(r.Body).Decode(&cd); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := CreateDestinationServices(r.Context(), cd.Destination_name, cd.Destination_address)

	if err != nil {
		http.Error(w, "failed to create the destination", 500)
		return
	}
}

func UpdateDestinationRoute(w http.ResponseWriter, r *http.Request) {

	var up UpdateDestination

	if err := json.NewDecoder(r.Body).Decode(&up); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := UpdateDestinationServices(r.Context(), up.Destination_name, up.Destination_address, up.Id)

	if err != nil {
		http.Error(w, "failed to update the destination", 500)
		return
	}
}

func UpdateDeliveryFlagRoute(w http.ResponseWriter, r *http.Request) {

	var udf UpdateDeliveryFlag

	if err := json.NewDecoder(r.Body).Decode(&udf); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := UpdateDeliveryFlagServices(r.Context(), udf.Id, udf.Delivery_flag)

	if err != nil {
		http.Error(w, "failed to update the delivery flag", 500)
		return
	}
}

func DeleteDestinationRoute(w http.ResponseWriter, r *http.Request) {

	var dd DeleteDestination

	if err := json.NewDecoder(r.Body).Decode(&dd); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := DeleteDestinationServices(r.Context(), dd.Id)

	if err != nil {
		http.Error(w, "failed to delete the destination", 500)
		return
	}
}

func AssignTopicToDestinationRoute(w http.ResponseWriter, r *http.Request) {

	var attd AssignTopicToDestination

	if err := json.NewDecoder(r.Body).Decode(&attd); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := AssignTopicToDestinationServices(r.Context(), attd.Destination_id, attd.Topic_id)

	if err != nil {
		http.Error(w, "failed to assign the topic to destination", 500)
		return
	}
}

func ListDestinationsRoute(w http.ResponseWriter, r *http.Request) {

	destinations, err := ListDestinationsServices(r.Context())
	if err != nil {
		http.Error(w, "failed to read destinations", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(destinations)
}

func ListDestinationTopicMappingsRoute(w http.ResponseWriter, r *http.Request) {

	mappings, err := ListDestinationTopicMappingsServices(r.Context())
	if err != nil {
		http.Error(w, "failed to read destination topic mappings", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mappings)
}

func DeleteDestinationTopicMappingRoute(w http.ResponseWriter, r *http.Request) {

	var ddtm DeleteDestinationTopicMapping

	if err := json.NewDecoder(r.Body).Decode(&ddtm); err != nil {
		http.Error(w, "failed to delete the topic for the destination", 500)
		return
	}

	err := DeleteDestinationTopicMappingServices(r.Context(), ddtm.Destination_id, ddtm.Topic_id)

	if err != nil {
		http.Error(w, "failed to delete the topic for the destination", 500)
		return
	}
}

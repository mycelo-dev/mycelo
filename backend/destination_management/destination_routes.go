package destination_management

import (
	"encoding/json"
	"errors"
	"net/http"
)

// CreateDestinationRoute decodes and creates a new destination.
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

// UpdateDestinationRoute decodes and updates an existing destination.
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

// UpdateDeliveryFlagRoute toggles delivery for a destination-topic mapping.
func UpdateDeliveryFlagRoute(w http.ResponseWriter, r *http.Request) {

	var udf UpdateDeliveryFlag

	if err := json.NewDecoder(r.Body).Decode(&udf); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := UpdateDeliveryFlagServices(r.Context(), udf.Destination_id, udf.Topic_id, udf.Delivery_flag)

	if err != nil {
		http.Error(w, "failed to update the delivery flag", 500)
		return
	}
}

// DeleteDestinationRoute deletes a destination when service rules allow it.
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

// AssignTopicToDestinationRoute links a topic to a destination.
func AssignTopicToDestinationRoute(w http.ResponseWriter, r *http.Request) {

	var attd AssignTopicToDestination

	if err := json.NewDecoder(r.Body).Decode(&attd); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := AssignTopicToDestinationServices(r.Context(), attd.Destination_id, attd.Topic_id, attd.DestinationTopicMappingPolicyInput)

	if err != nil {
		if isPolicyValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to assign the topic to destination", 500)
		return
	}
}

// ListDestinationsRoute returns all destinations as JSON.
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

// ListDestinationTopicMappingsRoute returns all destination-topic mappings as JSON.
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

// UpdateDestinationTopicMappingPolicyRoute updates customer-configurable delivery controls.
func UpdateDestinationTopicMappingPolicyRoute(w http.ResponseWriter, r *http.Request) {
	var update UpdateDestinationTopicMappingPolicy

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := UpdateDestinationTopicMappingPolicyServices(r.Context(), update.Destination_id, update.Topic_id, update.DestinationTopicMappingPolicyInput)
	if err != nil {
		if isPolicyValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to update destination topic mapping policy", 500)
		return
	}
}

// DeleteDestinationTopicMappingRoute removes a topic mapping from a destination.
func DeleteDestinationTopicMappingRoute(w http.ResponseWriter, r *http.Request) {

	var ddtm DeleteDestinationTopicMapping

	if err := json.NewDecoder(r.Body).Decode(&ddtm); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	err := DeleteDestinationTopicMappingServices(r.Context(), ddtm.Destination_id, ddtm.Topic_id)

	if err != nil {
		http.Error(w, "failed to delete the topic for the destination", 500)
		return
	}
}

func isPolicyValidationError(err error) bool {
	var validationError PolicyValidationError
	return errors.As(err, &validationError)
}

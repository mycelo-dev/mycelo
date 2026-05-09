package routes

import (
	"net/http"

	"github.com/mycelo-dev/mycelo/backend/api_key"
	"github.com/mycelo-dev/mycelo/backend/destination_management"
	"github.com/mycelo-dev/mycelo/backend/outbound"
	"github.com/mycelo-dev/mycelo/backend/stream"
	topics_routes "github.com/mycelo-dev/mycelo/backend/topics_management"
)

// NewMux registers the application's HTTP routes and returns a reusable mux.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	// stream routes
	mux.HandleFunc("/publish", stream.Publish)
	mux.HandleFunc("/events", stream.GetEvents)

	// topics routes
	mux.HandleFunc("/create_topic", topics_routes.CreateTopicRoute)
	mux.HandleFunc("/update_topic", topics_routes.UpdateTopicRoute)
	mux.HandleFunc("/topics", topics_routes.ListTopicsRoute)

	// destination routes
	mux.HandleFunc("/create_destination", destination_management.CreateDestinationRoute)
	mux.HandleFunc("/update_destination", destination_management.UpdateDestinationRoute)
	mux.HandleFunc("/update_destination_delivery_flag", destination_management.UpdateDeliveryFlagRoute)
	mux.HandleFunc("/delete_destination", destination_management.DeleteDestinationRoute)
	mux.HandleFunc("/assign_topic_to_destination", destination_management.AssignTopicToDestinationRoute)
	mux.HandleFunc("/destinations", destination_management.ListDestinationsRoute)
	mux.HandleFunc("/destination_topic_mappings", destination_management.ListDestinationTopicMappingsRoute)
	mux.HandleFunc("/update_destination_topic_mapping_policy", destination_management.UpdateDestinationTopicMappingPolicyRoute)
	mux.HandleFunc("/delete_topic_for_destination", destination_management.DeleteDestinationTopicMappingRoute)
	mux.HandleFunc("/dead_letter_events", outbound.ListDeadLetterEventsRoute)

	// api key routes
	mux.HandleFunc("/create_api_key", api_key.CreateApiKeyRoute)
	mux.HandleFunc("/revoke_api_key", api_key.RevokeApiKeyRoute)
	mux.HandleFunc("/rotate_api_key", api_key.RotateApiKeyRoute)

	return mux
}

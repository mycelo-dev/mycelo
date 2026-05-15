package routes

import (
	"expvar"
	"net/http"

	"github.com/mycelo-dev/mycelo/backend/account"
	"github.com/mycelo-dev/mycelo/backend/api_key"
	"github.com/mycelo-dev/mycelo/backend/auth"
	"github.com/mycelo-dev/mycelo/backend/auth/api_key_auth"
	"github.com/mycelo-dev/mycelo/backend/destination_management"
	"github.com/mycelo-dev/mycelo/backend/outbound"
	"github.com/mycelo-dev/mycelo/backend/stream"
	topics_routes "github.com/mycelo-dev/mycelo/backend/topics_management"
)

// NewMux registers the application's HTTP routes and returns a reusable mux.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	// stream routes
	mux.HandleFunc("/publish", requireApiKey(stream.Publish))
	mux.HandleFunc("/events", requireApiKey(stream.GetEvents))

	// topics routes
	mux.HandleFunc("/create_topic", requireApiKey(topics_routes.CreateTopicRoute))
	mux.HandleFunc("/update_topic", requireApiKey(topics_routes.UpdateTopicRoute))
	mux.HandleFunc("/topics", requireApiKey(topics_routes.ListTopicsRoute))

	// destination routes
	mux.HandleFunc("/create_destination", requireApiKey(destination_management.CreateDestinationRoute))
	mux.HandleFunc("/update_destination", requireApiKey(destination_management.UpdateDestinationRoute))
	mux.HandleFunc("/update_destination_delivery_flag", requireApiKey(destination_management.UpdateDeliveryFlagRoute))
	mux.HandleFunc("/delete_destination", requireApiKey(destination_management.DeleteDestinationRoute))
	mux.HandleFunc("/assign_topic_to_destination", requireApiKey(destination_management.AssignTopicToDestinationRoute))
	mux.HandleFunc("/destinations", requireApiKey(destination_management.ListDestinationsRoute))
	mux.HandleFunc("/destination_topic_mappings", requireApiKey(destination_management.ListDestinationTopicMappingsRoute))
	mux.HandleFunc("/update_destination_topic_mapping_policy", requireApiKey(destination_management.UpdateDestinationTopicMappingPolicyRoute))
	mux.HandleFunc("/delete_topic_for_destination", requireApiKey(destination_management.DeleteDestinationTopicMappingRoute))
	mux.HandleFunc("/dead_letter_events", requireApiKey(outbound.ListDeadLetterEventsRoute))
	mux.HandleFunc("/observability/outbound", requireApiKey(outbound.OutboundObservabilityRoute))
	mux.Handle("/debug/vars", expvar.Handler())

	// account and api key routes
	mux.HandleFunc("/signup", account.SignUpRoute)
	mux.HandleFunc("/login", account.LoginRoute)
	mux.HandleFunc("/teams", account.ListTeamsRoute)
	mux.HandleFunc("/create_team", account.CreateTeamRoute)
	mux.HandleFunc("/create_api_key", api_key.CreateApiKeyRoute)
	mux.HandleFunc("/revoke_api_key", requireApiKey(api_key.RevokeApiKeyRoute))
	mux.HandleFunc("/rotate_api_key", requireApiKey(api_key.RotateApiKeyRoute))

	return mux
}

func requireApiKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := api_key_auth.GetApiKeyFromRequest(r)
		if apiKey == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}

		authContext, err := api_key_auth.ApiKeyAuthenticator(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(auth.WithAuthContext(r.Context(), authContext)))
	}
}

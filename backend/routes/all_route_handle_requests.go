package routes

import (
	"expvar"
	"net/http"
	"strings"

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
	mux.HandleFunc("/console/events", requireSessionTeam(stream.GetEvents))
	mux.HandleFunc("/console/event_topics", requireSessionTeam(stream.ListEventTopicsRoute))

	// topics routes
	mux.HandleFunc("/create_topic", requireSessionTeam(topics_routes.CreateTopicRoute))
	mux.HandleFunc("/update_topic", requireSessionTeam(topics_routes.UpdateTopicRoute))
	mux.HandleFunc("/topics", requireSessionTeam(topics_routes.ListTopicsRoute))

	// destination routes
	mux.HandleFunc("/create_destination", requireSessionTeam(destination_management.CreateDestinationRoute))
	mux.HandleFunc("/update_destination", requireSessionTeam(destination_management.UpdateDestinationRoute))
	mux.HandleFunc("/update_destination_delivery_flag", requireSessionTeam(destination_management.UpdateDeliveryFlagRoute))
	mux.HandleFunc("/delete_destination", requireSessionTeam(destination_management.DeleteDestinationRoute))
	mux.HandleFunc("/assign_topic_to_destination", requireSessionTeam(destination_management.AssignTopicToDestinationRoute))
	mux.HandleFunc("/destinations", requireSessionTeam(destination_management.ListDestinationsRoute))
	mux.HandleFunc("/destination_topic_mappings", requireSessionTeam(destination_management.ListDestinationTopicMappingsRoute))
	mux.HandleFunc("/update_destination_topic_mapping_policy", requireSessionTeam(destination_management.UpdateDestinationTopicMappingPolicyRoute))
	mux.HandleFunc("/delete_topic_for_destination", requireSessionTeam(destination_management.DeleteDestinationTopicMappingRoute))
	mux.HandleFunc("/delivery_failures", requireSessionTeam(outbound.ListDeliveryFailuresRoute))
	mux.HandleFunc("/dead_letter_events", requireSessionTeam(outbound.ListDeadLetterEventsRoute))
	mux.HandleFunc("/observability/outbound", requireSessionTeam(outbound.OutboundObservabilityRoute))
	mux.Handle("/debug/vars", expvar.Handler())

	// account and api key routes
	mux.HandleFunc("/signup", account.SignUpRoute)
	mux.HandleFunc("/login", account.LoginRoute)
	mux.HandleFunc("/logout", account.LogoutRoute)
	mux.HandleFunc("/teams", account.ListTeamsRoute)
	mux.HandleFunc("/create_team", account.CreateTeamRoute)
	mux.HandleFunc("/create_api_key", api_key.CreateApiKeyRoute)
	mux.HandleFunc("/revoke_api_key", requireSessionTeam(api_key.RevokeApiKeyRoute))
	mux.HandleFunc("/rotate_api_key", requireSessionTeam(api_key.RotateApiKeyRoute))

	return mux
}

func requireSessionTeam(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := account.SessionContextFromRequest(r.Context(), r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		teamPublicID := strings.TrimSpace(r.Header.Get(account.TeamScopeHeader))
		if teamPublicID == "" {
			http.Error(w, "missing active team", http.StatusBadRequest)
			return
		}

		authContext, err := account.TeamAuthContextServices(r.Context(), session.TenantPublicId, session.UserPublicId, teamPublicID)
		if err != nil {
			http.Error(w, "invalid active team", http.StatusForbidden)
			return
		}

		next(w, r.WithContext(auth.WithAuthContext(r.Context(), authContext)))
	}
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

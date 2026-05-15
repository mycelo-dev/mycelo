package api_key

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/account"
	"github.com/mycelo-dev/mycelo/backend/auth"
)

// CreateApiKeyRoute issues or replaces an API key for a team selected by the signed-up account.
func CreateApiKeyRoute(w http.ResponseWriter, r *http.Request) {
	var payload CreateApiKeyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload.TeamPublicId = strings.TrimSpace(payload.TeamPublicId)
	if payload.TeamPublicId == "" {
		http.Error(w, "team_public_id is required", http.StatusBadRequest)
		return
	}

	session, err := account.SessionContextFromRequest(r.Context(), r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	apiKey, err := CreateApiKeyForTeamServices(r.Context(), session.TenantPublicId, session.UserPublicId, payload.TeamPublicId)
	if err != nil {
		http.Error(w, "error creating the api key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiKey)
}

// RevokeApiKeyRoute revokes the API key identified in the request body.
func RevokeApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	authContext, err := auth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = RevokeApiKeyServices(r.Context(), authContext.TenantPublicId, authContext.TeamPublicId)

	if err != nil {
		http.Error(w, "failed to revoke the API key", 500)
		return
	}
}

// RotateApiKeyRoute replaces the current API key and returns the new token.
func RotateApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	var rak RotateApiKeyResponse

	authContext, err := auth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rak, err = RotateApiKeyServices(r.Context(), authContext.TenantPublicId, authContext.TeamPublicId)

	if err != nil {
		http.Error(w, "error rotating the api key", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rak)

}

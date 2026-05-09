package api_key

import (
	"encoding/json"
	"net/http"
)

// CreateApiKeyRoute issues a new API key and returns it as JSON.
func CreateApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	var ak CreateApiKeyResponse

	ak, err := CreateApiKeyServices(r.Context())

	if err != nil {
		http.Error(w, "error creating the api key", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ak)

}

// RevokeApiKeyRoute revokes the API key identified in the request body.
func RevokeApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	var rak RevokeApiKeyPayload
	if err := json.NewDecoder(r.Body).Decode(&rak); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}

	err := RevokeApiKeyServices(r.Context(), rak.TenantPublicId, rak.TeamPublicId)

	if err != nil {
		http.Error(w, "failed to revoke the API key", 500)
		return
	}
}

// RotateApiKeyRoute replaces the current API key and returns the new token.
func RotateApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	var rak RotateApiKeyResponse

	rak, err := RotateApiKeyServices(r.Context())

	if err != nil {
		http.Error(w, "error rotating the api key", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rak)

}

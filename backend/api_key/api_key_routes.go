package api_key

import (
	"encoding/json"
	"net/http"
)

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

func RevokeApiKeyRoute(w http.ResponseWriter, r *http.Request) {

	var rak RevokeApiKeyPayload
	if err := json.NewDecoder(r.Body).Decode(&rak); err != nil {
		http.Error(w, "Invalid request body", 500)
		return
	}

	err := RevokeApiKeyServices(r.Context(), rak.TenantPublicId, rak.TeamPublicId)

	if err != nil {
		http.Error(w, "failed to revoke the API key", 500)
		return
	}
}

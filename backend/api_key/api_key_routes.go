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

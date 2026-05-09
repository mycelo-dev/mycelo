package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/api_key"
)

func TestRevokeApiKeyRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/revoke_api_key", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	api_key.RevokeApiKeyRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("RevokeApiKeyRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

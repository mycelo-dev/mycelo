package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/account"
	"github.com/mycelo-dev/mycelo/backend/api_key"
)

func TestSignUpRouteRequiresTenantUserAndEmail(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"tenant_name":"","user_name":"","email":"","password":""}`))
	rr := httptest.NewRecorder()

	account.SignUpRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("SignUpRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestLoginRouteRequiresEmail(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":""}`))
	rr := httptest.NewRecorder()

	account.LoginRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("LoginRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateApiKeyRouteRequiresTeamContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/create_api_key", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	api_key.CreateApiKeyRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("CreateApiKeyRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateApiKeyRouteRequiresSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/create_api_key", strings.NewReader(`{"team_public_id":"team-123"}`))
	rr := httptest.NewRecorder()

	api_key.CreateApiKeyRoute(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("CreateApiKeyRoute returned status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRevokeApiKeyRouteRequiresAuthContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/revoke_api_key", nil)
	rr := httptest.NewRecorder()

	api_key.RevokeApiKeyRoute(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("RevokeApiKeyRoute returned status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

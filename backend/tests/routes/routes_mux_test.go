package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/account"
	"github.com/mycelo-dev/mycelo/backend/routes"
)

func TestNewMuxRegistersKnownRoutes(t *testing.T) {
	mux := routes.NewMux()

	testCases := []string{
		"/publish",
		"/events",
		"/console/events",
		"/console/event_topics",
		"/console/topic_heads",
		"/create_topic",
		"/create_destination",
		"/destination_topic_mappings",
		"/update_destination_topic_mapping_policy",
		"/dead_letter_events",
		"/observability/outbound",
		"/debug/vars",
		"/signup",
		"/login",
		"/logout",
		"/teams",
		"/create_team",
		"/create_api_key",
		"/revoke_api_key",
		"/rotate_api_key",
	}

	for _, path := range testCases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("route %s was not registered", path)
		}

		if handler == nil {
			t.Fatalf("route %s resolved to nil handler", path)
		}
	}
}

func TestConsoleRouteRequiresSessionCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	rec := httptest.NewRecorder()

	routes.NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestConsoleRouteRequiresActiveTeamHeader(t *testing.T) {
	t.Setenv("MYCELO_SESSION_JWT_SECRET", "test-session-secret")

	token, err := account.CreateSession(context.Background(), account.SignUpResponse{
		TenantPublicId: "tenant-123",
		UserPublicId:   "user-456",
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	req.AddCookie(&http.Cookie{Name: account.SessionCookieName, Value: token})
	rec := httptest.NewRecorder()

	routes.NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestStreamRouteStillRequiresApiKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()

	routes.NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

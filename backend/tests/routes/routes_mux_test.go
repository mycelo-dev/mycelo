package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/routes"
)

func TestNewMuxRegistersKnownRoutes(t *testing.T) {
	mux := routes.NewMux()

	testCases := []string{
		"/publish",
		"/events",
		"/create_topic",
		"/create_destination",
		"/destination_topic_mappings",
		"/update_destination_topic_mapping_policy",
		"/dead_letter_events",
		"/create_api_key",
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

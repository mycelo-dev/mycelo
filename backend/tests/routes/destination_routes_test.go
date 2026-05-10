package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	destination_management "github.com/mycelo-dev/mycelo/backend/destination_management"
)

func TestCreateDestinationRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/create_destination", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.CreateDestinationRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("CreateDestinationRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateDestinationRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update_destination", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.UpdateDestinationRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("UpdateDestinationRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateDeliveryFlagRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update_destination_delivery_flag", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.UpdateDeliveryFlagRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("UpdateDeliveryFlagRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteDestinationRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/delete_destination", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.DeleteDestinationRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("DeleteDestinationRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAssignTopicToDestinationRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/assign_topic_to_destination", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.AssignTopicToDestinationRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("AssignTopicToDestinationRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteDestinationTopicMappingRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/delete_topic_for_destination", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.DeleteDestinationTopicMappingRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("DeleteDestinationTopicMappingRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateDestinationTopicMappingPolicyRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update_destination_topic_mapping_policy", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	destination_management.UpdateDestinationTopicMappingPolicyRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("UpdateDestinationTopicMappingPolicyRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

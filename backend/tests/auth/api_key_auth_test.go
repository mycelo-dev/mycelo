package tests

import (
	"testing"

	"github.com/mycelo-dev/mycelo/backend/auth/api_key_auth"
)

func TestCompareApiKeyHash(t *testing.T) {
	if !api_key_auth.CompareApiKeyHash("same", "same") {
		t.Fatal("CompareApiKeyHash should return true for equal strings")
	}

	if api_key_auth.CompareApiKeyHash("left", "right") {
		t.Fatal("CompareApiKeyHash should return false for different strings")
	}
}

func TestApiKeyParts(t *testing.T) {
	apiKey := "mc_tenant-123_team-456_secret-789"

	if got := api_key_auth.GetTenantPublicIdFromApiKey(apiKey); got != "tenant-123" {
		t.Fatalf("GetTenantPublicIdFromApiKey returned %q", got)
	}

	if got := api_key_auth.GetTeamPublicIdFromApiKey(apiKey); got != "team-456" {
		t.Fatalf("GetTeamPublicIdFromApiKey returned %q", got)
	}

	if got := api_key_auth.GetHashStringFromApiKey(apiKey); got != "secret-789" {
		t.Fatalf("GetHashStringFromApiKey returned %q", got)
	}
}

func TestApiKeyPartsMalformedInput(t *testing.T) {
	apiKey := "invalid"

	if got := api_key_auth.GetTenantPublicIdFromApiKey(apiKey); got != "" {
		t.Fatalf("GetTenantPublicIdFromApiKey returned %q, want empty string", got)
	}

	if got := api_key_auth.GetTeamPublicIdFromApiKey(apiKey); got != "" {
		t.Fatalf("GetTeamPublicIdFromApiKey returned %q, want empty string", got)
	}

	if got := api_key_auth.GetHashStringFromApiKey(apiKey); got != "" {
		t.Fatalf("GetHashStringFromApiKey returned %q, want empty string", got)
	}
}

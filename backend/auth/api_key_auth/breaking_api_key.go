package api_key_auth

import "strings"

// GetTenantPublicIdFromApiKey extracts the tenant identifier from a raw API key.
func GetTenantPublicIdFromApiKey(api_key string) string {
	parts := strings.Split(api_key, "_")
	if len(parts) <= 1 {
		return ""
	}

	return parts[1]
}

// GetTeamPublicIdFromApiKey extracts the team identifier from a raw API key.
func GetTeamPublicIdFromApiKey(api_key string) string {
	parts := strings.Split(api_key, "_")
	if len(parts) <= 2 {
		return ""
	}

	return parts[2]
}

// GetHashStringFromApiKey extracts the unhashed secret segment from a raw API key.
func GetHashStringFromApiKey(api_key string) string {
	parts := strings.Split(api_key, "_")
	if len(parts) <= 3 {
		return ""
	}

	return parts[3]
}

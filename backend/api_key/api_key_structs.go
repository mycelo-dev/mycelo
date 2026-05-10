package api_key

// CreateApiKeyResponse returns a newly generated API key to the caller.
type CreateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

// RevokeApiKeyPayload identifies the tenant and team whose key should be revoked.
type RevokeApiKeyPayload struct {
	TenantPublicId string `json:"tenant_public_id"`
	TeamPublicId   string `json:"team_public_id"`
}

// RotateApiKeyResponse returns the replacement key after rotation.
type RotateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

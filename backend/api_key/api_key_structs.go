package api_key

// CreateApiKeyResponse returns a newly generated API key to the caller.
type CreateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

// CreateApiKeyPayload identifies the team selected by the signed-up account.
type CreateApiKeyPayload struct {
	TeamPublicId string `json:"team_public_id"`
}

// RotateApiKeyResponse returns the replacement key after rotation.
type RotateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

package api_key

type CreateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

type RevokeApiKeyPayload struct {
	TenantPublicId string `json:"tenant_public_id"`
	TeamPublicId   string `json:"team_public_id"`
}

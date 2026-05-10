package auth

// AuthContext carries the tenant and team resolved during authentication.
type AuthContext struct {
	TenantPublicId string
	TeamPublicId   string
}

package account

// SignUpPayload contains the user-provided account setup fields.
type SignUpPayload struct {
	TenantName string `json:"tenant_name"`
	UserName   string `json:"user_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

// LoginPayload contains the email used to restore account context.
type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUpResponse returns the generated tenant and user scope.
type SignUpResponse struct {
	TenantPublicId string `json:"tenant_public_id"`
	UserPublicId   string `json:"user_public_id"`
	TenantName     string `json:"tenant_name"`
	UserName       string `json:"user_name"`
	Email          string `json:"email"`
	SessionToken   string `json:"session_token"`
}

// TeamRecord is a team visible to the signed-up account.
type TeamRecord struct {
	TeamPublicId string `json:"team_public_id"`
	TeamName     string `json:"team_name"`
}

// CreateTeamPayload contains the account context and team name for a new team.
type CreateTeamPayload struct {
	TeamName string `json:"team_name"`
}

type loginRecord struct {
	Account      SignUpResponse
	PasswordHash string
}

// SessionContext carries account scope restored from an operator-console session token.
type SessionContext struct {
	TenantPublicId string
	UserPublicId   string
}

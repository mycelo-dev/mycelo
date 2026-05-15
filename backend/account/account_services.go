package account

import "context"

// SignUpServices creates a tenant account and its first user.
func SignUpServices(ctx context.Context, tenantName string, userName string, email string, password string) (SignUpResponse, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return SignUpResponse{}, err
	}

	account, err := SignUpRepository(ctx, tenantName, userName, email, passwordHash)
	if err != nil {
		return SignUpResponse{}, err
	}

	return AccountWithSession(ctx, account)
}

// LoginServices restores account context from an email address.
func LoginServices(ctx context.Context, email string, password string) (SignUpResponse, error) {
	record, err := LoginRepository(ctx, email)
	if err != nil {
		return SignUpResponse{}, err
	}

	if !VerifyPassword(password, record.PasswordHash) {
		return SignUpResponse{}, errInvalidPassword
	}

	return AccountWithSession(ctx, record.Account)
}

// CreateTeamServices creates a team under a signed-up account.
func CreateTeamServices(ctx context.Context, tenantPublicId string, userPublicId string, teamName string) (TeamRecord, error) {
	return CreateTeamRepository(ctx, tenantPublicId, userPublicId, teamName)
}

// ListTeamsServices lists teams for a signed-up account.
func ListTeamsServices(ctx context.Context, tenantPublicId string, userPublicId string) ([]TeamRecord, error) {
	return ListTeamsRepository(ctx, tenantPublicId, userPublicId)
}

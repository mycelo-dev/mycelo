package account

import (
	"context"

	"github.com/mycelo-dev/mycelo/backend/auth"
	"github.com/mycelo-dev/mycelo/backend/internal/apikeytoken"
)

const defaultSignupTeamName = "Default team"

// SignUpServices creates a tenant account and its first user.
func SignUpServices(ctx context.Context, tenantName string, userName string, email string, password string) (SignUpResponse, error) {
	if err := requireSessionSigningSecret(); err != nil {
		return SignUpResponse{}, err
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return SignUpResponse{}, err
	}

	apiKeySecret, apiKeyHash, err := apikeytoken.CreateSecret()
	if err != nil {
		return SignUpResponse{}, err
	}

	account, err := SignUpRepository(ctx, tenantName, userName, email, passwordHash, defaultSignupTeamName, apiKeyHash)
	if err != nil {
		return SignUpResponse{}, err
	}
	account.ApiKey = apikeytoken.Build(account.TenantPublicId, account.TeamPublicId, apiKeySecret)

	return AccountWithSession(ctx, account)
}

// LoginServices restores account context from an email address.
func LoginServices(ctx context.Context, email string, password string) (SignUpResponse, error) {
	if err := requireSessionSigningSecret(); err != nil {
		return SignUpResponse{}, err
	}

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

// TeamAuthContextServices validates a selected team and returns repository auth scope.
func TeamAuthContextServices(ctx context.Context, tenantPublicId string, userPublicId string, teamPublicId string) (auth.AuthContext, error) {
	team, err := ReadTeamForTenantUserRepository(ctx, tenantPublicId, userPublicId, teamPublicId)
	if err != nil {
		return auth.AuthContext{}, err
	}

	return auth.AuthContext{
		TenantPublicId: tenantPublicId,
		TeamPublicId:   team.TeamPublicId,
	}, nil
}

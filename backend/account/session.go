package account

import (
	"context"
	"net/http"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/core"
)

const sessionHeader = "X-Mycelo-Session"

// CreateSession creates a new operator-console session token and stores its hash.
func CreateSession(ctx context.Context, account SignUpResponse) (string, error) {
	randomBytes, err := core.GetRandomBytes(32)
	if err != nil {
		return "", err
	}

	sessionToken := "ms_" + core.GetHexString(randomBytes)
	sessionHash := core.GetHashString(sessionToken)

	if err := StoreSessionRepository(ctx, account.TenantPublicId, account.UserPublicId, sessionHash); err != nil {
		return "", err
	}

	return sessionToken, nil
}

// AccountWithSession attaches a newly-created session token to an account response.
func AccountWithSession(ctx context.Context, account SignUpResponse) (SignUpResponse, error) {
	sessionToken, err := CreateSession(ctx, account)
	if err != nil {
		return SignUpResponse{}, err
	}

	account.SessionToken = sessionToken
	return account, nil
}

// SessionContextFromRequest restores account scope from the operator-console session header.
func SessionContextFromRequest(ctx context.Context, r *http.Request) (SessionContext, error) {
	sessionToken := strings.TrimSpace(r.Header.Get(sessionHeader))
	if sessionToken == "" {
		return SessionContext{}, errMissingSession
	}

	return ReadSessionRepository(ctx, core.GetHashString(sessionToken))
}

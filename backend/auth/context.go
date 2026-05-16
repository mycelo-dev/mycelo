package auth

import (
	"context"
	"errors"
)

type authContextKey struct{}

var ErrMissingAuthContext = errors.New("missing auth context")

// WithAuthContext stores the authenticated tenant-team scope on a request context.
func WithAuthContext(ctx context.Context, authContext AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, authContext)
}

// FromContext returns the authenticated tenant-team scope for scoped operations.
func FromContext(ctx context.Context) (AuthContext, error) {
	authContext, ok := ctx.Value(authContextKey{}).(AuthContext)
	if !ok {
		return AuthContext{}, ErrMissingAuthContext
	}

	return authContext, nil
}

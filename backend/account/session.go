package account

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mycelo-dev/mycelo/backend/configs"
)

const (
	SessionCookieName = "mycelo_session"
	TeamScopeHeader   = "X-Mycelo-Team"
	sessionTTL        = time.Hour
)

var jwtEncoding = base64.RawURLEncoding

type sessionClaims struct {
	TenantPublicID string `json:"tenant_public_id"`
	UserPublicID   string `json:"user_public_id"`
	ExpiresAt      int64  `json:"exp"`
}

// CreateSession creates a signed operator-console JWT.
func CreateSession(ctx context.Context, account SignUpResponse) (string, error) {
	claims := sessionClaims{
		TenantPublicID: account.TenantPublicId,
		UserPublicID:   account.UserPublicId,
		ExpiresAt:      time.Now().Add(sessionTTL).Unix(),
	}

	token, err := signSessionJWT(claims)
	if err != nil {
		return "", err
	}

	return token, nil
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

// SetSessionCookie stores the signed session token in a hardened browser cookie.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(sessionTTL),
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearSessionCookie expires the browser session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// SessionContextFromRequest restores account scope from the operator-console JWT cookie.
func SessionContextFromRequest(ctx context.Context, r *http.Request) (SessionContext, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return SessionContext{}, errMissingSession
	}

	sessionToken := strings.TrimSpace(cookie.Value)
	if sessionToken == "" {
		return SessionContext{}, errMissingSession
	}

	claims, err := verifySessionJWT(sessionToken)
	if err != nil {
		return SessionContext{}, err
	}

	return SessionContext{
		TenantPublicId: claims.TenantPublicID,
		UserPublicId:   claims.UserPublicID,
	}, nil
}

func signSessionJWT(claims sessionClaims) (string, error) {
	header, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signingInput := jwtEncoding.EncodeToString(header) + "." + jwtEncoding.EncodeToString(payload)
	signature, err := signJWTInput(signingInput)
	if err != nil {
		return "", err
	}

	return signingInput + "." + jwtEncoding.EncodeToString(signature), nil
}

func verifySessionJWT(token string) (sessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return sessionClaims{}, errInvalidSession
	}

	expectedSignature, err := signJWTInput(parts[0] + "." + parts[1])
	if err != nil {
		return sessionClaims{}, err
	}

	actualSignature, err := jwtEncoding.DecodeString(parts[2])
	if err != nil {
		return sessionClaims{}, errInvalidSession
	}

	if subtle.ConstantTimeCompare(actualSignature, expectedSignature) != 1 {
		return sessionClaims{}, errInvalidSession
	}

	payload, err := jwtEncoding.DecodeString(parts[1])
	if err != nil {
		return sessionClaims{}, errInvalidSession
	}

	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return sessionClaims{}, errInvalidSession
	}
	if claims.TenantPublicID == "" || claims.UserPublicID == "" || claims.ExpiresAt <= time.Now().Unix() {
		return sessionClaims{}, errInvalidSession
	}

	return claims, nil
}

func signJWTInput(signingInput string) ([]byte, error) {
	if err := requireSessionSigningSecret(); err != nil {
		return nil, err
	}

	secret := configs.GetSessionJWTSigningSecret()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return mac.Sum(nil), nil
}

func requireSessionSigningSecret() error {
	if strings.TrimSpace(configs.GetSessionJWTSigningSecret()) == "" {
		return errMissingSessionSigningSecret
	}

	return nil
}

var errInvalidSession = errors.New("invalid session token")
var errMissingSessionSigningSecret = errors.New("missing session JWT signing secret")

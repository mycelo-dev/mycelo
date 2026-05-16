package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/account"
)

func TestSessionContextFromRequestUsesSignedJWTCookie(t *testing.T) {
	t.Setenv("MYCELO_SESSION_JWT_SECRET", "test-session-secret")

	token, err := account.CreateSession(context.Background(), account.SignUpResponse{
		TenantPublicId: "tenant-123",
		UserPublicId:   "user-456",
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	req.AddCookie(&http.Cookie{Name: account.SessionCookieName, Value: token})

	session, err := account.SessionContextFromRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("SessionContextFromRequest returned error: %v", err)
	}

	if session.TenantPublicId != "tenant-123" || session.UserPublicId != "user-456" {
		t.Fatalf("session context = %#v", session)
	}
}

func TestSessionContextFromRequestRejectsTamperedJWTCookie(t *testing.T) {
	t.Setenv("MYCELO_SESSION_JWT_SECRET", "test-session-secret")

	token, err := account.CreateSession(context.Background(), account.SignUpResponse{
		TenantPublicId: "tenant-123",
		UserPublicId:   "user-456",
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	req.AddCookie(&http.Cookie{Name: account.SessionCookieName, Value: token + "tampered"})

	if _, err := account.SessionContextFromRequest(context.Background(), req); err == nil {
		t.Fatal("SessionContextFromRequest accepted a tampered token")
	}
}

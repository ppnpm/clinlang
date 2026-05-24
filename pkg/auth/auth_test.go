package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clinlang/pkg/auth"
)

func captureUserHandler(t *testing.T) (http.Handler, *string) {
	t.Helper()
	var captured string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = auth.UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	return h, &captured
}

func TestMiddleware_LocalModeAlwaysInjectsLocal(t *testing.T) {
	h, captured := captureUserHandler(t)
	mw := auth.Middleware("local", h)

	req := httptest.NewRequest(http.MethodGet, "/whatever", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("local mode: status: want 200, got %d", rec.Code)
	}
	if *captured != auth.LocalUser {
		t.Errorf("local mode: user: want %q, got %q", auth.LocalUser, *captured)
	}
}

func TestMiddleware_LocalModeIgnoresHeader(t *testing.T) {
	// Even if the header is set, local mode should not honor it.
	h, captured := captureUserHandler(t)
	mw := auth.Middleware("local", h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(auth.HostedUserHeader, "attacker")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != auth.LocalUser {
		t.Errorf("local mode must ignore %s header, got user=%q", auth.HostedUserHeader, *captured)
	}
}

func TestMiddleware_HostedModeRequiresHeader(t *testing.T) {
	h, _ := captureUserHandler(t)
	mw := auth.Middleware("hosted", h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("hosted no header: status: want 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "X-Forwarded-User") {
		t.Errorf("hosted 401 body should mention X-Forwarded-User; got %q", rec.Body.String())
	}
}

func TestMiddleware_HostedModeAcceptsHeader(t *testing.T) {
	h, captured := captureUserHandler(t)
	mw := auth.Middleware("hosted", h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(auth.HostedUserHeader, "dr.smith")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("hosted with header: status: want 200, got %d", rec.Code)
	}
	if *captured != "dr.smith" {
		t.Errorf("hosted user: want %q, got %q", "dr.smith", *captured)
	}
}

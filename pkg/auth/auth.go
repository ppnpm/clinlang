// Package auth provides per-request user identification middleware.
//
// ClinLang's HTTP API never handles passwords or sessions itself.
//
//   - In LOCAL mode (single-user laptop), every request is attributed
//     to the synthetic user "local". No header is required; there is
//     no trust boundary inside the local machine.
//
//   - In HOSTED mode (clinic deployment behind a reverse proxy), the
//     reverse proxy is the trust boundary. It performs the actual
//     authentication (SSO/OIDC/Basic/etc.) and forwards the username
//     in the X-Forwarded-User header. The Go server treats this
//     header as authoritative.
//
// If a hosted-mode request arrives without X-Forwarded-User, it is
// rejected with 401 — the assumption is that a misconfigured proxy is
// safer to fail closed than to silently allow anonymous access.
package auth

import (
	"context"
	"encoding/json"
	"net/http"
)

// HostedUserHeader is the header set by the reverse proxy in hosted mode.
const HostedUserHeader = "X-Forwarded-User"

// LocalUser is the synthetic identity used in local mode.
const LocalUser = "local"

type contextKey string

const userKey contextKey = "clinlang.user"

// Middleware extracts the acting user from the request and stores it
// in the request context for downstream handlers.
//
//	mode = "local"  → always sets user to LocalUser
//	mode = "hosted" → reads HostedUserHeader, 401 if missing/empty
func Middleware(mode string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := LocalUser
		if mode == "hosted" {
			user = r.Header.Get(HostedUserHeader)
			if user == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "unauthorized: X-Forwarded-User header missing",
				})
				return
			}
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext returns the user that the middleware attributed to
// this request. Returns "" if the middleware was not applied.
func UserFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(string); ok {
		return v
	}
	return ""
}

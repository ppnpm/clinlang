package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// corsMiddleware applies the deployment-mode-appropriate CORS posture.
//
// LOCAL: emits CORS headers reflecting the request's Origin if it is
// an http://localhost:* or http://127.0.0.1:* origin. This lets a
// Next.js dev server on :3000 talk to the API on :8080 during
// development without ad-hoc workarounds.
//
// HOSTED: emits no CORS headers at all. Browsers will refuse any
// cross-origin request, which is the correct behavior — the reverse
// proxy is the only intended caller and operates same-origin.
func corsMiddleware(mode string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mode == ModeLocal {
			origin := r.Header.Get("Origin")
			if isLocalOrigin(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		origin == "http://localhost" ||
		origin == "http://127.0.0.1"
}

// writeJSON serialises v as JSON and writes it to w with content-type
// set. Encoding errors fall back to a generic 500.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"internal encoding error"}`, http.StatusInternalServerError)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

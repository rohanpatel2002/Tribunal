package main

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// RequireAuth enforces a simple Bearer token check for Enterprise endpoints.
// CORS headers are now handled by the corsWrapper middleware in main.go
func RequireAuth(expectedKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests (CORS headers already set by corsWrapper)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if expectedKey == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "server misconfiguration: enterprise API key required but not set",
			})
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "authorization required",
			})
			return
		}

		// Bearer <token>
		scheme, token, found := strings.Cut(authHeader, " ")
		if !found || !strings.EqualFold(strings.TrimSpace(scheme), "bearer") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or missing api key",
			})
			return
		}

		token = strings.TrimSpace(token)
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedKey)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or missing api key",
			})
			return
		}

		next(w, r)
	}
}

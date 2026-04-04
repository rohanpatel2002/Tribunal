package main

import (
	"net/http"
	"strings"
)

// RequireAuth enforces a simple Bearer token check for Enterprise endpoints.
func RequireAuth(expectedKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] != expectedKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or missing api key",
			})
			return
		}

		next(w, r)
	}
}

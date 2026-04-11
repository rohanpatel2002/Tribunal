package main

import (
	"net/http"
	"strings"
)

// CORSMiddleware validates and applies CORS headers with origin whitelist
func CORSMiddleware(allowedOrigins string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse allowed origins from environment
		origins := strings.Split(allowedOrigins, ",")
		if allowedOrigins == "" {
			origins = []string{"http://localhost:3000"}
		}

		// Validate request origin
		origin := r.Header.Get("Origin")
		isAllowed := false
		for _, allowed := range origins {
			if strings.TrimSpace(allowed) == origin {
				isAllowed = true
				break
			}
		}

		// Only set origin header if allowed
		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// PoliciesHandler handles security policies CRUD operations
func PoliciesHandler(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repository := r.URL.Query().Get("repository")
		if repository == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "repository parameter required"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			getPoliciesHandler(w, r, repo, repository)
		case http.MethodPost:
			createPolicyHandler(w, r, repo, repository)
		case http.MethodDelete:
			deletePolicyHandler(w, r, repo, repository)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// getPoliciesHandler retrieves all active policies for a repository
func getPoliciesHandler(w http.ResponseWriter, r *http.Request, repo Repository, repository string) {
	ctx := r.Context()

	policies, err := repo.GetSecurityPolicies(ctx, repository)
	if err != nil {
		log.Printf("failed to fetch policies: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch policies"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  policies,
		"count": len(policies),
	})
}

// createPolicyHandler creates a new security policy
func createPolicyHandler(w http.ResponseWriter, r *http.Request, repo Repository, repository string) {
	var policy SecurityPolicy
	err := json.NewDecoder(r.Body).Decode(&policy)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if policy.PolicyName == "" || policy.PolicyType == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "policyName and policyType are required"})
		return
	}

	// Set metadata
	policy.Repository = repository
	policy.IsActive = true
	policy.CreatedAt = time.Now()
	policy.CreatedBy = r.Header.Get("X-User-ID")
	if policy.CreatedBy == "" {
		policy.CreatedBy = "system"
	}

	ctx := r.Context()
	err = repo.SaveSecurityPolicy(ctx, &policy)
	if err != nil {
		log.Printf("failed to save policy: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to save policy"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

// deletePolicyHandler deactivates a security policy
func deletePolicyHandler(w http.ResponseWriter, r *http.Request, repo Repository, repository string) {
	policyName := r.URL.Query().Get("policyName")
	if policyName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "policyName parameter required"})
		return
	}

	actor := r.Header.Get("X-User-ID")
	if actor == "" {
		actor = "system"
	}

	ctx := r.Context()
	err := repo.DeleteSecurityPolicy(ctx, repository, policyName, actor)
	if err != nil {
		log.Printf("failed to delete policy: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete policy"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheckHandler provides comprehensive health status
func HealthCheckHandler(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"checks": map[string]string{
				"database": "ok",
			},
		}

		// Quick database health check
		_, err := repo.GetRepositoryAuditSummary(ctx, "health-check")
		if err != nil {
			health["checks"].(map[string]string)["database"] = "degraded"
			health["status"] = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if health["status"] != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(health)
	}
}

// ErrorRecoveryMiddleware provides graceful error handling
func ErrorRecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "internal server error",
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()
		next(w, r)
	}
}

// RequestLoggingMiddleware logs all HTTP requests with correlation IDs
func RequestLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		startTime := time.Now()

		log.Printf("[%s] %s %s %s", correlationID, r.Method, r.RequestURI, r.RemoteAddr)

		// Call the handler
		next(w, r)

		duration := time.Since(startTime)
		log.Printf("[%s] completed in %d ms", correlationID, duration.Milliseconds())
	}
}

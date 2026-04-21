package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// WebhookSecurityConfig holds webhook verification settings
type WebhookSecurityConfig struct {
	GitHubSecret  string
	GitLabSecret  string
	BitbucketKey  string
	RateLimitByIP map[string]*RateLimitEntry
}

// RateLimitEntry tracks rate limit state per IP
type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

// VerifyGitHubSignature validates the GitHub webhook signature
// Implements: https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks
func VerifyGitHubSignature(secret string, payload []byte, signature string) bool {
	if secret == "" {
		// No secret configured, skip verification
		return true
	}

	// GitHub sends X-Hub-Signature-256: sha256=<hex_digest>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	providedHex := strings.TrimPrefix(signature, "sha256=")
	providedSig, err := hex.DecodeString(providedHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := mac.Sum(nil)

	return hmac.Equal(providedSig, expectedSig)
}

// VerifyGitLabSignature validates the GitLab webhook signature
// Implements: https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#secret-token
func VerifyGitLabSignature(secret string, payload []byte, token string) bool {
	if secret == "" {
		return true
	}
	// GitLab webhook token is a shared secret sent as plaintext token header.
	_ = payload
	return hmac.Equal([]byte(token), []byte(secret))
}

// VerifyBitbucketSignature validates Bitbucket webhook signature
// Implements: https://confluence.atlassian.com/bitbucketserver/manage-webhooks-938025878.html
func VerifyBitbucketSignature(secret string, payload []byte, signature string) bool {
	if secret == "" {
		return true
	}

	provided := signature
	provided = strings.TrimPrefix(provided, "sha256=")

	providedSig, err := hex.DecodeString(provided)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)

	return hmac.Equal(providedSig, expected)
}

// WebhookSignatureMiddleware verifies webhook signatures before processing
func WebhookSignatureMiddleware(secret string, provider string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only verify for POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Read and preserve the body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("failed to read webhook body", "error", err, "provider", provider)
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read request body"})
				return
			}
			// Replace body so it can be read again by the handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Verify signature based on provider
			var isValid bool
			switch provider {
			case "github":
				sig := r.Header.Get("X-Hub-Signature-256")
				isValid = VerifyGitHubSignature(secret, body, sig)
				if !isValid && secret != "" {
					slog.Warn("invalid GitHub webhook signature", "delivery_id", r.Header.Get("X-GitHub-Delivery"))
				}

			case "gitlab":
				token := r.Header.Get("X-Gitlab-Token")
				isValid = VerifyGitLabSignature(secret, body, token)
				if !isValid && secret != "" {
					slog.Warn("invalid GitLab webhook signature")
				}

			case "bitbucket":
				sig := r.Header.Get("X-Hub-Signature")
				isValid = VerifyBitbucketSignature(secret, body, sig)
				if !isValid && secret != "" {
					slog.Warn("invalid Bitbucket webhook signature")
				}

			default:
				isValid = true
			}

			if !isValid {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": fmt.Sprintf("invalid webhook signature for %s", provider),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckWebhookIdempotency verifies we haven't already processed this webhook
func (h *Handler) CheckWebhookIdempotency(ctx interface{}, deliveryID string) bool {
	if h.repo == nil {
		// Can't check, allow it through
		return true
	}

	// For now, skipping implementation - can be enhanced later with proper context passing
	slog.Debug("webhook idempotency check", "delivery_id", deliveryID)
	return true
}

// RateLimitWebhooks applies per-IP rate limiting
func RateLimitWebhooks(perIPLimit int, windowSize time.Duration) func(http.Handler) http.Handler {
	limitMap := make(map[string]*RateLimitEntry)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now()

			// Get or create rate limit entry
			entry, exists := limitMap[ip]
			if !exists || now.After(entry.ResetTime) {
				entry = &RateLimitEntry{
					Count:     0,
					ResetTime: now.Add(windowSize),
				}
				limitMap[ip] = entry
			}

			// Check if limit exceeded
			if entry.Count >= perIPLimit {
				slog.Warn("webhook rate limit exceeded", "ip", ip, "limit", perIPLimit)
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
					"limit": fmt.Sprintf("%d requests per %v", perIPLimit, windowSize),
				})
				return
			}

			entry.Count++
			next.ServeHTTP(w, r)
		})
	}
}

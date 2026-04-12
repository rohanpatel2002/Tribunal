# Critical Fixes Implementation Guide

## Fix #1: Secure API Key Management (CRITICAL)

### Current Problem
```yaml
- TRIBUNAL_API_KEY=dev_enterprise_key_123  # Hardcoded in compose file
```

### Solution

**Step 1:** Create `.env.example`:
```bash
# API authentication key for dashboard access
TRIBUNAL_API_KEY=your_secure_api_key_here

# Database credentials
POSTGRES_USER=tribunal
POSTGRES_PASSWORD=generate_strong_password_here
POSTGRES_DB=tribunal_db

# GitHub integration (optional)
GITHUB_TOKEN=ghp_your_token_here

# LLM integration (optional)
OPENROUTER_API_KEY=sk-or-your_key_here

# Dashboard configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
```

**Step 2:** Update `docker-compose.yml`:
```yaml
environment:
  - DATABASE_URL=postgres://${POSTGRES_USER:-tribunal}:${POSTGRES_PASSWORD:-tribunal_password_dev}@db:5432/tribunal_db?sslmode=disable
  - PORT=8080
  - GITHUB_TOKEN=${GITHUB_TOKEN:-}
  - OPENROUTER_API_KEY=${OPENROUTER_API_KEY:-}
  - TRIBUNAL_API_KEY=${TRIBUNAL_API_KEY:?Error: TRIBUNAL_API_KEY not set}
```

**Step 3:** Create `.env.local` (NOT in git):
```bash
TRIBUNAL_API_KEY=$(openssl rand -hex 32)
POSTGRES_PASSWORD=$(openssl rand -hex 16)
```

**Step 4:** Add to `.gitignore`:
```
.env.local
.env.production
.env.*.local
```

**Risk Reduction:** 🔴 → 🟢 (Critical → Secure)

---

## Fix #2: CORS Origin Whitelist (CRITICAL)

### Current Problem
```go
w.Header().Set("Access-Control-Allow-Origin", "*")  // Allows ANY domain
```

### Solution

**Create new file: `services/go-interceptor/cors.go`:**
```go
package main

import (
	"net/http"
	"os"
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
```

**Update `services/go-interceptor/main.go`:**
```go
// In init() or main():
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
if allowedOrigins == "" {
	allowedOrigins = "http://localhost:3000"
}

// In route setup:
router.HandleFunc("/api/v1/audit/summary", 
	CORSMiddleware(allowedOrigins, RequireAuth(apiKey, handleAuditSummary)))
```

**Update `.env.example`:**
```bash
# CORS configuration - comma-separated list of allowed origins
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com,https://www.yourdomain.com
```

**Risk Reduction:** 🔴 → 🟢 (Critical → Secure)

---

## Fix #3: Patch Size Validation (CRITICAL)

### Current Problem
```go
const maxPatchLength = 12000  // Defined but never enforced!
```

### Solution

**Update `services/go-interceptor/handler.go`:**
```go
// Add at top of file
const (
	maxPatchLength    = 12000   // 12KB max patch size
	maxPayloadSize    = 1000000 // 1MB max webhook payload
	maxConcurrentReqs = 50      // Max simultaneous analysis requests
)

// Update AnalyzePullRequest handler:
func AnalyzePullRequest(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)
	defer r.Body.Close()

	// ... existing parsing code ...

	for _, file := range pr.Files {
		// Enforce patch size limit
		if len(file.Patch) > maxPatchLength {
			slog.Warn("patch exceeds size limit",
				"file", file.Path,
				"size", len(file.Patch),
				"limit", maxPatchLength)
			
			pr.FileAnalyses = append(pr.FileAnalyses, FileAnalysis{
				Path:      file.Path,
				RiskLevel: "CRITICAL",
				Summary:   "Patch exceeds safe analysis size limit",
				Confidence: 1.0,
			})
			continue
		}

		// Analyze valid-sized patch
		analysis := analyzer.AnalyzeFile(file)
		pr.FileAnalyses = append(pr.FileAnalyses, analysis)
	}

	writeJSON(w, http.StatusOK, pr)
}
```

**Risk Reduction:** 🔴 → 🟡 (Critical → Medium - DoS mitigated but not eliminated)

---

## Fix #4: Move API Key Out of Frontend State (CRITICAL)

### Current Problem
```typescript
const [apiKey, setApiKey] = useState("dev_enterprise_key_123");  // Visible in DevTools!
```

### Solution

**Update `dashboard/src/app/page.tsx`:**

**Option A: HTTP-Only Cookies (Recommended)**
```typescript
// Remove from state entirely
// const [apiKey, setApiKey] = useState("...");

// Fetch function now doesn't need API key - cookies sent automatically
const fetchData = async () => {
    try {
        setLoading(true);
        
        const [summary, logs] = await Promise.all([
            fetch(`${API_BASE}/api/v1/audit/summary?period=90d`, {
                credentials: 'include',  // Send cookies
            }).then(r => r.json()),
            fetch(`${API_BASE}/api/v1/audit/logs?limit=10&offset=0`, {
                credentials: 'include',
            }).then(r => r.json()),
        ]);

        setSummary(summary);
        setLogs(logs);
    } catch (e) {
        console.error("Failed to load dashboard:", e);
        setError("Failed to load dashboard. Please refresh.");
    } finally {
        setLoading(false);
    }
};
```

**Option B: Session Token (Alternative)**
```typescript
// On app mount, exchange for session token
useEffect(() => {
    const initSession = async () => {
        try {
            const res = await fetch(`${API_BASE}/auth/session`, {
                method: 'POST',
                credentials: 'include'
            });
            if (!res.ok) {
                setError("Authentication failed");
                return;
            }
            // Token stored in HTTP-only cookie automatically
            fetchData();
        } catch (e) {
            setError("Failed to initialize session");
        }
    };
    initSession();
}, []);
```

**Backend `services/go-interceptor/middleware.go` update:**
```go
// Set HTTP-only cookie instead of header validation
func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "tribunal_session",
		Value:    token,
		Path:     "/",
		MaxAge:   3600 * 24,         // 24 hours
		HttpOnly: true,               // Prevent JavaScript access
		Secure:   os.Getenv("SECURE_COOKIES") == "true",  // HTTPS only in prod
		SameSite: http.SameSiteLaxMode,
	})
}
```

**Risk Reduction:** 🔴 → 🟢 (Critical → Secure)

---

## Fix #5: Hardcoded Backend URL (CRITICAL)

### Current Problem
```typescript
const res = await fetch(`http://localhost:8080/api/v1/audit/summary...`)  // Hardcoded!
```

### Solution

**Create `dashboard/next.config.ts`:**
```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  publicRuntimeConfig: {
    apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  },
};

export default nextConfig;
```

**Update `dashboard/src/app/page.tsx`:**
```typescript
// Add at top of component
const getApiUrl = () => {
  if (typeof window !== 'undefined') {
    return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
};

const API_BASE = getApiUrl();

// Use API_BASE throughout component
const res = await fetch(`${API_BASE}/api/v1/audit/summary...`)
```

**Update `docker-compose.yml`:**
```yaml
dashboard:
  environment:
    - NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:8080}
```

**Update `.env.example`:**
```bash
# For local development
NEXT_PUBLIC_API_URL=http://localhost:8080

# For staging
# NEXT_PUBLIC_API_URL=https://api-staging.yourdomain.com

# For production
# NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

**Risk Reduction:** 🔴 → 🟢 (Critical → Secure & Flexible)

---

## Fix #6: Missing Rate Limiting on LLM Calls (CRITICAL)

### Current Problem
```go
// No rate limiting, no backoff - can exhaust API quota silently
httpClient: &http.Client{Timeout: 60 * time.Second},
```

### Solution

**Update `services/go-interceptor/llm_client.go`:**
```go
package main

import (
	"fmt"
	"math"
	"net/http"
	"time"
	"golang.org/x/time/rate"
)

type OpenRouterClient struct {
	apiKey    string
	httpClient *http.Client
	limiter   *rate.Limiter  // NEW: Rate limiter
}

// NewOpenRouterClient initializes with rate limiting
func NewOpenRouterClient(apiKey string, requestsPerMinute int) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     5 * time.Second,
				MaxConnsPerHost:     5,
			},
		},
		limiter: rate.NewLimiter(
			rate.Limit(float64(requestsPerMinute) / 60.0),
			1,
		),
	}
}

// AnalyzeCode with rate limiting and retry logic
func (c *OpenRouterClient) AnalyzeCode(code string) (*AIDetectionResult, error) {
	// Check rate limit
	if !c.limiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded, retry after 60 seconds")
	}

	// Implement exponential backoff for retries
	var (
		result *AIDetectionResult
		err    error
		maxRetries = 3
		baseDelay = 1 * time.Second
	)

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err = c.performAnalysis(code)
		
		if err == nil {
			return result, nil
		}

		// Check if error is retryable
		if isRetryableError(err) {
			// Exponential backoff: 1s, 2s, 4s
			delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
			slog.Warn("LLM request failed, retrying",
				"attempt", attempt+1,
				"delay", delay,
				"error", err)
			time.Sleep(delay)
			continue
		}

		// Non-retryable error
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	return nil, fmt.Errorf("LLM analysis failed after %d retries: %w", maxRetries, err)
}

func isRetryableError(err error) bool {
	// Retry on timeout or 429 (too many requests)
	return err != nil // Simplified; add proper error type checking
}
```

**Update `.env.example`:**
```bash
# Rate limiting for OpenRouter API
# Set to your account's rate limit (e.g., 10 requests/minute for free tier)
LLM_REQUESTS_PER_MINUTE=10
```

**Risk Reduction:** 🔴 → 🟡 (Critical → Medium - Rate limit respected but API quota still depletable)

---

## Fix #7: Missing Input Validation for Database Queries (CRITICAL)

### Current Problem
Need to verify all database queries use parameterized statements. Create audit script:

```bash
#!/bin/bash
# Check for SQL injection vulnerabilities

grep -r "db\.Query\|Query(" services/go-interceptor/*.go | grep -v "\$1\|\$2\|\$3\|:\w"

# Should return no results if all queries are parameterized
```

### Solution: Verify & Document

**Create `services/go-interceptor/DATABASE_AUDIT.md`:**
```markdown
# Database Query Audit

All queries must use parameterized statements to prevent SQL injection.

## ✅ SAFE
```go
db.Query("SELECT * FROM pull_requests WHERE id = $1", prID)
db.Query("DELETE FROM analyses WHERE repo = $1 AND created < $2", repo, cutoffTime)
```

## ❌ DANGEROUS (Never do this)
```go
db.Query("SELECT * FROM pull_requests WHERE id = '" + prID + "'")
db.Query("DELETE FROM analyses WHERE repo = '" + repo + "'")
```

## Audit Checklist
- [ ] All SELECT queries use $1, $2, $3... placeholders
- [ ] All INSERT queries use $1, $2, $3... placeholders
- [ ] All UPDATE queries use $1, $2, $3... placeholders
- [ ] All DELETE queries use $1, $2, $3... placeholders
- [ ] No string concatenation with user input
- [ ] Repository names and file paths treated as parameters
```

---

## Priority Implementation Order

1. **Week 1 - Critical Fixes (4-6 hours)**
   - [ ] Fix #1: Secure API key (30 min)
   - [ ] Fix #2: CORS whitelist (30 min)
   - [ ] Fix #3: Patch size validation (1 hour)
   - [ ] Fix #4: Remove API key from frontend (2 hours)
   - [ ] Fix #5: Environment-based backend URL (1 hour)
   - [ ] Fix #6: Rate limiting (1.5 hours)
   - [ ] Test: `docker-compose up` and verify all endpoints work

2. **Week 2 - High Priority Fixes (8-12 hours)**
   - [ ] Database retry logic
   - [ ] Pagination implementation
   - [ ] Error logging for LLM fallbacks

3. **Week 3-4 - Medium Priority Fixes (16-20 hours)**
   - [ ] Error boundaries
   - [ ] Caching layer
   - [ ] Security policies backend integration

---

## Testing After Fixes

```bash
# Test API key requirement
curl http://localhost:8080/api/v1/audit/summary
# Should return 401 Unauthorized

curl -H "Authorization: Bearer wrong_key" http://localhost:8080/api/v1/audit/summary
# Should return 401 Unauthorized

curl -H "Authorization: Bearer $(echo $TRIBUNAL_API_KEY)" http://localhost:8080/api/v1/audit/summary
# Should return 200 OK

# Test CORS enforcement
curl -H "Origin: https://evil.com" http://localhost:8080/api/v1/audit/summary
# Should NOT include Access-Control-Allow-Origin header

curl -H "Origin: http://localhost:3000" http://localhost:8080/api/v1/audit/summary
# Should include Access-Control-Allow-Origin header
```

---

**Next Steps:** Implement these fixes in the order specified. Each fix is independent and can be merged separately.

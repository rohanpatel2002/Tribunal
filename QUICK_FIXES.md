# Quick-Start Fixes: Apply Immediately

This document contains copy-paste ready code to fix critical issues. **Apply these in the next 30 minutes.**

---

## Fix #1: CORS Middleware (30 minutes)

### Create file: `services/go-interceptor/cors.go`

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

### Update: `services/go-interceptor/main.go`

Replace this block:
```go
router.HandleFunc("/api/v1/audit/summary", RequireAuth(apiKey, handleAuditSummary))
router.HandleFunc("/api/v1/audit/logs", RequireAuth(apiKey, handleAuditLogs))
```

With this:
```go
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
if allowedOrigins == "" {
	allowedOrigins = "http://localhost:3000"
}

router.HandleFunc("/api/v1/audit/summary", 
	CORSMiddleware(allowedOrigins, RequireAuth(apiKey, handleAuditSummary)))
router.HandleFunc("/api/v1/audit/logs", 
	CORSMiddleware(allowedOrigins, RequireAuth(apiKey, handleAuditLogs)))
```

### Update: `docker-compose.yml` environment section

Add this line:
```yaml
- ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

---

## Fix #2: Patch Size Validation (20 minutes)

### Update: `services/go-interceptor/handler.go`

Add at the top of the file:
```go
const (
	maxPatchLength = 12000   // 12KB max patch size
	maxPayloadSize = 1000000 // 1MB max webhook payload
)
```

In the `AnalyzePullRequest` function, add this after parsing the body:
```go
// Limit request body size
r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)
defer r.Body.Close()
```

Then in the file analysis loop, replace:
```go
for _, file := range pr.Files {
    analysis := analyzer.AnalyzeFile(file)
    pr.FileAnalyses = append(pr.FileAnalyses, analysis)
}
```

With:
```go
for _, file := range pr.Files {
    // Enforce patch size limit
    if len(file.Patch) > maxPatchLength {
        slog.Warn("patch exceeds size limit",
            "file", file.Path,
            "size", len(file.Patch))
        
        pr.FileAnalyses = append(pr.FileAnalyses, FileAnalysis{
            Path:       file.Path,
            RiskLevel:  "CRITICAL",
            Summary:    "Patch exceeds safe analysis size limit",
            Confidence: 1.0,
        })
        continue
    }

    analysis := analyzer.AnalyzeFile(file)
    pr.FileAnalyses = append(pr.FileAnalyses, analysis)
}
```

---

## Fix #3: API Key Out of Docker Compose (10 minutes)

### Update: `docker-compose.yml`

Change this line:
```yaml
- TRIBUNAL_API_KEY=dev_enterprise_key_123 # Demo API key for dashboard
```

To this:
```yaml
- TRIBUNAL_API_KEY=${TRIBUNAL_API_KEY:?Error: TRIBUNAL_API_KEY not set}
```

### Create: `.env.local` (NOT committed to git)

```bash
TRIBUNAL_API_KEY=$(openssl rand -hex 32)
POSTGRES_PASSWORD=$(openssl rand -hex 16)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Update: `.gitignore`

Add these lines:
```
.env.local
.env.production.local
.env.*.local
```

---

## Fix #4: Rate Limiting for LLM (30 minutes)

### Update: `services/go-interceptor/llm_client.go`

Add this import at the top:
```go
import (
	// ... existing imports ...
	"golang.org/x/time/rate"
)
```

Update the struct:
```go
type OpenRouterClient struct {
	apiKey     string
	httpClient *http.Client
	limiter    *rate.Limiter
}
```

Update the constructor:
```go
func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(0.5), 1), // 0.5 req/sec = 30/min
	}
}
```

Update AnalyzeCode method:
```go
func (c *OpenRouterClient) AnalyzeCode(code string) (*AIDetectionResult, error) {
	// Check rate limit
	if !c.limiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded, retry after 60 seconds")
	}

	// ... rest of existing code ...
}
```

---

## Fix #5: Environment-Based Backend URL (20 minutes)

### Update: `dashboard/src/app/page.tsx`

At the top of the file, replace:
```typescript
const API_BASE = "http://localhost:8080";
```

With:
```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
```

Replace all instances of:
```typescript
`http://localhost:8080/api/v1/...`
```

With:
```typescript
`${API_BASE}/api/v1/...`
```

### Update: `docker-compose.yml` dashboard section

Update the environment:
```yaml
environment:
  - NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:8080}
```

---

## Fix #6: Remove API Key from Frontend State (15 minutes)

### Update: `dashboard/src/app/page.tsx`

Remove this line if it exists:
```typescript
const [apiKey, setApiKey] = useState("dev_enterprise_key_123");
```

Remove API key from any fetch calls. Replace:
```typescript
fetch(`${API_BASE}/api/v1/audit/summary`, {
    headers: { "Authorization": `Bearer ${apiKey}` }
})
```

With (cookies will be sent automatically):
```typescript
fetch(`${API_BASE}/api/v1/audit/summary`, {
    credentials: 'include'
})
```

---

## Verification Steps

### After applying all fixes:

```bash
# 1. Stop existing containers
docker-compose down

# 2. Generate new API key
export TRIBUNAL_API_KEY=$(openssl rand -hex 32)
echo $TRIBUNAL_API_KEY  # Save this somewhere safe!

# 3. Start fresh
docker-compose up -d

# 4. Wait 10 seconds for services to start
sleep 10

# 5. Test CORS rejection (should NOT have Access-Control-Allow-Origin header)
curl -i -H "Origin: https://evil.com" http://localhost:8080/api/v1/audit/summary

# 6. Test CORS acceptance (should have Access-Control-Allow-Origin header)
curl -i -H "Origin: http://localhost:3000" http://localhost:8080/api/v1/audit/summary

# 7. Test authentication (should return 401)
curl -i http://localhost:8080/api/v1/audit/summary

# 8. Test with correct API key (should return 200)
curl -i -H "Authorization: Bearer $TRIBUNAL_API_KEY" http://localhost:8080/api/v1/audit/summary

# 9. Check frontend loads
curl http://localhost:3000

# 10. View Go logs for rate limiting
docker logs tribunal_interceptor | grep -i "rate\|limit"
```

---

## Files Changed

- ✅ `services/go-interceptor/cors.go` (NEW)
- ✅ `services/go-interceptor/main.go` (UPDATED)
- ✅ `services/go-interceptor/handler.go` (UPDATED)
- ✅ `services/go-interceptor/llm_client.go` (UPDATED)
- ✅ `docker-compose.yml` (UPDATED)
- ✅ `dashboard/src/app/page.tsx` (UPDATED)
- ✅ `.gitignore` (UPDATED)
- ✅ `.env.local` (NEW - NOT committed)

---

## Time to Deploy

**Total Time:** ~2 hours (including testing)

**Deploy Now?** YES ✅ - These are all non-breaking changes
- Frontend fetches work the same (better configuration)
- Backend endpoints unchanged
- Docker setup improved
- CORS more restrictive (only improves security)
- Rate limiting improves stability

---

## Rollback If Needed

```bash
# If something breaks, revert with:
git checkout .
docker-compose down
rm .env.local
docker-compose up
```

---

*Ready to apply? Start with Fix #1 and work down the list.*

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func parseDurationEnv(key string) (time.Duration, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		slog.Warn("invalid duration config", "key", key, "value", value, "error", err)
		return 0, false
	}
	return parsed, true
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	customRisk := os.Getenv("ENTERPRISE_CUSTOM_RISK_KEYWORDS")
	customCritical := os.Getenv("ENTERPRISE_CUSTOM_CRITICAL_KEYWORDS")

	if customRisk != "" || customCritical != "" {
		InitCustomRules(customRisk, customCritical)
		slog.Info("Loaded custom enterprise rules engine", "customRisk", customRisk, "customCritical", customCritical)
	}

	enterpriseAPIKey := os.Getenv("TRIBUNAL_API_KEY")
	if enterpriseAPIKey != "" {
		slog.Info("Enterprise API Authorization enabled")
	} else {
		slog.Warn("TRIBUNAL_API_KEY not set! Sensitive endpoints like /analyze will be unprotected (development mode).")
	}

	redisURL := os.Getenv("REDIS_URL")

	dbURL := os.Getenv("DATABASE_URL")
	var repo Repository
	var oauthStore GitHubOAuthStore
	if dbURL != "" {
		pgRepo, err := NewPostgresRepository(context.Background(), dbURL)
		if err != nil {
			slog.Warn("failed to connect to Postgres. Falling back to in-memory storage.", "error", err)
			repo = NewInMemoryRepository()
			oauthStore = NewInMemoryGitHubOAuthStore()
		} else {
			slog.Info("connected to Postgres database")
			repo = pgRepo
			oauthStore = NewPostgresGitHubOAuthStore(pgRepo)
			defer pgRepo.Close()
		}
	} else {
		slog.Info("DATABASE_URL not set, running interceptor with in-memory persistence (data will reset on restart)")
		repo = NewInMemoryRepository()
		oauthStore = NewInMemoryGitHubOAuthStore()
	}

	var redisChecker RedisHealthChecker
	var redisMetrics RedisMetricsProvider
	if strings.TrimSpace(redisURL) != "" {
		sessionTTL := defaultGitHubSessionTTL
		if parsed, ok := parseDurationEnv("REDIS_GITHUB_SESSION_TTL"); ok {
			sessionTTL = parsed
		}

		oauthStateTTL := time.Duration(0)
		if parsed, ok := parseDurationEnv("REDIS_OAUTH_STATE_TTL"); ok {
			oauthStateTTL = parsed
		}

		redisStore, err := NewRedisGitHubOAuthStore(redisURL, RedisGitHubOAuthOptions{
			SessionTTL:    sessionTTL,
			OAuthStateTTL: oauthStateTTL,
		})
		if err != nil {
			slog.Warn("failed to connect to Redis, falling back to persistent store", "error", err)
		} else {
			slog.Info("Redis session store initialized")
			oauthStore = redisStore
			redisChecker = redisStore
			redisMetrics = redisStore
		}
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	var ghClient GitHubIntegrator
	if githubToken != "" {
		ghClient = NewGitHubClient(githubToken)
		slog.Info("GitHub client initialized")
	} else {
		slog.Info("GITHUB_TOKEN not set, running without GitHub integrations")
	}

	openrouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	var llmClient LLMIntegrator
	if openrouterAPIKey != "" {
		llmClient = NewOpenRouterClient(openrouterAPIKey)
		slog.Info("OpenRouter LLM client initialized")
	} else {
		slog.Info("OPENROUTER_API_KEY not set, using heuristic analysis only")
	}

	h := NewHandler(repo, ghClient, llmClient, oauthStore)

	// Get allowed origins from environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	slog.Info("CORS origins configured", "origins", allowedOrigins)

	// Configure webhook security
	githubWebhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	gitlabWebhookSecret := os.Getenv("GITLAB_WEBHOOK_SECRET")
	bitbucketWebhookSecret := os.Getenv("BITBUCKET_WEBHOOK_SECRET")

	if githubWebhookSecret != "" {
		slog.Info("GitHub webhook signature verification enabled")
	} else {
		slog.Warn("GITHUB_WEBHOOK_SECRET not set, webhook signature verification disabled (development mode)")
	}

	mux := http.NewServeMux()

	// Public Health Check
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/health/detailed", HealthCheckHandler(repo, redisChecker))
	mux.HandleFunc("/metrics", MetricsHandler(redisMetrics))
	mux.HandleFunc("/metrics/prometheus", PrometheusMetricsHandler(redisMetrics))

	// Webhooks with signature verification
	githubWebhookHandler := http.Handler(http.HandlerFunc(h.githubWebhookHandler))
	githubWebhookHandler = WebhookSignatureMiddleware(githubWebhookSecret, "github")(githubWebhookHandler)
	githubWebhookHandler = RateLimitWebhooks(120, time.Minute)(githubWebhookHandler)
	mux.Handle("/webhook/github", githubWebhookHandler)

	gitlabWebhookHandler := http.Handler(http.HandlerFunc(h.gitlabWebhookHandler))
	gitlabWebhookHandler = WebhookSignatureMiddleware(gitlabWebhookSecret, "gitlab")(gitlabWebhookHandler)
	gitlabWebhookHandler = RateLimitWebhooks(120, time.Minute)(gitlabWebhookHandler)
	mux.Handle("/webhook/gitlab", gitlabWebhookHandler)

	bitbucketWebhookHandler := http.Handler(http.HandlerFunc(h.bitbucketWebhookHandler))
	bitbucketWebhookHandler = WebhookSignatureMiddleware(bitbucketWebhookSecret, "bitbucket")(bitbucketWebhookHandler)
	bitbucketWebhookHandler = RateLimitWebhooks(120, time.Minute)(bitbucketWebhookHandler)
	mux.Handle("/webhook/bitbucket", bitbucketWebhookHandler)

	// Public read endpoint for specific PRs
	mux.HandleFunc("/analysis", h.getAnalysisHandler)

	// Protect internal tools and audit summaries with Enterprise API Keys if configured.
	corsWrapper := func(next http.HandlerFunc) http.HandlerFunc {
		return CORSMiddleware(allowedOrigins, next)
	}
	mux.HandleFunc("/api/v1/github/connect/callback", corsWrapper(h.githubConnectCallbackHandler))

	if enterpriseAPIKey != "" {
		mux.HandleFunc("/analyze", corsWrapper(RequireAuth(enterpriseAPIKey, h.analyzeHandler)))
		mux.HandleFunc("/api/v1/audit/summary", corsWrapper(RequireAuth(enterpriseAPIKey, h.getAuditSummaryHandler)))
		mux.HandleFunc("/api/v1/audit/logs", corsWrapper(RequireAuth(enterpriseAPIKey, h.getAuditLogsHandler)))
		mux.HandleFunc("/api/v1/audit/export", corsWrapper(RequireAuth(enterpriseAPIKey, ExportHandler(repo))))
		mux.HandleFunc("/api/v1/github/connect/start", corsWrapper(RequireAuth(enterpriseAPIKey, h.githubConnectStartHandler)))
		mux.HandleFunc("/api/v1/github/connect/status", corsWrapper(RequireAuth(enterpriseAPIKey, h.githubConnectionStatusHandler)))
		mux.HandleFunc("/api/v1/github/connect/disconnect", corsWrapper(RequireAuth(enterpriseAPIKey, h.githubDisconnectHandler)))
		mux.HandleFunc("/api/v1/policies", corsWrapper(RequireAuth(enterpriseAPIKey, PoliciesHandler(repo))))
		mux.HandleFunc("/api/v1/api-keys", corsWrapper(RequireAuth(enterpriseAPIKey, h.listAPIKeysHandler)))
		mux.HandleFunc("/api/v1/api-keys/rotate", corsWrapper(RequireAuth(enterpriseAPIKey, h.rotateAPIKeyHandler)))
	} else {
		mux.HandleFunc("/analyze", corsWrapper(h.analyzeHandler))
		mux.HandleFunc("/api/v1/audit/summary", corsWrapper(h.getAuditSummaryHandler))
		mux.HandleFunc("/api/v1/audit/logs", corsWrapper(h.getAuditLogsHandler))
		mux.HandleFunc("/api/v1/audit/export", corsWrapper(ExportHandler(repo)))
		mux.HandleFunc("/api/v1/github/connect/start", corsWrapper(h.githubConnectStartHandler))
		mux.HandleFunc("/api/v1/github/connect/status", corsWrapper(h.githubConnectionStatusHandler))
		mux.HandleFunc("/api/v1/github/connect/disconnect", corsWrapper(h.githubDisconnectHandler))
		mux.HandleFunc("/api/v1/policies", corsWrapper(PoliciesHandler(repo)))
		mux.HandleFunc("/api/v1/api-keys", corsWrapper(h.listAPIKeysHandler))
		mux.HandleFunc("/api/v1/api-keys/rotate", corsWrapper(h.rotateAPIKeyHandler))
	}

	addr := ":" + port
	srv := &http.Server{
		Addr:         addr,
		Handler:      SecurityHeadersMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	idleConnsClosed := make(chan struct{})
	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer stop()

		<-ctx.Done()
		slog.Info("shutting down gracefully, pressing Ctrl+C again will force exit")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("go-interceptor starting", "addr", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server failed critically", "error", err)
	}

	<-idleConnsClosed
	slog.Info("go-interceptor gracefully stopped")
}

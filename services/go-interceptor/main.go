package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
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

	dbURL := os.Getenv("DATABASE_URL")
	var repo Repository
	if dbURL != "" {
		pgRepo, err := NewPostgresRepository(context.Background(), dbURL)
		if err != nil {
			slog.Warn("failed to connect to Postgres (running without persistence)", "error", err)
		} else {
			slog.Info("connected to Postgres database")
			repo = pgRepo
			defer pgRepo.Close()
		}
	} else {
		slog.Info("DATABASE_URL not set, running interceptor without database persistence")
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	var ghClient GitHubIntegrator
	if githubToken != "" {
		ghClient = NewGitHubClient(githubToken)
		slog.Info("GitHub client initialized")
	} else {
		slog.Info("GITHUB_TOKEN not set, running without GitHub integrations")
	}

	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	var llmClient LLMIntegrator
	if anthropicAPIKey != "" {
		llmClient = NewClaudeClient(anthropicAPIKey)
		slog.Info("Anthropic LLM client initialized")
	} else {
		slog.Info("ANTHROPIC_API_KEY not set, using heuristic analysis only")
	}

	h := NewHandler(repo, ghClient, llmClient)

	mux := http.NewServeMux()

	// Public Health Check
	mux.HandleFunc("/health", h.healthHandler)

	// Webhooks (Authorized internally mostly by GitHub signature, though not implemented yet; let's keep it open for now)
	mux.HandleFunc("/webhook/github", h.githubWebhookHandler)
	mux.HandleFunc("/webhook/gitlab", h.gitlabWebhookHandler)
	mux.HandleFunc("/webhook/bitbucket", h.bitbucketWebhookHandler)

	// Public read endpoint for specific PRs
	mux.HandleFunc("/analysis", h.getAnalysisHandler)

	// Protect internal tools and audit summaries with Enterprise API Keys if configured.
	if enterpriseAPIKey != "" {
		mux.HandleFunc("/analyze", RequireAuth(enterpriseAPIKey, h.analyzeHandler))
		mux.HandleFunc("/api/v1/audit/summary", RequireAuth(enterpriseAPIKey, h.getAuditSummaryHandler))
	} else {
		mux.HandleFunc("/analyze", h.analyzeHandler)
		mux.HandleFunc("/api/v1/audit/summary", h.getAuditSummaryHandler)
	}

	addr := ":" + port
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
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

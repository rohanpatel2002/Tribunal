package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
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
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/analyze", h.analyzeHandler)
	mux.HandleFunc("/webhook/github", h.githubWebhookHandler)
	mux.HandleFunc("/analysis", h.getAnalysisHandler)
	mux.HandleFunc("/api/v1/audit/summary", h.getAuditSummaryHandler)

	addr := ":" + port
	slog.Info("go-interceptor starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

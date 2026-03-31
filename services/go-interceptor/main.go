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

	h := NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/analyze", h.analyzeHandler)
	mux.HandleFunc("/webhook/github", h.githubWebhookHandler)
	mux.HandleFunc("/analysis", h.getAnalysisHandler)

	addr := ":" + port
	slog.Info("go-interceptor starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

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
			log.Printf("Warning: failed to connect to Postgres: %v (running without persistence)", err)
		} else {
			log.Printf("Connected to Postgres database.")
			repo = pgRepo
			defer pgRepo.Close()
		}
	} else {
		log.Printf("DATABASE_URL not set. Running interceptor without database persistence.")
	}

	h := NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/analyze", h.analyzeHandler)
	mux.HandleFunc("/webhook/github", h.githubWebhookHandler)
	mux.HandleFunc("/analysis", h.getAnalysisHandler)

	addr := ":" + port
	log.Printf("go-interceptor listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

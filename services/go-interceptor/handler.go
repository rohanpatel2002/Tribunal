package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GitHubWebhookPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	TribunalFiles []ChangedFile `json:"tribunal_files"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "go-interceptor",
	})
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	if err := validateAnalyzeRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp := BuildResponse(req)
	writeJSON(w, http.StatusOK, resp)
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var payload GitHubWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook payload"})
		return
	}

	if payload.PullRequest.Number == 0 || strings.TrimSpace(payload.Repository.FullName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields: repository.full_name or pull_request.number",
		})
		return
	}

	if len(payload.TribunalFiles) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "webhook payload missing tribunal_files; use /analyze for direct analysis or include file patches",
		})
		return
	}

	req := AnalyzeRequest{
		Repository: payload.Repository.FullName,
		PRNumber:   payload.PullRequest.Number,
		Files:      payload.TribunalFiles,
	}

	resp := BuildResponse(req)
	writeJSON(w, http.StatusOK, resp)
}

func validateAnalyzeRequest(req AnalyzeRequest) error {
	if strings.TrimSpace(req.Repository) == "" {
		return fmt.Errorf("repository is required")
	}
	if req.PRNumber <= 0 {
		return fmt.Errorf("prNumber must be > 0")
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("files must not be empty")
	}
	for i, f := range req.Files {
		if strings.TrimSpace(f.Path) == "" {
			return fmt.Errorf("files[%d].path is required", i)
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

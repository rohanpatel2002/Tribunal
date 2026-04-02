package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type GitHubWebhookPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest struct {
		Number int `json:"number"`
		Head   struct {
			Sha string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
	TribunalFiles []ChangedFile `json:"tribunal_files"`
}

// Handler holds the application's external dependencies (like the database).
type Handler struct {
	repo         Repository
	githubClient GitHubIntegrator
}

// NewHandler creates a new HTTP handler with the given repository and GitHub client.
func NewHandler(repo Repository, gh GitHubIntegrator) *Handler {
	return &Handler{repo: repo, githubClient: gh}
}

func (h *Handler) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "go-interceptor",
	})
}

func (h *Handler) getAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	repoName := r.URL.Query().Get("repo")
	prStr := r.URL.Query().Get("pr")

	if repoName == "" || prStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'repo' or 'pr' query parameters"})
		return
	}

	prNum, err := strconv.Atoi(prStr)
	if err != nil || prNum <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid 'pr' number"})
		return
	}

	if h.repo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database persistence is not configured"})
		return
	}

	analysis, err := h.repo.GetAnalysisByPR(r.Context(), repoName, prNum)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || err == ErrAnalysisNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "analysis not found"})
			return
		}
		slog.Error("database error fetching analysis", "repo", repoName, "pr", prNum, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

func (h *Handler) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req AnalyzeRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := validateAnalyzeRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp := BuildResponse(req)

	// Persist the analysis if a repository is configured
	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save analysis to DB", "error", err)
			// We do not fail the request if the DB save fails, just log it.
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	deliveryID := r.Header.Get("X-GitHub-Delivery")
	if deliveryID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-GitHub-Delivery header"})
		return
	}

	var payload GitHubWebhookPayload
	if err := decodeJSONBody(r.Body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
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

	for i, f := range payload.TribunalFiles {
		if strings.TrimSpace(f.Path) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("tribunal_files[%d].path is required", i),
			})
			return
		}
	}

	if h.repo != nil {
		processed, err := h.repo.MarkWebhookProcessed(r.Context(), deliveryID, payload.Repository.FullName)
		if err != nil {
			slog.Warn("failed to verify webhook idempotency", "error", err)
		} else if !processed {
			slog.Info("webhook already processed, ignoring", "deliveryID", deliveryID)
			writeJSON(w, http.StatusOK, map[string]string{"message": "webhook already processed"})
			return
		}
	}

	req := AnalyzeRequest{
		Repository: payload.Repository.FullName,
		PRNumber:   payload.PullRequest.Number,
		Files:      payload.TribunalFiles,
	}

	resp := BuildResponse(req)

	// Persist to database
	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save webhook analysis to DB", "error", err)
		}
	}

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
	const maxFilesPerRequest = 300
	if len(req.Files) > maxFilesPerRequest {
		return fmt.Errorf("files must not exceed %d entries", maxFilesPerRequest)
	}

	allowedStatus := map[string]struct{}{
		"added":    {},
		"modified": {},
		"removed":  {},
		"deleted":  {},
	}

	for i, f := range req.Files {
		if strings.TrimSpace(f.Path) == "" {
			return fmt.Errorf("files[%d].path is required", i)
		}

		status := strings.ToLower(strings.TrimSpace(f.Status))
		if status == "" {
			return fmt.Errorf("files[%d].status is required", i)
		}
		if _, ok := allowedStatus[status]; !ok {
			return fmt.Errorf("files[%d].status must be one of added, modified, removed, deleted", i)
		}
	}
	return nil
}

func decodeJSONBody(body io.Reader, target any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON payload")
	}

	if decoder.More() {
		return fmt.Errorf("invalid JSON payload: multiple JSON values are not allowed")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

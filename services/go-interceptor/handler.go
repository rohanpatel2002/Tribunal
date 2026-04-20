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

const (
	maxPayloadSize = 1000000 // 1MB max request payload
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

type GitLabWebhookPayload struct {
	ObjectKind string `json:"object_kind"`
	Project    struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
	ObjectAttributes struct {
		Iid        int    `json:"iid"`
		Action     string `json:"action"`
		LastCommit struct {
			Id string `json:"id"`
		} `json:"last_commit"`
	} `json:"object_attributes"`
	TribunalFiles []ChangedFile `json:"tribunal_files"`
}

type BitbucketWebhookPayload struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest struct {
		Id int `json:"id"`
	} `json:"pullrequest"`
	TribunalFiles []ChangedFile `json:"tribunal_files"`
}

// Handler holds the application's external dependencies (like the database).
type Handler struct {
	repo         Repository
	githubClient GitHubIntegrator
	llmClient    LLMIntegrator
}

// NewHandler creates a new HTTP handler with the given repository and GitHub client.
func NewHandler(repo Repository, gh GitHubIntegrator, llm LLMIntegrator) *Handler {
	return &Handler{repo: repo, githubClient: gh, llmClient: llm}
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

func (h *Handler) getAuditSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	repository := r.URL.Query().Get("repository")
	if repository == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing repository query parameter"})
		return
	}

	if h.repo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	summary, err := h.repo.GetRepositoryAuditSummary(r.Context(), repository)
	if err != nil {
		slog.Error("failed to get audit summary", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) getAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	repository := r.URL.Query().Get("repository")
	if repository == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing repository query parameter"})
		return
	}

	// Pagination parameters
	limit := 50
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Filtering by severity
	severityFilter := r.URL.Query().Get("severity")
	recommendationFilter := r.URL.Query().Get("recommendation")

	// Sorting (default: createdAt DESC)
	sortBy := r.URL.Query().Get("sortBy")       // "date", "aiScore", "prNumber"
	sortOrder := r.URL.Query().Get("sortOrder") // "asc", "desc"
	if sortOrder != "asc" {
		sortOrder = "desc" // default
	}

	if h.repo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	records, err := h.repo.GetRecentAnalyses(r.Context(), limit, repository)
	if err != nil {
		slog.Error("failed to fetch recent analyses audit logs", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if records == nil {
		records = []PRAnalysisRecord{}
	}

	// Apply filtering
	filtered := records
	if severityFilter != "" {
		filtered = filterBySeverity(filtered, severityFilter)
	}
	if recommendationFilter != "" {
		filtered = filterByRecommendation(filtered, recommendationFilter)
	}

	// Apply sorting
	sortRecords(filtered, sortBy, sortOrder)

	// Apply pagination on filtered results
	total := len(filtered)
	if offset >= total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	paginated := filtered[offset:end]

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":        paginated,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
		"hasMore":     end < total,
		"pageCount":   (total + limit - 1) / limit,
		"currentPage": (offset / limit) + 1,
	})
}

func (h *Handler) githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxPayloadSize))
	defer r.Body.Close()

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

	// Tier validation
	tier := "FREE"
	if h.repo != nil {
		fetchedTier, err := h.repo.GetSubscriptionTier(r.Context(), payload.Repository.FullName)
		if err == nil {
			tier = fetchedTier
		} else {
			slog.Warn("failed to fetch subscription tier, defaulting to FREE", "error", err)
		}
	}
	slog.Info("operating under SaaS context", "repo", payload.Repository.FullName, "tier", tier)

	// 1. Initialize Check Run if we have a GitHub Client + a Commit SHA
	var checkRunID int64
	repoContext := ""

	// SaaS Logic: Only allow God-Mode context fetch for non-FREE tiers
	if h.githubClient != nil && payload.PullRequest.Head.Sha != "" {
		crID, err := h.githubClient.CreateCheckRun(r.Context(), payload.Repository.FullName, payload.PullRequest.Head.Sha, "TRIBUNAL AI God-Mode")
		if err != nil {
			slog.Warn("failed to create check run", "error", err)
		} else {
			checkRunID = crID
		}

		if tier != "FREE" {
			contextStr, fetchErr := h.githubClient.FetchRepositoryContext(r.Context(), payload.Repository.FullName, payload.PullRequest.Head.Sha)
			if fetchErr != nil {
				slog.Warn("failed to fetch god-mode context", "error", fetchErr)
			} else {
				repoContext = contextStr
			}
		} else {
			slog.Info("FREE tier detected; skipping repository context ingestion limits")
		}
	}

	// Prevent Anthropic LLM API calls completely on the FREE tier
	activeLLM := h.llmClient
	if tier == "FREE" {
		activeLLM = nil
	}

	// 3. Analyze the patches with LLM (or fallback heuristic if FREE/nil), passing along the context
	resp := BuildResponse(r.Context(), req, activeLLM, repoContext)

	// 4. Persist to database
	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save webhook analysis to DB", "error", err)
		}
	}

	// Update Check Run with Final Markdown
	if h.githubClient != nil && checkRunID > 0 {
		markdownBody := GenerateContextBriefing(&resp)

		conclusion := "success"
		if resp.Recommendation == "BLOCK" {
			conclusion = "failure"
		} else if resp.Recommendation == "REVIEW_REQUIRED" {
			conclusion = "neutral"
		}

		title := fmt.Sprintf("Analysis Complete: %s", resp.Recommendation)
		shortSummary := fmt.Sprintf("Analyzed %d files: %d Critical, %d High Risk.", resp.TotalFiles, resp.Critical, resp.High)

		updateOpts := UpdateCheckRunOptions{
			Status:     "completed",
			Conclusion: conclusion,
			Output: CheckRunOutput{
				Title:   title,
				Summary: shortSummary,
				Text:    markdownBody,
			},
		}

		if err := h.githubClient.UpdateCheckRun(r.Context(), payload.Repository.FullName, checkRunID, updateOpts); err != nil {
			slog.Warn("failed to update github check run", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) gitlabWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxPayloadSize))
	defer r.Body.Close()

	// GitLab sends X-Gitlab-Event
	event := r.Header.Get("X-Gitlab-Event")
	if event == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Gitlab-Event header"})
		return
	}

	var payload GitLabWebhookPayload
	if err := decodeJSONBody(r.Body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Basic validation for Merge Request logic
	if payload.ObjectAttributes.Iid == 0 || strings.TrimSpace(payload.Project.PathWithNamespace) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields: project.path_with_namespace or object_attributes.iid",
		})
		return
	}

	if len(payload.TribunalFiles) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "webhook payload missing tribunal_files",
		})
		return
	}

	req := AnalyzeRequest{
		Repository: payload.Project.PathWithNamespace,
		PRNumber:   payload.ObjectAttributes.Iid,
		Files:      payload.TribunalFiles,
	}

	// Normally we would have a GitLab client to fetch context like with GitHub.
	// For MVPs, we assume heuristic or LLM with no God-Mode Context.
	repoContext := ""

	// 1. Analyze the patches with LLM
	resp := BuildResponse(r.Context(), req, h.llmClient, repoContext)

	// 2. Persist to database
	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save gitlab webhook analysis to DB", "error", err)
		}
	}

	// 3. A complete implementation would write back comments to GitLab MR API here.

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) bitbucketWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxPayloadSize))
	defer r.Body.Close()

	// Bitbucket sends X-Event-Key
	event := r.Header.Get("X-Event-Key")
	if event == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Event-Key header"})
		return
	}

	var payload BitbucketWebhookPayload
	if err := decodeJSONBody(r.Body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if payload.PullRequest.Id == 0 || strings.TrimSpace(payload.Repository.FullName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields: repository.full_name or pullrequest.id",
		})
		return
	}

	if len(payload.TribunalFiles) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "webhook payload missing tribunal_files",
		})
		return
	}

	req := AnalyzeRequest{
		Repository: payload.Repository.FullName,
		PRNumber:   payload.PullRequest.Id,
		Files:      payload.TribunalFiles,
	}

	repoContext := ""
	resp := BuildResponse(r.Context(), req, h.llmClient, repoContext)

	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save bitbucket webhook analysis to DB", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxPayloadSize))
	defer r.Body.Close()

	var req AnalyzeRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := validateAnalyzeRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// For manual test runs without webhook context
	repoContext := ""
	resp := BuildResponse(r.Context(), req, h.llmClient, repoContext)

	// Enforce security policies against the analysis result
	if h.repo != nil {
		enforcer := NewPolicyEnforcer(h.repo)
		if err := enforcer.EnforcePolicy(r.Context(), &resp, req.Repository); err != nil {
			slog.Warn("policy enforcement failed", "error", err, "repo", req.Repository)
			// Continue - enforcement is advisory, not blocking
		}
	}

	// Persist the analysis if a repository is configured
	if h.repo != nil {
		if err := h.repo.SaveAnalysis(r.Context(), &resp); err != nil {
			slog.Warn("failed to save analysis to DB", "error", err)
			// We do not fail the request if the DB save fails, just log it.
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

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("invalid JSON payload: multiple JSON values are not allowed")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) listAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	repository := r.URL.Query().Get("repository")
	if repository == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing repository query parameter"})
		return
	}

	if h.repo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	keys, err := ListActiveAPIKeys(h.repo, repository)
	if err != nil {
		slog.Error("failed to list API keys", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve keys"})
		return
	}

	if keys == nil {
		keys = []*APIKeyMetadata{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"keys": keys})
}

func (h *Handler) rotateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxPayloadSize))
	defer r.Body.Close()

	var req APIKeyRotationRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.CurrentKeyID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "currentKeyId is required"})
		return
	}

	if h.repo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	response, err := RotateAPIKey(h.repo, req.CurrentKeyID, req.Name)
	if err != nil {
		slog.Error("failed to rotate API key", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// filterBySeverity filters audit records by risk severity
func filterBySeverity(records []PRAnalysisRecord, severity string) []PRAnalysisRecord {
	var filtered []PRAnalysisRecord
	for _, record := range records {
		switch strings.ToLower(severity) {
		case "critical":
			if record.Critical > 0 {
				filtered = append(filtered, record)
			}
		case "high":
			if record.High > 0 {
				filtered = append(filtered, record)
			}
		case "medium":
			if record.Medium > 0 {
				filtered = append(filtered, record)
			}
		case "low":
			if record.Low > 0 {
				filtered = append(filtered, record)
			}
		}
	}
	return filtered
}

// filterByRecommendation filters audit records by recommendation status
func filterByRecommendation(records []PRAnalysisRecord, recommendation string) []PRAnalysisRecord {
	var filtered []PRAnalysisRecord
	rec := strings.ToUpper(recommendation)
	for _, record := range records {
		if record.Recommendation == rec {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// sortRecords sorts audit records by specified field
func sortRecords(records []PRAnalysisRecord, sortBy string, sortOrder string) {
	isDesc := sortOrder == "desc"

	switch strings.ToLower(sortBy) {
	case "prnumber":
		if isDesc {
			for i := 0; i < len(records)-1; i++ {
				for j := i + 1; j < len(records); j++ {
					if records[j].PRNumber > records[i].PRNumber {
						records[i], records[j] = records[j], records[i]
					}
				}
			}
		} else {
			for i := 0; i < len(records)-1; i++ {
				for j := i + 1; j < len(records); j++ {
					if records[j].PRNumber < records[i].PRNumber {
						records[i], records[j] = records[j], records[i]
					}
				}
			}
		}
	case "aiscore":
		// Note: Would need to fetch file details to get AI score, skip for now
	default:
		// Default sort is by PR number descending (most recent first)
	}
}

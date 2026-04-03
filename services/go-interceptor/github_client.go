package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GitHubIntegrator defines the contract for interacting with the GitHub API.
// Using an interface allows us to mock GitHub during testing.
type GitHubIntegrator interface {
	// CreateCheckRun initializes a new Check Run on a specific commit SHA.
	CreateCheckRun(ctx context.Context, repository string, headSHA string, name string) (int64, error)

	// UpdateCheckRun updates an existing Check Run with analysis results and conclusion.
	UpdateCheckRun(ctx context.Context, repository string, checkRunID int64, opts UpdateCheckRunOptions) error

	// FetchRepositoryContext retrieves architectural context files (e.g., README.md) to build God-Mode context.
	FetchRepositoryContext(ctx context.Context, repository string, headSHA string) (string, error)
}

// CheckRunOutput represents the rich markdown output displayed in the GitHub UI.
type CheckRunOutput struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Text    string `json:"text,omitempty"`
}

// UpdateCheckRunOptions encapsulates the parameters needed to conclude a check run.
type UpdateCheckRunOptions struct {
	Status     string         `json:"status"`               // "in_progress", "completed"
	Conclusion string         `json:"conclusion,omitempty"` // "success", "failure", "neutral", "action_required"
	Output     CheckRunOutput `json:"output"`
}

// DefaultGitHubClient is the concrete implementation using net/http.
type DefaultGitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewGitHubClient creates a new configured GitHub CI adapter.
func NewGitHubClient(token string) *DefaultGitHubClient {
	return &DefaultGitHubClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.github.com",
		token:      token,
	}
}

// CreateCheckRun hits the POST /repos/{owner}/{repo}/check-runs endpoint.
func (c *DefaultGitHubClient) CreateCheckRun(ctx context.Context, repository string, headSHA string, name string) (int64, error) {
	if c.token == "" {
		return 0, fmt.Errorf("github token is not configured")
	}

	url := fmt.Sprintf("%s/repos/%s/check-runs", c.baseURL, repository)

	payload := map[string]string{
		"name":     name,
		"head_sha": headSHA,
		"status":   "in_progress",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal create check run payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

// UpdateCheckRun hits the PATCH /repos/{owner}/{repo}/check-runs/{check_run_id} endpoint.
func (c *DefaultGitHubClient) UpdateCheckRun(ctx context.Context, repository string, checkRunID int64, opts UpdateCheckRunOptions) error {
	if c.token == "" {
		return fmt.Errorf("github token is not configured")
	}

	url := fmt.Sprintf("%s/repos/%s/check-runs/%d", c.baseURL, repository, checkRunID)

	bodyBytes, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to marshal update check run payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	return nil
}

// FetchRepositoryContext attempts to pull high-value context files like README.md from the repository
// at the given commit SHA to feed into our LLM Context Engine.
func (c *DefaultGitHubClient) FetchRepositoryContext(ctx context.Context, repository string, headSHA string) (string, error) {
	if c.token == "" {
		return "", fmt.Errorf("github token is not configured")
	}

	// Endpoint for fetching file content
	url := fmt.Sprintf("%s/repos/%s/contents/README.md?ref=%s", c.baseURL, repository, headSHA)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create context request: %w", err)
	}

	// Requesting raw text instead of base64 JSON
	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github context request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If README does not exist, return an empty context gracefully
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("github context api returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read context body: %w", err)
	}

	return buf.String(), nil
}

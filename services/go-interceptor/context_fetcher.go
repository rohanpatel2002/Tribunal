package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

// RepositoryContext holds enriched context about a repository for briefing generation.
type RepositoryContext struct {
	ReadmeContent         string            `json:"readme_content"`
	Incidents             []IncidentPattern `json:"incidents"`
	RunbookSummary        string            `json:"runbook_summary"`
	DeploymentHistory     []DeploymentEvent `json:"deployment_history"`
	ArchitectureNotes     string            `json:"architecture_notes"`
	SecurityPolicies      string            `json:"security_policies"`
	RelatedDocumentation  []string          `json:"related_documentation"`
	RecentFailures        []string          `json:"recent_failures"`
	ServiceDependencies   []ServiceDep      `json:"service_dependencies"`
	ContextBriefing       string            `json:"context_briefing"`
	ContextRelevanceScore float64           `json:"context_relevance_score"`
}

// IncidentPattern represents a historical incident or failure pattern.
type IncidentPattern struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	AffectedCode []string `json:"affected_code"`
	RootCause    string   `json:"root_cause"`
	DateOccurred string   `json:"date_occurred"`
	Resolution   string   `json:"resolution"`
	Severity     string   `json:"severity"`
}

// DeploymentEvent tracks deployment history and incidents.
type DeploymentEvent struct {
	Version        string `json:"version"`
	Date           string `json:"date"`
	Status         string `json:"status"`
	Changes        int    `json:"changes"`
	Incidents      int    `json:"incidents"`
	RollbackReason string `json:"rollback_reason,omitempty"`
}

// ServiceDep describes service dependencies.
type ServiceDep struct {
	ServiceName string `json:"service_name"`
	Dependency  string `json:"dependency"`
	Type        string `json:"type"` // db, cache, queue, api
	Critical    bool   `json:"critical"`
}

// ContextFetcher retrieves repository context from GitHub and synthesizes briefings.
type ContextFetcher struct {
	httpClient  *http.Client
	githubToken string
	baseURL     string
}

// NewContextFetcher creates a new context fetcher.
func NewContextFetcher(githubToken string) *ContextFetcher {
	return &ContextFetcher{
		httpClient:  &http.Client{},
		githubToken: githubToken,
		baseURL:     "https://api.github.com",
	}
}

// FetchRepositoryContext retrieves README, releases, and incident patterns from GitHub.
func (cf *ContextFetcher) FetchRepositoryContext(ctx context.Context, repoFullName string) (*RepositoryContext, error) {
	owner, repo, err := parseRepoName(repoFullName)
	if err != nil {
		return nil, err
	}

	context := &RepositoryContext{
		Incidents:            []IncidentPattern{},
		DeploymentHistory:    []DeploymentEvent{},
		RelatedDocumentation: []string{},
		RecentFailures:       []string{},
		ServiceDependencies:  []ServiceDep{},
	}

	// Fetch README
	readmeContent, err := cf.fetchFileContent(ctx, owner, repo, "README.md")
	if err == nil {
		context.ReadmeContent = readmeContent
	} else {
		slog.Warn("failed to fetch README", "repo", repoFullName, "error", err)
	}

	// Fetch releases and parse for incidents
	releases, err := cf.fetchReleases(ctx, owner, repo)
	if err == nil {
		context.DeploymentHistory = parseDeploymentHistory(releases)
		context.Incidents = extractIncidentPatterns(releases)
	} else {
		slog.Warn("failed to fetch releases", "repo", repoFullName, "error", err)
	}

	// Fetch security policy
	securityPolicy, err := cf.fetchFileContent(ctx, owner, repo, ".github/SECURITY.md")
	if err == nil {
		context.SecurityPolicies = securityPolicy
	}

	// Fetch runbook or RUNBOOK.md
	runbook, err := cf.fetchFileContent(ctx, owner, repo, "docs/RUNBOOK.md")
	if err != nil {
		runbook, err = cf.fetchFileContent(ctx, owner, repo, "RUNBOOK.md")
	}
	if err == nil {
		context.RunbookSummary = summarizeRunbook(runbook)
	}

	// Try to fetch architecture docs
	archDoc, err := cf.fetchFileContent(ctx, owner, repo, "docs/ARCHITECTURE.md")
	if err == nil {
		context.ArchitectureNotes = archDoc
	}

	// Parse dependencies from context
	context.ServiceDependencies = parseServiceDependencies(context.ReadmeContent, context.ArchitectureNotes)

	// Calculate context relevance score
	context.ContextRelevanceScore = calculateContextScore(context)

	// Generate context briefing for this repository
	context.ContextBriefing = generateContextBriefingText(context)

	return context, nil
}

// fetchFileContent retrieves a file's content from GitHub.
func (cf *ContextFetcher) fetchFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", cf.baseURL, owner, repo, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	if cf.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+cf.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	resp, err := cf.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("file not found")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// fetchReleases retrieves GitHub releases.
func (cf *ContextFetcher) fetchReleases(ctx context.Context, owner, repo string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=10&sort=created&direction=desc", cf.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if cf.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+cf.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := cf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var releases []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// parseRepoName extracts owner and repo from "owner/repo" format.
func parseRepoName(fullName string) (owner, repo string, err error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository name format: %s", fullName)
	}
	return parts[0], parts[1], nil
}

// parseDeploymentHistory extracts deployment events from GitHub releases.
func parseDeploymentHistory(releases []map[string]interface{}) []DeploymentEvent {
	var events []DeploymentEvent

	for _, release := range releases {
		tagName, ok := release["tag_name"].(string)
		if !ok {
			continue
		}

		createdAt, _ := release["created_at"].(string)
		body, _ := release["body"].(string)

		// Parse body for incident count (look for keywords)
		incidentCount := countKeywords(body, []string{"incident", "hotfix", "rollback", "fix", "critical"})
		changeCount := extractChangeCount(body)

		event := DeploymentEvent{
			Version:   tagName,
			Date:      createdAt,
			Status:    "success",
			Changes:   changeCount,
			Incidents: incidentCount,
		}

		if incidentCount > 0 || strings.Contains(strings.ToLower(body), "rollback") {
			event.Status = "partial_rollback"
		}

		events = append(events, event)
	}

	return events
}

// extractIncidentPatterns parses release notes for incident descriptions.
func extractIncidentPatterns(releases []map[string]interface{}) []IncidentPattern {
	var patterns []IncidentPattern

	for _, release := range releases {
		tagName, _ := release["tag_name"].(string)
		body, _ := release["body"].(string)
		createdAt, _ := release["created_at"].(string)

		// Look for incident patterns in body
		if strings.Contains(strings.ToLower(body), "incident") || strings.Contains(strings.ToLower(body), "hotfix") {
			lines := strings.Split(body, "\n")
			var description strings.Builder

			for _, line := range lines {
				if len(description.String()) < 200 && strings.TrimSpace(line) != "" {
					description.WriteString(line + "\n")
				}
			}

			pattern := IncidentPattern{
				Title:        "Incident in " + tagName,
				Description:  description.String(),
				DateOccurred: createdAt,
				Severity:     "MEDIUM",
				RootCause:    "See release notes for details",
			}

			// Detect severity
			if strings.Contains(strings.ToLower(body), "critical") || strings.Contains(strings.ToLower(body), "severe") {
				pattern.Severity = "CRITICAL"
			} else if strings.Contains(strings.ToLower(body), "rollback") {
				pattern.Severity = "HIGH"
			}

			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// summarizeRunbook creates a brief summary of runbook content.
func summarizeRunbook(runbook string) string {
	lines := strings.Split(runbook, "\n")
	var summary strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			summary.WriteString(trimmed + " ")
			if summary.Len() > 300 {
				break
			}
		}
	}

	result := summary.String()
	if len(result) > 300 {
		result = result[:300] + "..."
	}
	return result
}

// parseServiceDependencies extracts service dependencies from documentation.
func parseServiceDependencies(readme, archNotes string) []ServiceDep {
	var deps []ServiceDep

	// Simple pattern matching for common service names
	servicePatterns := map[string]bool{
		"postgres":      true,
		"redis":         true,
		"mongodb":       true,
		"mysql":         true,
		"kafka":         true,
		"rabbitmq":      true,
		"elasticsearch": true,
		"dynamodb":      true,
		"s3":            true,
		"datadog":       true,
	}

	combined := readme + " " + archNotes

	for service := range servicePatterns {
		if strings.Contains(strings.ToLower(combined), service) {
			depType := "db"
			if service == "redis" || service == "elasticsearch" {
				depType = "cache"
			} else if service == "kafka" || service == "rabbitmq" {
				depType = "queue"
			}

			deps = append(deps, ServiceDep{
				ServiceName: "application",
				Dependency:  service,
				Type:        depType,
				Critical:    true,
			})
		}
	}

	return deps
}

// calculateContextScore determines how relevant the context is (0-1).
func calculateContextScore(ctx *RepositoryContext) float64 {
	score := 0.0

	if ctx.ReadmeContent != "" {
		score += 0.2
	}
	if len(ctx.Incidents) > 0 {
		score += 0.25
	}
	if ctx.RunbookSummary != "" {
		score += 0.15
	}
	if len(ctx.DeploymentHistory) > 0 {
		score += 0.2
	}
	if ctx.ArchitectureNotes != "" {
		score += 0.2
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// generateContextBriefingText creates a human-readable briefing about the repository context.
func generateContextBriefingText(ctx *RepositoryContext) string {
	var buf strings.Builder

	if len(ctx.ServiceDependencies) > 0 {
		buf.WriteString("**Service Dependencies**: ")
		deps := []string{}
		for _, dep := range ctx.ServiceDependencies {
			deps = append(deps, dep.Dependency)
		}
		buf.WriteString(strings.Join(deps, ", "))
		buf.WriteString("\n")
	}

	if len(ctx.Incidents) > 0 {
		buf.WriteString("**Recent Incidents**: ")
		buf.WriteString(fmt.Sprintf("%d known incidents in release history.\n", len(ctx.Incidents)))

		for i, incident := range ctx.Incidents {
			if i < 2 { // Show top 2
				buf.WriteString(fmt.Sprintf("  - %s (%s): %s\n", incident.Title, incident.Severity, incident.RootCause))
			}
		}
	}

	if ctx.ContextRelevanceScore > 0.5 {
		buf.WriteString("**⚠️ Context Available**: This repository has documented context. Changes affecting dependencies or previously-incident services require extra scrutiny.\n")
	}

	return buf.String()
}

// countKeywords counts occurrences of keywords in text.
func countKeywords(text string, keywords []string) int {
	count := 0
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		count += strings.Count(lower, strings.ToLower(kw))
	}
	return count
}

// extractChangeCount attempts to extract the number of changes from release body.
func extractChangeCount(body string) int {
	// Simple regex to find patterns like "10 files changed"
	re := regexp.MustCompile(`(\d+)\s+files?\s+changed`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		var count int
		fmt.Sscanf(matches[1], "%d", &count)
		return count
	}
	return 0
}

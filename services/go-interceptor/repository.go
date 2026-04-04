package main

import (
	"context"
	"errors"
)

// Common repository errors
var (
	ErrAnalysisNotFound = errors.New("analysis not found")
)

// Repository defines the data access layer for PR analysis results.
// This interface allows us to easily mock the database during tests
// and abstract the underlying storage mechanism.
type Repository interface {
	// SaveAnalysis persists the full PR analysis summary and its associated
	// file-level results. It is recommended that implementations execute
	// this operation within a single database transaction.
	SaveAnalysis(ctx context.Context, response *AnalyzeResponse) error

	// GetAnalysisByPR retrieves a previously completed analysis by repository name
	// and PR number. Returns nil if no analysis is found.
	GetAnalysisByPR(ctx context.Context, repository string, prNumber int) (*AnalyzeResponse, error)

	// MarkWebhookProcessed attempts to record a webhook delivery ID.
	// Returns true if successfully recorded, or false if it was already processed.
	MarkWebhookProcessed(ctx context.Context, deliveryID string, repoFullName string) (bool, error)

	// GetRepositoryAuditSummary aggregates high-level historical analytics for enterprise reporting.
	GetRepositoryAuditSummary(ctx context.Context, repository string) (*AuditSummary, error)
}

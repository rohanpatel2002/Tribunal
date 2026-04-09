package main

import (
	"context"
	"fmt"
	"sync"
) // InMemoryRepository stores data in RAM. Ideal for local dev without a Postgres DB.
type InMemoryRepository struct {
	mu           sync.RWMutex
	analyses     map[string][]*AnalyzeResponse
	processedWeb map[string]bool
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		analyses:     make(map[string][]*AnalyzeResponse),
		processedWeb: make(map[string]bool),
	}
}

func (r *InMemoryRepository) SaveAnalysis(ctx context.Context, response *AnalyzeResponse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	repoName := response.Repository
	r.analyses[repoName] = append(r.analyses[repoName], response)
	return nil
}

func (r *InMemoryRepository) GetAnalysisByPR(ctx context.Context, repository string, prNumber int) (*AnalyzeResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.analyses[repository]
	for _, a := range list {
		if a.PRNumber == prNumber {
			return a, nil
		}
	}
	return nil, ErrAnalysisNotFound
}

func (r *InMemoryRepository) MarkWebhookProcessed(ctx context.Context, deliveryID string, repoFullName string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := repoFullName + ":" + deliveryID
	if r.processedWeb[key] {
		return false, nil
	}
	r.processedWeb[key] = true
	return true, nil
}

func (r *InMemoryRepository) GetRepositoryAuditSummary(ctx context.Context, repository string) (*AuditSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := &AuditSummary{
		Repository: repository,
	}

	list := r.analyses[repository]
	if len(list) == 0 {
		return summary, nil
	}

	var sumAIScore float64
	var fileCount int

	for _, a := range list {
		summary.TotalPRs++
		summary.TotalFiles += a.Summary.TotalFiles
		if a.Summary.AIGenerated > 0 {
			summary.AIGeneratedPRs++
		}
		summary.CriticalRisks += a.Summary.Critical
		summary.HighRisks += a.Summary.High

		for _, f := range a.Results {
			sumAIScore += f.AIScore
			fileCount++
		}
	}

	if fileCount > 0 {
		summary.AverageAIScore = sumAIScore / float64(fileCount)
	}

	return summary, nil
}

func (r *InMemoryRepository) GetSubscriptionTier(ctx context.Context, repoFullName string) (string, error) {
	return "FREE", nil
}

func (r *InMemoryRepository) GetRecentAnalyses(ctx context.Context, limit int, repository string) ([]PRAnalysisRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := r.analyses[repository]
	var records []PRAnalysisRecord

	// traverse backwards to get the most recent appended items first
	for i := len(list) - 1; i >= 0; i-- {
		if len(records) >= limit {
			break
		}
		a := list[i]
		records = append(records, PRAnalysisRecord{
			ID:             fmt.Sprintf("repo-%s-pr-%d", a.Repository, a.PRNumber),
			Repository:     a.Repository,
			PRNumber:       a.PRNumber,
			Recommendation: a.Summary.Recommendation,
			TotalFiles:     a.Summary.TotalFiles,
			AIGenerated:    a.Summary.AIGenerated,
			Critical:       a.Summary.Critical,
			High:           a.Summary.High,
			Medium:         a.Summary.Medium,
			Low:            a.Summary.Low,
		})
	}

	return records, nil
}

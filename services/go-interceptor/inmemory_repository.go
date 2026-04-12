package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// InMemoryRepository stores data in RAM, backed up to a JSON file.
// Ideal for local dev without a Postgres DB while surviving restarts.
type InMemoryRepository struct {
	mu           sync.RWMutex
	analyses     map[string][]*AnalyzeResponse
	policies     map[string][]*SecurityPolicy
	events       []*SecurityEvent
	processedWeb map[string]bool
	dbFile       string
}

func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		analyses:     make(map[string][]*AnalyzeResponse),
		processedWeb: make(map[string]bool),
		dbFile:       "local_database.json",
	}
	repo.loadFromFile()
	return repo
}

// DataStruct holds the state we serialize to the hard drive.
type dbState struct {
	Analyses     map[string][]*AnalyzeResponse `json:"analyses"`
	Policies     map[string][]*SecurityPolicy  `json:"policies"`
	ProcessedWeb map[string]bool               `json:"processedWeb"`
}

func (r *InMemoryRepository) loadFromFile() {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.dbFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("failed to load local DB: %v\n", err)
		}
		return
	}

	var state dbState
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("failed to parse local DB: %v\n", err)
		return
	}

	if state.Analyses != nil {
		r.analyses = state.Analyses
	}
	if state.Policies != nil {
		r.policies = state.Policies
	}
	if state.ProcessedWeb != nil {
		r.processedWeb = state.ProcessedWeb
	}
}

func (r *InMemoryRepository) saveToFile() {
	state := dbState{
		Analyses:     r.analyses,
		Policies:     r.policies,
		ProcessedWeb: r.processedWeb,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("failed to serialize state: %v\n", err)
		return
	}

	if err := os.WriteFile(r.dbFile, data, 0644); err != nil {
		fmt.Printf("failed to write local DB: %v\n", err)
	}
}

func (r *InMemoryRepository) SaveAnalysis(ctx context.Context, response *AnalyzeResponse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	repoName := response.Repository
	r.analyses[repoName] = append(r.analyses[repoName], response)
	r.saveToFile()
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
	r.saveToFile()
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
		summary.TotalFiles += a.TotalFiles
		if a.AIGenerated > 0 {
			summary.AIGeneratedPRs++
		}
		summary.CriticalRisks += a.Critical
		summary.HighRisks += a.High

		for _, f := range a.Files {
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
			Recommendation: a.Recommendation,
			TotalFiles:     a.TotalFiles,
			AIGenerated:    a.AIGenerated,
			Critical:       a.Critical,
			High:           a.High,
			Medium:         a.Medium,
			Low:            a.Low,
		})
	}

	return records, nil
}

// SaveSecurityPolicy saves a policy to in-memory storage.
func (r *InMemoryRepository) SaveSecurityPolicy(ctx context.Context, policy *SecurityPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.policies == nil {
		r.policies = make(map[string][]*SecurityPolicy)
	}

	// Append or update the policy
	r.policies[policy.Repository] = append(r.policies[policy.Repository], policy)
	r.saveToFile()
	return nil
}

// GetSecurityPolicies retrieves active policies for a repository.
func (r *InMemoryRepository) GetSecurityPolicies(ctx context.Context, repository string) ([]SecurityPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var policies []SecurityPolicy
	if list, exists := r.policies[repository]; exists {
		for _, p := range list {
			if p.IsActive {
				policies = append(policies, *p)
			}
		}
	}
	return policies, nil
}

// DeleteSecurityPolicy deactivates a policy.
func (r *InMemoryRepository) DeleteSecurityPolicy(ctx context.Context, repository string, policyName string, actor string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if list, exists := r.policies[repository]; exists {
		for _, p := range list {
			if p.PolicyName == policyName {
				p.IsActive = false
				break
			}
		}
	}
	r.saveToFile()
	return nil
}

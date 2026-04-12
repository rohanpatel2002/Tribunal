package main

import (
	"context"
	"log/slog"
	"strings"
)

// PolicyEnforcer applies security policies to analysis results
type PolicyEnforcer struct {
	repo Repository
}

// NewPolicyEnforcer creates a new policy enforcer
func NewPolicyEnforcer(repo Repository) *PolicyEnforcer {
	return &PolicyEnforcer{repo: repo}
}

// EnforcePolicy applies a security policy to an analysis result
func (pe *PolicyEnforcer) EnforcePolicy(ctx context.Context, analysis *AnalyzeResponse, repository string) error {
	if pe.repo == nil {
		return nil
	}

	// Fetch all active policies for this repository
	policies, err := pe.repo.GetSecurityPolicies(ctx, repository)
	if err != nil {
		slog.Error("failed to fetch policies for enforcement", "repo", repository, "error", err)
		return err
	}

	if len(policies) == 0 {
		return nil // No policies to enforce
	}

	slog.Debug("enforcing policies", "repo", repository, "policyCount", len(policies), "prNumber", analysis.PRNumber)

	// Apply each active policy
	for _, policy := range policies {
		if !policy.IsActive {
			continue
		}

		switch policy.PolicyType {
		case "AI_DETECTION":
			pe.enforceAIDetectionPolicy(analysis, &policy)
		case "VULNERABILITY_SCAN":
			pe.enforceVulnerabilityPolicy(analysis, &policy)
		case "CODE_STYLE":
			pe.enforceCodeStylePolicy(analysis, &policy)
		case "COMPLIANCE":
			pe.enforceCompliancePolicy(analysis, &policy)
		}
	}

	return nil
}

// enforceAIDetectionPolicy enforces AI generation threshold
func (pe *PolicyEnforcer) enforceAIDetectionPolicy(analysis *AnalyzeResponse, policy *SecurityPolicy) {
	threshold, ok := policy.Rules["threshold"].(float64)
	if !ok {
		threshold = 0.5 // Default: 50% AI threshold
	}

	avgAIScore := 0.0
	if analysis.TotalFiles > 0 {
		totalScore := 0.0
		for _, file := range analysis.Files {
			totalScore += file.AIScore
		}
		avgAIScore = totalScore / float64(analysis.TotalFiles)
	}

	slog.Debug("AI detection policy check",
		"policy", policy.PolicyName,
		"threshold", threshold,
		"avgAIScore", avgAIScore,
		"severityThreshold", policy.SeverityThreshold,
	)

	// If average AI score exceeds threshold, upgrade recommendation
	if avgAIScore > threshold {
		if analysis.Recommendation == "APPROVE" {
			analysis.Recommendation = "REVIEW_REQUIRED"
		}

		// If severity is high/critical, potentially block
		if policy.SeverityThreshold == "CRITICAL" || policy.SeverityThreshold == "HIGH" {
			analysis.Recommendation = "BLOCK"
		}
	}
}

// enforceVulnerabilityPolicy enforces vulnerability severity thresholds
func (pe *PolicyEnforcer) enforceVulnerabilityPolicy(analysis *AnalyzeResponse, policy *SecurityPolicy) {
	maxAllowedCritical := int64(0)
	maxAllowedHigh := int64(999)

	// Extract thresholds from policy rules
	if critical, ok := policy.Rules["maxCritical"].(float64); ok {
		maxAllowedCritical = int64(critical)
	}
	if high, ok := policy.Rules["maxHigh"].(float64); ok {
		maxAllowedHigh = int64(high)
	}

	slog.Debug("vulnerability policy check",
		"policy", policy.PolicyName,
		"maxCritical", maxAllowedCritical,
		"maxHigh", maxAllowedHigh,
		"foundCritical", analysis.Critical,
		"foundHigh", analysis.High,
	)

	// Block if critical violations exceed threshold
	if int64(analysis.Critical) > maxAllowedCritical {
		analysis.Recommendation = "BLOCK"
		return
	}

	// Require review if high violations exceed threshold
	if int64(analysis.High) > maxAllowedHigh && analysis.Recommendation == "APPROVE" {
		analysis.Recommendation = "REVIEW_REQUIRED"
	}
}

// enforceCodeStylePolicy enforces code style consistency
func (pe *PolicyEnforcer) enforceCodeStylePolicy(analysis *AnalyzeResponse, policy *SecurityPolicy) {
	// For now, a simple pass/fail on medium+ issues
	totalMediumOrHigher := analysis.Medium + analysis.High + analysis.Critical

	if totalMediumOrHigher > 0 && analysis.Recommendation == "APPROVE" {
		analysis.Recommendation = "REVIEW_REQUIRED"
	}

	slog.Debug("code style policy check",
		"policy", policy.PolicyName,
		"mediumOrHigher", totalMediumOrHigher,
	)
}

// enforceCompliancePolicy enforces compliance rules
func (pe *PolicyEnforcer) enforceCompliancePolicy(analysis *AnalyzeResponse, policy *SecurityPolicy) {
	// Check if any file patterns violate compliance
	if bannedPatterns, ok := policy.Rules["bannedPatterns"].([]interface{}); ok {
		for _, file := range analysis.Files {
			for _, pattern := range bannedPatterns {
				if patternStr, ok := pattern.(string); ok {
					if strings.Contains(file.Path, patternStr) {
						slog.Warn("compliance violation detected",
							"policy", policy.PolicyName,
							"file", file.Path,
							"pattern", patternStr,
						)
						analysis.Recommendation = "BLOCK"
						return
					}
				}
			}
		}
	}

	slog.Debug("compliance policy check",
		"policy", policy.PolicyName,
		"filesChecked", len(analysis.Files),
	)
}

// GetPolicySummary returns a summary of which policies affected the recommendation
func (pe *PolicyEnforcer) GetPolicySummary(ctx context.Context, analysis *AnalyzeResponse, repository string) map[string]interface{} {
	policies, err := pe.repo.GetSecurityPolicies(ctx, repository)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	applied := []string{}
	for _, p := range policies {
		if p.IsActive {
			applied = append(applied, p.PolicyName)
		}
	}

	return map[string]interface{}{
		"recommendedAction": analysis.Recommendation,
		"policiesApplied":   applied,
		"totalRisks":        analysis.Critical + analysis.High + analysis.Medium + analysis.Low,
		"criticalFindings":  analysis.Critical,
		"requiresReview":    analysis.Recommendation == "REVIEW_REQUIRED",
		"blockedByPolicy":   analysis.Recommendation == "BLOCK",
	}
}

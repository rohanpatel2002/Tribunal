package main

import (
	"strings"
	"testing"
)

func TestGenerateContextBriefing(t *testing.T) {
	resp := &AnalyzeResponse{
		Repository:     "rohanpatel2002/tribunal",
		PRNumber:       12,
		TotalFiles:     2,
		AIGenerated:    1,
		Critical:       1,
		High:           0,
		Medium:         0,
		Low:            1,
		Recommendation: "BLOCK",
		Files: []FileAnalysis{
			{
				Path:          "core.go",
				AIScore:       0.99,
				IsAIGenerated: true,
				Confidence:    0.95,
				RiskLevel:     "CRITICAL",
				Summary:       "Missing idempotency key on distributed lock.",
				Signals: SignalBreakdown{
					Style:   0.8,
					Pattern: 0.9,
					Risk:    1.0,
				},
			},
		},
	}

	markdown := GenerateContextBriefing(resp)

	if !strings.Contains(markdown, "# Context Briefing: PR #12") {
		t.Errorf("Expected PR title in briefing")
	}
	if !strings.Contains(markdown, "🚨 **OVERALL STATUS: BLOCK**") {
		t.Errorf("Expected BLOCK status with emoji")
	}
	if !strings.Contains(markdown, "core.go") {
		t.Errorf("Expected file path in actionable findings")
	}
	if !strings.Contains(markdown, "Missing idempotency key") {
		t.Errorf("Expected summary in file breakdown")
	}
	if !strings.Contains(markdown, "🛑 **BLOCK MERGE**") {
		t.Errorf("Expected block recommendation block")
	}
}

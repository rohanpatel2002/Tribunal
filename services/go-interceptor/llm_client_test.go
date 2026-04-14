package main

import (
	"testing"
)

func TestStripCodeFence(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "json fenced block",
			in:   "```json\n{\"riskLevel\":\"HIGH\"}\n```",
			want: "{\"riskLevel\":\"HIGH\"}",
		},
		{
			name: "plain fenced block",
			in:   "```\n{\"riskLevel\":\"LOW\"}\n```",
			want: "{\"riskLevel\":\"LOW\"}",
		},
		{
			name: "already clean json",
			in:   "{\"riskLevel\":\"MEDIUM\"}",
			want: "{\"riskLevel\":\"MEDIUM\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeFence(tt.in)
			if got != tt.want {
				t.Fatalf("stripCodeFence() mismatch\nwant: %q\n got: %q", tt.want, got)
			}
		})
	}
}

func TestClamp01(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{in: -0.25, want: 0},
		{in: 0, want: 0},
		{in: 0.42, want: 0.42},
		{in: 1, want: 1},
		{in: 1.9, want: 1},
	}

	for _, tt := range tests {
		got := clamp01(tt.in)
		if got != tt.want {
			t.Fatalf("clamp01(%v): want %v, got %v", tt.in, tt.want, got)
		}
	}
}

func TestNormalizeLLMAnalysisResult(t *testing.T) {
	result := &LLMAnalysisResult{
		AIScore:       1.2,
		IsAIGenerated: false,
		Confidence:    -0.4,
		RiskLevel:     " severe ",
		Summary:       "  suspicious patterns  ",
		SuggestedFix:  "\n\nfunc safe() {}\n",
	}

	normalizeLLMAnalysisResult(result)

	if result.AIScore != 1 {
		t.Fatalf("expected AIScore to be clamped to 1, got %v", result.AIScore)
	}
	if result.Confidence != 0 {
		t.Fatalf("expected Confidence to be clamped to 0, got %v", result.Confidence)
	}
	if result.RiskLevel != "MEDIUM" {
		t.Fatalf("expected invalid risk level to fallback to MEDIUM, got %q", result.RiskLevel)
	}
	if result.Summary != "suspicious patterns" {
		t.Fatalf("expected Summary to be trimmed, got %q", result.Summary)
	}
	if result.SuggestedFix != "func safe() {}" {
		t.Fatalf("expected SuggestedFix to be trimmed, got %q", result.SuggestedFix)
	}
	if !result.IsAIGenerated {
		t.Fatalf("expected IsAIGenerated to be true when AIScore >= 0.70")
	}
}

func TestParseAndNormalizeLLMResult(t *testing.T) {
	raw := "```json\n{\"aiScore\":0.95,\"isAIGenerated\":false,\"confidence\":0.8,\"riskLevel\":\"high\",\"summary\":\"  ok  \",\"suggestedFix\":\"  patch  \"}\n```"

	result, err := parseAndNormalizeLLMResult(raw)
	if err != nil {
		t.Fatalf("parseAndNormalizeLLMResult() returned error: %v", err)
	}

	if result.RiskLevel != "HIGH" {
		t.Fatalf("expected risk level to normalize to HIGH, got %q", result.RiskLevel)
	}
	if result.Summary != "ok" {
		t.Fatalf("expected trimmed summary, got %q", result.Summary)
	}
	if result.SuggestedFix != "patch" {
		t.Fatalf("expected trimmed suggestedFix, got %q", result.SuggestedFix)
	}
	if !result.IsAIGenerated {
		t.Fatalf("expected IsAIGenerated=true due to high AIScore")
	}
}

func TestParseAndNormalizeLLMResult_InvalidJSON(t *testing.T) {
	_, err := parseAndNormalizeLLMResult("not-json")
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

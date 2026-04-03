package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMIntegrator defines the interface for interacting with Large Language Models.
type LLMIntegrator interface {
	// AnalyzeCode asynchronously analyzes a single file patch.
	AnalyzeCode(ctx context.Context, filename string, patch string) (*LLMAnalysisResult, error)
}

// LLMAnalysisResult matches the JSON schema we will force the LLM to output.
type LLMAnalysisResult struct {
	AIScore       float64 `json:"aiScore"`       // 0.0 to 1.0 likelihood
	IsAIGenerated bool    `json:"isAIGenerated"` // final boolean flag
	Confidence    float64 `json:"confidence"`    // 0.0 to 1.0 confidence in assessment
	RiskLevel     string  `json:"riskLevel"`     // LOW, MEDIUM, HIGH, CRITICAL
	Summary       string  `json:"summary"`       // Short contextual briefing for the reviewer
}

// DefaultClaudeClient implements LLMIntegrator using the Anthropic Messages API.
type DefaultClaudeClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
}

// NewClaudeClient creates an LLM integrator for Anthropic.
func NewClaudeClient(apiKey string) *DefaultClaudeClient {
	return &DefaultClaudeClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.anthropic.com/v1/messages",
		apiKey:     apiKey,
		model:      "claude-3-sonnet-20240229", // Standard high-context model
	}
}

// AnalyzeCode submits the diff to Claude and parses the JSON response.
func (c *DefaultClaudeClient) AnalyzeCode(ctx context.Context, filename string, patch string) (*LLMAnalysisResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("anthropic api key not configured")
	}

	prompt := fmt.Sprintf(`Analyze the following code patch for file '%s'. 
Tell me two things: is this code likely AI-generated, and does it introduce any hidden business logic or semantic risks? 
Please respond ONLY with valid JSON strictly matching this structure:
{
  "aiScore": 0.85,
  "isAIGenerated": true,
  "confidence": 0.90,
  "riskLevel": "HIGH",
  "summary": "Explanation of risks or AI artifacts here."
}

Do not include any markdown blocks containing "json" or other text outside of the raw JSON object.

Code Patch:
%s`, filename, patch)

	payload := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal anthropic payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create anthropic request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic api returned status %d: %s", resp.StatusCode, string(b))
	}

	// Parse the wrapped anthropic response format
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response wrapper: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("empty text content returned from anthropic")
	}

	rawJSONString := anthropicResp.Content[0].Text

	var finalResult LLMAnalysisResult
	if err := json.Unmarshal([]byte(rawJSONString), &finalResult); err != nil {
		return nil, fmt.Errorf("failed to decode inner JSON structure from LLM text: %w. Raw string: %s", err, rawJSONString)
	}

	// Normalize Risk Level
	if finalResult.RiskLevel != "LOW" && finalResult.RiskLevel != "MEDIUM" && finalResult.RiskLevel != "HIGH" && finalResult.RiskLevel != "CRITICAL" {
		finalResult.RiskLevel = "MEDIUM" // Fallback fallback standard
	}

	return &finalResult, nil
}

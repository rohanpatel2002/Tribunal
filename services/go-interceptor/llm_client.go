package main

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"strings"
"time"
)

// LLMIntegrator defines the interface for interacting with Large Language Models.
type LLMIntegrator interface {
	// AnalyzeCode asynchronously analyzes a single file patch.
	AnalyzeCode(ctx context.Context, filename string, patch string, repoContext string) (*LLMAnalysisResult, error)
}

// LLMAnalysisResult matches the JSON schema we will force the LLM to output.
type LLMAnalysisResult struct {
	AIScore       float64 `json:"aiScore"`       // 0.0 to 1.0 likelihood
	IsAIGenerated bool    `json:"isAIGenerated"` // final boolean flag
	Confidence    float64 `json:"confidence"`    // 0.0 to 1.0 confidence in assessment
	RiskLevel     string  `json:"riskLevel"`     // LOW, MEDIUM, HIGH, CRITICAL
	Summary       string  `json:"summary"`       // Short contextual briefing for the reviewer
	SuggestedFix  string  `json:"suggestedFix"`  // A raw code block providing the remediated code
}

// OpenRouterClient implements LLMIntegrator using the OpenRouter API (OpenAI format).
type OpenRouterClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
}

// NewOpenRouterClient creates an LLM integrator for OpenRouter.
func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		baseURL:    "https://openrouter.ai/api/v1/chat/completions",
		model:      "meta-llama/llama-3.3-70b-instruct:free", // One of the best free models!
		apiKey:     apiKey,
	}
}

// AnalyzeCode submits the diff to OpenRouter and parses the JSON response.
func (c *OpenRouterClient) AnalyzeCode(ctx context.Context, filename string, patch string, repoContext string) (*LLMAnalysisResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("openrouter api key not configured")
	}

	contextAddendum := ""
	if repoContext != "" {
		// If we have architectural context, inject it into the prompt.
		contextAddendum = fmt.Sprintf("\nArchitectural Context (README):\n%s\n", repoContext)
	}

	prompt := fmt.Sprintf(`Analyze the following code patch for file '%s'.
Tell me two things: is this code likely AI-generated, and does it introduce any hidden business logic or semantic risks?%s
If there is a severe semantic risk or architectural violation, provide the exact valid code block to fix the developer's PR in the 'suggestedFix' field. If no fix is required, leave 'suggestedFix' empty.
Please respond ONLY with valid JSON strictly matching this structure:
{
  "aiScore": 0.85,
  "isAIGenerated": true,
  "confidence": 0.90,
  "riskLevel": "HIGH",
  "summary": "Explanation of risks or AI artifacts here.",
  "suggestedFix": "func retry() {\n  // secure idempotent implementation\n}"
}

Do not include any markdown blocks containing "json" or other text outside of the raw JSON object.

Code Patch:
%s`, filename, contextAddendum, patch)

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"}, // Force JSON response
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openrouter payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create openrouter request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "http://localhost:3000") // Required by OpenRouter
	req.Header.Set("X-Title", "Tribunal Local Dev")        // Required by OpenRouter
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter api returned status %d: %s", resp.StatusCode, string(b))
	}

	var openRouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode openrouter response wrapper: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("empty text content returned from openrouter")
	}

	rawJSONString := openRouterResp.Choices[0].Message.Content

	// Clean up markdown if the LLM hallucinated it despite our instructions
	rawJSONString = strings.TrimPrefix(rawJSONString, "```json")
rawJSONString = strings.TrimPrefix(rawJSONString, "```")
	rawJSONString = strings.TrimSuffix(strings.TrimSpace(rawJSONString), "```")

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

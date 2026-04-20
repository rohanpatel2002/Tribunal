package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalyzeHandler_RejectsUnknownFields(t *testing.T) {
	h := NewHandler(nil, &MockGitHubClient{}, nil, nil)
	body := `{"repository":"rohanpatel2002/tribunal","prNumber":12,"files":[{"path":"main.go","status":"modified","patch":"x"}],"unexpected":true}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.analyzeHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestAnalyzeHandler_RejectsOversizedPayload(t *testing.T) {
	h := NewHandler(nil, &MockGitHubClient{}, nil, nil)
	largePatch := strings.Repeat("a", maxPayloadSize+1024)
	body := `{"repository":"rohanpatel2002/tribunal","prNumber":12,"files":[{"path":"main.go","status":"modified","patch":"` + largePatch + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.analyzeHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreatePolicyHandler_RejectsUnknownFields(t *testing.T) {
	repo := &MockRepository{}
	handler := PoliciesHandler(repo)

	body := `{"policyName":"max-ai-threshold","policyType":"AI_DETECTION","rules":{"threshold":0.5},"unknown":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies?repository=rohanpatel2002/tribunal", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestCreatePolicyHandler_RejectsOversizedPayload(t *testing.T) {
	repo := &MockRepository{}
	handler := PoliciesHandler(repo)

	largeDescription := strings.Repeat("x", maxPayloadSize+1024)
	body := `{"policyName":"max-ai-threshold","policyType":"AI_DETECTION","description":"` + largeDescription + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies?repository=rohanpatel2002/tribunal", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

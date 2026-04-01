package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateCheckRun_Success(t *testing.T) {
	// Mock a successful GitHub API response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/repos/rohanpatel2002/tribunal/check-runs" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		if payload["name"] != "TRIBUNAL Code Review" || payload["head_sha"] != "dummy-sha" {
			t.Fatalf("unexpected payload: %v", payload)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": 42}`))
	}))
	defer mockServer.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = mockServer.URL // Override for testing

	id, err := client.CreateCheckRun(context.Background(), "rohanpatel2002/tribunal", "dummy-sha", "TRIBUNAL Code Review")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
}

func TestCreateCheckRun_MissingToken(t *testing.T) {
	client := NewGitHubClient("")
	_, err := client.CreateCheckRun(context.Background(), "rohanpatel2002/tribunal", "dummy-sha", "TRIBUNAL Code Review")

	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	if !strings.Contains(err.Error(), "github token is not configured") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestUpdateCheckRun_Success(t *testing.T) {
	// Mock a successful GitHub API PATCH response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/repos/rohanpatel2002/tribunal/check-runs/42" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload UpdateCheckRunOptions
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		if payload.Conclusion != "success" || payload.Status != "completed" {
			t.Fatalf("unexpected payload: %v", payload)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = mockServer.URL // Override for testing

	opts := UpdateCheckRunOptions{
		Status:     "completed",
		Conclusion: "success",
		Output: CheckRunOutput{
			Title:   "Passed",
			Summary: "Everything is fine.",
		},
	}

	err := client.UpdateCheckRun(context.Background(), "rohanpatel2002/tribunal", 42, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestUpdateCheckRun_MissingToken(t *testing.T) {
	client := NewGitHubClient("")
	opts := UpdateCheckRunOptions{Status: "completed"}
	err := client.UpdateCheckRun(context.Background(), "rohanpatel2002/tribunal", 42, opts)
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	if !strings.Contains(err.Error(), "github token is not configured") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

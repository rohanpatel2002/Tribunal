package main

import (
	"context"
	"testing"
	"time"
)

func newTestInMemoryRepository(t *testing.T) *InMemoryRepository {
	t.Helper()
	return &InMemoryRepository{
		analyses:     make(map[string][]*AnalyzeResponse),
		policies:     make(map[string][]*SecurityPolicy),
		apiKeys:      make(map[string]*APIKeyMetadata),
		processedWeb: make(map[string]bool),
		dbFile:       t.TempDir() + "/test_local_database.json",
	}
}

func TestInMemoryAPIKeyLifecycle(t *testing.T) {
	repo := newTestInMemoryRepository(t)
	ctx := context.Background()

	meta := &APIKeyMetadata{
		KeyID:       "kid_test_lifecycle",
		Repository:  "rohanpatel2002/tribunal",
		Name:        "Lifecycle Key",
		Permissions: []string{"read:audit", "write:policy"},
		ExpiresAt:   time.Now().AddDate(0, 0, 45),
	}

	if err := CreateAPIKey(repo, meta, "key_secret_for_hashing"); err != nil {
		t.Fatalf("CreateAPIKey() returned error: %v", err)
	}

	stored, err := GetAPIKeyMetadata(repo, meta.KeyID)
	if err != nil {
		t.Fatalf("GetAPIKeyMetadata() returned error: %v", err)
	}
	if stored.KeyHash == "" {
		t.Fatal("expected KeyHash to be populated")
	}
	if stored.Repository != meta.Repository {
		t.Fatalf("expected repository %q, got %q", meta.Repository, stored.Repository)
	}

	active, err := ListActiveAPIKeys(repo, meta.Repository)
	if err != nil {
		t.Fatalf("ListActiveAPIKeys() returned error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active key, got %d", len(active))
	}

	rotated, err := RotateAPIKey(repo, meta.KeyID, "Rotated Key")
	if err != nil {
		t.Fatalf("RotateAPIKey() returned error: %v", err)
	}
	if rotated.NewKeyID == "" || rotated.NewKey == "" {
		t.Fatal("expected new key id and secret from rotation response")
	}

	if err := DeprecateAPIKey(repo, meta.KeyID, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("DeprecateAPIKey() returned error: %v", err)
	}

	activeAfter, err := ListActiveAPIKeys(repo, meta.Repository)
	if err != nil {
		t.Fatalf("ListActiveAPIKeys() after deprecation returned error: %v", err)
	}
	if len(activeAfter) != 1 {
		t.Fatalf("expected only the rotated key to remain active, got %d active keys", len(activeAfter))
	}
	if activeAfter[0].KeyID != rotated.NewKeyID {
		t.Fatalf("expected active key to be rotated key %q, got %q", rotated.NewKeyID, activeAfter[0].KeyID)
	}

	_ = ctx // keep ctx for future repository API expansion without lint noise
}

func TestCreateAPIKeyRejectsDuplicateIDs(t *testing.T) {
	repo := newTestInMemoryRepository(t)
	meta := &APIKeyMetadata{KeyID: "kid_duplicate", Repository: "rohanpatel2002/tribunal", Name: "dup"}

	if err := CreateAPIKey(repo, meta, "key_one"); err != nil {
		t.Fatalf("initial CreateAPIKey failed: %v", err)
	}

	err := CreateAPIKey(repo, meta, "key_two")
	if err == nil {
		t.Fatal("expected duplicate key creation to fail")
	}
}

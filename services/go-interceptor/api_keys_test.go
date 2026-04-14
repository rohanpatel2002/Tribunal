package main

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateAPIKey_FormatAndLength(t *testing.T) {
	keyID, keySecret, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() returned error: %v", err)
	}

	if !strings.HasPrefix(keyID, "kid_") {
		t.Fatalf("expected keyID to start with 'kid_', got %q", keyID)
	}
	if len(keyID) != 20 { // kid_ + 16 hex chars
		t.Fatalf("expected keyID length 20, got %d (%q)", len(keyID), keyID)
	}

	if !strings.HasPrefix(keySecret, "key_") {
		t.Fatalf("expected keySecret to start with 'key_', got %q", keySecret)
	}
	if len(keySecret) != 68 { // key_ + 64 hex chars
		t.Fatalf("expected keySecret length 68, got %d", len(keySecret))
	}
}

func TestGenerateAPIKey_UniqueAcrossCalls(t *testing.T) {
	firstID, firstSecret, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("first GenerateAPIKey() returned error: %v", err)
	}

	secondID, secondSecret, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("second GenerateAPIKey() returned error: %v", err)
	}

	if firstID == secondID {
		t.Fatalf("expected unique key IDs, got duplicates: %q", firstID)
	}
	if firstSecret == secondSecret {
		t.Fatalf("expected unique key secrets, got duplicates")
	}
}

func TestAPIKeyMetadata_CheckKeyExpiry(t *testing.T) {
	t.Run("expired key", func(t *testing.T) {
		meta := &APIKeyMetadata{ExpiresAt: time.Now().Add(-1 * time.Hour)}
		expired, days := meta.CheckKeyExpiry()

		if !expired {
			t.Fatalf("expected expired=true")
		}
		if days != 0 {
			t.Fatalf("expected daysUntilExpiry=0 for expired key, got %d", days)
		}
	})

	t.Run("active key", func(t *testing.T) {
		meta := &APIKeyMetadata{ExpiresAt: time.Now().Add(48 * time.Hour)}
		expired, days := meta.CheckKeyExpiry()

		if expired {
			t.Fatalf("expected expired=false")
		}
		if days < 1 {
			t.Fatalf("expected positive daysUntilExpiry for active key, got %d", days)
		}
	})
}

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type apiKeyStore interface {
	GetAPIKeyMetadata(keyID string) (*APIKeyMetadata, error)
	CreateAPIKey(metadata *APIKeyMetadata, keySecret string) error
	DeprecateAPIKey(keyID string, expiresAt time.Time) error
	ListActiveAPIKeys(repository string) ([]*APIKeyMetadata, error)
}

// APIKeyMetadata stores API key information with rotation tracking
type APIKeyMetadata struct {
	KeyID           string    `json:"keyId" db:"key_id"`
	Repository      string    `json:"repository,omitempty" db:"repository"`
	KeyHash         string    `json:"keyHash" db:"key_hash"`
	Name            string    `json:"name" db:"name"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	LastUsedAt      time.Time `json:"lastUsedAt" db:"last_used_at"`
	ExpiresAt       time.Time `json:"expiresAt" db:"expires_at"`
	IsActive        bool      `json:"isActive" db:"is_active"`
	Permissions     []string  `json:"permissions" db:"permissions"`
	RotationDue     bool      `json:"rotationDue"`
	DaysUntilExpiry int       `json:"daysUntilExpiry"`
}

// APIKeyRotationRequest represents a key rotation request
type APIKeyRotationRequest struct {
	CurrentKeyID string `json:"currentKeyId"`
	Name         string `json:"name"`
}

// APIKeyRotationResponse contains new key after rotation
type APIKeyRotationResponse struct {
	OldKeyID     string          `json:"oldKeyId"`
	NewKeyID     string          `json:"newKeyId"`
	NewKey       string          `json:"newKey"`
	NewMetadata  *APIKeyMetadata `json:"newMetadata"`
	GracePeriod  time.Duration   `json:"gracePeriod"`
	DeprecatedAt time.Time       `json:"deprecatedAt"`
}

// GenerateAPIKey creates a new cryptographically secure API key
func GenerateAPIKey() (keyID string, keySecret string, err error) {
	// Generate Key ID (8 bytes = 16 hex chars)
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate key ID: %w", err)
	}
	keyID = "kid_" + hex.EncodeToString(idBytes)

	// Generate Key Secret (32 bytes = 64 hex chars)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate key secret: %w", err)
	}
	keySecret = "key_" + hex.EncodeToString(secretBytes)

	return keyID, keySecret, nil
}

// RotateAPIKey rotates an existing API key with grace period
// This would be called on the concrete implementation (PostgresRepository or InMemoryRepository)
func RotateAPIKey(repo Repository, currentKeyID, name string) (*APIKeyRotationResponse, error) {
	// Fetch current key metadata
	currentKey, err := GetAPIKeyMetadata(repo, currentKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current key: %w", err)
	}

	if !currentKey.IsActive {
		return nil, fmt.Errorf("cannot rotate inactive key")
	}

	// Generate new key
	newKeyID, newKeySecret, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	// New key expires in 90 days
	expiresAt := time.Now().AddDate(0, 0, 90)

	// Create new key metadata
	newMetadata := &APIKeyMetadata{
		KeyID:       newKeyID,
		Repository:  currentKey.Repository,
		Name:        name,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		ExpiresAt:   expiresAt,
		IsActive:    true,
		Permissions: currentKey.Permissions,
	}

	// Store new key in database (PostgreSQL or in-memory)
	if err := CreateAPIKey(repo, newMetadata, newKeySecret); err != nil {
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	// Mark old key as inactive after 14-day grace period
	gracePeriod := 14 * time.Hour * 24
	deprecatedAt := time.Now()
	oldKeyExpiresAt := time.Now().Add(gracePeriod)

	if err := DeprecateAPIKey(repo, currentKeyID, oldKeyExpiresAt); err != nil {
		return nil, fmt.Errorf("failed to deprecate old key: %w", err)
	}

	return &APIKeyRotationResponse{
		OldKeyID:     currentKeyID,
		NewKeyID:     newKeyID,
		NewKey:       newKeySecret,
		NewMetadata:  newMetadata,
		GracePeriod:  gracePeriod,
		DeprecatedAt: deprecatedAt,
	}, nil
}

// GetAPIKeyMetadata retrieves API key metadata (should be implemented by Repository)
func GetAPIKeyMetadata(repo Repository, keyID string) (*APIKeyMetadata, error) {
	if store, ok := repo.(apiKeyStore); ok {
		return store.GetAPIKeyMetadata(keyID)
	}

	// For now, return placeholder metadata
	// In production, this would query the database
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 90)
	daysUntil := int(expiresAt.Sub(now).Hours() / 24)

	return &APIKeyMetadata{
		KeyID:           keyID,
		Repository:      "global",
		Name:            "Default API Key",
		CreatedAt:       now.AddDate(0, 0, -30),
		LastUsedAt:      now.Add(-1 * time.Hour),
		ExpiresAt:       expiresAt,
		IsActive:        true,
		RotationDue:     daysUntil <= 14,
		DaysUntilExpiry: daysUntil,
	}, nil
}

// CreateAPIKey stores a new API key in the database (should be implemented by Repository)
func CreateAPIKey(repo Repository, metadata *APIKeyMetadata, keySecret string) error {
	if store, ok := repo.(apiKeyStore); ok {
		return store.CreateAPIKey(metadata, keySecret)
	}

	return fmt.Errorf("repository does not support API key creation")
}

// DeprecateAPIKey marks a key as deprecated with an expiry time (should be implemented by Repository)
func DeprecateAPIKey(repo Repository, keyID string, expiresAt time.Time) error {
	if store, ok := repo.(apiKeyStore); ok {
		return store.DeprecateAPIKey(keyID, expiresAt)
	}

	return fmt.Errorf("repository does not support API key deprecation")
}

// ListActiveAPIKeys retrieves all active API keys for a repository (should be implemented by Repository)
func ListActiveAPIKeys(repo Repository, repository string) ([]*APIKeyMetadata, error) {
	if store, ok := repo.(apiKeyStore); ok {
		return store.ListActiveAPIKeys(repository)
	}

	// Placeholder implementation for non-API-key-aware repositories.
	return []*APIKeyMetadata{}, nil
}

// CheckKeyExpiry checks if a key is expired or approaching expiry
func (meta *APIKeyMetadata) CheckKeyExpiry() (expired bool, daysUntilExpiry int) {
	now := time.Now()
	if now.After(meta.ExpiresAt) {
		return true, 0
	}

	daysUntil := int(meta.ExpiresAt.Sub(now).Hours() / 24)
	return false, daysUntil
}

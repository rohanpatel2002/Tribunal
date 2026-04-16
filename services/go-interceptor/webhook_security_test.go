package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyGitHubSignature_ValidAndInvalid(t *testing.T) {
	secret := "super-secret"
	payload := []byte(`{"action":"opened"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !VerifyGitHubSignature(secret, payload, signature) {
		t.Fatal("expected GitHub signature to be valid")
	}

	if VerifyGitHubSignature(secret, payload, "sha256=deadbeef") {
		t.Fatal("expected invalid GitHub signature to be rejected")
	}
}

func TestVerifyGitLabSignature_UsesSharedToken(t *testing.T) {
	secret := "gitlab-token"
	payload := []byte(`{"object_kind":"merge_request"}`)

	if !VerifyGitLabSignature(secret, payload, "gitlab-token") {
		t.Fatal("expected matching GitLab token to be valid")
	}

	if VerifyGitLabSignature(secret, payload, "wrong-token") {
		t.Fatal("expected non-matching GitLab token to be rejected")
	}
}

func TestVerifyBitbucketSignature_SupportsSha256Prefix(t *testing.T) {
	secret := "bitbucket-secret"
	payload := []byte(`{"event":"pr:opened"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	hexSig := hex.EncodeToString(mac.Sum(nil))

	if !VerifyBitbucketSignature(secret, payload, "sha256="+hexSig) {
		t.Fatal("expected prefixed bitbucket signature to be valid")
	}

	if !VerifyBitbucketSignature(secret, payload, hexSig) {
		t.Fatal("expected raw hex bitbucket signature to be valid")
	}

	if VerifyBitbucketSignature(secret, payload, "sha256=bad") {
		t.Fatal("expected malformed bitbucket signature to be rejected")
	}
}

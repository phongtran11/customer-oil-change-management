package auth

import (
	"encoding/hex"
	"testing"
)

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if token1 == token2 {
		t.Errorf("expected generated tokens to be unique, got identical tokens")
	}

	// 32 bytes should result in 64 hex characters
	if len(token1) != 64 {
		t.Errorf("expected token length to be 64, got %d", len(token1))
	}

	if _, err := hex.DecodeString(token1); err != nil {
		t.Errorf("expected token to be a valid hex string, got error: %v", err)
	}
}

func TestHashToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	hashed1 := HashToken(token)
	hashed2 := HashToken(token)

	if hashed1 != hashed2 {
		t.Errorf("expected identical inputs to produce identical hashes")
	}

	if hashed1 == token {
		t.Errorf("expected hashed token to be different from the raw token")
	}

	// SHA-256 is 32 bytes, which is 64 hex characters
	if len(hashed1) != 64 {
		t.Errorf("expected hashed token length to be 64, got %d", len(hashed1))
	}

	if _, err := hex.DecodeString(hashed1); err != nil {
		t.Errorf("expected hashed token to be a valid hex string, got error: %v", err)
	}
}

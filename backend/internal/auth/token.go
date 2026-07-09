package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const refreshTokenBytes = 32 // 256-bit token

// GenerateRefreshToken produces a cryptographically secure random hex string
// suitable for use as a refresh token stored in the database.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken hashes a token using SHA-256 and returns its hex encoded representation.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

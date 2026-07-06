package auth

import (
	"crypto/rand"
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

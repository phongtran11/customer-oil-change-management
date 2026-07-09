package config

import (
	"os"
	"testing"
)

func TestLoad_EnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("DB_URL", "postgres://postgres:password@localhost:5432/db")
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ACCESS_TOKEN_EXPIRY_MINUTES", "30")
	os.Setenv("REFRESH_TOKEN_EXPIRY_DAYS", "14")

	// Ensure cleanup
	defer func() {
		os.Unsetenv("DB_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("ACCESS_TOKEN_EXPIRY_MINUTES")
		os.Unsetenv("REFRESH_TOKEN_EXPIRY_DAYS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DBURL != "postgres://postgres:password@localhost:5432/db" {
		t.Errorf("Expected DBURL to be %q, got %q", "postgres://postgres:password@localhost:5432/db", cfg.DBURL)
	}
	if cfg.JWTSecret != "supersecret" {
		t.Errorf("Expected JWTSecret to be %q, got %q", "supersecret", cfg.JWTSecret)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("Expected ServerPort to be %q, got %q", "9090", cfg.ServerPort)
	}
	if cfg.AccessTokenExpiryMinutes != 30 {
		t.Errorf("Expected AccessTokenExpiryMinutes to be %d, got %d", 30, cfg.AccessTokenExpiryMinutes)
	}
	if cfg.RefreshTokenExpiryDays != 14 {
		t.Errorf("Expected RefreshTokenExpiryDays to be %d, got %d", 14, cfg.RefreshTokenExpiryDays)
	}
}

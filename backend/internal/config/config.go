package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment / .env file.
type Config struct {
	ServerPort               string        `mapstructure:"SERVER_PORT"`
	DBURL                    string        `mapstructure:"DB_URL"`
	JWTSecret                string        `mapstructure:"JWT_SECRET"`
	AppEnv                   string        `mapstructure:"APP_ENV"`
	AccessTokenExpiryMinutes int           `mapstructure:"ACCESS_TOKEN_EXPIRY_MINUTES"`
	RefreshTokenExpiryDays   int           `mapstructure:"REFRESH_TOKEN_EXPIRY_DAYS"`
	AccessTokenExpiry        time.Duration `mapstructure:"-"`
	RefreshTokenExpiry       time.Duration `mapstructure:"-"`
}

// Load reads configuration from environment variables and an optional .env file.
func Load() (*Config, error) {
	v := viper.New()

	// Read from .env file if it exists; non-fatal if absent.
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		// Not fatal — env vars may be injected directly (e.g., Docker).
		_ = err
	}

	// Allow env vars to override file values.
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind env variables explicitly so they can be parsed even if .env is missing (e.g. inside Docker)
	_ = v.BindEnv("SERVER_PORT")
	_ = v.BindEnv("DB_URL")
	_ = v.BindEnv("JWT_SECRET")
	_ = v.BindEnv("APP_ENV")
	_ = v.BindEnv("ACCESS_TOKEN_EXPIRY_MINUTES")
	_ = v.BindEnv("REFRESH_TOKEN_EXPIRY_DAYS")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	
	// Validate required fields.
	if cfg.DBURL == "" {
		return nil, fmt.Errorf("config: DB_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}

	cfg.AccessTokenExpiry = time.Duration(cfg.AccessTokenExpiryMinutes) * time.Minute
	cfg.RefreshTokenExpiry = time.Duration(cfg.RefreshTokenExpiryDays) * 24 * time.Hour

	return &cfg, nil
}

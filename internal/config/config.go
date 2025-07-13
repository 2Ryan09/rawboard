package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port        string
	Environment string

	// Database configuration
	DatabaseURL     string
	DatabaseTimeout time.Duration

	// Authentication configuration
	APIKey string

	// Sentry configuration
	SentryDSN        string
	SentrySampleRate float64

	// Leaderboard configuration
	MaxScoreEntries int
	MaxScoreValue   int64
	MaxGameIDLength int
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	config := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database defaults
		DatabaseURL:     getEnv("DATABASE_URL", "localhost:6379"),
		DatabaseTimeout: getDurationEnv("DATABASE_TIMEOUT", 5*time.Second),

		// Authentication
		APIKey: getEnv("RAWBOARD_API_KEY", ""),

		// Sentry defaults
		SentryDSN:        getEnv("SENTRY_DSN", ""),
		SentrySampleRate: getFloatEnv("SENTRY_SAMPLE_RATE", 1.0),

		// Leaderboard defaults (traditional arcade values)
		MaxScoreEntries: getIntEnv("MAX_SCORE_ENTRIES", 10),
		MaxScoreValue:   getInt64Env("MAX_SCORE_VALUE", 999999999),
		MaxGameIDLength: getIntEnv("MAX_GAME_ID_LENGTH", 50),
	}

	// Validate critical configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	if c.DatabaseTimeout <= 0 {
		return fmt.Errorf("DATABASE_TIMEOUT must be positive")
	}

	if c.SentrySampleRate < 0 || c.SentrySampleRate > 1 {
		return fmt.Errorf("SENTRY_SAMPLE_RATE must be between 0 and 1")
	}

	if c.MaxScoreEntries <= 0 || c.MaxScoreEntries > 100 {
		return fmt.Errorf("MAX_SCORE_ENTRIES must be between 1 and 100")
	}

	if c.MaxScoreValue <= 0 {
		return fmt.Errorf("MAX_SCORE_VALUE must be positive")
	}

	if c.MaxGameIDLength <= 0 || c.MaxGameIDLength > 100 {
		return fmt.Errorf("MAX_GAME_ID_LENGTH must be between 1 and 100")
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// HasAPIKey returns true if API key authentication is configured
func (c *Config) HasAPIKey() bool {
	return c.APIKey != ""
}

// HasSentry returns true if Sentry monitoring is configured
func (c *Config) HasSentry() bool {
	return c.SentryDSN != ""
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

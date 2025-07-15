package database

import (
	"os"
	"strings"
	"testing"
)

func TestNewValkeyDB_EnvironmentVariablePrecedence(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"VALKEY_URI":   os.Getenv("VALKEY_URI"),
		"REDIS_URL":    os.Getenv("REDIS_URL"),
		"DATABASE_URL": os.Getenv("DATABASE_URL"),
		"VALKEY_URL":   os.Getenv("VALKEY_URL"),
		"REDIS_HOST":   os.Getenv("REDIS_HOST"),
		"REDIS_PORT":   os.Getenv("REDIS_PORT"),
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("uses VALKEY_URI when available", func(t *testing.T) {
		// Clean environment for this subtest
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("VALKEY_URI", "redis://test-valkey:6379")
		os.Setenv("REDIS_URL", "redis://should-not-use:6379")

		// This will fail connection but we can check the error message indicates correct URI
		_, err := NewValkeyDB()
		if err == nil {
			t.Skip("Test connection succeeded unexpectedly")
		}

		// The error should reference the VALKEY_URI, not REDIS_URL
		if !containsHost(err.Error(), "test-valkey") {
			t.Errorf("Expected error to reference test-valkey host, got: %v", err)
		}
	})

	t.Run("falls back to REDIS_URL when VALKEY_URI not set", func(t *testing.T) {
		// Clean environment for this subtest
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("REDIS_URL", "redis://test-redis:6379")
		os.Setenv("DATABASE_URL", "redis://should-not-use:6379")

		_, err := NewValkeyDB()
		if err == nil {
			t.Skip("Test connection succeeded unexpectedly")
		}

		if !containsHost(err.Error(), "test-redis") {
			t.Errorf("Expected error to reference test-redis host, got: %v", err)
		}
	})

	t.Run("builds URL from REDIS_HOST and REDIS_PORT components", func(t *testing.T) {
		// Clean environment for this subtest
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		os.Setenv("REDIS_HOST", "component-host")
		os.Setenv("REDIS_PORT", "9999")

		_, err := NewValkeyDB()
		if err == nil {
			t.Skip("Test connection succeeded unexpectedly")
		}

		if !containsHost(err.Error(), "component-host") {
			t.Errorf("Expected error to reference component-host, got: %v", err)
		}
	})
}

func containsHost(errorStr, host string) bool {
	return strings.Contains(errorStr, host)
}

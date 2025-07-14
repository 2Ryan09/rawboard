package database

import (
	"context"
	"os"
	"testing"
)

func TestDatabaseOperations(t *testing.T) {
	// Skip if no database available
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping database tests - database tests disabled")
	}

	db, err := NewValkeyDB()
	if err != nil {
		t.Skip("Skipping database tests - no database available")
	}
	defer db.Close()

	// Test the connection
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Skip("Skipping database tests - database connection failed")
	}

	t.Run("can store and retrieve values", func(t *testing.T) {
		key := "player:test:score"
		value := "1500"

		// Should be able to store a value
		if err := db.Set(ctx, key, value); err != nil {
			t.Fatalf("Should be able to store value: %v", err)
		}

		// Should be able to retrieve the same value
		got, err := db.Get(ctx, key)
		if err != nil {
			t.Fatalf("Should be able to retrieve stored value: %v", err)
		}
		if got != value {
			t.Errorf("Retrieved value %q should match stored value %q", got, value)
		}
	})

	t.Run("returns error for non-existent keys", func(t *testing.T) {
		// Should return an error when key doesn't exist
		_, err := db.Get(ctx, "player:nonexistent:score")
		if err == nil {
			t.Error("Should return error when retrieving non-existent key")
		}
	})

	t.Run("can overwrite existing values", func(t *testing.T) {
		key := "player:update:score"

		// Store initial value
		if err := db.Set(ctx, key, "1000"); err != nil {
			t.Fatalf("Should be able to store initial value: %v", err)
		}

		// Update with new value
		newValue := "2000"
		if err := db.Set(ctx, key, newValue); err != nil {
			t.Fatalf("Should be able to update existing value: %v", err)
		}

		// Should get the updated value
		got, err := db.Get(ctx, key)
		if err != nil {
			t.Fatalf("Should be able to retrieve updated value: %v", err)
		}
		if got != newValue {
			t.Errorf("Should get updated value %q, got %q", newValue, got)
		}
	})
}

package database

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestDatabaseBehaviors focuses on database layer behaviors and resilience
func TestDatabaseBehaviors(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping database behavioral tests - database tests disabled")
	}

	// Setup test environment
	db, err := NewValkeyDB()
	if err != nil {
		t.Skip("Skipping database behavioral tests - no database available")
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Skip("Skipping database behavioral tests - database connection failed")
	}

	t.Run("Database Connection Resilience", func(t *testing.T) {
		// Behavior: Database should handle connection issues gracefully

		// Test multiple ping operations
		for i := 0; i < 5; i++ {
			err := db.Ping(ctx)
			if err != nil {
				t.Errorf("Ping %d failed: %v", i+1, err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	})

	t.Run("Concurrent Database Operations", func(t *testing.T) {
		// Behavior: Database should handle concurrent operations safely
		gameID := fmt.Sprintf("concurrent_%d", time.Now().UnixNano())

		var wg sync.WaitGroup
		numGoroutines := 10

		// Launch concurrent operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				key := fmt.Sprintf("test:concurrent:%s:%d", gameID, routineID)
				value := fmt.Sprintf("value_%d", routineID)

				// Set a value
				err := db.Set(ctx, key, value)
				if err != nil {
					t.Errorf("Failed to set key %s: %v", key, err)
					return
				}

				// Get the value back
				retrieved, err := db.Get(ctx, key)
				if err != nil {
					t.Errorf("Failed to get key %s: %v", key, err)
					return
				}

				if retrieved != value {
					t.Errorf("Value mismatch for key %s: expected %s, got %s",
						key, value, retrieved)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("Data Persistence and Retrieval", func(t *testing.T) {
		// Behavior: Data should persist correctly and be retrievable
		gameID := fmt.Sprintf("persistence_%d", time.Now().UnixNano())

		testData := map[string]string{
			fmt.Sprintf("test:persistence:%s:string", gameID): "test_string_value",
			fmt.Sprintf("test:persistence:%s:number", gameID): "12345",
			fmt.Sprintf("test:persistence:%s:json", gameID):   `{"test": true, "score": 1000}`,
		}

		// Store all test data
		for key, value := range testData {
			err := db.Set(ctx, key, value)
			if err != nil {
				t.Fatalf("Failed to store key %s: %v", key, err)
			}
		}

		// Retrieve and verify all test data
		for key, expectedValue := range testData {
			retrievedValue, err := db.Get(ctx, key)
			if err != nil {
				t.Errorf("Failed to retrieve key %s: %v", key, err)
				continue
			}

			if retrievedValue != expectedValue {
				t.Errorf("Data mismatch for key %s: expected %s, got %s",
					key, expectedValue, retrievedValue)
			}
		}
	})

	t.Run("Error Handling and Recovery", func(t *testing.T) {
		// Behavior: Database should handle invalid operations gracefully

		// Test getting non-existent key
		_, err := db.Get(ctx, "non:existent:key:12345")
		if err == nil {
			t.Error("Getting non-existent key should return an error")
		}

		// Database should still be functional after error
		testKey := fmt.Sprintf("test:recovery:%d", time.Now().UnixNano())
		err = db.Set(ctx, testKey, "recovery_test")
		if err != nil {
			t.Errorf("Database should be functional after handling error: %v", err)
		}
	})

	t.Run("Memory and Resource Management", func(t *testing.T) {
		// Behavior: Database should handle many operations without resource leaks
		gameID := fmt.Sprintf("memory_%d", time.Now().UnixNano())

		// Perform many operations
		numOperations := 50 // Reduced for simpler test

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("test:memory:%s:%d", gameID, i)
			value := fmt.Sprintf("value_%d", i)

			// Set
			err := db.Set(ctx, key, value)
			if err != nil {
				t.Errorf("Operation %d failed: %v", i, err)
				break
			}

			// Get
			retrieved, err := db.Get(ctx, key)
			if err != nil {
				t.Errorf("Retrieval %d failed: %v", i, err)
				break
			}

			if retrieved != value {
				t.Errorf("Value mismatch at operation %d", i)
			}
		}
	})

	t.Run("Connection Stability", func(t *testing.T) {
		// Behavior: Database connection should remain stable over time
		startTime := time.Now()
		duration := 200 * time.Millisecond
		iterations := 0

		for time.Since(startTime) < duration {
			key := fmt.Sprintf("test:stability:%d:%d", time.Now().UnixNano(), iterations)
			value := fmt.Sprintf("stability_test_%d", iterations)

			err := db.Set(ctx, key, value)
			if err != nil {
				t.Errorf("Stability test failed at iteration %d: %v", iterations, err)
				break
			}

			retrieved, err := db.Get(ctx, key)
			if err != nil {
				t.Errorf("Stability retrieval failed at iteration %d: %v", iterations, err)
				break
			}

			if retrieved != value {
				t.Errorf("Stability value mismatch at iteration %d", iterations)
			}

			iterations++
			time.Sleep(time.Millisecond)
		}

		if iterations < 10 {
			t.Errorf("Expected at least 10 stability iterations, got %d", iterations)
		}
	})
}

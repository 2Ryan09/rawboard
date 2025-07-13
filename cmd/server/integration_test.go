package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rawboard/internal/database"
	"rawboard/internal/handlers"
	"rawboard/internal/leaderboard"
	"rawboard/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyIntegration(t *testing.T) {
	// Skip if no database available
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping integration tests - database tests disabled")
	}

	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := database.NewValkeyDB()
	if err != nil {
		t.Skip("Skipping integration tests - no database available")
	}
	defer db.Close()

	// Setup services
	leaderboardService := leaderboard.NewService(db)
	apiKey := "test-integration-key-123"
	apiKeyMiddleware := middleware.APIKeyMiddleware(apiKey)

	// Setup router
	router := gin.New()
	handlers.SetupRoutes(router, leaderboardService, apiKeyMiddleware)

	t.Run("public endpoint works without API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/games/test-game/leaderboard", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 404 (no leaderboard yet) but not 401 (unauthorized)
		if w.Code == http.StatusUnauthorized {
			t.Errorf("Public endpoint should not require authentication, got status %d", w.Code)
		}
	})

	t.Run("protected endpoint requires API key", func(t *testing.T) {
		scoreData := map[string]interface{}{
			"initials": "TST",
			"score":    1000,
		}
		jsonData, _ := json.Marshal(scoreData)

		req := httptest.NewRequest("POST", "/api/v1/games/test-game/scores", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Protected endpoint should require authentication, got status %d", w.Code)
		}
	})

	t.Run("protected endpoint works with valid API key", func(t *testing.T) {
		scoreData := map[string]interface{}{
			"initials": "TST",
			"score":    1500,
		}
		jsonData, _ := json.Marshal(scoreData)

		req := httptest.NewRequest("POST", "/api/v1/games/test-integration/scores", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Protected endpoint should work with valid API key, got status %d: %s", w.Code, w.Body.String())
		}

		// Verify the score was actually saved
		req = httptest.NewRequest("GET", "/api/v1/games/test-integration/leaderboard", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Should be able to retrieve leaderboard after submitting score, got status %d", w.Code)
		}
	})
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

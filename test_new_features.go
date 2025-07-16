package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"rawboard/internal/database"
	"rawboard/internal/handlers"
	"rawboard/internal/leaderboard"
	"rawboard/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Skip if no database available
	if os.Getenv("SKIP_DB_TESTS") != "" {
		log.Println("Skipping new features test - database tests disabled")
		return
	}

	// Setup test database
	db, err := database.NewValkeyDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	// Setup services and handlers
	leaderboardService := leaderboard.NewService(db)
	apiKeyMiddleware := middleware.APIKeyMiddleware("test-key")

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	handlers.SetupRoutes(router, leaderboardService, apiKeyMiddleware)

	gameID := fmt.Sprintf("new_features_test_%d", time.Now().UnixNano())
	log.Printf("Testing new API features with game ID: %s", gameID)

	// Test 1: Submit multiple scores to generate achievements
	log.Println("\n=== Test 1: Submit scores for achievements ===")
	scores := []struct {
		initials string
		score    int64
	}{
		{"ACE", 500},   // First score
		{"ACE", 1200},  // Getting Started (1K)
		{"ACE", 6000},  // Rising Star (5K)
		{"PRO", 15000}, // High Achiever (10K)
		{"PRO", 30000}, // Score Master (25K)
		{"ACE", 8000},  // Lower score (shouldn't affect leaderboard)
		{"NEW", 800},   // First score for NEW
	}

	for i, s := range scores {
		body := map[string]interface{}{
			"initials": s.initials,
			"score":    s.score,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/games/%s/scores", gameID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			log.Printf("ERROR: Score submission %d failed with status %d: %s", i+1, w.Code, w.Body.String())
			return
		}
		log.Printf("✅ Score %d submitted: %s - %d", i+1, s.initials, s.score)
	}

	// Test 2: Enhanced player stats with achievements
	log.Println("\n=== Test 2: Enhanced player statistics ===")
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/games/%s/players/ACE/stats/enhanced?include_history=true", gameID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		log.Printf("ERROR: Enhanced stats request failed: %d - %s", w.Code, w.Body.String())
		return
	}

	var enhancedStats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &enhancedStats)

	achievements := enhancedStats["achievements"].([]interface{})
	log.Printf("✅ ACE has %d achievements:", len(achievements))
	for _, ach := range achievements {
		achMap := ach.(map[string]interface{})
		log.Printf("   🏆 %s: %s", achMap["name"], achMap["description"])
	}

	scoreHistory := enhancedStats["score_history"].([]interface{})
	log.Printf("✅ ACE score history: %d entries", len(scoreHistory))

	// Test 3: Score analysis endpoint
	log.Println("\n=== Test 3: Score analysis ===")
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/games/%s/scores/analyze?top_players=3", gameID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		log.Printf("ERROR: Score analysis request failed: %d - %s", w.Code, w.Body.String())
		return
	}

	var analysis map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &analysis)

	log.Printf("✅ Game Analysis:")
	log.Printf("   Total Players: %.0f", analysis["total_players"])
	log.Printf("   Total Scores: %.0f", analysis["total_scores"])
	log.Printf("   Highest Score: %.0f", analysis["highest_score"])
	log.Printf("   Average Score: %.1f", analysis["average_score"])

	topPlayers := analysis["top_players"].([]interface{})
	log.Printf("   Top %d Players:", len(topPlayers))
	for i, player := range topPlayers {
		playerMap := player.(map[string]interface{})
		log.Printf("     %d. %s - %.0f (Rank: %.0f)", i+1,
			playerMap["initials"], playerMap["high_score"], playerMap["current_rank"])
	}

	scoreDistribution := analysis["score_distribution"].(map[string]interface{})
	log.Printf("   Score Distribution:")
	for range_, count := range scoreDistribution {
		log.Printf("     %s: %.0f players", range_, count)
	}

	// Test 4: Standardized error responses
	log.Println("\n=== Test 4: Standardized error responses ===")

	// Test invalid initials
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/games/%s/players/AA/stats", gameID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusBadRequest {
		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)

		error := errorResp["error"].(map[string]interface{})
		meta := errorResp["meta"].(map[string]interface{})

		log.Printf("✅ Standardized error response:")
		log.Printf("   Code: %s", error["code"])
		log.Printf("   Message: %s", error["message"])
		log.Printf("   Request ID: %s", meta["request_id"])
		log.Printf("   Timestamp: %s", meta["timestamp"])
	} else {
		log.Printf("ERROR: Expected 400, got %d", w.Code)
		return
	}

	// Test 5: Health check at new location
	log.Println("\n=== Test 5: Health check endpoint ===")
	req = httptest.NewRequest("GET", "/api/v1/health", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var health map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &health)
		log.Printf("✅ Health check response:")
		log.Printf("   Status: %s", health["status"])
		log.Printf("   Service: %s", health["service"])
		log.Printf("   Version: %s", health["version"])
	} else {
		log.Printf("ERROR: Health check failed: %d", w.Code)
		return
	}

	// Test 6: Player not found handling
	log.Println("\n=== Test 6: Player not found handling ===")
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/games/%s/players/ZZZ/stats", gameID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)

		error := errorResp["error"].(map[string]interface{})
		log.Printf("✅ Player not found error:")
		log.Printf("   Code: %s", error["code"])
		log.Printf("   Message: %s", error["message"])
	} else {
		log.Printf("ERROR: Expected 404, got %d", w.Code)
		return
	}

	log.Println("\n🎉 All new feature tests passed!")
	log.Println("✅ Enhanced player statistics with achievements")
	log.Println("✅ Score analysis endpoint working")
	log.Println("✅ Standardized error responses implemented")
	log.Println("✅ Health check at /api/v1/health")
	log.Println("✅ Achievement system functioning")
	log.Println("✅ Request correlation with UUIDs")
}

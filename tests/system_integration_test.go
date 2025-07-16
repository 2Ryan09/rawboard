package tests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"rawboard/internal/database"
	"rawboard/internal/leaderboard"
)

// TestSystemIntegration tests the complete system integration behaviors
func TestSystemIntegration(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping system integration tests - database tests disabled")
	}

	// Setup test environment
	db, err := database.NewValkeyDB()
	if err != nil {
		t.Skip("Skipping system integration tests - no database available")
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		t.Skip("Skipping system integration tests - database connection failed")
	}

	service := leaderboard.NewService(db)
	ctx := context.Background()

	t.Run("End-to-End User Journey: Complete Score Lifecycle", func(t *testing.T) {
		// Behavior: Complete user journey from first score to achievements
		gameID := fmt.Sprintf("journey_%d", time.Now().UnixNano())
		initials := "JRN"

		// Step 1: Submit first score
		err := service.SubmitScore(ctx, gameID, initials, 500)
		if err != nil {
			t.Fatalf("Failed to submit first score: %v", err)
		}

		// Step 2: Verify score appears in leaderboard
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		if len(leaderboard.Entries) != 1 {
			t.Errorf("Expected 1 entry in leaderboard, got %d", len(leaderboard.Entries))
		}

		if leaderboard.Entries[0].Initials != initials {
			t.Errorf("Expected player %s in leaderboard, got %s", initials, leaderboard.Entries[0].Initials)
		}

		// Step 3: Check basic player stats
		stats, err := service.GetPlayerStats(ctx, gameID, initials)
		if err != nil {
			t.Fatalf("Failed to get player stats: %v", err)
		}

		if stats.HighScore != 500 {
			t.Errorf("Expected high score 500, got %d", stats.HighScore)
		}

		// Step 4: Check enhanced stats with achievements
		enhanced, err := service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
		if err != nil {
			t.Fatalf("Failed to get enhanced stats: %v", err)
		}

		if len(enhanced.Achievements) == 0 {
			t.Error("Expected at least one achievement after first score")
		}

		// Step 5: Submit improvement to unlock new achievement
		err = service.SubmitScore(ctx, gameID, initials, 1500)
		if err != nil {
			t.Fatalf("Failed to submit improved score: %v", err)
		}

		// Step 6: Verify improvement is reflected everywhere
		enhanced, err = service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
		if err != nil {
			t.Fatalf("Failed to get enhanced stats after improvement: %v", err)
		}

		if enhanced.HighScore != 1500 {
			t.Errorf("Expected improved high score 1500, got %d", enhanced.HighScore)
		}

		if len(enhanced.Achievements) <= 1 {
			t.Error("Expected additional achievements after score improvement")
		}

		// Step 7: Verify leaderboard shows only highest score
		leaderboard, err = service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard after improvement: %v", err)
		}

		if len(leaderboard.Entries) != 1 {
			t.Errorf("Player should appear only once in leaderboard, got %d entries", len(leaderboard.Entries))
		}

		if leaderboard.Entries[0].Score != 1500 {
			t.Errorf("Leaderboard should show highest score 1500, got %d", leaderboard.Entries[0].Score)
		}
	})

	t.Run("Multi-Player Competition: Leaderboard Dynamics", func(t *testing.T) {
		// Behavior: Multiple players should compete properly on leaderboard
		gameID := fmt.Sprintf("competition_%d", time.Now().UnixNano())

		// Create a competitive scenario
		players := []struct {
			initials string
			scores   []int64
		}{
			{"ACE", []int64{1000, 2000, 5000}}, // Will be #1
			{"PRO", []int64{3000, 4000, 3500}}, // Will be #2
			{"AVG", []int64{1500, 2500, 1800}}, // Will be #3
			{"NEW", []int64{800, 1200, 900}},   // Will be #4
			{"TRY", []int64{500, 600, 700}},    // Will be #5
		}

		// Submit all scores
		for _, player := range players {
			for _, score := range player.scores {
				err := service.SubmitScore(ctx, gameID, player.initials, score)
				if err != nil {
					t.Fatalf("Failed to submit score %d for %s: %v", score, player.initials, err)
				}
			}
		}

		// Verify final leaderboard
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get final leaderboard: %v", err)
		}

		if len(leaderboard.Entries) != len(players) {
			t.Errorf("Expected %d players on leaderboard, got %d", len(players), len(leaderboard.Entries))
		}

		// Verify ranking order (highest to lowest)
		expectedOrder := []struct {
			initials string
			score    int64
		}{
			{"ACE", 5000},
			{"PRO", 4000},
			{"AVG", 2500},
			{"NEW", 1200},
			{"TRY", 700},
		}

		for i, expected := range expectedOrder {
			if i >= len(leaderboard.Entries) {
				t.Errorf("Missing entry at rank %d", i+1)
				break
			}

			entry := leaderboard.Entries[i]
			if entry.Initials != expected.initials {
				t.Errorf("Rank %d: expected %s, got %s", i+1, expected.initials, entry.Initials)
			}

			if entry.Score != expected.score {
				t.Errorf("Rank %d (%s): expected score %d, got %d",
					i+1, expected.initials, expected.score, entry.Score)
			}
		}

		// Verify each player has correct individual stats
		for _, expected := range expectedOrder {
			stats, err := service.GetPlayerStats(ctx, gameID, expected.initials)
			if err != nil {
				t.Errorf("Failed to get stats for %s: %v", expected.initials, err)
				continue
			}

			if stats.HighScore != expected.score {
				t.Errorf("Player %s: expected high score %d, got %d",
					expected.initials, expected.score, stats.HighScore)
			}
		}
	})

	t.Run("System Resilience: Concurrent Multi-Player Operations", func(t *testing.T) {
		// Behavior: System should handle concurrent operations from multiple players
		gameID := fmt.Sprintf("resilience_%d", time.Now().UnixNano())

		var wg sync.WaitGroup
		numPlayers := 8
		scoresPerPlayer := 4

		// Launch concurrent player sessions
		for playerID := 0; playerID < numPlayers; playerID++ {
			wg.Add(1)
			go func(pID int) {
				defer wg.Done()
				initials := fmt.Sprintf("P%02d", pID)

				// Each player submits multiple scores
				for scoreNum := 0; scoreNum < scoresPerPlayer; scoreNum++ {
					score := int64((pID+1)*1000 + scoreNum*250)

					err := service.SubmitScore(ctx, gameID, initials, score)
					if err != nil {
						t.Logf("Player %s score submission failed: %v", initials, err)
					}

					// Sometimes check their own stats
					if scoreNum == scoresPerPlayer-1 {
						stats, err := service.GetPlayerStats(ctx, gameID, initials)
						if err != nil {
							t.Logf("Player %s stats retrieval failed: %v", initials, err)
						} else {
							expectedHighest := int64((pID+1)*1000 + (scoresPerPlayer-1)*250)
							if stats.HighScore != expectedHighest {
								t.Logf("Player %s: expected high score %d, got %d",
									initials, expectedHighest, stats.HighScore)
							}
						}
					}

					time.Sleep(time.Millisecond * 5) // Simulate real-world timing
				}
			}(playerID)
		}

		wg.Wait()
		time.Sleep(100 * time.Millisecond) // Allow all operations to complete

		// Verify final system state
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get final leaderboard: %v", err)
		}

		// Should have most/all players
		if len(leaderboard.Entries) < numPlayers/2 {
			t.Errorf("Expected at least %d players, got %d", numPlayers/2, len(leaderboard.Entries))
		}

		// Verify no duplicate players
		seenPlayers := make(map[string]bool)
		for _, entry := range leaderboard.Entries {
			if seenPlayers[entry.Initials] {
				t.Errorf("Player %s appears multiple times in leaderboard", entry.Initials)
			}
			seenPlayers[entry.Initials] = true
		}

		// Verify proper sorting
		for i := 0; i < len(leaderboard.Entries)-1; i++ {
			currentScore := leaderboard.Entries[i].Score
			nextScore := leaderboard.Entries[i+1].Score

			if currentScore < nextScore {
				t.Errorf("Leaderboard not sorted: rank %d (%d) < rank %d (%d)",
					i+1, currentScore, i+2, nextScore)
			}
		}
	})

	t.Run("Cross-Game Isolation: Data Boundary Verification", func(t *testing.T) {
		// Behavior: Different games should not interfere with each other
		timestamp := time.Now().UnixNano()
		gameA := fmt.Sprintf("isolation_a_%d", timestamp)
		gameB := fmt.Sprintf("isolation_b_%d", timestamp)
		gameC := fmt.Sprintf("isolation_c_%d", timestamp)

		// Set up different scenarios in each game
		gameData := map[string][]struct {
			initials string
			score    int64
		}{
			gameA: {{"AAA", 1000}, {"ABB", 1500}, {"ACC", 2000}},
			gameB: {{"BBB", 3000}, {"BCC", 3500}, {"BAA", 4000}},
			gameC: {{"CCC", 5000}, {"CAA", 5500}, {"CBB", 6000}},
		}

		// Submit all data
		for gameID, players := range gameData {
			for _, player := range players {
				err := service.SubmitScore(ctx, gameID, player.initials, player.score)
				if err != nil {
					t.Fatalf("Failed to submit score for %s in game %s: %v",
						player.initials, gameID, err)
				}
			}
		}

		// Verify each game has only its own data
		for gameID, expectedPlayers := range gameData {
			leaderboard, err := service.GetLeaderboard(ctx, gameID)
			if err != nil {
				t.Fatalf("Failed to get leaderboard for game %s: %v", gameID, err)
			}

			if len(leaderboard.Entries) != len(expectedPlayers) {
				t.Errorf("Game %s: expected %d players, got %d",
					gameID, len(expectedPlayers), len(leaderboard.Entries))
			}

			// Verify all expected players are present with correct scores
			playerScores := make(map[string]int64)
			for _, entry := range leaderboard.Entries {
				playerScores[entry.Initials] = entry.Score
			}

			for _, expected := range expectedPlayers {
				if score, exists := playerScores[expected.initials]; !exists {
					t.Errorf("Game %s: missing player %s", gameID, expected.initials)
				} else if score != expected.score {
					t.Errorf("Game %s: player %s expected score %d, got %d",
						gameID, expected.initials, expected.score, score)
				}
			}
		}
	})

	t.Run("System Analytics: Comprehensive Data Analysis", func(t *testing.T) {
		// Behavior: Analytics should provide accurate insights across the system
		gameID := fmt.Sprintf("analytics_%d", time.Now().UnixNano())

		// Create a diverse dataset
		players := []struct {
			initials string
			score    int64
		}{
			{"LOW", 500},  // Lowest
			{"AVG", 2500}, // Average
			{"MID", 3000}, // Middle-high
			{"HGH", 4500}, // High
			{"TOP", 5000}, // Highest
		}

		totalScores := int64(0)
		for _, player := range players {
			err := service.SubmitScore(ctx, gameID, player.initials, player.score)
			if err != nil {
				t.Fatalf("Failed to submit score for analytics test: %v", err)
			}
			totalScores += player.score
		}

		// Get analytics
		analysis, err := service.GetScoreAnalysis(ctx, gameID, len(players))
		if err != nil {
			t.Fatalf("Failed to get score analysis: %v", err)
		}

		// Verify analytics accuracy
		if int(analysis.TotalPlayers) != len(players) {
			t.Errorf("Expected %d total players, got %d", len(players), analysis.TotalPlayers)
		}

		if analysis.HighestScore != 5000 {
			t.Errorf("Expected highest score 5000, got %d", analysis.HighestScore)
		}

		expectedAverage := float64(totalScores) / float64(len(players))
		if analysis.AverageScore != expectedAverage {
			t.Errorf("Expected average score %f, got %f", expectedAverage, analysis.AverageScore)
		}

		// Verify leaderboard consistency with analytics
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard for analytics verification: %v", err)
		}

		if len(leaderboard.Entries) != int(analysis.TotalPlayers) {
			t.Errorf("Leaderboard entries (%d) don't match analytics total players (%d)",
				len(leaderboard.Entries), analysis.TotalPlayers)
		}

		if leaderboard.Entries[0].Score != analysis.HighestScore {
			t.Errorf("Leaderboard top score (%d) doesn't match analytics highest (%d)",
				leaderboard.Entries[0].Score, analysis.HighestScore)
		}
	})

	t.Run("Achievement System: Progressive Unlocking", func(t *testing.T) {
		// Behavior: Achievements should unlock progressively as players improve
		gameID := fmt.Sprintf("achievements_%d", time.Now().UnixNano())
		initials := "ACH"

		// Track achievement progression
		milestones := []struct {
			score                   int64
			expectedMinAchievements int
			description             string
		}{
			{100, 1, "First Score"},
			{1000, 2, "Getting Started"},
			{5000, 3, "Rising Star"},
			{10000, 4, "High Scorer"},
		}

		for _, milestone := range milestones {
			err := service.SubmitScore(ctx, gameID, initials, milestone.score)
			if err != nil {
				t.Fatalf("Failed to submit milestone score %d: %v", milestone.score, err)
			}

			enhanced, err := service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
			if err != nil {
				t.Fatalf("Failed to get enhanced stats for milestone %d: %v", milestone.score, err)
			}

			if len(enhanced.Achievements) < milestone.expectedMinAchievements {
				t.Errorf("Milestone %d (%s): expected at least %d achievements, got %d",
					milestone.score, milestone.description, milestone.expectedMinAchievements,
					len(enhanced.Achievements))
			}

			// Verify achievement quality - should have relevant milestones
			hasRelevantAchievement := false
			for _, ach := range enhanced.Achievements {
				if ach.UnlockedAt.After(time.Now().Add(-time.Minute)) {
					hasRelevantAchievement = true
					break
				}
			}

			if !hasRelevantAchievement && milestone.score > 100 {
				t.Errorf("Milestone %d should unlock a recent achievement", milestone.score)
			}
		}
	})
}

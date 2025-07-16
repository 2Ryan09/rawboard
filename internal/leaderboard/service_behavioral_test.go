package leaderboard

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"rawboard/internal/database"
)

// TestLeaderboardBehaviors focuses on key leaderboard service behaviors
func TestLeaderboardBehaviors(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping leaderboard behavioral tests - database tests disabled")
	}

	// Setup test environment
	db, err := database.NewValkeyDB()
	if err != nil {
		t.Skip("Skipping leaderboard behavioral tests - no database available")
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		t.Skip("Skipping leaderboard behavioral tests - database connection failed")
	}

	service := NewService(db)
	ctx := context.Background()

	t.Run("Player Score Progression: High Score Tracking", func(t *testing.T) {
		// Behavior: Service should track only the highest score per player
		gameID := fmt.Sprintf("progression_%d", time.Now().UnixNano())
		initials := "PLR"

		// Submit multiple scores for same player
		scores := []int64{1000, 3000, 2000, 5000, 1500} // non-monotonic
		for _, score := range scores {
			err := service.SubmitScore(ctx, gameID, initials, score)
			if err != nil {
				t.Fatalf("Failed to submit score %d: %v", score, err)
			}
		}

		// Get player stats
		stats, err := service.GetPlayerStats(ctx, gameID, initials)
		if err != nil {
			t.Fatalf("Failed to get player stats: %v", err)
		}

		// Should have highest score (5000)
		if stats.HighScore != 5000 {
			t.Errorf("Expected highest score 5000, got %d", stats.HighScore)
		}

		// Check leaderboard
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		// Player should appear only once with highest score
		plrCount := 0
		var plrScore int64
		for _, entry := range leaderboard.Entries {
			if entry.Initials == initials {
				plrCount++
				plrScore = entry.Score
			}
		}

		if plrCount != 1 {
			t.Errorf("Player should appear exactly once on leaderboard, appeared %d times", plrCount)
		}

		if plrScore != 5000 {
			t.Errorf("Expected player's highest score (5000) on leaderboard, got %d", plrScore)
		}
	})

	t.Run("Achievement Progression: Milestone Unlocking", func(t *testing.T) {
		// Behavior: Achievements should unlock as players reach milestones
		gameID := fmt.Sprintf("achievements_%d", time.Now().UnixNano())
		initials := "ACH"

		// First score - should unlock "First Score"
		err := service.SubmitScore(ctx, gameID, initials, 500)
		if err != nil {
			t.Fatalf("Failed to submit first score: %v", err)
		}

		stats, err := service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
		if err != nil {
			t.Fatalf("Failed to get enhanced stats: %v", err)
		}

		if len(stats.Achievements) != 1 {
			t.Errorf("Expected 1 achievement after first score, got %d", len(stats.Achievements))
		}

		// Reach 1K - should unlock "Getting Started"
		err = service.SubmitScore(ctx, gameID, initials, 1200)
		if err != nil {
			t.Fatalf("Failed to submit 1K score: %v", err)
		}

		stats, err = service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
		if err != nil {
			t.Fatalf("Failed to get enhanced stats: %v", err)
		}

		if len(stats.Achievements) != 2 {
			t.Errorf("Expected 2 achievements after 1K score, got %d", len(stats.Achievements))
		}

		// Reach 5K - should unlock "Rising Star"
		err = service.SubmitScore(ctx, gameID, initials, 5500)
		if err != nil {
			t.Fatalf("Failed to submit 5K score: %v", err)
		}

		stats, err = service.GetEnhancedPlayerStats(ctx, gameID, initials, false)
		if err != nil {
			t.Fatalf("Failed to get enhanced stats: %v", err)
		}

		if len(stats.Achievements) != 3 {
			t.Errorf("Expected 3 achievements after 5K score, got %d", len(stats.Achievements))
		}

		// Verify specific achievement names
		achievementNames := make([]string, len(stats.Achievements))
		for i, ach := range stats.Achievements {
			achievementNames[i] = ach.Name
		}
		sort.Strings(achievementNames)

		expected := []string{"First Score", "Getting Started", "Rising Star"}
		sort.Strings(expected)

		for i, name := range expected {
			if i >= len(achievementNames) || achievementNames[i] != name {
				t.Errorf("Missing achievement: %s. Got: %v", name, achievementNames)
				break
			}
		}
	})

	t.Run("Concurrent Score Submission: Data Consistency", func(t *testing.T) {
		// Behavior: Service should handle concurrent submissions safely
		gameID := fmt.Sprintf("concurrent_%d", time.Now().UnixNano())

		var wg sync.WaitGroup
		numGoroutines := 5
		scoresPerGoroutine := 3

		// Launch concurrent score submissions
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				initials := fmt.Sprintf("P%02d", routineID)

				for j := 0; j < scoresPerGoroutine; j++ {
					score := int64((routineID+1)*1000 + (j+1)*100)
					err := service.SubmitScore(ctx, gameID, initials, score)
					if err != nil {
						// Log but don't fail - some contention is expected
						t.Logf("Score submission failed for %s: %v", initials, err)
					}
					time.Sleep(2 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
		time.Sleep(300 * time.Millisecond) // Allow operations to complete

		// Verify final state
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		// Basic consistency checks
		if len(leaderboard.Entries) == 0 {
			t.Error("Service should have processed at least some scores")
		}

		// Verify leaderboard integrity: scores should be in descending order
		for i := 0; i < len(leaderboard.Entries)-1; i++ {
			currentScore := leaderboard.Entries[i].Score
			nextScore := leaderboard.Entries[i+1].Score

			if currentScore < nextScore {
				t.Errorf("Leaderboard not properly sorted: rank %d (%d) < rank %d (%d)",
					i+1, currentScore, i+2, nextScore)
			}
		}

		// Verify each player appears only once
		seenPlayers := make(map[string]bool)
		for _, entry := range leaderboard.Entries {
			if seenPlayers[entry.Initials] {
				t.Errorf("Player %s appears multiple times in leaderboard", entry.Initials)
			}
			seenPlayers[entry.Initials] = true
		}
	})

	t.Run("Game Isolation: Cross-Game Data Separation", func(t *testing.T) {
		// Behavior: Games should not interfere with each other's data
		timestamp := time.Now().UnixNano()
		gameA := fmt.Sprintf("isolation_a_%d", timestamp)
		gameB := fmt.Sprintf("isolation_b_%d", timestamp)

		// Submit different scores to different games
		err := service.SubmitScore(ctx, gameA, "AAA", 1000)
		if err != nil {
			t.Fatalf("Failed to submit to game A: %v", err)
		}

		err = service.SubmitScore(ctx, gameB, "BBB", 2000)
		if err != nil {
			t.Fatalf("Failed to submit to game B: %v", err)
		}

		err = service.SubmitScore(ctx, gameA, "AAA", 1500) // improvement in game A
		if err != nil {
			t.Fatalf("Failed to improve score in game A: %v", err)
		}

		// Check game A leaderboard
		leaderboardA, err := service.GetLeaderboard(ctx, gameA)
		if err != nil {
			t.Fatalf("Failed to get leaderboard A: %v", err)
		}

		if len(leaderboardA.Entries) != 1 {
			t.Errorf("Game A should have 1 player, got %d", len(leaderboardA.Entries))
		}
		if leaderboardA.Entries[0].Initials != "AAA" {
			t.Errorf("Game A should have player AAA, got %s", leaderboardA.Entries[0].Initials)
		}
		if leaderboardA.Entries[0].Score != 1500 {
			t.Errorf("Game A player should have score 1500, got %d", leaderboardA.Entries[0].Score)
		}

		// Check game B leaderboard
		leaderboardB, err := service.GetLeaderboard(ctx, gameB)
		if err != nil {
			t.Fatalf("Failed to get leaderboard B: %v", err)
		}

		if len(leaderboardB.Entries) != 1 {
			t.Errorf("Game B should have 1 player, got %d", len(leaderboardB.Entries))
		}
		if leaderboardB.Entries[0].Initials != "BBB" {
			t.Errorf("Game B should have player BBB, got %s", leaderboardB.Entries[0].Initials)
		}
		if leaderboardB.Entries[0].Score != 2000 {
			t.Errorf("Game B player should have score 2000, got %d", leaderboardB.Entries[0].Score)
		}
	})

	t.Run("Score Analysis: Statistical Accuracy", func(t *testing.T) {
		// Behavior: Analytics should provide accurate statistical insights
		gameID := fmt.Sprintf("analytics_%d", time.Now().UnixNano())

		// Submit known dataset
		scores := []struct {
			initials string
			score    int64
		}{
			{"P01", 1000}, {"P02", 2000}, {"P03", 3000},
			{"P04", 4000}, {"P05", 5000},
		}

		totalScores := int64(0)
		for _, s := range scores {
			err := service.SubmitScore(ctx, gameID, s.initials, s.score)
			if err != nil {
				t.Fatalf("Failed to submit score for %s: %v", s.initials, err)
			}
			totalScores += s.score
		}

		// Get analysis
		analysis, err := service.GetScoreAnalysis(ctx, gameID, len(scores))
		if err != nil {
			t.Fatalf("Failed to get score analysis: %v", err)
		}

		// Verify statistics
		if int(analysis.TotalPlayers) != len(scores) {
			t.Errorf("Expected %d total players, got %d", len(scores), analysis.TotalPlayers)
		}

		if int(analysis.TotalScores) != len(scores) {
			t.Errorf("Expected %d total scores, got %d", len(scores), analysis.TotalScores)
		}

		if analysis.HighestScore != 5000 {
			t.Errorf("Expected highest score 5000, got %d", analysis.HighestScore)
		}

		expectedAverage := float64(totalScores) / float64(len(scores))
		if analysis.AverageScore != expectedAverage {
			t.Errorf("Expected average score %f, got %f", expectedAverage, analysis.AverageScore)
		}
	})

	t.Run("Performance Under Load: Large Dataset Handling", func(t *testing.T) {
		// Behavior: Service should handle large datasets efficiently
		gameID := fmt.Sprintf("load_%d", time.Now().UnixNano())

		// Submit many scores
		numPlayers := 50
		start := time.Now()

		for i := 0; i < numPlayers; i++ {
			initials := fmt.Sprintf("P%02d", i%100) // Some duplicate initials
			score := int64((i + 1) * 100)           // Start from 100
			err := service.SubmitScore(ctx, gameID, initials, score)
			if err != nil {
				t.Fatalf("Failed to submit score %d: %v", score, err)
			}
		}

		submissionTime := time.Since(start)

		// Verify leaderboard response time
		start = time.Now()
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}
		retrievalTime := time.Since(start)

		// Verify constraints
		if len(leaderboard.Entries) > 10 {
			t.Errorf("Leaderboard should show max 10 entries, got %d", len(leaderboard.Entries))
		}

		// Performance should be reasonable
		if submissionTime > 10*time.Second {
			t.Errorf("Score submissions took too long: %v", submissionTime)
		}

		if retrievalTime > 200*time.Millisecond {
			t.Errorf("Leaderboard retrieval took too long: %v", retrievalTime)
		}

		// Verify top scores are properly sorted
		for i := 0; i < len(leaderboard.Entries)-1; i++ {
			currentScore := leaderboard.Entries[i].Score
			nextScore := leaderboard.Entries[i+1].Score

			if currentScore < nextScore {
				t.Errorf("Leaderboard not properly sorted: rank %d (%d) < rank %d (%d)",
					i+1, currentScore, i+2, nextScore)
			}
		}
	})
}

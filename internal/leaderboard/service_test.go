package leaderboard

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"rawboard/internal/database"
)

func TestLeaderboardService(t *testing.T) {
	// Skip if no database available
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping leaderboard tests - database tests disabled")
	}

	ctx := context.Background()
	t.Run("stores and retrieves player scores correctly", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		gameID := "test_store_retrieve_" + generateTestID()
		initials := "AAA"
		score := int64(15000)

		// When a player submits a score
		err := service.SubmitScore(ctx, gameID, initials, score)
		if err != nil {
			t.Fatalf("Failed to submit score: %v", err)
		}

		// Then the score should appear on the leaderboard
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		if len(leaderboard.Entries) != 1 {
			t.Fatalf("Expected 1 entry on leaderboard, got %d", len(leaderboard.Entries))
		}

		entry := leaderboard.Entries[0]
		if entry.Initials != initials {
			t.Errorf("Expected player initials %s on leaderboard, got %s", initials, entry.Initials)
		}
		if entry.Score != score {
			t.Errorf("Expected score %d on leaderboard, got %d", score, entry.Score)
		}
	})
	t.Run("ranks players by highest score first", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		gameID := "test_ranking_" + generateTestID()

		// When multiple players submit different scores
		players := []struct {
			initials string
			score    int64
		}{
			{"LOW", 1000}, // lowest score
			{"HIG", 5000}, // highest score
			{"MID", 3000}, // middle score
			{"SEC", 4000}, // second highest
			{"MIN", 2000}, // second lowest
		}

		for _, player := range players {
			service.SubmitScore(ctx, gameID, player.initials, player.score)
		}

		// Then the leaderboard should rank them by score (highest first)
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		expectedRanking := []string{"HIG", "SEC", "MID", "MIN", "LOW"}
		expectedScores := []int64{5000, 4000, 3000, 2000, 1000}

		for i, entry := range leaderboard.Entries {
			if entry.Initials != expectedRanking[i] {
				t.Errorf("Rank %d: expected %s but got %s", i+1, expectedRanking[i], entry.Initials)
			}
			if entry.Score != expectedScores[i] {
				t.Errorf("Rank %d: expected score %d but got %d", i+1, expectedScores[i], entry.Score)
			}
		}
	})
	t.Run("maintains only the top 10 highest scores", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		gameID := "test_top10_" + generateTestID()

		// When more than 10 players submit scores
		for i := 0; i < 15; i++ {
			initials := fmt.Sprintf("P%02d", i)
			score := int64(i * 100) // scores: 0, 100, 200, ..., 1400
			service.SubmitScore(ctx, gameID, initials, score)
		}

		// Then only the top 10 scores should remain
		leaderboard, err := service.GetLeaderboard(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to get leaderboard: %v", err)
		}

		if len(leaderboard.Entries) != 10 {
			t.Errorf("Expected leaderboard to show only top 10 scores, got %d", len(leaderboard.Entries))
		}

		// And the lowest score on the leaderboard should be higher than the excluded scores
		lowestDisplayedScore := leaderboard.Entries[9].Score // 10th place (0-indexed)
		expectedLowestScore := int64(500)                    // score for player P05

		if lowestDisplayedScore != expectedLowestScore {
			t.Errorf("Expected lowest displayed score to be %d, got %d", expectedLowestScore, lowestDisplayedScore)
		}
	})
	t.Run("rejects invalid player initials", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		gameID := "test_invalid_" + generateTestID()
		score := int64(1000)

		// When players try to submit scores with invalid initials
		invalidInitials := []string{"", "A", "AB", "ABCD", "A B", "12", "a b"}

		for _, initials := range invalidInitials {
			err := service.SubmitScore(ctx, gameID, initials, score)
			// Then the submission should be rejected
			if err == nil {
				t.Errorf("Expected rejection for invalid initials '%s', but submission was accepted", initials)
			}
		}
	})
	t.Run("accepts valid three-letter initials", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		gameID := "test_valid_" + generateTestID()
		score := int64(1000)

		// When a player submits a score with valid 3-letter initials
		validInitials := []string{"ABC", "XYZ", "AAA", "123"}

		for _, initials := range validInitials {
			err := service.SubmitScore(ctx, gameID, initials, score)
			// Then the submission should be accepted
			if err != nil {
				t.Errorf("Expected acceptance for valid initials '%s', but got error: %v", initials, err)
			}
		}
	})
	t.Run("keeps separate leaderboards for different games", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer db.Close()
		service := NewService(db)

		testID := generateTestID()
		tetrisGameID := "tetris_" + testID
		snakeGameID := "snake_" + testID

		// When players submit scores to different games
		service.SubmitScore(ctx, tetrisGameID, "TET", 1000)
		service.SubmitScore(ctx, snakeGameID, "SNK", 2000)
		service.SubmitScore(ctx, tetrisGameID, "TE2", 1500)

		// Then each game should have its own leaderboard
		tetrisBoard, err := service.GetLeaderboard(ctx, tetrisGameID)
		if err != nil {
			t.Fatalf("Failed to get tetris leaderboard: %v", err)
		}

		snakeBoard, err := service.GetLeaderboard(ctx, snakeGameID)
		if err != nil {
			t.Fatalf("Failed to get snake leaderboard: %v", err)
		}

		if len(tetrisBoard.Entries) != 2 {
			t.Errorf("Expected tetris leaderboard to have 2 entries, got %d", len(tetrisBoard.Entries))
		}

		if len(snakeBoard.Entries) != 1 {
			t.Errorf("Expected snake leaderboard to have 1 entry, got %d", len(snakeBoard.Entries))
		}

		// And the top score in each game should be correct
		if tetrisBoard.Entries[0].Score != 1500 {
			t.Errorf("Expected top tetris score to be 1500, got %d", tetrisBoard.Entries[0].Score)
		}

		if snakeBoard.Entries[0].Score != 2000 {
			t.Errorf("Expected top snake score to be 2000, got %d", snakeBoard.Entries[0].Score)
		}
	})
}

func setupTestDatabase(t *testing.T) database.DB {
	db, err := database.NewValkeyDB()
	if err != nil {
		t.Skip("Skipping test - failed to connect to database")
	}
	
	// Test the connection
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Skip("Skipping test - database connection failed")
	}
	
	return db
}

func generateTestID() string {
	return fmt.Sprintf("%d_%d", time.Now().Unix(), rand.Intn(10000))
}

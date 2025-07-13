package leaderboard

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"rawboard/internal/database"
	"rawboard/internal/models"
)

// Service handles leaderboard operations
type Service struct {
	db database.DB
}

// NewService creates a new leaderboard service
func NewService(db database.DB) *Service {
	return &Service{db: db}
}

// SubmitScore submits a new score entry (traditional arcade style)
func (s *Service) SubmitScore(ctx context.Context, gameID, initials string, score int64) error {
	// Validate initials (should be 3 characters, no spaces allowed)
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 || strings.Contains(initials, " ") {
		return fmt.Errorf("initials must be exactly 3 characters with no spaces")
	}

	// Create the score entry
	entry := models.ScoreEntry{
		Initials:  initials,
		Score:     score,
		Timestamp: time.Now(),
	}

	// Get current leaderboard
	leaderboard, err := s.GetLeaderboard(ctx, gameID)
	if err != nil {
		// If no leaderboard exists yet, create a new one
		leaderboard = &models.Leaderboard{
			GameID:  gameID,
			Entries: []models.ScoreEntry{},
		}
	}

	// Add new entry
	leaderboard.Entries = append(leaderboard.Entries, entry)

	// Sort by score (highest first)
	sort.Slice(leaderboard.Entries, func(i, j int) bool {
		return leaderboard.Entries[i].Score > leaderboard.Entries[j].Score
	})

	// Keep only top 10 scores (traditional arcade limit)
	if len(leaderboard.Entries) > 10 {
		leaderboard.Entries = leaderboard.Entries[:10]
	}

	// Save back to database
	return s.saveLeaderboard(ctx, leaderboard)
}

// GetLeaderboard returns the current leaderboard for a game
func (s *Service) GetLeaderboard(ctx context.Context, gameID string) (*models.Leaderboard, error) {
	key := fmt.Sprintf("leaderboard:%s", gameID)

	data, err := s.db.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("no leaderboard found for game")
	}

	var leaderboard models.Leaderboard
	if err := json.Unmarshal([]byte(data), &leaderboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaderboard: %w", err)
	}

	return &leaderboard, nil
}

// saveLeaderboard saves a leaderboard to the database
func (s *Service) saveLeaderboard(ctx context.Context, leaderboard *models.Leaderboard) error {
	data, err := json.Marshal(leaderboard)
	if err != nil {
		return fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	key := fmt.Sprintf("leaderboard:%s", leaderboard.GameID)
	return s.db.Set(ctx, key, string(data))
}

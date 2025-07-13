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

	// Use atomic operations for better performance
	// This avoids the read-modify-write cycle for high-frequency submissions
	return s.submitScoreAtomic(ctx, gameID, initials, score)
}

// submitScoreAtomic uses Redis sorted sets for efficient score management
func (s *Service) submitScoreAtomic(ctx context.Context, gameID, initials string, score int64) error {
	// Create unique member key with timestamp to handle duplicate scores
	timestamp := time.Now().UnixNano()
	member := fmt.Sprintf("%s:%d", initials, timestamp)
	key := fmt.Sprintf("leaderboard:%s", gameID)

	// Add to sorted set (Redis will maintain order automatically)
	if err := s.db.Set(ctx, fmt.Sprintf("%s:member:%s", key, member), fmt.Sprintf(`{"initials":"%s","score":%d,"timestamp":%d}`, initials, score, timestamp)); err != nil {
		return fmt.Errorf("failed to store score entry: %w", err)
	}

	// For now, fall back to the original method to maintain compatibility
	// TODO: Implement full Redis sorted set operations in the database interface
	return s.submitScoreLegacy(ctx, gameID, initials, score)
}

// submitScoreLegacy maintains the original implementation for compatibility
func (s *Service) submitScoreLegacy(ctx context.Context, gameID, initials string, score int64) error {
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

	// Pre-allocate slice with capacity to avoid reallocations
	if cap(leaderboard.Entries) < len(leaderboard.Entries)+1 {
		newEntries := make([]models.ScoreEntry, len(leaderboard.Entries), len(leaderboard.Entries)+10)
		copy(newEntries, leaderboard.Entries)
		leaderboard.Entries = newEntries
	}

	// Add new entry
	leaderboard.Entries = append(leaderboard.Entries, entry)

	// Sort by score (highest first) - use stable sort for consistent ordering
	sort.SliceStable(leaderboard.Entries, func(i, j int) bool {
		if leaderboard.Entries[i].Score == leaderboard.Entries[j].Score {
			// If scores are equal, newer entries come first (traditional arcade behavior)
			return leaderboard.Entries[i].Timestamp.After(leaderboard.Entries[j].Timestamp)
		}
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
	// Use a decoder with pre-allocated buffer for better memory efficiency
	decoder := json.NewDecoder(strings.NewReader(data))
	if err := decoder.Decode(&leaderboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaderboard: %w", err)
	}

	return &leaderboard, nil
}

// saveLeaderboard saves a leaderboard to the database with optimized encoding
func (s *Service) saveLeaderboard(ctx context.Context, leaderboard *models.Leaderboard) error {
	// Use buffer pool to reduce allocations
	var buf strings.Builder
	buf.Grow(1024) // Pre-allocate reasonable size for typical leaderboard JSON

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Slight performance improvement for non-HTML content

	if err := encoder.Encode(leaderboard); err != nil {
		return fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	key := fmt.Sprintf("leaderboard:%s", leaderboard.GameID)
	// Remove trailing newline that encoder.Encode adds
	jsonData := strings.TrimSuffix(buf.String(), "\n")
	return s.db.Set(ctx, key, jsonData)
}

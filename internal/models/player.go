package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ScoreEntry represents a simple arcade-style score entry
type ScoreEntry struct {
	Initials  string    `json:"initials" example:"AAA"`                       // Three letter initials (e.g., "AAA")
	Score     int64     `json:"score" example:"12500"`                        // Player's score
	Timestamp time.Time `json:"timestamp" example:"2025-07-13T15:30:00.000Z"` // When this score was achieved
}

// Validate ensures the ScoreEntry meets arcade standards
func (se *ScoreEntry) Validate() error {
	// Normalize initials
	se.Initials = strings.ToUpper(strings.TrimSpace(se.Initials))

	if len(se.Initials) != 3 {
		return fmt.Errorf("initials must be exactly 3 characters, got %d", len(se.Initials))
	}

	if strings.Contains(se.Initials, " ") {
		return fmt.Errorf("initials cannot contain spaces")
	}

	if se.Score < 0 {
		return fmt.Errorf("score cannot be negative")
	}

	if se.Score > 999999999 { // Traditional arcade max
		return fmt.Errorf("score too high - maximum allowed is 999,999,999")
	}

	if se.Timestamp.IsZero() {
		se.Timestamp = time.Now()
	}

	return nil
}

// Leaderboard represents a simple arcade leaderboard
type Leaderboard struct {
	GameID  string       `json:"game_id" example:"pacman"` // Unique identifier for the game
	Entries []ScoreEntry `json:"entries"`                  // Top scores (max 10, sorted by score desc)
}

// Validate ensures the Leaderboard meets arcade standards
func (lb *Leaderboard) Validate() error {
	if strings.TrimSpace(lb.GameID) == "" {
		return fmt.Errorf("game_id cannot be empty")
	}

	if len(lb.GameID) > 50 {
		return fmt.Errorf("game_id too long - maximum 50 characters")
	}

	if len(lb.Entries) > 10 {
		return fmt.Errorf("leaderboard cannot have more than 10 entries")
	}

	// Validate each entry
	for i, entry := range lb.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry %d invalid: %w", i, err)
		}
	}

	return nil
}

// MarshalJSON provides custom JSON marshaling with validation
func (lb *Leaderboard) MarshalJSON() ([]byte, error) {
	if err := lb.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Use anonymous struct to avoid infinite recursion
	type leaderboardAlias Leaderboard
	return json.Marshal((*leaderboardAlias)(lb))
}

// PlayerStats represents comprehensive statistics for a player (initials)
type PlayerStats struct {
	Initials     string    `json:"initials" example:"AAA"`                      // Three letter initials
	HighScore    int64     `json:"high_score" example:"15000"`                  // Player's highest score
	TotalScores  int       `json:"total_scores" example:"5"`                    // Number of scores submitted
	LastPlayed   time.Time `json:"last_played" example:"2025-07-16T15:30:00Z"`  // Last time this player submitted a score
	AverageScore float64   `json:"average_score" example:"12000.5"`             // Average of all scores
	FirstPlayed  time.Time `json:"first_played" example:"2025-07-15T10:15:00Z"` // First time this player submitted a score
}

// AllScoresRecord represents the complete score history for a game
type AllScoresRecord struct {
	GameID  string       `json:"game_id" example:"pacman"`
	Scores  []ScoreEntry `json:"scores"`  // All scores ever submitted
	Updated time.Time    `json:"updated"` // Last update timestamp
}

// PlayerHighScores represents a mapping of initials to their highest scores
type PlayerHighScores struct {
	GameID      string                 `json:"game_id" example:"pacman"`
	HighScores  map[string]ScoreEntry  `json:"high_scores"`  // initials -> highest score
	Updated     time.Time              `json:"updated"`      // Last update timestamp
}

// Achievement represents a player achievement
type Achievement struct {
	ID          string    `json:"id" example:"first_score"`
	Name        string    `json:"name" example:"First Score"`
	Description string    `json:"description" example:"Submit your first score"`
	UnlockedAt  time.Time `json:"unlocked_at" example:"2025-07-16T15:30:00Z"`
	Icon        string    `json:"icon,omitempty" example:"🎯"`
}

// EnhancedPlayerStats represents comprehensive statistics with achievements
type EnhancedPlayerStats struct {
	Initials      string        `json:"initials" example:"AAA"`
	HighScore     int64         `json:"high_score" example:"15000"`
	TotalScores   int           `json:"total_scores" example:"5"`
	LastPlayed    time.Time     `json:"last_played" example:"2025-07-16T15:30:00Z"`
	AverageScore  float64       `json:"average_score" example:"12000.5"`
	FirstPlayed   time.Time     `json:"first_played" example:"2025-07-15T10:15:00Z"`
	CurrentRank   *int          `json:"current_rank,omitempty" example:"3"`
	Achievements  []Achievement `json:"achievements"`
	ScoreHistory  []ScoreEntry  `json:"score_history,omitempty"` // Optional, only if requested
}

// ScoreAnalysisResponse represents bulk analysis for a game
type ScoreAnalysisResponse struct {
	GameID           string                         `json:"game_id" example:"pacman"`
	TotalPlayers     int                           `json:"total_players" example:"25"`
	TotalScores      int                           `json:"total_scores" example:"150"`
	HighestScore     int64                         `json:"highest_score" example:"50000"`
	AverageScore     float64                       `json:"average_score" example:"12500.5"`
	LastActivity     time.Time                     `json:"last_activity" example:"2025-07-16T15:30:00Z"`
	TopPlayers       []EnhancedPlayerStats         `json:"top_players"`
	ScoreDistribution map[string]int               `json:"score_distribution"` // e.g., "0-1000": 5, "1000-5000": 10
	RecentAchievements []Achievement                `json:"recent_achievements"`
	Updated          time.Time                     `json:"updated"`
}

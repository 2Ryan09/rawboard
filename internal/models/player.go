package models

import (
	"time"
)

// ScoreEntry represents a simple arcade-style score entry
type ScoreEntry struct {
	Initials  string    `json:"initials"`  // Three letter initials (e.g., "AAA")
	Score     int64     `json:"score"`     // Player's score
	Timestamp time.Time `json:"timestamp"` // When this score was achieved
}

// Leaderboard represents a simple arcade leaderboard
type Leaderboard struct {
	GameID  string       `json:"game_id"`
	Entries []ScoreEntry `json:"entries"`
}

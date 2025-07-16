package handlers

import (
	"rawboard/internal/models"
)

// ScoreSubmissionRequest represents a request to submit a new score
// This is the only input-specific type we need, as it doesn't include
// system-generated fields like timestamp
type ScoreSubmissionRequest struct {
	Initials string `json:"initials" binding:"required" example:"AAA" minLength:"3" maxLength:"3"`
	Score    int64  `json:"score" binding:"required,min=0" example:"12500" minimum:"0" maximum:"999999999"`
}

// ToScoreEntry converts a submission request to a models.ScoreEntry
func (r *ScoreSubmissionRequest) ToScoreEntry() *models.ScoreEntry {
	return &models.ScoreEntry{
		Initials: r.Initials,
		Score:    r.Score,
		// Timestamp will be set during validation
	}
}

// ScoreSubmissionResponse represents the response after submitting a score
// This includes both the submitted entry and the current leaderboard state
type ScoreSubmissionResponse struct {
	Message     string              `json:"message" example:"Score submitted successfully"`
	Entry       *models.ScoreEntry  `json:"entry"`
	Leaderboard *models.Leaderboard `json:"leaderboard"`
	Rank        *int                `json:"rank,omitempty" example:"3"` // Position in leaderboard (1-10), nil if not in top 10
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error" example:"Invalid request format"`
	Details map[string]interface{} `json:"details,omitempty"` // Additional error context
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(message string, details ...map[string]interface{}) *ErrorResponse {
	resp := &ErrorResponse{Error: message}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	return resp
}

// HealthResponse represents a standardized health check response
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`                 // Service status
	Service   string `json:"service" example:"rawboard"`               // Service name
	Version   string `json:"version" example:"1.0.0"`                  // Service version
	Timestamp string `json:"timestamp" example:"2025-07-13T19:30:00Z"` // Current server time
}

// NewHealthResponse creates a standardized health response
func NewHealthResponse(status, service, version, timestamp string) *HealthResponse {
	return &HealthResponse{
		Status:    status,
		Service:   service,
		Version:   version,
		Timestamp: timestamp,
	}
}

// WelcomeResponse represents a standardized API welcome response
type WelcomeResponse struct {
	Message      string                 `json:"message" example:"🎮 Welcome to Rawboard"`
	Version      string                 `json:"version" example:"1.0.0"`
	Endpoints    map[string]interface{} `json:"endpoints"`
	Features     []string               `json:"features"`
	ArcadeSpirit string                 `json:"arcade_spirit" example:"🕹️ Bringing back the classic arcade experience!"`
}

// NewWelcomeResponse creates a standardized welcome response
func NewWelcomeResponse() *WelcomeResponse {
	return &WelcomeResponse{
		Message: "🎮 Welcome to Rawboard - Traditional Arcade Leaderboard Service",
		Version: "1.0.0",
		Endpoints: map[string]interface{}{
			"health":                            "GET /health",
			"get_leaderboard":                   "GET /api/v1/games/{gameId}/leaderboard",
			"submit_score (requires API key)":   "POST /api/v1/games/{gameId}/scores",
			"get_player_stats":                  "GET /api/v1/games/{gameId}/players/{initials}/stats",
			"get_all_scores (requires API key)": "GET /api/v1/games/{gameId}/scores/all",
		},
		Features: []string{
			"3-character initials (traditional arcade style)",
			"Top-10 score tracking with highest score per player",
			"Complete score history tracking",
			"Individual player statistics",
			"API key authentication for score submission",
			"Rate limiting and request timeouts",
			"Real-time leaderboard updates",
		},
		ArcadeSpirit: "🕹️ Bringing back the classic arcade experience!",
	}
}

// PlayerStatsResponse represents the response when getting player statistics
type PlayerStatsResponse struct {
	*models.PlayerStats
	Rank *int `json:"rank,omitempty" example:"3"` // Current rank in leaderboard (1-10), nil if not in top 10
}

// AllScoresResponse represents the response when getting all scores for a game
type AllScoresResponse struct {
	*models.AllScoresRecord
	TotalPlayers int     `json:"total_players" example:"25"`      // Number of unique players
	HighestScore int64   `json:"highest_score" example:"50000"`   // Highest score across all players
	AverageScore float64 `json:"average_score" example:"12500.5"` // Average score across all submissions
}

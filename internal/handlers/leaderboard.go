package handlers

import (
	"net/http"

	"rawboard/internal/leaderboard"

	"github.com/gin-gonic/gin"
)

// LeaderboardHandler handles HTTP requests for leaderboard operations
type LeaderboardHandler struct {
	service *leaderboard.Service
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(service *leaderboard.Service) *LeaderboardHandler {
	return &LeaderboardHandler{service: service}
}

// SubmitScore handles POST /api/v1/games/:gameId/scores
func (h *LeaderboardHandler) SubmitScore(c *gin.Context) {
	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	// Validate gameID format (prevent injection attacks and ensure reasonable length)
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID must be between 1 and 50 characters"})
		return
	}

	var req struct {
		Initials string `json:"initials" binding:"required"`
		Score    int64  `json:"score" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Additional validation for score range (prevent unrealistic scores)
	if req.Score > 999999999 { // 9 digits max for traditional arcade feel
		c.JSON(http.StatusBadRequest, gin.H{"error": "Score too high - maximum allowed is 999,999,999"})
		return
	}

	// Submit the score
	err := h.service.SubmitScore(c.Request.Context(), gameID, req.Initials, req.Score)
	if err != nil {
		// Log the error for debugging but don't expose internal details
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Score submitted successfully",
		"game_id":  gameID,
		"initials": req.Initials,
		"score":    req.Score,
	})
}

// GetLeaderboard handles GET /api/v1/games/:gameId/leaderboard
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No leaderboard found for this game"})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

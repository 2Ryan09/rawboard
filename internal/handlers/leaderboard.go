package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"rawboard/internal/leaderboard"
	"rawboard/internal/models"

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
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	// Validate gameID format (prevent injection attacks and ensure reasonable length)
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	var req ScoreSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidRequest, "Invalid request format",
			map[string]interface{}{"validation_error": err.Error()}))
		return
	}

	// Convert to score entry and validate
	entry := req.ToScoreEntry()
	if err := entry.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeValidationFailed, err.Error()))
		return
	}

	// Submit the score
	err := h.service.SubmitScore(c.Request.Context(), gameID, entry.Initials, entry.Score)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInternalError, err.Error()))
		return
	}

	// Get updated leaderboard to include in response
	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), gameID)
	if err != nil {
		// If we can't get the leaderboard, still return success for the submission
		c.JSON(http.StatusCreated, ScoreSubmissionResponse{
			Message: "Score submitted successfully",
			Entry:   entry,
		})
		return
	}

	// Find the rank of the submitted score or the player's current position
	var rank *int
	for i, scoreEntry := range leaderboard.Entries {
		if scoreEntry.Initials == entry.Initials {
			// Player is on the leaderboard - return their current rank
			// This could be either the just-submitted score (if it's their new high score)
			// or their existing high score (if this submission was lower)
			rankValue := i + 1
			rank = &rankValue
			break
		}
	}
	// If rank is still nil, the player is not in the top 10

	c.JSON(http.StatusCreated, ScoreSubmissionResponse{
		Message:     "Score submitted successfully",
		Entry:       entry,
		Leaderboard: leaderboard,
		Rank:        rank,
	})
}

// GetLeaderboard handles GET /api/v1/games/:gameId/leaderboard
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, NewStandardErrorResponse(
			ErrorCodeGameNotFound, "No leaderboard found for this game",
			map[string]interface{}{"game_id": gameID}))
		return
	}

	// Return the models.Leaderboard directly - no need for conversion
	// Ensure it's typed as models.Leaderboard for documentation
	var response *models.Leaderboard = leaderboard
	c.JSON(http.StatusOK, response)
}

// GetPlayerStats handles GET /api/v1/games/:gameId/players/:initials/stats
func (h *LeaderboardHandler) GetPlayerStats(c *gin.Context) {
	gameID := c.Param("gameId")
	initials := c.Param("initials")

	if gameID == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	if initials == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidInitials, "Player initials are required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	// Validate initials format
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"initials", initials, "exactly 3 characters"))
		return
	}

	stats, err := h.service.GetPlayerStats(c.Request.Context(), gameID, initials)
	if err != nil {
		c.JSON(http.StatusNotFound, NewStandardErrorResponse(
			ErrorCodePlayerNotFound, "No stats found for this player",
			map[string]interface{}{
				"game_id":  gameID,
				"initials": initials,
			}))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllScores handles GET /api/v1/games/:gameId/scores/all (admin endpoint)
func (h *LeaderboardHandler) GetAllScores(c *gin.Context) {
	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	allScores, err := h.service.GetAllScoresForGame(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, NewStandardErrorResponse(
			ErrorCodeScoreHistoryEmpty, "No score history found for this game",
			map[string]interface{}{"game_id": gameID}))
		return
	}

	c.JSON(http.StatusOK, allScores)
}

// GetEnhancedPlayerStats handles GET /api/v1/games/:gameId/players/:initials/stats/enhanced
func (h *LeaderboardHandler) GetEnhancedPlayerStats(c *gin.Context) {
	gameID := c.Param("gameId")
	initials := c.Param("initials")

	if gameID == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	if initials == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidInitials, "Player initials are required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	// Validate initials format
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"initials", initials, "exactly 3 characters"))
		return
	}

	// Check if score history should be included
	includeHistory := c.Query("include_history") == "true"

	stats, err := h.service.GetEnhancedPlayerStats(c.Request.Context(), gameID, initials, includeHistory)
	if err != nil {
		c.JSON(http.StatusNotFound, NewStandardErrorResponse(
			ErrorCodePlayerNotFound, "No stats found for this player",
			map[string]interface{}{
				"game_id":  gameID,
				"initials": initials,
			}))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetScoreAnalysis handles GET /api/v1/games/:gameId/scores/analyze
func (h *LeaderboardHandler) GetScoreAnalysis(c *gin.Context) {
	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, NewStandardErrorResponse(
			ErrorCodeInvalidGameID, "Game ID is required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
			"gameId", gameID, "length between 1 and 50 characters"))
		return
	}

	// Parse top players limit (default to 5, max 10)
	topPlayersLimit := 5
	if limitStr := c.Query("top_players"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 10 {
			topPlayersLimit = limit
		}
	}

	analysis, err := h.service.GetScoreAnalysis(c.Request.Context(), gameID, topPlayersLimit)
	if err != nil {
		c.JSON(http.StatusNotFound, NewStandardErrorResponse(
			ErrorCodeScoreHistoryEmpty, "No score analysis available for this game",
			map[string]interface{}{"game_id": gameID}))
		return
	}

	c.JSON(http.StatusOK, analysis)
}

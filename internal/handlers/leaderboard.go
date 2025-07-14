package handlers

import (
	"net/http"

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
		c.JSON(http.StatusBadRequest, NewErrorResponse("Game ID is required"))
		return
	}

	// Validate gameID format (prevent injection attacks and ensure reasonable length)
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Game ID must be between 1 and 50 characters"))
		return
	}

	var req ScoreSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request format", map[string]interface{}{
			"validation_error": err.Error(),
		}))
		return
	}

	// Convert to score entry and validate
	entry := req.ToScoreEntry()
	if err := entry.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	// Submit the score
	err := h.service.SubmitScore(c.Request.Context(), gameID, entry.Initials, entry.Score)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
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

	// Find the rank of the submitted score
	var rank *int
	for i, scoreEntry := range leaderboard.Entries {
		if scoreEntry.Initials == entry.Initials && scoreEntry.Score == entry.Score {
			rankValue := i + 1
			rank = &rankValue
			break
		}
	}

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
		c.JSON(http.StatusBadRequest, NewErrorResponse("Game ID is required"))
		return
	}

	// Validate gameID format
	if len(gameID) > 50 || len(gameID) < 1 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Game ID must be between 1 and 50 characters"))
		return
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("No leaderboard found for this game", map[string]interface{}{
			"game_id": gameID,
		}))
		return
	}

	// Return the models.Leaderboard directly - no need for conversion
	// Ensure it's typed as models.Leaderboard for documentation
	var response *models.Leaderboard = leaderboard
	c.JSON(http.StatusOK, response)
}

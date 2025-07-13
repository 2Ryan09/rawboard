package handlers

import (
	"net/http"

	"rawboard/internal/leaderboard"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the API routes
func SetupRoutes(r *gin.Engine, leaderboardService *leaderboard.Service) {
	leaderboardHandler := NewLeaderboardHandler(leaderboardService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Welcome endpoint
		v1.GET("/", welcomeHandler)

		// Game routes (simplified arcade style)
		games := v1.Group("/games")
		{
			games.POST("/:gameId/scores", leaderboardHandler.SubmitScore)        // POST /api/v1/games/:gameId/scores
			games.GET("/:gameId/leaderboard", leaderboardHandler.GetLeaderboard) // GET /api/v1/games/:gameId/leaderboard
		}
	}
}

func welcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "Welcome to Rawboard Arcade API!",
		"service":     "rawboard-arcade",
		"version":     "1.0.0",
		"api_version": "v1",
		"description": "Traditional arcade-style leaderboard service",
		"endpoints": gin.H{
			"health":          "/health",
			"submit_score":    "POST /api/v1/games/:gameId/scores",
			"get_leaderboard": "GET /api/v1/games/:gameId/leaderboard",
		},
		"usage": gin.H{
			"submit_score": gin.H{
				"method": "POST",
				"url":    "/api/v1/games/tetris/scores",
				"body": gin.H{
					"initials": "AAA",
					"score":    15000,
				},
			},
			"get_leaderboard": gin.H{
				"method": "GET",
				"url":    "/api/v1/games/tetris/leaderboard",
			},
		},
	})
}

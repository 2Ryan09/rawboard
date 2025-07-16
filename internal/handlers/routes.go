package handlers

import (
	"net/http"

	"rawboard/internal/leaderboard"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the API routes
func SetupRoutes(r *gin.Engine, leaderboardService *leaderboard.Service, apiKeyMiddleware gin.HandlerFunc) {
	leaderboardHandler := NewLeaderboardHandler(leaderboardService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Welcome endpoint (public)
		v1.GET("/", welcomeHandler)

		// Game routes
		games := v1.Group("/games")
		{
			// Public endpoints (no authentication required)
			games.GET("/:gameId/leaderboard", leaderboardHandler.GetLeaderboard)             // GET /api/v1/games/:gameId/leaderboard
			games.GET("/:gameId/players/:initials/stats", leaderboardHandler.GetPlayerStats) // GET /api/v1/games/:gameId/players/:initials/stats

			// Protected endpoints (API key required)
			protected := games.Group("")
			protected.Use(apiKeyMiddleware)
			{
				protected.POST("/:gameId/scores", leaderboardHandler.SubmitScore)     // POST /api/v1/games/:gameId/scores
				protected.GET("/:gameId/scores/all", leaderboardHandler.GetAllScores) // GET /api/v1/games/:gameId/scores/all (admin)
			}
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
			"health":           "/health",
			"submit_score":     "POST /api/v1/games/:gameId/scores (API key required)",
			"get_leaderboard":  "GET /api/v1/games/:gameId/leaderboard (public)",
			"get_player_stats": "GET /api/v1/games/:gameId/players/:initials/stats (public)",
			"get_all_scores":   "GET /api/v1/games/:gameId/scores/all (API key required, admin)",
		},
		"authentication": gin.H{
			"type": "API Key",
			"headers": []string{
				"X-API-Key: <your-api-key>",
				"Authorization: Bearer <your-api-key>",
			},
			"required_for": []string{
				"POST /api/v1/games/:gameId/scores",
				"GET /api/v1/games/:gameId/scores/all",
			},
			"public_endpoints": []string{
				"GET /api/v1/games/:gameId/leaderboard",
				"GET /api/v1/games/:gameId/players/:initials/stats",
				"GET /health",
			},
		},
		"usage": gin.H{
			"submit_score": gin.H{
				"method": "POST",
				"url":    "/api/v1/games/tetris/scores",
				"headers": gin.H{
					"Content-Type": "application/json",
					"X-API-Key":    "<your-api-key>",
				},
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

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"

	"rawboard/internal/database"
	"rawboard/internal/handlers"
	"rawboard/internal/leaderboard"
	"rawboard/internal/middleware"
)

func main() {
	// Sentry initialization
	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn != "" {
		env := getEnvironment()

		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDsn,
			Environment:      env,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
			SampleRate:       1.0,
			ServerName:       "rawboard",
		}); err != nil {
			fmt.Printf("❌ Sentry initialization failed: %v\n", err)
		} else {
			fmt.Printf("✅ Sentry monitoring enabled\n")
		}
	}

	// Set Gin mode based on environment
	if getEnvironment() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))

	// Initialize database
	db, err := database.NewValkeyDB()
	if err != nil {
		fmt.Printf("❌ Database initialization failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Printf("✅ Database connected\n")

	// Initialize services
	leaderboardService := leaderboard.NewService(db)

	// Setup API key authentication
	apiKey := os.Getenv("RAWBOARD_API_KEY")
	if apiKey == "" {
		if getEnvironment() == "production" {
			fmt.Printf("❌ FATAL: API key is required in production environment\n")
			fmt.Printf("❌ Please set the RAWBOARD_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Printf("⚠️  Warning: No RAWBOARD_API_KEY set - authentication disabled\n")
		fmt.Printf("⚠️  This is only allowed in development mode\n")
	} else {
		fmt.Printf("✅ API key authentication enabled\n")
	}
	apiKeyMiddleware := middleware.APIKeyMiddleware(apiKey)

	// Infrastructure health check
	router.GET("/health", healthCheck)

	// Welcome endpoint with API documentation
	router.GET("/", apiWelcomeHandler)

	// API routes
	v1 := router.Group("/api/v1")

	// Initialize handlers
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)

	// Public routes (no authentication required)
	v1.GET("/games/:gameId/leaderboard", leaderboardHandler.GetLeaderboard)

	// Protected routes (API key required)
	protected := v1.Group("/games/:gameId")
	protected.Use(apiKeyMiddleware)
	{
		protected.POST("/scores", leaderboardHandler.SubmitScore)
	}

	// Start server
	fmt.Printf("🚀 Starting Rawboard server on port 8080\n")
	fmt.Printf("🎮 Traditional arcade leaderboard service ready!\n")

	if err := router.Run(":8080"); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

func healthCheck(c *gin.Context) {
	response := handlers.NewHealthResponse(
		"healthy",
		"rawboard",
		"1.0.0",
		time.Now().UTC().Format(time.RFC3339),
	)
	c.JSON(http.StatusOK, response)
}

func apiWelcomeHandler(c *gin.Context) {
	response := handlers.NewWelcomeResponse()
	c.JSON(http.StatusOK, response)
}

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

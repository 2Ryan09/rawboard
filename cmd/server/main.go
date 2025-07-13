package main

import (
	"fmt"
	"net/http"
	"os"

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
		}
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

	// Initialize services
	leaderboardService := leaderboard.NewService(db)

	// Setup API key authentication
	apiKey := os.Getenv("RAWBOARD_API_KEY")
	if apiKey == "" {
		fmt.Printf("⚠️  Warning: No RAWBOARD_API_KEY set - authentication disabled\n")
	} else {
		fmt.Printf("✅ API key authentication enabled\n")
	}
	apiKeyMiddleware := middleware.APIKeyMiddleware(apiKey)

	// Infrastructure health check
	router.GET("/health", healthCheck)

	// Setup API routes
	handlers.SetupRoutes(router, leaderboardService, apiKeyMiddleware)

	// Start server
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rawboard",
	})
}

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

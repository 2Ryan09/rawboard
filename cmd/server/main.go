package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	bugsnaggin "github.com/bugsnag/bugsnag-go-gin"
	"github.com/bugsnag/bugsnag-go/v2"

	"rawboard/internal/database"
	"rawboard/internal/handlers"
	"rawboard/internal/leaderboard"
	"rawboard/internal/middleware"
)

func main() {
	// Bugsnag initialization
	bugsnagAPIKey := os.Getenv("BUGSNAG_API_KEY")

	// Set Gin mode based on environment
	if getEnvironment() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add Bugsnag middleware if API key is provided
	if bugsnagAPIKey != "" {
		env := getEnvironment()
		router.Use(bugsnaggin.AutoNotify(bugsnag.Configuration{
			APIKey:          bugsnagAPIKey,
			ReleaseStage:    env,
			AppVersion:      "1.0.0",
			Hostname:        "rawboard",
			ProjectPackages: []string{"main", "github.com/2ryan09/rawboard"},
		}))
		fmt.Printf("✅ Bugsnag monitoring enabled\n")
	}

	// Initialize database - required for operation
	fmt.Printf("🔌 Attempting database connection...\n")
	db, err := database.NewValkeyDB()
	if err != nil {
		fmt.Printf("❌ Database initialization failed: %v\n", err)
		fmt.Printf("❌ Rawboard requires a Redis/Valkey database to operate\n")
		os.Exit(1)
	}
	fmt.Printf("✅ Database connected\n")
	defer db.Close()

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

	// Setup all API routes using the handlers package
	handlers.SetupRoutes(router, leaderboardService, apiKeyMiddleware)

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

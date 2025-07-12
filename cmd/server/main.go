package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
)

func main() {
	// Sentry initialization
	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn != "" {
		fmt.Println("🔧 Initializing Sentry...")
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDsn,
			Environment:      getEnvironment(),
			Debug:            true, // Enable debug output
			TracesSampleRate: 1.0,  // Capture 100% of transactions for performance monitoring
			SampleRate:       1.0,  // Capture 100% of errors
		}); err != nil {
			fmt.Printf("❌ Sentry initialization failed: %v\n", err)
		} else {
			fmt.Printf("✅ Sentry initialized successfully for environment: %s\n", getEnvironment())
		}
	} else {
		fmt.Println("⚠️  SENTRY_DSN not set - Sentry monitoring disabled")
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))

	// Infrastructure health check
	router.GET("/health", healthCheck)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/", welcomeHandler)
		v1.GET("/test-error", testErrorHandler) // Test endpoint for Sentry
	}

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

func welcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "Welcome to Rawboard Arcade API!",
		"service":     "rawboard-arcade",
		"version":     "1.0.0",
		"api_version": "v1",
		"endpoints": gin.H{
			"health":   "/health",
			"api_root": "/api/v1/",
		},
	})
}

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

func testErrorHandler(c *gin.Context) {
	// Test Sentry error capture
	err := fmt.Errorf("test error for Sentry monitoring")
	sentry.CaptureException(err)
	
	// Flush to ensure the error is sent immediately
	sentry.Flush(2 * time.Second)

	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "This is a test error for Sentry",
		"message": "Check your Sentry dashboard for this error",
	})
}

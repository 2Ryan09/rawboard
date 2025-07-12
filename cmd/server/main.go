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
		v1.GET("/test-panic", testPanicHandler) // Test panic for Sentry
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
	// Add some context to Sentry
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("test_type", "manual_error")
		scope.SetLevel(sentry.LevelError)
		scope.SetContext("request", map[string]interface{}{
			"url":    c.Request.URL.String(),
			"method": c.Request.Method,
		})
	})

	// Test Sentry error capture with more context
	err := fmt.Errorf("CRITICAL: Test error for Sentry monitoring - timestamp: %d", time.Now().Unix())
	eventID := sentry.CaptureException(err)
	
	fmt.Printf("🔍 Captured Sentry error with ID: %s\n", *eventID)

	// Flush to ensure the error is sent immediately
	if sentry.Flush(5 * time.Second) {
		fmt.Println("✅ Sentry flush successful")
	} else {
		fmt.Println("❌ Sentry flush failed")
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error":     "This is a test error for Sentry",
		"message":   "Check your Sentry dashboard for this error",
		"sentry_id": *eventID,
	})
}

func testPanicHandler(c *gin.Context) {
	// This will trigger a panic that should be caught by Sentry
	panic("This is a test panic for Sentry monitoring!")
}

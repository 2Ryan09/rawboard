package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Infrastructure health check
	router.GET("/health", healthCheck)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/", welcomeHandler)
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

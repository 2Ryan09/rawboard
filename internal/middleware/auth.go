package middleware

import (
	"net/http"
	"strings"

	"rawboard/internal/handlers"

	"github.com/gin-gonic/gin"
)

// APIKeyMiddleware validates API key for protected endpoints
func APIKeyMiddleware(validAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation if no API key is configured (development)
		if validAPIKey == "" {
			c.Next()
			return
		}

		// Check X-API-Key header first
		apiKey := c.GetHeader("X-API-Key")

		// If not found, check Authorization header with Bearer format
		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("API key required", map[string]interface{}{
				"message": "Please provide API key in X-API-Key header or Authorization: Bearer <key>",
			}))
			c.Abort()
			return
		}

		if apiKey != validAPIKey {
			c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid API key"))
			c.Abort()
			return
		}

		// API key is valid, continue to next handler
		c.Next()
	}
}

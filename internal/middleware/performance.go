package middleware

import (
	"context"
	"sync"
	"time"

	"rawboard/internal/handlers"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// TimeoutMiddleware adds request timeout to prevent hanging requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

// RateLimitMiddleware implements simple in-memory rate limiting
// For production, consider using Redis-based rate limiting
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	mu := sync.RWMutex{}

	return gin.HandlerFunc(func(c *gin.Context) {
		// Use client IP as the key for rate limiting
		key := c.ClientIP()

		mu.RLock()
		limiter, exists := limiters[key]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			// Double-check pattern to avoid race conditions
			if limiter, exists = limiters[key]; !exists {
				limiter = rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize)
				limiters[key] = limiter
			}
			mu.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(429, handlers.NewErrorResponse("Rate limit exceeded", map[string]interface{}{
				"retry_after": "1s",
			}))
			c.Abort()
			return
		}

		c.Next()
	})
}

// CleanupOldLimiters removes old rate limiters to prevent memory leaks
// This should be called periodically in a background goroutine
func CleanupOldLimiters(limiters map[string]*rate.Limiter, mu *sync.RWMutex) {
	mu.Lock()
	defer mu.Unlock()

	// Simple cleanup: clear all limiters every hour
	// In production, implement more sophisticated cleanup based on last access time
	for key := range limiters {
		delete(limiters, key)
	}
}

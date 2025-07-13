package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validAPIKey := "test-api-key-123"

	// Test handler that should only be reached with valid auth
	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	t.Run("allows request with valid X-API-Key header", func(t *testing.T) {
		router := gin.New()
		router.Use(APIKeyMiddleware(validAPIKey))
		router.POST("/test", testHandler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("X-API-Key", validAPIKey)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("allows request with valid Authorization Bearer header", func(t *testing.T) {
		router := gin.New()
		router.Use(APIKeyMiddleware(validAPIKey))
		router.POST("/test", testHandler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Authorization", "Bearer "+validAPIKey)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("rejects request with invalid API key", func(t *testing.T) {
		router := gin.New()
		router.Use(APIKeyMiddleware(validAPIKey))
		router.POST("/test", testHandler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("X-API-Key", "invalid-key")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("rejects request with no API key", func(t *testing.T) {
		router := gin.New()
		router.Use(APIKeyMiddleware(validAPIKey))
		router.POST("/test", testHandler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("allows request when no API key is configured (development mode)", func(t *testing.T) {
		router := gin.New()
		router.Use(APIKeyMiddleware("")) // Empty API key = development mode
		router.POST("/test", testHandler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 in development mode, got %d", w.Code)
		}
	})
}

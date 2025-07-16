package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMiddlewareBehaviors focuses on middleware security and functionality behaviors
func TestMiddlewareBehaviors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("API Key Authentication: Access Control Behavior", func(t *testing.T) {
		// Behavior: Middleware should enforce API key requirements correctly
		validAPIKey := "test-api-key-12345"
		middleware := APIKeyMiddleware(validAPIKey)

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})

		// Test 1: No API key provided
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 without API key, got %d", w.Code)
		}

		// Test 2: Invalid API key
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", "invalid-key")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 with invalid API key, got %d", w.Code)
		}

		// Test 3: Valid API key via X-API-Key header
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", validAPIKey)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 with valid X-API-Key, got %d", w.Code)
		}

		// Test 4: Valid API key via Authorization header
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validAPIKey)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 with valid Authorization bearer, got %d", w.Code)
		}
	})

	t.Run("Development Mode: Bypass Behavior", func(t *testing.T) {
		// Behavior: When no API key is configured, middleware should allow access
		middleware := APIKeyMiddleware("") // Empty API key = development mode

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "development access"})
		})

		// Should allow access without API key in development mode
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 in development mode, got %d", w.Code)
		}
	})

	t.Run("Error Response Consistency", func(t *testing.T) {
		// Behavior: Middleware should return consistent error responses
		middleware := APIKeyMiddleware("test-key")

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Test error response format
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Error("Error response should be JSON")
		}

		// Response should contain error message
		body := w.Body.String()
		if body == "" {
			t.Error("Error response should have body")
		}
	})

	t.Run("Header Parsing: Case Insensitive Behavior", func(t *testing.T) {
		// Behavior: Middleware should handle various header case formats
		validAPIKey := "test-api-key-12345"
		middleware := APIKeyMiddleware(validAPIKey)

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})

		// Test different header case variations
		headerVariations := []string{
			"X-API-Key",
			"x-api-key",
			"X-Api-Key",
		}

		for _, header := range headerVariations {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set(header, validAPIKey)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Header variation %s should work, got status %d", header, w.Code)
			}
		}
	})

	t.Run("Authorization Bearer Token: Format Handling", func(t *testing.T) {
		// Behavior: Middleware should properly parse Bearer token format
		validAPIKey := "test-api-key-12345"
		middleware := APIKeyMiddleware(validAPIKey)

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})

		// Test valid Bearer format
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validAPIKey)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Valid Bearer token should work, got %d", w.Code)
		}

		// Test invalid Bearer format (wrong token)
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Invalid Bearer token should fail, got %d", w.Code)
		}

		// Test malformed Authorization header
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "NotBearer "+validAPIKey)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Malformed Authorization header should fail, got %d", w.Code)
		}
	})

	t.Run("Security Headers: Response Enhancement", func(t *testing.T) {
		// Behavior: Middleware should not interfere with response headers
		middleware := APIKeyMiddleware("test-key")

		router := gin.New()
		router.Use(middleware)
		router.GET("/protected", func(c *gin.Context) {
			c.Header("Custom-Header", "test-value")
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		if w.Header().Get("Custom-Header") != "test-value" {
			t.Error("Middleware should not interfere with custom response headers")
		}
	})
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS_DefaultConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test router with CORS
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request with allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	// Create test recorder
	w := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin to be 'http://localhost:3000', got %s",
			w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be 'true', got %s",
			w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test router with CORS
	router := gin.New()
	router.Use(CORS())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,X-Telegram-Init-Data")

	// Create test recorder
	w := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Check CORS headers
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set")
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Error("Expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORS_UnallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test router with CORS
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request with unallowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://malicious-site.com")

	// Create test recorder
	w := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that CORS origin header is not set for unallowed origins
	if w.Header().Get("Access-Control-Allow-Origin") == "https://malicious-site.com" {
		t.Error("Expected Access-Control-Allow-Origin to not be set for unallowed origin")
	}
}

func TestCORS_CustomConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create custom config
	config := CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"X-Custom-Header"},
		AllowCredentials: false,
		MaxAge:           3600,
	}

	// Create test router with custom CORS config
	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	// Create test recorder
	w := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin to be 'https://example.com', got %s",
			w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") == "true" {
		t.Error("Expected Access-Control-Allow-Credentials to not be 'true'")
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{"http://localhost:3000", []string{"http://localhost:3000"}, true},
		{"https://example.com", []string{"*"}, true},
		{"https://api.example.com", []string{"*.example.com"}, true},
		{"https://malicious.com", []string{"http://localhost:3000"}, false},
		{"https://subdomain.evil.com", []string{"*.example.com"}, false},
	}

	for _, tt := range tests {
		result := isOriginAllowed(tt.origin, tt.allowedOrigins)
		if result != tt.expected {
			t.Errorf("isOriginAllowed(%s, %v) = %t, want %t",
				tt.origin, tt.allowedOrigins, result, tt.expected)
		}
	}
}

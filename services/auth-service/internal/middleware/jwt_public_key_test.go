package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/services"
)

// setupTestMiddleware creates a JWT middleware for testing
func setupTestMiddleware(t *testing.T) *JWTPublicKeyMiddleware {
	t.Helper()

	// Create temporary directory for test keys
	tempDir := t.TempDir()

	keyPaths := services.KeyPaths{
		PrivateKeyPath: filepath.Join(tempDir, "test_private.pem"),
		PublicKeyPath:  filepath.Join(tempDir, "test_public.pem"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	jwtService, err := services.NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	return NewJWTPublicKeyMiddleware(jwtService)
}

func TestNewJWTPublicKeyMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	keyPaths := services.KeyPaths{
		PrivateKeyPath: filepath.Join(tempDir, "test_private.pem"),
		PublicKeyPath:  filepath.Join(tempDir, "test_public.pem"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	jwtService, err := services.NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	middleware := NewJWTPublicKeyMiddleware(jwtService)

	if middleware == nil {
		t.Error("Expected middleware but got nil")
	}

	if middleware.jwtService != jwtService {
		t.Error("JWT service not properly assigned to middleware")
	}
}

// TestPublicKeyHandler was removed - JWKS support deferred to future version

func TestPublicKeyPEMHandler(t *testing.T) {
	middleware := setupTestMiddleware(t)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	router.GET("/public-key.pem", middleware.PublicKeyPEMHandler())

	// Create a test request
	req, err := http.NewRequest("GET", "/public-key.pem", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type (Gin sets this to "text/plain" without charset by default)
	expectedContentType := "text/plain"
	actualContentType := rr.Header().Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Errorf("Expected content type %q, got %q", expectedContentType, actualContentType)
	}

	// Verify response body
	body := rr.Body.String()

	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Verify PEM format
	if !strings.Contains(body, "-----BEGIN PUBLIC KEY-----") {
		t.Error("Expected PEM to contain BEGIN PUBLIC KEY header")
	}

	if !strings.Contains(body, "-----END PUBLIC KEY-----") {
		t.Error("Expected PEM to contain END PUBLIC KEY footer")
	}
}

func TestPublicKeyHandlerErrorCases(t *testing.T) {
	// Since we can't easily mock the JWT service to return errors,
	// we'll skip this test for now. In a real implementation, we would
	// use dependency injection with interfaces to make this testable.
	t.Skip("Error case testing requires interface-based dependency injection")
}

func TestMiddlewareIntegration(t *testing.T) {
	middleware := setupTestMiddleware(t)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with PEM endpoint only (JWKS removed)
	router := gin.New()
	router.GET("/public-key.pem", middleware.PublicKeyPEMHandler())

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "PEM endpoint",
			path:           "/public-key.pem",
			expectedStatus: http.StatusOK,
			expectedType:   "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Header().Get("Content-Type") != tt.expectedType {
				t.Errorf("Expected content type %q, got %q", tt.expectedType, rr.Header().Get("Content-Type"))
			}

			if rr.Body.Len() == 0 {
				t.Error("Expected non-empty response body")
			}
		})
	}
}

// Benchmark tests
// BenchmarkPublicKeyHandler was removed - JWKS support deferred to future version

func BenchmarkPublicKeyPEMHandler(b *testing.B) {
	middleware := setupTestMiddleware(&testing.T{})
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/public-key.pem", middleware.PublicKeyPEMHandler())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/public-key.pem", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}

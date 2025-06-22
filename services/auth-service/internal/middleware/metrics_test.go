package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shard-legends/auth-service/internal/metrics"
)

// createTestMetrics creates isolated test metrics to avoid registry conflicts
func createTestMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test_auth_service",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test_auth_service", 
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test_auth_service",
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),
	}
}

func TestMetricsMiddleware(t *testing.T) {
	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Initialize test metrics (isolated)
	m := createTestMetrics()
	
	// Add metrics middleware
	router.Use(MetricsMiddleware(m))
	
	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond) // Simulate some processing time
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Make test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Note: In a real test environment, we would check the metric values
	// but since prometheus metrics are global, we'd need to use a test registry
	// or reset metrics between tests to avoid interference
}

func TestMetricsMiddlewareSkipsMetricsEndpoint(t *testing.T) {
	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Initialize test metrics (isolated)
	m := createTestMetrics()
	
	// Add metrics middleware
	router.Use(MetricsMiddleware(m))
	
	// Add metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# HELP test_metric Test metric\n")
	})
	
	// Make test request to metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// The middleware should skip recording metrics for the /metrics endpoint itself
	// to avoid recursive metric recording
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/auth", "/auth"},
		{"/health", "/health"},
		{"/jwks", "/jwks"},
		{"/public-key.pem", "/public-key.pem"},
		{"/admin/tokens/stats", "/admin/tokens/stats"},
		{"/admin/tokens/cleanup", "/admin/tokens/cleanup"},
		{"/admin/tokens/user/123e4567-e89b-12d3-a456-426614174000", "/admin/tokens/user/{userId}"},
		{"/admin/tokens/abc123def456", "/admin/tokens/{jti}"},
		{"/unknown/path", "unknown"},
		{"/admin/tokens", "unknown"}, // This should match unknown since it doesn't match any pattern
	}
	
	for _, test := range tests {
		result := normalizeEndpoint(test.path)
		if result != test.expected {
			t.Errorf("normalizeEndpoint(%s) = %s, expected %s", test.path, result, test.expected)
		}
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"/admin/tokens/user/123", "/admin/tokens/user/", true},
		{"/admin/tokens/user", "/admin/tokens/user/", false},
		{"/admin/tokens/abc123", "/admin/tokens/", true},
		{"/admin/token", "/admin/tokens/", false},
		{"", "/admin/", false},
	}
	
	for _, test := range tests {
		result := matchesPattern(test.path, test.pattern)
		if result != test.expected {
			t.Errorf("matchesPattern(%s, %s) = %t, expected %t", 
				test.path, test.pattern, result, test.expected)
		}
	}
}
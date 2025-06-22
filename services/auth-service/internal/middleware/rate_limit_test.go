package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shard-legends/auth-service/internal/metrics"
)

// createTestRateLimitMetrics creates test metrics instance to avoid conflicts
func createTestRateLimitMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		AuthRateLimitHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test_rate_limit_auth_service",
				Name:      "auth_rate_limit_hits_total",
				Help:      "Total number of requests blocked by rate limiting",
			},
			[]string{"ip"},
		),
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	testMetrics := createTestRateLimitMetrics()
	
	// Create rate limiter with 2 requests per minute for testing
	rl := NewRateLimiter(2, logger, testMetrics)
	defer rl.Close()
	
	// Test IP
	testIP := "192.168.1.1"
	
	// First request should be allowed
	if !rl.allow(testIP) {
		t.Error("First request should be allowed")
	}
	
	// Second request should be allowed
	if !rl.allow(testIP) {
		t.Error("Second request should be allowed")
	}
	
	// Third request should be blocked
	if rl.allow(testIP) {
		t.Error("Third request should be blocked")
	}
}

func TestRateLimiter_Middleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	testMetrics := createTestRateLimitMetrics()
	
	// Create rate limiter with 10 requests per minute
	rl := NewRateLimiter(10, logger, testMetrics)
	defer rl.Close()
	
	// Create test router
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	
	// Create test recorder
	w := httptest.NewRecorder()
	
	// Make request
	router.ServeHTTP(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_Middleware_RateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	testMetrics := createTestRateLimitMetrics()
	
	// Create rate limiter with 1 request per minute for testing
	rl := NewRateLimiter(1, logger, testMetrics)
	defer rl.Close()
	
	// Create test router
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.2:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	if w1.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w1.Code)
	}
	
	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345" // Same IP
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", w2.Code)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	testMetrics := createTestRateLimitMetrics()
	
	// Create rate limiter with 1 request per minute
	rl := NewRateLimiter(1, logger, testMetrics)
	defer rl.Close()
	
	// Create test router
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Request from first IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	if w1.Code != http.StatusOK {
		t.Errorf("First IP: expected status 200, got %d", w1.Code)
	}
	
	// Request from second IP should also succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345" // Different IP
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusOK {
		t.Errorf("Second IP: expected status 200, got %d", w2.Code)
	}
}

func TestTokenBucket_Consume(t *testing.T) {
	bucket := &TokenBucket{
		tokens:     2,
		lastRefill: time.Now(),
	}
	
	rate := 30 * time.Second // 2 requests per minute
	capacity := 2
	
	// First consumption should succeed
	if !bucket.consume(rate, capacity) {
		t.Error("First consume should succeed")
	}
	
	// Second consumption should succeed
	if !bucket.consume(rate, capacity) {
		t.Error("Second consume should succeed")
	}
	
	// Third consumption should fail
	if bucket.consume(rate, capacity) {
		t.Error("Third consume should fail")
	}
	
	// Verify token count
	if bucket.tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", bucket.tokens)
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	bucket := &TokenBucket{
		tokens:     0,
		lastRefill: time.Now().Add(-time.Minute), // 1 minute ago
	}
	
	rate := 30 * time.Second // 2 requests per minute
	capacity := 2
	
	// Should refill 2 tokens (60 seconds / 30 seconds per token = 2 tokens)
	if !bucket.consume(rate, capacity) {
		t.Error("Consume should succeed after refill")
	}
	
	// Should still have 1 token left
	if bucket.tokens != 1 {
		t.Errorf("Expected 1 token after consume, got %d", bucket.tokens)
	}
}
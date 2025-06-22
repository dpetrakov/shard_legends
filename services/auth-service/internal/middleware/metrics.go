package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/metrics"
)

// MetricsMiddleware creates a middleware that collects HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid self-recording
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		
		// Increment in-flight requests
		m.HTTPRequestsInFlight.Inc()
		defer m.HTTPRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Record metrics after request completion
		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())
		
		// Normalize endpoint for better grouping
		endpoint := normalizeEndpoint(c.Request.URL.Path)
		
		m.RecordHTTPRequest(
			c.Request.Method,
			endpoint,
			statusCode,
			duration,
		)
	}
}

// normalizeEndpoint normalizes URL paths for better metric grouping
// Replaces dynamic segments with placeholders to avoid high cardinality
func normalizeEndpoint(path string) string {
	switch {
	case path == "/auth":
		return "/auth"
	case path == "/health":
		return "/health"
	case path == "/jwks":
		return "/jwks"
	case path == "/public-key.pem":
		return "/public-key.pem"
	case path == "/admin/tokens/stats":
		return "/admin/tokens/stats"
	case path == "/admin/tokens/cleanup":
		return "/admin/tokens/cleanup"
	case matchesPattern(path, "/admin/tokens/user/"):
		return "/admin/tokens/user/{userId}"
	case matchesPattern(path, "/admin/tokens/") && !matchesPattern(path, "/admin/tokens/user/"):
		return "/admin/tokens/{jti}"
	default:
		return "unknown"
	}
}

// matchesPattern checks if path starts with the given pattern
func matchesPattern(path, pattern string) bool {
	if len(path) < len(pattern) {
		return false
	}
	return path[:len(pattern)] == pattern
}
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// MetricsMiddleware creates a middleware that collects HTTP metrics
func MetricsMiddleware(metricsCollector *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if metricsCollector == nil {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()

		// If path is empty, use the raw path (useful for 404s)
		if path == "" {
			path = c.Request.URL.Path
		}

		// Increment in-flight requests
		metricsCollector.HTTPRequestsInFlight.Inc()
		defer metricsCollector.HTTPRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metricsCollector.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metricsCollector.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	}
}

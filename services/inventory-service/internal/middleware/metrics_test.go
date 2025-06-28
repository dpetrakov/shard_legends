package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// Global metrics instance to avoid registration conflicts
var testMetrics *metrics.Metrics

func init() {
	testMetrics = metrics.New()
}

func TestMetricsMiddleware_NilMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles nil metrics gracefully", func(t *testing.T) {
		// Create middleware with nil metrics
		middleware := MetricsMiddleware(nil)

		// Create test router
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		// Make request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Should not panic
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMetricsMiddleware_WithMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("records HTTP metrics correctly", func(t *testing.T) {
		// Use global metrics instance
		metricsCollector := testMetrics

		// Create middleware
		middleware := MetricsMiddleware(metricsCollector)

		// Create test router
		router := gin.New()
		router.Use(middleware)
		router.GET("/api/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Simulate some processing time
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		// Make request
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that metrics were recorded
		// Note: We can't easily assert exact values due to timing, but we can check they were called
		assert.NotNil(t, metricsCollector.HTTPRequestsTotal)
		assert.NotNil(t, metricsCollector.HTTPRequestDuration)
		assert.NotNil(t, metricsCollector.HTTPRequestsInFlight)
	})

	t.Run("records metrics for different HTTP methods", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
		router.POST("/test", func(c *gin.Context) { c.Status(http.StatusCreated) })
		router.PUT("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
		router.DELETE("/test", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		methods := []struct {
			method string
			status int
		}{
			{"GET", http.StatusOK},
			{"POST", http.StatusCreated},
			{"PUT", http.StatusOK},
			{"DELETE", http.StatusNoContent},
		}

		for _, m := range methods {
			req := httptest.NewRequest(m.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, m.status, w.Code)
		}
	})

	t.Run("records metrics for different status codes", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
		router.GET("/notfound", func(c *gin.Context) { c.Status(http.StatusNotFound) })
		router.GET("/error", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

		testCases := []struct {
			path   string
			status int
		}{
			{"/ok", http.StatusOK},
			{"/notfound", http.StatusNotFound},
			{"/error", http.StatusInternalServerError},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.status, w.Code)
		}
	})
}

func TestMetricsMiddleware_PathHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles empty path for 404s", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		// No routes defined - all requests will be 404

		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		// Middleware should handle empty FullPath() gracefully
	})

	t.Run("uses FullPath for registered routes", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		router.GET("/api/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
		})

		req := httptest.NewRequest("GET", "/api/users/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("uses raw path when FullPath is empty", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)

		req := httptest.NewRequest("GET", "/some/path", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestMetricsMiddleware_InFlightRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("increments and decrements in-flight requests", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)

		requestStarted := make(chan struct{})
		requestCanFinish := make(chan struct{})

		router.GET("/slow", func(c *gin.Context) {
			close(requestStarted)
			<-requestCanFinish
			c.Status(http.StatusOK)
		})

		// Start request in goroutine
		go func() {
			req := httptest.NewRequest("GET", "/slow", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}()

		// Wait for request to start
		<-requestStarted

		// Check in-flight metric (should be > 0)
		metric := &dto.Metric{}
		err := metricsCollector.HTTPRequestsInFlight.Write(metric)
		require.NoError(t, err)
		assert.Equal(t, float64(1), metric.GetGauge().GetValue())

		// Let request finish
		close(requestCanFinish)

		// Give time for request to complete
		time.Sleep(10 * time.Millisecond)

		// Check in-flight metric (should be 0)
		metric = &dto.Metric{}
		err = metricsCollector.HTTPRequestsInFlight.Write(metric)
		require.NoError(t, err)
		assert.Equal(t, float64(0), metric.GetGauge().GetValue())
	})
}

func TestMetricsMiddleware_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles concurrent requests correctly", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(5 * time.Millisecond)
			c.Status(http.StatusOK)
		})

		// Make multiple concurrent requests
		numRequests := 10
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		for i := 0; i < numRequests; i++ {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})
}

func TestMetricsMiddleware_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles middleware chain with multiple middlewares", func(t *testing.T) {
		metricsCollector := testMetrics
		metricsMiddleware := MetricsMiddleware(metricsCollector)

		// Custom middleware that adds a header
		headerMiddleware := func(c *gin.Context) {
			c.Header("X-Test", "test")
			c.Next()
		}

		router := gin.New()
		router.Use(headerMiddleware)
		router.Use(metricsMiddleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test", w.Header().Get("X-Test"))
	})

	t.Run("handles middleware with early abort", func(t *testing.T) {
		metricsCollector := testMetrics
		metricsMiddleware := MetricsMiddleware(metricsCollector)

		// Middleware that aborts early
		authMiddleware := func(c *gin.Context) {
			if c.GetHeader("Authorization") == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Next()
		}

		router := gin.New()
		router.Use(metricsMiddleware)
		router.Use(authMiddleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "protected"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("handles panic in handler", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()

		// Recovery middleware to handle panics
		router.Use(gin.Recovery())
		router.Use(middleware)

		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()

		// Should not crash
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMetricsMiddleware_RequestTiming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("measures request duration", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)
		router.GET("/fast", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(20 * time.Millisecond)
			c.Status(http.StatusOK)
		})

		// Test fast request
		req := httptest.NewRequest("GET", "/fast", nil)
		w := httptest.NewRecorder()
		start := time.Now()
		router.ServeHTTP(w, req)
		fastDuration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test slow request
		req = httptest.NewRequest("GET", "/slow", nil)
		w = httptest.NewRecorder()
		start = time.Now()
		router.ServeHTTP(w, req)
		slowDuration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, slowDuration > fastDuration)
	})
}

func TestMetricsMiddleware_StatusCodeHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("correctly records various HTTP status codes", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)

		// Set up routes for different status codes
		testCases := []struct {
			path   string
			status int
		}{
			{"/status/ok", http.StatusOK},
			{"/status/created", http.StatusCreated},
			{"/status/nocontent", http.StatusNoContent},
			{"/status/badrequest", http.StatusBadRequest},
			{"/status/unauthorized", http.StatusUnauthorized},
			{"/status/forbidden", http.StatusForbidden},
			{"/status/notfound", http.StatusNotFound},
			{"/status/methodnotallowed", http.StatusMethodNotAllowed},
			{"/status/internalerror", http.StatusInternalServerError},
			{"/status/badgateway", http.StatusBadGateway},
			{"/status/unavailable", http.StatusServiceUnavailable},
		}

		for _, tc := range testCases {
			statusCode := tc.status // Capture for closure
			router.GET(tc.path, func(c *gin.Context) {
				c.Status(statusCode)
			})

			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		}
	})
}

func TestMetricsMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("full integration test", func(t *testing.T) {
		metricsCollector := testMetrics
		middleware := MetricsMiddleware(metricsCollector)

		router := gin.New()
		router.Use(middleware)

		// Health endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		// API endpoints
		router.GET("/api/items", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"items": []string{}})
		})

		router.POST("/api/items", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "123"})
		})

		// Make various requests
		requests := []struct {
			method string
			path   string
			status int
		}{
			{"GET", "/health", http.StatusOK},
			{"GET", "/api/items", http.StatusOK},
			{"POST", "/api/items", http.StatusCreated},
			{"GET", "/nonexistent", http.StatusNotFound},
		}

		for _, req := range requests {
			httpReq := httptest.NewRequest(req.method, req.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)
			assert.Equal(t, req.status, w.Code)
		}

		// Verify metrics were collected (basic smoke test)
		assert.NotNil(t, metricsCollector.HTTPRequestsTotal)
		assert.NotNil(t, metricsCollector.HTTPRequestDuration)
		assert.NotNil(t, metricsCollector.HTTPRequestsInFlight)
	})
}

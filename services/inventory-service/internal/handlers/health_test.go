package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthHandler(t *testing.T) {
	logger := slog.Default()

	t.Run("create handler with all dependencies", func(t *testing.T) {
		handler := NewHealthHandler(logger, nil, nil)

		assert.NotNil(t, handler)
		assert.Equal(t, logger, handler.logger)
		assert.Nil(t, handler.postgres)
		assert.Nil(t, handler.redis)
	})

	t.Run("create handler with nil logger", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, nil)

		assert.NotNil(t, handler)
		assert.Nil(t, handler.logger)
		assert.Nil(t, handler.postgres)
		assert.Nil(t, handler.redis)
	})
}

func TestHealthHandler_Health_NilDependencies(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("both databases nil", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/health", nil)
		c.Request = req

		handler.Health(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "not_configured", response.Dependencies["postgresql"])
		assert.Equal(t, "not_configured", response.Dependencies["redis"])
	})

	t.Run("nil logger panics", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/health", nil)
		c.Request = req

		// Should panic with nil logger
		assert.Panics(t, func() {
			handler.Health(c)
		})
	})
}

func TestHealthResponse_Structure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := slog.Default()
	handler := NewHealthHandler(logger, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/health", nil)
	c.Request = req

	handler.Health(c)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	t.Run("required fields present", func(t *testing.T) {
		// Check all required fields are present
		assert.NotEmpty(t, response.Status)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotEmpty(t, response.Version)
		assert.NotNil(t, response.Dependencies)
	})

	t.Run("timestamp format", func(t *testing.T) {
		// Check timestamp format
		_, err = time.Parse(time.RFC3339, response.Timestamp)
		assert.NoError(t, err, "timestamp should be in RFC3339 format")
	})

	t.Run("dependencies structure", func(t *testing.T) {
		// Check dependencies structure
		assert.Contains(t, response.Dependencies, "postgresql")
		assert.Contains(t, response.Dependencies, "redis")
	})

	t.Run("version is set", func(t *testing.T) {
		// Check version is set
		assert.Equal(t, "1.0.0", response.Version)
	})
}

func TestHealthResponse_JSONSerialization(t *testing.T) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Dependencies: map[string]string{
			"postgresql": "healthy",
			"redis":      "healthy",
		},
	}

	t.Run("marshal and unmarshal", func(t *testing.T) {
		// Test JSON marshaling
		data, err := json.Marshal(response)
		require.NoError(t, err)

		// Test JSON unmarshaling
		var unmarshaled HealthResponse
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, response.Status, unmarshaled.Status)
		assert.Equal(t, response.Version, unmarshaled.Version)
		assert.Equal(t, response.Timestamp, unmarshaled.Timestamp)
		assert.Equal(t, response.Dependencies, unmarshaled.Dependencies)
	})

	t.Run("JSON field names", func(t *testing.T) {
		data, err := json.Marshal(response)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(data, &jsonMap)
		require.NoError(t, err)

		// Check JSON field names match struct tags
		assert.Contains(t, jsonMap, "status")
		assert.Contains(t, jsonMap, "timestamp")
		assert.Contains(t, jsonMap, "version")
		assert.Contains(t, jsonMap, "dependencies")
	})
}

func TestHealthHandler_HTTPStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 when healthy", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/health", nil)
		c.Request = req

		handler.Health(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response.Status)
	})

	t.Run("content type is JSON", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/health", nil)
		c.Request = req

		handler.Health(c)

		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})
}

func TestHealthHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("multiple concurrent requests", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		// Test concurrent access
		results := make(chan int, 10)
		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				req := httptest.NewRequest("GET", "/health", nil)
				c.Request = req

				handler.Health(c)
				results <- w.Code
			}()
		}

		// Collect results
		for i := 0; i < 10; i++ {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})

	t.Run("timestamp precision", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		// Make multiple requests quickly
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		req1 := httptest.NewRequest("GET", "/health", nil)
		c1.Request = req1

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		req2 := httptest.NewRequest("GET", "/health", nil)
		c2.Request = req2

		handler.Health(c1)
		handler.Health(c2)

		var response1, response2 HealthResponse
		err := json.Unmarshal(w1.Body.Bytes(), &response1)
		require.NoError(t, err)
		err = json.Unmarshal(w2.Body.Bytes(), &response2)
		require.NoError(t, err)

		// Timestamps should be different (or at least parseable)
		time1, err1 := time.Parse(time.RFC3339, response1.Timestamp)
		time2, err2 := time.Parse(time.RFC3339, response2.Timestamp)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.True(t, time2.After(time1) || time2.Equal(time1))
	})
}

func TestHealthHandler_ContextTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handler completes within reasonable time", func(t *testing.T) {
		logger := slog.Default()
		handler := NewHealthHandler(logger, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/health", nil)
		c.Request = req

		// Should complete quickly with nil dependencies
		start := time.Now()
		handler.Health(c)
		duration := time.Since(start)

		// Should complete in well under a second
		assert.Less(t, duration, 1*time.Second)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthResponse_FieldTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := slog.Default()
	handler := NewHealthHandler(logger, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/health", nil)
	c.Request = req

	handler.Health(c)

	// Parse as generic map to check field types
	var jsonMap map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jsonMap)
	require.NoError(t, err)

	t.Run("status is string", func(t *testing.T) {
		status, ok := jsonMap["status"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, status)
	})

	t.Run("timestamp is string", func(t *testing.T) {
		timestamp, ok := jsonMap["timestamp"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, timestamp)
	})

	t.Run("version is string", func(t *testing.T) {
		version, ok := jsonMap["version"].(string)
		assert.True(t, ok)
		assert.Equal(t, "1.0.0", version)
	})

	t.Run("dependencies is object", func(t *testing.T) {
		deps, ok := jsonMap["dependencies"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, deps)
		
		// Check dependency values are strings
		for key, value := range deps {
			assert.NotEmpty(t, key)
			_, isString := value.(string)
			assert.True(t, isString, "dependency value should be string")
		}
	})
}
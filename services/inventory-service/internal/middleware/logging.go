package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/internal/auth"
)

// LoggingMiddleware provides request logging
type LoggingMiddleware struct {
	logger *slog.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// LogRequests logs all HTTP requests with user context
func (m *LoggingMiddleware) LogRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Get user context if available
		var userID, telegramID interface{}
		if user, exists := auth.GetUserFromContext(c); exists {
			userID = user.UserID
			telegramID = user.TelegramID
		}

		// Log request with all relevant information
		m.logger.Info("HTTP request processed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"user_id", userID,
			"telegram_id", telegramID,
			"request_size", c.Request.ContentLength,
			"response_size", c.Writer.Size(),
		)

		// Log errors separately for better visibility
		if c.Writer.Status() >= 400 {
			m.logger.Error("HTTP request error",
				"method", c.Request.Method,
				"path", path,
				"status", c.Writer.Status(),
				"latency", latency,
				"user_id", userID,
				"error_details", c.Errors.String(),
			)
		}
	}
}

package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/inventory-service/pkg/logger"
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
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)

		// Add request ID to context for logger
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get request logger with context
		reqLogger := logger.WithContext(ctx, m.logger)

		// Log request start at debug level
		reqLogger.Debug("HTTP request started",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Get user context if available (from JWT middleware)
		var userID, telegramID interface{}
		if userIDValue, exists := c.Get("user_id"); exists {
			userID = userIDValue
			// Add user ID to context
			ctx = context.WithValue(ctx, logger.UserIDKey, userID)
			reqLogger = logger.WithContext(ctx, m.logger)
		}
		if telegramIDValue, exists := c.Get("telegram_id"); exists {
			telegramID = telegramIDValue
		}

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if c.Writer.Status() >= 500 {
			logLevel = slog.LevelError
		} else if c.Writer.Status() >= 400 {
			logLevel = slog.LevelWarn
		}

		// Log request with appropriate level
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.Float64("latency_ms", latency.Seconds()*1000),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Any("user_id", userID),
			slog.Any("telegram_id", telegramID),
			slog.Int64("request_size", c.Request.ContentLength),
			slog.Int("response_size", c.Writer.Size()),
		}

		// Add error details if present
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		reqLogger.LogAttrs(ctx, logLevel, "HTTP request completed", attrs...)
	}
}

package logger

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"
)

// ContextKey type for context values
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

// NewLogger creates a new structured logger with JSON output
func NewLogger(level string) *slog.Logger {
	// Convert level string to slog.Level
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create JSON handler with specified level
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attribute formatting
			if a.Key == slog.TimeKey {
				// Use RFC3339 format with milliseconds
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.RFC3339Nano))
			}
			if a.Key == slog.SourceKey {
				// Simplify source location
				source := a.Value.Any().(*slog.Source)
				source.File = shortFile(source.File)
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler).With(
		slog.String("service", "deck-game-service"),
		slog.String("version", "1.0.0"),
	)

	return logger
}

// WithContext returns a logger with context values
func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return nil
	}

	attrs := []slog.Attr{}

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add user ID if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	if len(attrs) > 0 {
		anyAttrs := make([]any, len(attrs))
		for i, attr := range attrs {
			anyAttrs[i] = attr
		}
		return logger.With(anyAttrs...)
	}

	return logger
}

// LogError logs an error with additional context
func LogError(logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	if logger == nil {
		return
	}

	allAttrs := append([]slog.Attr{
		slog.String("error", err.Error()),
		slog.String("error_type", typeOf(err)),
	}, attrs...)

	logger.LogAttrs(context.Background(), slog.LevelError, msg, allAttrs...)
}

// LogWithDuration logs a message with duration
func LogWithDuration(logger *slog.Logger, level slog.Level, msg string, duration time.Duration, attrs ...slog.Attr) {
	if logger == nil {
		return
	}

	allAttrs := append([]slog.Attr{
		slog.Duration("duration", duration),
		slog.Float64("duration_ms", duration.Seconds()*1000),
	}, attrs...)

	logger.LogAttrs(context.Background(), level, msg, allAttrs...)
}

// Helper functions

func shortFile(file string) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	return short
}

func typeOf(v interface{}) string {
	return strings.TrimPrefix(reflect.TypeOf(v).String(), "*")
}

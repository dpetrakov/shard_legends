package logger

import (
	"log/slog"
	"os"
	"strings"
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
		Level: logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}
package utils

import (
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger with English messages
func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// Use JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}

// NewDevelopmentLogger creates a logger with debug level for development
func NewDevelopmentLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	// Use text handler for better readability in development
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}

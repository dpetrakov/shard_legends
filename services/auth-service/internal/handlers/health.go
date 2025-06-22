package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/storage"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger       *slog.Logger
	userRepo     storage.UserRepository
	tokenStorage storage.TokenStorage
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *slog.Logger, userRepo storage.UserRepository, tokenStorage storage.TokenStorage) *HealthHandler {
	return &HealthHandler{
		logger:       logger,
		userRepo:     userRepo,
		tokenStorage: tokenStorage,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

// Health handles GET /health requests
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check database health
	dbStatus := "healthy"
	if err := h.userRepo.Health(ctx); err != nil {
		h.logger.Error("Database health check failed", "error", err)
		dbStatus = "unhealthy"
	}

	// Check Redis health
	redisStatus := "healthy"
	if h.tokenStorage != nil {
		if err := h.tokenStorage.Health(ctx); err != nil {
			h.logger.Error("Redis health check failed", "error", err)
			redisStatus = "unhealthy"
		}
	} else {
		redisStatus = "not_configured"
	}

	// Determine overall status
	overallStatus := "healthy"
	if dbStatus == "unhealthy" || redisStatus == "unhealthy" {
		overallStatus = "unhealthy"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Dependencies: map[string]string{
			"postgresql": dbStatus,
			"redis":      redisStatus,
			"jwt_keys":   "loaded", // JWT keys are loaded since we have JWT service running
		},
	}

	h.logger.Debug("Health check requested", 
		"status", response.Status,
		"postgresql", dbStatus,
		"redis", redisStatus)

	// Return appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/internal/database"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger   *slog.Logger
	postgres *database.PostgresDB
	redis    *database.RedisDB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *slog.Logger, postgres *database.PostgresDB, redis *database.RedisDB) *HealthHandler {
	return &HealthHandler{
		logger:   logger,
		postgres: postgres,
		redis:    redis,
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
	if h.postgres != nil {
		if err := h.postgres.Health(ctx); err != nil {
			h.logger.Error("Database health check failed", "error", err)
			dbStatus = "unhealthy"
		}
	} else {
		dbStatus = "not_configured"
	}

	// Check Redis health
	redisStatus := "healthy"
	if h.redis != nil {
		if err := h.redis.Health(ctx); err != nil {
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
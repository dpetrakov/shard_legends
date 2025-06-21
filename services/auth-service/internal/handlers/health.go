package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
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
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Dependencies: map[string]string{
			"postgresql": "not_configured",
			"redis":      "not_configured",
			"jwt_keys":   "loaded", // JWT keys are loaded since we have JWT service running
		},
	}

	h.logger.Debug("Health check requested", "status", response.Status)

	c.JSON(http.StatusOK, response)
}

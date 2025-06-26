package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"user-service/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// GetHealth returns the health status of the service
// GET /health
func (h *HealthHandler) GetHealth(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		Service:   "user-service",
	}

	h.logger.Info("Health check requested", map[string]interface{}{
		"status": response.Status,
		"service": response.Service,
	})

	c.JSON(http.StatusOK, response)
}
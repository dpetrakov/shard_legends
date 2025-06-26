package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shard-legends/production-service/internal/database"
)

type HealthHandler struct {
	db    *database.DB
	redis *database.RedisClient
}

func NewHealthHandler(db *database.DB, redis *database.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

type HealthResponse struct {
	Status       string            `json:"status"`
	Services     map[string]string `json:"services"`
	DatabasePool DatabasePoolStats `json:"database_pool"`
}

type DatabasePoolStats struct {
	TotalConns    int32 `json:"total_connections"`
	IdleConns     int32 `json:"idle_connections"`
	AcquiredConns int32 `json:"acquired_connections"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := HealthResponse{
		Status:   "ok",
		Services: make(map[string]string),
	}

	// Check database
	if err := h.db.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "down: " + err.Error()
	} else {
		response.Services["database"] = "ok"
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["redis"] = "down: " + err.Error()
	} else {
		response.Services["redis"] = "ok"
	}

	// Get database pool stats
	stats := h.db.Stats()
	response.DatabasePool = DatabasePoolStats{
		TotalConns:    stats.TotalConns(),
		IdleConns:     stats.IdleConns(),
		AcquiredConns: stats.AcquiredConns(),
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	if response.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if all services are ready
	if err := h.db.Health(ctx); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}

	if err := h.redis.Health(ctx); err != nil {
		http.Error(w, "Redis not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
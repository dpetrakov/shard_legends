package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Mode      string    `json:"mode"`
}

// NewHealthHandler creates a health check handler
func NewHealthHandler(mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
			Service:   "telegram-bot-service",
			Mode:      mode,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
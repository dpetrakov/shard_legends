package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Service   string            `json:"service"`
	Uptime    string            `json:"uptime,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

var startTime = time.Now()

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return simple pong message
	response := Response{
		Message: "pong",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Calculate uptime
	uptime := time.Since(startTime).Round(time.Second).String()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Service:   "ping-service",
		Uptime:    uptime,
		Details: map[string]string{
			"description": "Simple ping service for testing",
			"environment": "development",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/health", healthHandler)

	port := "8080"
	fmt.Printf("Ping service starting on port %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /ping   - Returns pong message\n")
	fmt.Printf("  GET  /health - Health check endpoint\n")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

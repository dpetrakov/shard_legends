package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// Prometheus метрики
var (
	// Счетчик HTTP запросов
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Время обработки запросов
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Бизнес-метрики: количество ping запросов
	pingRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ping_requests_total",
			Help: "Total number of ping requests",
		},
	)

	// Uptime сервиса
	serviceUptimeSeconds = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(pingRequestsTotal)
	prometheus.MustRegister(serviceUptimeSeconds)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Only accept POST requests
	if r.Method != http.MethodPost {
		// Метрика для неправильных методов
		httpRequestsTotal.WithLabelValues(r.Method, "/ping", "405").Inc()
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Увеличиваем счетчик ping запросов
	pingRequestsTotal.Inc()

	// Return simple pong message
	response := Response{
		Message: "pong",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		// Метрика для ошибок
		httpRequestsTotal.WithLabelValues(r.Method, "/ping", "500").Inc()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Метрики для успешных запросов
	duration := time.Since(start).Seconds()
	httpRequestsTotal.WithLabelValues(r.Method, "/ping", "200").Inc()
	httpRequestDuration.WithLabelValues(r.Method, "/ping").Observe(duration)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Only accept GET requests
	if r.Method != http.MethodGet {
		httpRequestsTotal.WithLabelValues(r.Method, "/health", "405").Inc()
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
		httpRequestsTotal.WithLabelValues(r.Method, "/health", "500").Inc()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Метрики для успешных health checks
	duration := time.Since(start).Seconds()
	httpRequestsTotal.WithLabelValues(r.Method, "/health", "200").Inc()
	httpRequestDuration.WithLabelValues(r.Method, "/health").Observe(duration)
}

func main() {
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/health", healthHandler)
	http.Handle("/metrics", promhttp.Handler())

	port := "8080"
	fmt.Printf("Ping service starting on port %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /ping   - Returns pong message\n")
	fmt.Printf("  GET  /health - Health check endpoint\n")
	fmt.Printf("  GET  /metrics - Prometheus metrics\n")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

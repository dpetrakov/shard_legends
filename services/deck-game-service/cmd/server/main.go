package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shard-legends/deck-game-service/internal/config"
	"github.com/shard-legends/deck-game-service/internal/database"
	"github.com/shard-legends/deck-game-service/pkg/logger"
	"github.com/shard-legends/deck-game-service/pkg/metrics"
)

func main() {
	// Load .env file if it exists (for development)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel)
	log.Info("Starting deck-game-service", "config", cfg.String())

	// Initialize metrics
	metricsCollector := metrics.New()
	metricsCollector.Initialize()
	defer metricsCollector.Shutdown()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize PostgreSQL
	postgres, err := database.NewPostgresDB(cfg.DatabaseURL, cfg.DatabaseMaxConns, log, metricsCollector)
	if err != nil {
		log.Error("Failed to initialize PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer postgres.Close()

	// Initialize Redis (for JWT auth only)
	redis, err := database.NewRedisDB(cfg.RedisAuthURL, cfg.RedisMaxConns, log, metricsCollector)
	if err != nil {
		log.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close()

	// Create Gin router for public API
	publicRouter := gin.New()

	// Add global middleware for public API
	publicRouter.Use(gin.Recovery())
	publicRouter.Use(requestLogger(log))
	publicRouter.Use(metricsMiddleware(metricsCollector))

	// Create Gin router for internal API
	internalRouter := gin.New()
	internalRouter.Use(gin.Recovery())
	internalRouter.Use(requestLogger(log))

	// Initialize health handler
	healthHandler := NewHealthHandler(log, postgres, redis)

	// Register internal routes
	internalRouter.GET("/health", healthHandler.Health)

	// Metrics endpoint for Prometheus scraping
	internalRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// TODO: Register public API routes here
	// This will be implemented in later tasks (D-Deck-004, D-Deck-005)

	// Create public HTTP server
	publicServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort),
		Handler:      publicRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create internal HTTP server
	internalServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.InternalServicePort),
		Handler:      internalRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start health monitoring goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			// Update dependency health metrics
			postgresHealthy := postgres.Health(ctx) == nil
			redisHealthy := redis.Health(ctx) == nil
			metricsCollector.UpdateDependencyHealth("postgres", postgresHealthy)
			metricsCollector.UpdateDependencyHealth("redis", redisHealthy)

			// Update connection metrics
			postgres.UpdateMetrics()
			redis.UpdateMetrics()

			cancel()
		}
	}()

	// Start public server in a goroutine
	go func() {
		log.Info("Public server starting", "address", publicServer.Addr)
		if err := publicServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Public server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Start internal server in a goroutine
	go func() {
		log.Info("Internal server starting", "address", internalServer.Addr)
		if err := internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Internal server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown of both servers
	go func() {
		if err := publicServer.Shutdown(ctx); err != nil {
			log.Error("Public server forced to shutdown", "error", err)
		}
	}()

	if err := internalServer.Shutdown(ctx); err != nil {
		log.Error("Internal server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server exited")
}

// requestLogger is a middleware that logs HTTP requests
func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// metricsMiddleware records HTTP metrics
func metricsMiddleware(metrics *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		metrics.HTTPRequestsInFlight.Inc()

		// Process request
		c.Next()

		// Decrement in-flight requests
		metrics.HTTPRequestsInFlight.Dec()

		// Record metrics
		duration := time.Since(start)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		metrics.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			fmt.Sprintf("%d", c.Writer.Status()),
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(duration.Seconds())
	}
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger   *slog.Logger
	postgres *database.PostgresDB
	redis    *database.RedisDB
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(logger *slog.Logger, postgres *database.PostgresDB, redis *database.RedisDB) *HealthHandler {
	return &HealthHandler{
		logger:   logger,
		postgres: postgres,
		redis:    redis,
	}
}

// Health handles GET /health requests
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status := "healthy"
	details := map[string]string{}

	// Check PostgreSQL
	if err := h.postgres.Health(ctx); err != nil {
		status = "unhealthy"
		details["postgres"] = "unhealthy: " + err.Error()
	} else {
		details["postgres"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		status = "unhealthy"
		details["redis"] = "unhealthy: " + err.Error()
	} else {
		details["redis"] = "healthy"
	}

	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":  status,
		"service": "deck-game-service",
		"details": details,
	})
}

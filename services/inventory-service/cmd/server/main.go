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
	"github.com/shard-legends/inventory-service/internal/config"
	"github.com/shard-legends/inventory-service/internal/database"
	"github.com/shard-legends/inventory-service/internal/handlers"
	"github.com/shard-legends/inventory-service/internal/middleware"
	"github.com/shard-legends/inventory-service/pkg/logger"
	"github.com/shard-legends/inventory-service/pkg/metrics"
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
	log.Info("Starting inventory-service", "config", cfg.String())

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

	// Initialize Redis
	redis, err := database.NewRedisDB(cfg.RedisURL, cfg.RedisMaxConns, log, metricsCollector)
	if err != nil {
		log.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close()

	// Create Gin router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger(log))
	router.Use(middleware.MetricsMiddleware(metricsCollector))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(log, postgres, redis)

	// Register routes
	router.GET("/health", healthHandler.Health)

	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start health monitoring goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				
				// Update dependency health metrics
				postgresHealthy := postgres.Health(ctx) == nil
				redisHealthy := redis.Health(ctx) == nil
				metricsCollector.UpdateDependencyHealth("postgres", postgresHealthy)
				metricsCollector.UpdateDependencyHealth("redis", redisHealthy)

				cancel()
			}
		}
	}()

	// Start server in a goroutine
	go func() {
		log.Info("Server starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", "error", err)
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

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
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
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
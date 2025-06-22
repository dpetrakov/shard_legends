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
	"github.com/shard-legends/auth-service/internal/config"
	"github.com/shard-legends/auth-service/internal/handlers"
	"github.com/shard-legends/auth-service/internal/metrics"
	"github.com/shard-legends/auth-service/internal/middleware"
	"github.com/shard-legends/auth-service/internal/services"
	"github.com/shard-legends/auth-service/internal/storage"
	"github.com/shard-legends/auth-service/pkg/utils"
)

func main() {
	// Load .env file if it exists (for development)
	_ = godotenv.Load()

	// Initialize logger
	logger := utils.NewLogger()

	logger.Info("Starting auth-service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded", "config", cfg.String())

	// Initialize metrics
	metricsCollector := metrics.New()
	metricsCollector.Initialize()
	
	// Defer shutdown of metrics
	defer metricsCollector.Shutdown()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))
	router.Use(middleware.CORS()) // Add CORS support
	router.Use(middleware.MetricsMiddleware(metricsCollector)) // Add metrics collection

	// Create rate limiter for auth endpoint (10 requests per minute)
	authRateLimiter := middleware.NewRateLimiter(10, logger, metricsCollector)
	defer authRateLimiter.Close()

	// Initialize services
	telegramValidator := services.NewTelegramValidator(cfg.TelegramBotTokens, logger)
	
	// Initialize JWT service
	keyPaths := services.KeyPaths{
		PrivateKeyPath: cfg.JWTPrivateKeyPath,
		PublicKeyPath:  cfg.JWTPublicKeyPath,
	}
	jwtService, err := services.NewJWTService(keyPaths, cfg.JWTIssuer, cfg.JWTExpiryHours, logger)
	if err != nil {
		logger.Error("Failed to initialize JWT service", "error", err)
		os.Exit(1)
	}

	// Initialize PostgreSQL storage
	userRepo, err := storage.NewPostgresStorage(cfg.DatabaseURL, cfg.DatabaseMaxConns, logger, metricsCollector)
	if err != nil {
		logger.Error("Failed to initialize PostgreSQL storage", "error", err)
		os.Exit(1)
	}
	defer userRepo.Close()

	// Initialize Redis token storage
	tokenStorage, err := storage.NewRedisTokenStorage(cfg.RedisURL, cfg.RedisMaxConns, logger, metricsCollector)
	if err != nil {
		logger.Error("Failed to initialize Redis token storage", "error", err)
		os.Exit(1)
	}
	defer tokenStorage.Close()

	// Initialize middleware
	jwtPublicKeyMiddleware := middleware.NewJWTPublicKeyMiddleware(jwtService)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(logger, userRepo, tokenStorage)
	authHandler := handlers.NewAuthHandler(logger, telegramValidator, jwtService, userRepo, tokenStorage, metricsCollector)
	adminHandler := handlers.NewAdminHandler(logger, tokenStorage)

	// Register routes
	router.GET("/health", healthHandler.Health)
	router.POST("/auth", authRateLimiter.Limit(), authHandler.Auth) // Apply rate limiting to auth endpoint
	
	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// JWT public key endpoints for other services
	router.GET("/jwks", jwtPublicKeyMiddleware.PublicKeyHandler())
	router.GET("/public-key.pem", jwtPublicKeyMiddleware.PublicKeyPEMHandler())
	
	// Admin endpoints for token management (should be protected in production)
	adminGroup := router.Group("/admin")
	{
		adminGroup.GET("/tokens/stats", adminHandler.GetTokenStats)
		adminGroup.GET("/tokens/user/:userId", adminHandler.GetUserTokens)
		adminGroup.DELETE("/tokens/user/:userId", adminHandler.RevokeUserTokens)
		adminGroup.DELETE("/tokens/:jti", adminHandler.RevokeToken)
		adminGroup.POST("/tokens/cleanup", adminHandler.CleanupExpiredTokens)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start token cleanup and metrics update goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.TokenCleanupIntervalHours) * time.Hour)
		metricsTicker := time.NewTicker(30 * time.Second) // Update metrics every 30 seconds
		defer ticker.Stop()
		defer metricsTicker.Stop()

		// Update metrics on startup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		activeCount, err := tokenStorage.GetActiveTokenCount(ctx)
		cancel()
		if err == nil {
			metricsCollector.UpdateActiveTokensCount(float64(activeCount))
		}

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TokenCleanupTimeoutMins)*time.Minute)
				cleanedCount, err := tokenStorage.CleanupExpiredTokens(ctx)
				cancel()

				if err != nil {
					logger.Error("Token cleanup failed", "error", err)
				} else {
					logger.Info("Token cleanup completed", "cleaned_tokens", cleanedCount)
				}

				// Update active tokens count after cleanup
				ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
				activeCount, err := tokenStorage.GetActiveTokenCount(ctx)
				cancel()
				if err == nil {
					metricsCollector.UpdateActiveTokensCount(float64(activeCount))
				}

			case <-metricsTicker.C:
				// Periodically update active tokens count and dependency health
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				
				// Update active tokens count
				activeCount, err := tokenStorage.GetActiveTokenCount(ctx)
				if err == nil {
					metricsCollector.UpdateActiveTokensCount(float64(activeCount))
				}

				// Update dependency health
				postgresHealth := userRepo.Health(ctx) == nil
				redisHealth := tokenStorage.Health(ctx) == nil
				metricsCollector.UpdateDependencyHealth("postgres", postgresHealth)
				metricsCollector.UpdateDependencyHealth("redis", redisHealth)

				cancel()
			}
		}
	}()

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
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

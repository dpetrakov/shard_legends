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
	"github.com/shard-legends/auth-service/internal/config"
	"github.com/shard-legends/auth-service/internal/handlers"
	"github.com/shard-legends/auth-service/internal/middleware"
	"github.com/shard-legends/auth-service/internal/services"
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

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))

	// Initialize services
	telegramValidator := services.NewTelegramValidator(cfg.TelegramBotToken, logger)
	
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

	// Initialize middleware
	jwtPublicKeyMiddleware := middleware.NewJWTPublicKeyMiddleware(jwtService)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(logger)
	authHandler := handlers.NewAuthHandler(logger, telegramValidator)

	// Register routes
	router.GET("/health", healthHandler.Health)
	router.POST("/auth", authHandler.Auth)
	
	// JWT public key endpoints for other services
	router.GET("/jwks", jwtPublicKeyMiddleware.PublicKeyHandler())
	router.GET("/public-key.pem", jwtPublicKeyMiddleware.PublicKeyPEMHandler())

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

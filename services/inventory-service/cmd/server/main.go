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
	"github.com/shard-legends/inventory-service/internal/service"
	"github.com/shard-legends/inventory-service/internal/storage"
	"github.com/shard-legends/inventory-service/pkg/jwt"
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
	redis, err := database.NewRedisDB(cfg.RedisURL, cfg.RedisAuthURL, cfg.RedisMaxConns, log, metricsCollector)
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
	publicRouter.Use(middleware.MetricsMiddleware(metricsCollector))

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

	// Initialize API with JWT authentication for both routers
	if err := setupAPIWithJWT(cfg, publicRouter, internalRouter, postgres, redis, log, metricsCollector); err != nil {
		log.Error("Failed to setup API with JWT", "error", err)
		os.Exit(1)
	}

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
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// setupAPIWithJWT initializes the API with JWT authentication and real services
func setupAPIWithJWT(cfg *config.Config, publicRouter *gin.Engine, internalRouter *gin.Engine, postgres *database.PostgresDB, redis *database.RedisDB, logger *slog.Logger, metricsCollector *metrics.Metrics) error {
	// Load JWT public key from auth-service using configured URL
	publicKey, err := jwt.LoadPublicKeyFromAuthService(cfg.AuthServiceURL)
	if err != nil {
		logger.Error("Failed to load JWT public key from auth-service - this is required for authentication", "error", err)
		return fmt.Errorf("JWT public key is required: %w", err)
	}
	logger.Info("Successfully loaded JWT public key from auth-service")

	// Create service implementations for interface compatibility
	cacheImpl := service.NewRedisCache(redis)
	var metricsImpl service.MetricsInterface = metrics.NewServiceMetrics(metricsCollector)

	// Initialize storage layer
	inventoryRepo := storage.NewInventoryStorage(postgres.Pool(), redis, logger, metricsCollector)
	classifierRepo := storage.NewClassifierStorage(postgres.Pool(), redis, logger, metricsCollector)
	itemRepo := storage.NewItemStorage(postgres.Pool(), logger, metricsImpl, cacheImpl)

	// Create repository interfaces wrapper
	repoInterfaces := &service.RepositoryInterfaces{
		Classifier: classifierRepo,
		Item:       itemRepo,
		Inventory:  inventoryRepo,
	}

	// Create service dependencies
	serviceDeps := &service.ServiceDependencies{
		Repositories: repoInterfaces,
		Cache:        cacheImpl,
		Metrics:      metricsImpl,
	}

	// Initialize additional components and add them to dependencies
	serviceDeps.CodeConverter = service.NewCodeConverter(serviceDeps)
	serviceDeps.BalanceChecker = service.NewBalanceChecker(serviceDeps)

	// Initialize service layer
	inventoryService := service.NewInventoryService(serviceDeps)
	classifierService := service.NewClassifierService(serviceDeps)

	// Initialize JWT middleware (required for public endpoints only)
	jwtMiddleware := middleware.NewJWTAuthMiddleware(publicKey, redis, logger)

	// Initialize handlers
	inventoryHandler := handlers.NewInventoryHandler(inventoryService, classifierService, logger)

	// Setup public API routes
	publicAPI := publicRouter.Group("/api/inventory")

	// Public endpoints (require JWT authentication)
	public := publicAPI.Group("")
	public.Use(jwtMiddleware.AuthenticateJWT())
	{
		public.GET("", inventoryHandler.GetUserInventory)
		public.GET("/items", inventoryHandler.GetUserInventory)         // alias for compatibility
		public.POST("/items/details", inventoryHandler.GetItemsDetails) // localized item details
	}


	// Setup internal API routes
	internalAPI := internalRouter.Group("/api/inventory")

	// Internal endpoints (no authentication required - isolated by network)
	internal := internalAPI.Group("")
	{
		internal.POST("/add-items", inventoryHandler.AddItems)
		internal.POST("/reserve", inventoryHandler.ReserveItems)
		internal.POST("/return-reserve", inventoryHandler.ReturnReservedItems)
		internal.POST("/consume-reserve", inventoryHandler.ConsumeReservedItems)
		internal.GET("/reservation/:operationID", inventoryHandler.GetReservationStatus)
	}

	// Administrative endpoints on internal API (no authentication required - isolated by network)
	admin := internalAPI.Group("/admin")
	{
		admin.POST("/adjust", inventoryHandler.AdjustInventory)
	}

	return nil
}

// Simple health handler
type HealthHandler struct {
	logger   *slog.Logger
	postgres *database.PostgresDB
	redis    *database.RedisDB
}

func NewHealthHandler(logger *slog.Logger, postgres *database.PostgresDB, redis *database.RedisDB) *HealthHandler {
	return &HealthHandler{
		logger:   logger,
		postgres: postgres,
		redis:    redis,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	dependencies := gin.H{}

	// Check PostgreSQL
	if err := h.postgres.Health(ctx); err != nil {
		dependencies["postgresql"] = gin.H{"status": "unhealthy", "error": err.Error()}
		status["status"] = "degraded"
	} else {
		dependencies["postgresql"] = gin.H{"status": "healthy"}
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		dependencies["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
		status["status"] = "degraded"
	} else {
		dependencies["redis"] = gin.H{"status": "healthy"}
	}

	status["dependencies"] = dependencies

	if status["status"] == "degraded" {
		c.JSON(503, status)
	} else {
		c.JSON(200, status)
	}
}

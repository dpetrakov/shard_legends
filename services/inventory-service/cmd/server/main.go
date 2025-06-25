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
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shard-legends/inventory-service/internal/config"
	"github.com/shard-legends/inventory-service/internal/database"
	"github.com/shard-legends/inventory-service/internal/middleware"
	"github.com/shard-legends/inventory-service/internal/models"
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

	// Initialize health handler
	healthHandler := NewHealthHandler(log, postgres, redis)

	// Register health check routes
	router.GET("/health", healthHandler.Health)

	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize API with JWT authentication
	if err := setupAPIWithJWT(router, cfg, postgres, redis, log, metricsCollector); err != nil {
		log.Error("Failed to setup API with JWT", "error", err)
		os.Exit(1)
	}

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

// setupAPIWithJWT initializes the API with JWT authentication and real services
func setupAPIWithJWT(router *gin.Engine, cfg *config.Config, postgres *database.PostgresDB, redis *database.RedisDB, logger *slog.Logger, metricsCollector *metrics.Metrics) error {
	// Load JWT public key directly from auth-service container
	publicKey, err := jwt.LoadPublicKeyFromAuthService("http://auth-service:8080")
	if err != nil {
		logger.Warn("Failed to load JWT public key from auth-service container, JWT auth will be disabled", "error", err)
		publicKey = nil
	} else {
		logger.Info("Successfully loaded JWT public key from auth-service container")
	}

	// Initialize storage layer
	inventoryRepo := storage.NewInventoryStorage(postgres.Pool(), redis.Client(), logger, metricsCollector)
	classifierRepo := storage.NewClassifierStorage(postgres.Pool(), redis.Client(), logger, metricsCollector)
	itemRepo := storage.NewItemStorage(postgres.Pool(), logger, metricsCollector)

	// Create service implementations for interface compatibility
	cacheImpl := service.NewRedisCache(redis.Client())
	var metricsImpl service.MetricsInterface = metrics.NewServiceMetrics(metricsCollector)

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

	// Initialize service layer
	inventoryService := service.NewInventoryService(serviceDeps)
	classifierService := service.NewClassifierService(serviceDeps)

	// Initialize JWT middleware
	var jwtMiddleware *middleware.JWTAuthMiddleware
	if publicKey != nil {
		jwtMiddleware = middleware.NewJWTAuthMiddleware(publicKey, redis.Client(), logger)
	}

	// Initialize handlers
	inventoryHandler := NewInventoryHandler(inventoryService, classifierService, logger)

	// Setup API routes
	api := router.Group("/api/inventory")

	// Public endpoint (requires JWT authentication)
	if jwtMiddleware != nil {
		public := api.Group("")
		public.Use(jwtMiddleware.AuthenticateJWT())
		{
			public.GET("", inventoryHandler.GetUserInventory)
		}
	} else {
		// Fallback without JWT for development
		api.GET("", inventoryHandler.GetUserInventoryNoAuth)
	}

	// Internal endpoints (no authentication, used by other microservices)
	internal := api.Group("")
	{
		internal.POST("/add-items", inventoryHandler.AddItems)
		internal.POST("/adjust", inventoryHandler.AdjustInventory)
	}

	return nil
}

// Simple inventory handler implementation
type InventoryHandler struct {
	inventoryService  service.InventoryService
	classifierService service.ClassifierService
	logger            *slog.Logger
	validator         *validator.Validate
}

func NewInventoryHandler(inventoryService service.InventoryService, classifierService service.ClassifierService, logger *slog.Logger) *InventoryHandler {
	return &InventoryHandler{
		inventoryService:  inventoryService,
		classifierService: classifierService,
		logger:            logger,
		validator:         validator.New(),
	}
}

func (h *InventoryHandler) GetUserInventory(c *gin.Context) {
	h.logger.Info("GetUserInventory with JWT called")

	// Extract user ID from JWT token (middleware should set it)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in token",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get section from query params (default to "main")
	sectionParam := c.DefaultQuery("section", "main")

	// Get section UUID from classifier
	sectionMapping, err := h.classifierService.GetClassifierMapping(c.Request.Context(), "inventory_section")
	if err != nil {
		h.logger.Error("Failed to get section mapping", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve section",
		})
		return
	}

	sectionID, exists := sectionMapping[sectionParam]
	if !exists {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_section",
			Message: "Invalid section specified",
		})
		return
	}

	// Get user inventory
	inventory, err := h.inventoryService.GetUserInventory(c.Request.Context(), userID, sectionID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve inventory",
		})
		return
	}

	c.JSON(200, models.InventoryResponse{
		Items: func() []models.InventoryItemResponse {
			result := make([]models.InventoryItemResponse, len(inventory))
			for i, item := range inventory {
				result[i] = *item
			}
			return result
		}(),
	})
}

func (h *InventoryHandler) GetUserInventoryNoAuth(c *gin.Context) {
	h.logger.Info("GetUserInventory without JWT called")

	// For development only - user_id should be provided as query param
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(400, models.ErrorResponse{
			Error:   "missing_user_id",
			Message: "user_id query parameter is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get section from query params (default to "main")
	sectionParam := c.DefaultQuery("section", "main")

	// Get section UUID from classifier
	sectionMapping, err := h.classifierService.GetClassifierMapping(c.Request.Context(), "inventory_section")
	if err != nil {
		h.logger.Error("Failed to get section mapping", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve section",
		})
		return
	}

	sectionID, exists := sectionMapping[sectionParam]
	if !exists {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_section",
			Message: "Invalid section specified",
		})
		return
	}

	// Get user inventory
	inventory, err := h.inventoryService.GetUserInventory(c.Request.Context(), userID, sectionID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve inventory",
		})
		return
	}

	c.JSON(200, models.InventoryResponse{
		Items: func() []models.InventoryItemResponse {
			result := make([]models.InventoryItemResponse, len(inventory))
			for i, item := range inventory {
				result[i] = *item
			}
			return result
		}(),
	})
}

func (h *InventoryHandler) AddItems(c *gin.Context) {
	h.logger.Info("AddItems called")

	var req models.AddItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	operationIDs, err := h.inventoryService.AddItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add items", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to add items",
		})
		return
	}

	c.JSON(200, models.OperationResponse{
		Success:      true,
		OperationIDs: operationIDs,
	})
}

func (h *InventoryHandler) AdjustInventory(c *gin.Context) {
	h.logger.Info("AdjustInventory called")

	var req models.AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(400, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	response, err := h.inventoryService.AdjustInventory(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to adjust inventory", "error", err)
		c.JSON(500, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to adjust inventory",
		})
		return
	}

	c.JSON(200, response)
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

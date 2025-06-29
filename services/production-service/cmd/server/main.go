package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shard-legends/production-service/internal/adapters"
	"github.com/shard-legends/production-service/internal/config"
	"github.com/shard-legends/production-service/internal/database"
	"github.com/shard-legends/production-service/internal/handlers"
	"github.com/shard-legends/production-service/internal/handlers/public"
	customMiddleware "github.com/shard-legends/production-service/internal/middleware"
	"github.com/shard-legends/production-service/internal/service"
	"github.com/shard-legends/production-service/internal/storage"
	"github.com/shard-legends/production-service/pkg/jwt"
	"github.com/shard-legends/production-service/pkg/logger"
	"github.com/shard-legends/production-service/pkg/metrics"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Set service start time for metrics
	startTime := time.Now()
	go func() {
		for {
			metrics.ServiceUptime.Set(time.Since(startTime).Seconds())
			time.Sleep(cfg.Metrics.UpdateInterval)
		}
	}()

	// Set service info
	metrics.ServiceInfo.WithLabelValues("1.0.0", time.Now().Format(time.RFC3339)).Set(1)

	// Initialize database
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redis, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redis.Close()

	// Initialize JWT validator
	jwtValidator := jwt.NewValidator(cfg.Auth.PublicKeyURL, redis, cfg.Auth.CacheTTL)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.JWTValidatorClient)
	defer cancel()

	if err := jwtValidator.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize JWT validator", zap.Error(err))
	}

	// Refresh JWT public key periodically
	go func() {
		ticker := time.NewTicker(cfg.Auth.RefreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			if err := jwtValidator.RefreshPublicKey(ctx); err != nil {
				logger.Error("Failed to refresh JWT public key", zap.Error(err))
			}
		}
	}()

	// Initialize repository dependencies
	dbAdapter := adapters.NewDatabaseAdapter(db)
	cacheAdapter := adapters.NewCacheAdapter(redis)
	metricsAdapter := adapters.NewMetricsAdapter()

	repositoryDeps := &storage.RepositoryDependencies{
		DB:               dbAdapter,
		Cache:            cacheAdapter,
		MetricsCollector: metricsAdapter,
	}

	// Initialize repository
	repository := storage.NewRepository(repositoryDeps)

	// Initialize service dependencies
	serviceDeps := &service.ServiceDependencies{
		Repository: repository,
		Cache:      cacheAdapter,
		Metrics:    metricsAdapter,
	}

	// Initialize service layer
	serviceLayer := service.NewService(serviceDeps)

	// Initialize external service clients with configurable timeouts
	inventoryClient := service.NewHTTPInventoryClientWithTimeout(cfg.ExternalServices.InventoryService.BaseURL, cfg.ExternalServices.InventoryService.Timeout, logger.Get())
	userClient := service.NewHTTPUserClientWithTimeout(cfg.ExternalServices.UserService.BaseURL, cfg.ExternalServices.UserService.Timeout, logger.Get())

	// Initialize task service
	taskService := service.NewTaskService(
		repository.Task,
		repository.Recipe,
		repository.Classifier,
		serviceLayer.CodeConverter,
		inventoryClient,
		userClient,
		logger.Get(),
	)

	// Initialize cleanup service for orphaned draft tasks
	cleanupConfig := service.CleanupConfig{
		OrphanedTaskTimeout: cfg.Cleanup.OrphanedTaskTimeout,
		CleanupInterval:     cfg.Cleanup.CleanupInterval,
	}
	cleanupService := service.NewCleanupService(
		repository.Task,
		inventoryClient,
		logger.Get(),
		cleanupConfig,
	)

	// Start cleanup service in background
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	go func() {
		cleanupService.Start(cleanupCtx)
	}()

	// Initialize handlers
	handlerDeps := &handlers.HandlerDependencies{
		Service: serviceLayer,
		DB:      db,
		Redis:   redis,
		Logger:  logger.Get(),
	}
	allHandlers := handlers.NewHandlers(handlerDeps)

	// Initialize additional handlers
	factoryHandler := public.NewFactoryHandler(taskService, logger.Get())

	// Setup public router
	publicRouter := chi.NewRouter()

	// Public router middleware
	publicRouter.Use(middleware.RequestID)
	publicRouter.Use(middleware.RealIP)
	publicRouter.Use(customMiddleware.Recovery())
	publicRouter.Use(customMiddleware.Logging())
	publicRouter.Use(customMiddleware.Metrics())
	publicRouter.Use(middleware.Timeout(cfg.Timeouts.HTTPMiddleware))

	// CORS for public endpoints
	publicRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Setup internal router
	internalRouter := chi.NewRouter()

	// Internal router middleware
	internalRouter.Use(middleware.RequestID)
	internalRouter.Use(middleware.RealIP)
	internalRouter.Use(customMiddleware.Recovery())
	internalRouter.Use(customMiddleware.Logging())
	internalRouter.Use(customMiddleware.Metrics())
	internalRouter.Use(middleware.Timeout(cfg.Timeouts.HTTPMiddleware))

	// Internal endpoints - health, metrics
	internalRouter.Get("/health", allHandlers.Health.Health)
	internalRouter.Handle("/metrics", promhttp.Handler())

	// Public API routes
	publicRouter.Route("/production", func(r chi.Router) {
		// All public endpoints require JWT authentication
		r.Use(customMiddleware.Auth(jwtValidator))

		r.Get("/recipes", allHandlers.Recipe.GetRecipes)

		r.Route("/factory", func(r chi.Router) {
			r.Get("/queue", factoryHandler.GetQueue)
			r.Get("/completed", factoryHandler.GetCompleted)
			r.Post("/start", factoryHandler.StartProduction)
			r.Post("/claim", factoryHandler.Claim)
			r.Post("/cancel", factoryHandler.Cancel)
		})
	})

	// Create public HTTP server
	publicServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      publicRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Create internal HTTP server
	internalServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.InternalPort),
		Handler:      internalRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start public server in a goroutine
	go func() {
		logger.Info("Starting Production Service public server",
			zap.String("host", cfg.Server.Host),
			zap.String("port", cfg.Server.Port),
		)

		if err := publicServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start public server", zap.Error(err))
		}
	}()

	// Start internal server in a goroutine
	go func() {
		logger.Info("Starting Production Service internal server",
			zap.String("host", cfg.Server.Host),
			zap.String("port", cfg.Server.InternalPort),
		)

		if err := internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start internal server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Timeouts.GracefulShutdown)
	defer cancel()

	// Shutdown both servers
	shutdownErr := make(chan error, 2)

	go func() {
		if err := publicServer.Shutdown(ctx); err != nil {
			shutdownErr <- fmt.Errorf("public server shutdown error: %w", err)
		} else {
			shutdownErr <- nil
		}
	}()

	go func() {
		if err := internalServer.Shutdown(ctx); err != nil {
			shutdownErr <- fmt.Errorf("internal server shutdown error: %w", err)
		} else {
			shutdownErr <- nil
		}
	}()

	// Wait for both servers to shut down
	for i := 0; i < 2; i++ {
		if err := <-shutdownErr; err != nil {
			logger.Error("Server forced to shutdown", zap.Error(err))
		}
	}

	logger.Info("Servers exited")
}

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
			time.Sleep(10 * time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := jwtValidator.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize JWT validator", zap.Error(err))
	}

	// Refresh JWT public key periodically
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
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

	// Initialize external service clients
	inventoryClient := service.NewHTTPInventoryClient(cfg.ExternalServices.InventoryService.BaseURL, logger.Get())
	userClient := service.NewHTTPUserClient(cfg.ExternalServices.UserService.BaseURL, logger.Get())

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

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Recovery())
	r.Use(customMiddleware.Logging())
	r.Use(customMiddleware.Metrics())
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/health", allHandlers.Health.Health)
	r.Get("/ready", allHandlers.Health.Ready)

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints (require JWT)
		r.Group(func(r chi.Router) {
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

		// Internal endpoints (no auth required - internal services only)
		r.Route("/internal", func(r chi.Router) {
			r.Get("/task/{taskId}", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message":"Internal task endpoint not implemented yet"}`))
			})
			r.Get("/recipe/{recipeId}", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message":"Internal recipe endpoint not implemented yet"}`))
			})
		})

		// Admin endpoints (require JWT with admin role)
		r.Route("/admin", func(r chi.Router) {
			r.Use(customMiddleware.AdminAuth(jwtValidator))

			r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message":"Admin tasks endpoint not implemented yet"}`))
			})
			r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message":"Admin stats endpoint not implemented yet"}`))
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Production Service",
			zap.String("host", cfg.Server.Host),
			zap.String("port", cfg.Server.Port),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

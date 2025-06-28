package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"user-service/internal/config"
	"user-service/internal/handlers"
	"user-service/internal/handlers/api"
	"user-service/internal/handlers/public"
	"user-service/internal/middleware"
	"user-service/internal/services"
	"user-service/pkg/jwt"
	"user-service/pkg/logger"
)

func main() {
	// Инициализация логгера
	logger := logger.New()
	logger.Info("Starting User Service")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Загрузка публичного ключа от Auth Service
	logger.Info("Loading public key from Auth Service", map[string]interface{}{"url": cfg.Auth.PublicKeyURL})
	publicKey, err := jwt.LoadPublicKeyFromAuthService(cfg.Auth.PublicKeyURL)
	if err != nil {
		logger.Error("Failed to load public key from Auth Service", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	logger.Info("Public key loaded successfully")

	// Подключение к Redis
	logger.Info("Connecting to Redis", map[string]interface{}{"url": cfg.Auth.RedisURL})
	opt, err := redis.ParseURL(cfg.Auth.RedisURL)
	if err != nil {
		logger.Error("Failed to parse Redis URL", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	redisClient := redis.NewClient(opt)
	
	// Проверка подключения к Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	logger.Info("Redis connection established")

	// Инициализация сервисов
	mockDataService := services.NewMockDataService()

	// Инициализация middleware
	authMiddleware := middleware.NewJWTAuthMiddleware(publicKey, redisClient, logger)

	// Инициализация handlers
	healthHandler := handlers.NewHealthHandler(logger)
	profileHandler := public.NewProfileHandler(mockDataService, logger)
	internalUserHandler := api.NewUserHandler(mockDataService, logger)

	// Настройка Gin
	gin.SetMode(gin.ReleaseMode)
	
	// Create public router
	publicRouter := gin.New()
	publicRouter.Use(gin.Recovery())
	publicRouter.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request (Public)", map[string]interface{}{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency.String(),
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		})
		return ""
	}))

	// Create internal router
	internalRouter := gin.New()
	internalRouter.Use(gin.Recovery())
	internalRouter.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request (Internal)", map[string]interface{}{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency.String(),
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		})
		return ""
	}))

	// Register internal routes (no authentication)
	internalRouter.GET("/health", healthHandler.GetHealth)
	internalRouter.GET("/ready", healthHandler.GetHealth) // Readiness check
	internalRouter.GET("/users/:user_id/production-slots", internalUserHandler.GetProductionSlots)
	internalRouter.GET("/users/:user_id/production-modifiers", internalUserHandler.GetProductionModifiers)

	// Register public routes (with JWT authentication)
	publicRoutes := publicRouter.Group("/")
	publicRoutes.Use(authMiddleware.AuthenticateJWT())
	{
		publicRoutes.GET("/profile", profileHandler.GetProfile)
		publicRoutes.GET("/production-slots", profileHandler.GetProductionSlots)
	}

	// Create public HTTP server
	publicServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      publicRouter,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create internal HTTP server
	internalServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.InternalPort),
		Handler:      internalRouter,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск публичного сервера в отдельной горутине
	go func() {
		logger.Info("Starting public HTTP server", map[string]interface{}{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		})
		
		if err := publicServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start public HTTP server", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
	}()

	// Запуск внутреннего сервера в отдельной горутине
	go func() {
		logger.Info("Starting internal HTTP server", map[string]interface{}{
			"host": cfg.Server.Host,
			"port": cfg.Server.InternalPort,
		})
		
		if err := internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start internal HTTP server", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
	}()

	logger.Info("User Service started successfully")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down User Service")

	// Контекст для graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Остановка публичного HTTP сервера
	if err := publicServer.Shutdown(ctx); err != nil {
		logger.Error("Public HTTP server forced to shutdown", map[string]interface{}{"error": err.Error()})
	}

	// Остановка внутреннего HTTP сервера
	if err := internalServer.Shutdown(ctx); err != nil {
		logger.Error("Internal HTTP server forced to shutdown", map[string]interface{}{"error": err.Error()})
	}

	// Закрытие Redis подключения
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection", map[string]interface{}{"error": err.Error()})
	}

	logger.Info("User Service stopped")
}
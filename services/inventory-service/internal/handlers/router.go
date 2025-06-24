package handlers

import (
	"crypto/rsa"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shard-legends/inventory-service/internal/middleware"
	"github.com/shard-legends/inventory-service/internal/service"
)

// RouterConfig contains configuration for setting up routes
type RouterConfig struct {
	InventoryService  service.InventoryService
	ClassifierService service.ClassifierService
	PublicKey         *rsa.PublicKey
	RedisClient       *redis.Client
	Logger            *slog.Logger
}

// SetupRoutes configures all routes for the inventory service
func SetupRoutes(router *gin.Engine, config *RouterConfig) {
	// Initialize handlers
	inventoryHandler := NewInventoryHandler(
		config.InventoryService,
		config.ClassifierService,
		config.Logger,
	)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTAuthMiddleware(config.PublicKey, config.RedisClient, config.Logger)
	adminMiddleware := middleware.NewAdminMiddleware(config.Logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(config.Logger)

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware.LogRequests())

	// API base group
	api := router.Group("/api/inventory")

	// Public endpoints (require authentication)
	public := api.Group("")
	public.Use(jwtMiddleware.AuthenticateJWT())
	{
		// GET /inventory - получить инвентарь пользователя
		public.GET("/inventory", inventoryHandler.GetUserInventory)
	}

	// Internal endpoints (require authentication, used by other services)
	internal := api.Group("")
	internal.Use(jwtMiddleware.AuthenticateJWT())
	{
		// POST /inventory/reserve - зарезервировать предметы
		internal.POST("/inventory/reserve", inventoryHandler.ReserveItems)

		// POST /inventory/return-reserve - вернуть зарезервированные предметы
		internal.POST("/inventory/return-reserve", inventoryHandler.ReturnReservedItems)

		// POST /inventory/consume-reserve - потребить зарезервированные предметы
		internal.POST("/inventory/consume-reserve", inventoryHandler.ConsumeReservedItems)

		// POST /inventory/add-items - добавить предметы в инвентарь
		internal.POST("/inventory/add-items", inventoryHandler.AddItems)
	}

	// Admin endpoints (require authentication + admin permissions)
	admin := api.Group("/admin")
	admin.Use(jwtMiddleware.AuthenticateJWT())
	admin.Use(adminMiddleware.RequireAdmin())
	{
		// POST /admin/inventory/adjust - административная корректировка инвентаря
		admin.POST("/inventory/adjust", inventoryHandler.AdjustInventory)
	}

	// Health endpoint (no authentication required)
	router.GET("/health", inventoryHandler.GetHealthInfo)
}

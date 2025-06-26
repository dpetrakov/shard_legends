package handlers

import (
	"github.com/shard-legends/production-service/internal/database"
	"github.com/shard-legends/production-service/internal/handlers/public"
	"github.com/shard-legends/production-service/internal/service"
	"go.uber.org/zap"
)

// Handlers содержит все HTTP обработчики
type Handlers struct {
	Health *HealthHandler
	Recipe *public.RecipeHandler
}

// HandlerDependencies содержит зависимости для создания handlers
type HandlerDependencies struct {
	Service *service.Service
	DB      *database.DB
	Redis   *database.RedisClient
	Logger  *zap.Logger
}

// NewHandlers создает новый экземпляр Handlers со всеми обработчиками
func NewHandlers(deps *HandlerDependencies) *Handlers {
	return &Handlers{
		Health: NewHealthHandler(deps.DB, deps.Redis),
		Recipe: public.NewRecipeHandler(deps.Service.Recipe, deps.Logger),
	}
}

package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
)

// RecipeService определяет интерфейс сервиса для работы с рецептами
type RecipeService interface {
	// GetRecipesForUser возвращает рецепты доступные пользователю с учетом лимитов
	GetRecipesForUser(ctx context.Context, userID uuid.UUID, filters *models.RecipeFilters) ([]models.ProductionRecipeWithLimits, error)
	
	// GetRecipeByID возвращает рецепт по ID с информацией о лимитах для пользователя
	GetRecipeByID(ctx context.Context, recipeID uuid.UUID, userID *uuid.UUID) (*models.ProductionRecipeWithLimits, error)
	
	// ValidateRecipeAccess проверяет доступность рецепта для пользователя
	ValidateRecipeAccess(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, requestedExecutions int) error
}

// CodeConverterService определяет интерфейс для преобразования кодов в UUID и обратно
type CodeConverterService interface {
	// ConvertRecipeFromCodes преобразует коды в UUID для рецепта
	ConvertRecipeFromCodes(ctx context.Context, recipe *models.ProductionRecipe) error
	
	// ConvertRecipeToCodes преобразует UUID в коды для рецепта
	ConvertRecipeToCodes(ctx context.Context, recipe *models.ProductionRecipe) error
	
	// ConvertCodeToUUID преобразует код в UUID
	ConvertCodeToUUID(ctx context.Context, classifierName, code string) (*uuid.UUID, error)
	
	// ConvertUUIDToCode преобразует UUID в код
	ConvertUUIDToCode(ctx context.Context, classifierName string, id uuid.UUID) (*string, error)
}

// ServiceDependencies содержит зависимости для создания сервисов
type ServiceDependencies struct {
	Repository *storage.Repository
	Cache      storage.CacheInterface
	Metrics    storage.MetricsInterface
}

// Service объединяет все сервисы
type Service struct {
	Recipe        RecipeService
	CodeConverter CodeConverterService
}

// NewService создает новый экземпляр Service со всеми сервисами
func NewService(deps *ServiceDependencies) *Service {
	return &Service{
		Recipe:        NewRecipeService(deps),
		CodeConverter: NewCodeConverterService(deps),
	}
}
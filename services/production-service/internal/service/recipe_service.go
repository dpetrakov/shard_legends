package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
)

// recipeService реализует RecipeService
type recipeService struct {
	repository    *storage.Repository
	cache         storage.CacheInterface
	metrics       storage.MetricsInterface
	codeConverter CodeConverterService
}

// NewRecipeService создает новый экземпляр сервиса рецептов
func NewRecipeService(deps *ServiceDependencies) RecipeService {
	return &recipeService{
		repository: deps.Repository,
		cache:      deps.Cache,
		metrics:    deps.Metrics,
		codeConverter: NewCodeConverterService(deps),
	}
}

// GetRecipesForUser возвращает рецепты доступные пользователю с учетом лимитов
func (s *recipeService) GetRecipesForUser(ctx context.Context, userID uuid.UUID, filters *models.RecipeFilters) ([]models.ProductionRecipeWithLimits, error) {
	_ = time.Now() // TODO: добавить метрику для времени выполнения сервиса

	// Получаем все активные рецепты с фильтрацией
	recipes, err := s.repository.Recipe.GetActiveRecipes(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get active recipes: %w", err)
	}

	var recipesWithLimits []models.ProductionRecipeWithLimits
	for _, recipe := range recipes {
		// Преобразуем UUID в коды для API
		if err := s.codeConverter.ConvertRecipeToCodes(ctx, &recipe); err != nil {
			return nil, fmt.Errorf("failed to convert recipe codes for %s: %w", recipe.ID, err)
		}

		// Получаем информацию о лимитах пользователя
		userLimits, err := s.repository.Recipe.CheckRecipeLimits(ctx, userID, recipe.ID, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to check recipe limits for %s: %w", recipe.ID, err)
		}

		recipeWithLimits := models.ProductionRecipeWithLimits{
			ProductionRecipe: recipe,
			UserLimits:       userLimits,
		}

		recipesWithLimits = append(recipesWithLimits, recipeWithLimits)
	}

	return recipesWithLimits, nil
}

// GetRecipeByID возвращает рецепт по ID с информацией о лимитах для пользователя
func (s *recipeService) GetRecipeByID(ctx context.Context, recipeID uuid.UUID, userID *uuid.UUID) (*models.ProductionRecipeWithLimits, error) {
	_ = time.Now() // TODO: добавить метрику для времени выполнения сервиса

	// Получаем рецепт из репозитория
	recipe, err := s.repository.Recipe.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe %s: %w", recipeID, err)
	}

	// Преобразуем UUID в коды для API
	if err := s.codeConverter.ConvertRecipeToCodes(ctx, recipe); err != nil {
		return nil, fmt.Errorf("failed to convert recipe codes: %w", err)
	}

	recipeWithLimits := &models.ProductionRecipeWithLimits{
		ProductionRecipe: *recipe,
	}

	// Если указан пользователь, получаем информацию о лимитах
	if userID != nil {
		userLimits, err := s.repository.Recipe.CheckRecipeLimits(ctx, *userID, recipeID, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to check recipe limits: %w", err)
		}
		recipeWithLimits.UserLimits = userLimits
	}

	return recipeWithLimits, nil
}

// ValidateRecipeAccess проверяет доступность рецепта для пользователя
func (s *recipeService) ValidateRecipeAccess(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, requestedExecutions int) error {
	_ = time.Now() // TODO: добавить метрику для времени выполнения сервиса

	// Проверяем что рецепт существует и активен
	recipe, err := s.repository.Recipe.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return fmt.Errorf("recipe not found: %w", err)
	}

	if !recipe.IsActive {
		return fmt.Errorf("recipe %s is not active", recipeID)
	}

	// Проверяем лимиты пользователя
	userLimits, err := s.repository.Recipe.CheckRecipeLimits(ctx, userID, recipeID, requestedExecutions)
	if err != nil {
		return fmt.Errorf("failed to check recipe limits: %w", err)
	}

	// Проверяем что ни один лимит не превышен
	for _, limit := range userLimits {
		if limit.IsExceeded {
			return &RecipeLimitExceededError{
				RecipeID:     recipeID,
				LimitType:    limit.LimitType,
				LimitObject:  limit.LimitObject,
				CurrentUsage: limit.CurrentUsage,
				MaxAllowed:   limit.MaxAllowed,
				ResetTime:    limit.ResetTime,
			}
		}
	}

	return nil
}

// RecipeLimitExceededError представляет ошибку превышения лимита рецепта
type RecipeLimitExceededError struct {
	RecipeID     uuid.UUID
	LimitType    string
	LimitObject  string
	CurrentUsage int
	MaxAllowed   int
	ResetTime    *string
}

func (e *RecipeLimitExceededError) Error() string {
	return fmt.Sprintf("recipe limit exceeded for %s: %s %s (%d/%d)", 
		e.RecipeID, e.LimitType, e.LimitObject, e.CurrentUsage, e.MaxAllowed)
}

// IsRecipeLimitExceededError проверяет является ли ошибка ошибкой превышения лимита
func IsRecipeLimitExceededError(err error) (*RecipeLimitExceededError, bool) {
	if limitErr, ok := err.(*RecipeLimitExceededError); ok {
		return limitErr, true
	}
	return nil, false
}
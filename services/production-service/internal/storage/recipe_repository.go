package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
)

// recipeRepository реализует RecipeRepository
type recipeRepository struct {
	db      DatabaseInterface
	cache   CacheInterface
	metrics MetricsInterface
}

// NewRecipeRepository создает новый экземпляр репозитория рецептов
func NewRecipeRepository(deps *RepositoryDependencies) RecipeRepository {
	return &recipeRepository{
		db:      deps.DB,
		cache:   deps.Cache,
		metrics: deps.MetricsCollector,
	}
}

// GetActiveRecipes возвращает список активных рецептов с возможностью фильтрации
func (r *recipeRepository) GetActiveRecipes(ctx context.Context, filters *models.RecipeFilters) ([]models.ProductionRecipe, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_active_recipes", time.Since(start))
	}()
	r.metrics.IncDBQuery("get_active_recipes")

	// Базовый запрос
	query := `
		SELECT 
			r.id, r.name, r.operation_class_code, r.description, 
			r.is_active, r.production_time_seconds, r.created_at, r.updated_at
		FROM production.recipes r
		WHERE r.is_active = true
	`
	args := []interface{}{}
	argIndex := 1

	// Добавляем фильтры
	if filters != nil {
		if filters.OperationClassCode != nil {
			query += fmt.Sprintf(" AND r.operation_class_code = $%d", argIndex)
			args = append(args, *filters.OperationClassCode)
			argIndex++
		}
	}

	query += " ORDER BY r.name"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipes: %w", err)
	}
	defer rows.Close()

	var recipes []models.ProductionRecipe
	for rows.Next() {
		var recipe models.ProductionRecipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.OperationClassCode,
			&recipe.Description,
			&recipe.IsActive,
			&recipe.ProductionTimeSeconds,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}

		// Загружаем связанные данные
		if err := r.loadRecipeDetails(ctx, &recipe); err != nil {
			return nil, fmt.Errorf("failed to load recipe details for %s: %w", recipe.ID, err)
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return recipes, nil
}

// GetRecipeByID возвращает рецепт по ID с полной информацией
func (r *recipeRepository) GetRecipeByID(ctx context.Context, recipeID uuid.UUID) (*models.ProductionRecipe, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_recipe_by_id", time.Since(start))
	}()
	r.metrics.IncDBQuery("get_recipe_by_id")

	query := `
		SELECT 
			id, name, operation_class_code, description, 
			is_active, production_time_seconds, created_at, updated_at
		FROM production.recipes 
		WHERE id = $1
	`

	var recipe models.ProductionRecipe
	row := r.db.QueryRow(ctx, query, recipeID)
	err := row.Scan(
		&recipe.ID,
		&recipe.Name,
		&recipe.OperationClassCode,
		&recipe.Description,
		&recipe.IsActive,
		&recipe.ProductionTimeSeconds,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe %s: %w", recipeID, err)
	}

	// Загружаем связанные данные
	if err := r.loadRecipeDetails(ctx, &recipe); err != nil {
		return nil, fmt.Errorf("failed to load recipe details: %w", err)
	}

	return &recipe, nil
}

// loadRecipeDetails загружает входные предметы, выходные предметы и лимиты для рецепта
func (r *recipeRepository) loadRecipeDetails(ctx context.Context, recipe *models.ProductionRecipe) error {
	// Загружаем входные предметы
	inputQuery := `
		SELECT item_id, collection_code, quality_level_code, quantity
		FROM production.recipe_input_items 
		WHERE recipe_id = $1
		ORDER BY item_id
	`
	inputRows, err := r.db.Query(ctx, inputQuery, recipe.ID)
	if err != nil {
		return fmt.Errorf("failed to query input items: %w", err)
	}
	defer inputRows.Close()

	for inputRows.Next() {
		var input models.RecipeInputItem
		input.RecipeID = recipe.ID
		err := inputRows.Scan(
			&input.ItemID,
			&input.CollectionCode,
			&input.QualityLevelCode,
			&input.Quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to scan input item: %w", err)
		}
		recipe.InputItems = append(recipe.InputItems, input)
	}

	// Загружаем выходные предметы
	outputQuery := `
		SELECT 
			item_id, min_quantity, max_quantity, probability_percent, output_group,
			collection_source_input_index, quality_source_input_index,
			fixed_collection_code, fixed_quality_level_code
		FROM production.recipe_output_items 
		WHERE recipe_id = $1
		ORDER BY output_group NULLS LAST, probability_percent DESC
	`
	outputRows, err := r.db.Query(ctx, outputQuery, recipe.ID)
	if err != nil {
		return fmt.Errorf("failed to query output items: %w", err)
	}
	defer outputRows.Close()

	for outputRows.Next() {
		var output models.RecipeOutputItem
		output.RecipeID = recipe.ID
		err := outputRows.Scan(
			&output.ItemID,
			&output.MinQuantity,
			&output.MaxQuantity,
			&output.ProbabilityPercent,
			&output.OutputGroup,
			&output.CollectionSourceInputIndex,
			&output.QualitySourceInputIndex,
			&output.FixedCollectionCode,
			&output.FixedQualityLevelCode,
		)
		if err != nil {
			return fmt.Errorf("failed to scan output item: %w", err)
		}
		recipe.OutputItems = append(recipe.OutputItems, output)
	}

	// Загружаем лимиты
	limits, err := r.GetRecipeLimits(ctx, recipe.ID)
	if err != nil {
		return fmt.Errorf("failed to load recipe limits: %w", err)
	}
	recipe.Limits = limits

	return nil
}

// GetRecipeLimits возвращает все лимиты для рецепта
func (r *recipeRepository) GetRecipeLimits(ctx context.Context, recipeID uuid.UUID) ([]models.RecipeLimit, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_recipe_limits", time.Since(start))
	}()
	r.metrics.IncDBQuery("get_recipe_limits")

	query := `
		SELECT limit_type, limit_object, target_item_id, limit_quantity
		FROM production.recipe_limits 
		WHERE recipe_id = $1
		ORDER BY limit_type, limit_object
	`

	rows, err := r.db.Query(ctx, query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipe limits: %w", err)
	}
	defer rows.Close()

	var limits []models.RecipeLimit
	for rows.Next() {
		var limit models.RecipeLimit
		limit.RecipeID = recipeID
		err := rows.Scan(
			&limit.LimitType,
			&limit.LimitObject,
			&limit.TargetItemID,
			&limit.LimitQuantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe limit: %w", err)
		}
		limits = append(limits, limit)
	}

	return limits, nil
}

// GetRecipeUsageStats возвращает статистику использования рецепта пользователем
func (r *recipeRepository) GetRecipeUsageStats(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, limitType string, limitObject string, targetItemID *uuid.UUID, periodStart, periodEnd time.Time) (*models.RecipeUsageStats, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_recipe_usage_stats", time.Since(start))
	}()
	r.metrics.IncDBQuery("get_recipe_usage_stats")

	var query string
	var args []interface{}

	switch limitObject {
	case models.LimitObjectRecipeExecution:
		// Считаем количество выполнений рецепта
		query = `
			SELECT COUNT(*) 
			FROM production.production_tasks 
			WHERE user_id = $1 AND recipe_id = $2 
			  AND created_at >= $3 AND created_at < $4
			  AND status != 'cancelled'
		`
		args = []interface{}{userID, recipeID, periodStart, periodEnd}

	case models.LimitObjectItemReward:
		// Считаем количество полученных предметов
		if targetItemID == nil {
			return nil, fmt.Errorf("target_item_id is required for item_reward limits")
		}
		query = `
			SELECT COALESCE(SUM(toi.quantity), 0)
			FROM production.production_tasks pt
			JOIN production.task_output_items toi ON pt.id = toi.task_id
			WHERE pt.user_id = $1 AND pt.recipe_id = $2 
			  AND toi.item_id = $3
			  AND pt.created_at >= $4 AND pt.created_at < $5
			  AND pt.status IN ('completed', 'claimed')
		`
		args = []interface{}{userID, recipeID, *targetItemID, periodStart, periodEnd}

	default:
		return nil, fmt.Errorf("unsupported limit object: %s", limitObject)
	}

	var currentUsage int
	row := r.db.QueryRow(ctx, query, args...)
	if err := row.Scan(&currentUsage); err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// Получаем лимит из рецепта
	limitQuery := `
		SELECT limit_quantity 
		FROM production.recipe_limits 
		WHERE recipe_id = $1 AND limit_type = $2 AND limit_object = $3
	`
	limitArgs := []interface{}{recipeID, limitType, limitObject}
	
	if targetItemID != nil {
		limitQuery += " AND target_item_id = $4"
		limitArgs = append(limitArgs, *targetItemID)
	} else {
		limitQuery += " AND target_item_id IS NULL"
	}

	var limitQuantity int
	row = r.db.QueryRow(ctx, limitQuery, limitArgs...)
	if err := row.Scan(&limitQuantity); err != nil {
		return nil, fmt.Errorf("failed to get limit quantity: %w", err)
	}

	return &models.RecipeUsageStats{
		RecipeID:      recipeID,
		UserID:        userID,
		LimitType:     limitType,
		LimitObject:   limitObject,
		TargetItemID:  targetItemID,
		CurrentUsage:  currentUsage,
		LimitQuantity: limitQuantity,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
	}, nil
}

// CheckRecipeLimits проверяет все лимиты рецепта для пользователя
func (r *recipeRepository) CheckRecipeLimits(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, requestedExecutions int) ([]models.UserRecipeLimit, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("check_recipe_limits", time.Since(start))
	}()
	r.metrics.IncDBQuery("check_recipe_limits")

	// Получаем все лимиты рецепта
	limits, err := r.GetRecipeLimits(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe limits: %w", err)
	}

	var userLimits []models.UserRecipeLimit
	now := time.Now()

	for _, limit := range limits {
		// Определяем период для лимита
		periodStart, periodEnd := r.calculateLimitPeriod(now, limit.LimitType)

		// Получаем текущую статистику использования
		stats, err := r.GetRecipeUsageStats(ctx, userID, recipeID, limit.LimitType, limit.LimitObject, limit.TargetItemID, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get usage stats: %w", err)
		}

		// Рассчитываем использование после запрашиваемых выполнений
		futureUsage := stats.CurrentUsage + requestedExecutions
		isExceeded := futureUsage > stats.LimitQuantity

		// Определяем время сброса лимита
		var resetTime *string
		if limit.LimitType != models.LimitTypeTotal {
			resetTimeStr := periodEnd.Format(time.RFC3339)
			resetTime = &resetTimeStr
		}

		userLimit := models.UserRecipeLimit{
			LimitType:    limit.LimitType,
			LimitObject:  limit.LimitObject,
			TargetItemID: limit.TargetItemID,
			CurrentUsage: stats.CurrentUsage,
			MaxAllowed:   stats.LimitQuantity,
			IsExceeded:   isExceeded,
			ResetTime:    resetTime,
		}

		userLimits = append(userLimits, userLimit)
	}

	return userLimits, nil
}

// calculateLimitPeriod рассчитывает период для проверки лимита
func (r *recipeRepository) calculateLimitPeriod(now time.Time, limitType string) (time.Time, time.Time) {
	var start, end time.Time

	switch limitType {
	case models.LimitTypeDaily:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 1)

	case models.LimitTypeWeekly:
		// Неделя начинается с понедельника
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday-1))
		end = start.AddDate(0, 0, 7)

	case models.LimitTypeMonthly:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)

	case models.LimitTypeSeasonal:
		// Сезон длится 3 месяца, начинается с первого месяца квартала
		month := now.Month()
		seasonStart := ((int(month)-1)/3)*3 + 1
		start = time.Date(now.Year(), time.Month(seasonStart), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 3, 0)

	case models.LimitTypeTotal, models.LimitTypePerEvent:
		// Для total и per_event лимитов период не ограничен
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	default:
		// По умолчанию - дневной лимит
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 1)
	}

	return start, end
}
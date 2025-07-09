package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/shard-legends/deck-game-service/internal/service"
)

// RecipeStorage implements recipe data access using PostgreSQL
type RecipeStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewRecipeStorage creates a new recipe storage
func NewRecipeStorage(pool *pgxpool.Pool, logger *slog.Logger) service.RecipeRepository {
	return &RecipeStorage{
		pool:   pool,
		logger: logger,
	}
}

// FindRecipesByOutputItem finds recipes by output item characteristics
func (s *RecipeStorage) FindRecipesByOutputItem(ctx context.Context, itemCode string, qualityLevelCode, collectionCode *string) ([]service.Recipe, error) {
	query := `
		SELECT DISTINCT 
			r.id,
			r.code,
			r.operation_class_code,
			r.production_time_seconds,
			r.is_active,
			ro.fixed_quality_level_code,
			ro.fixed_collection_code
		FROM production.recipes r
		INNER JOIN production.recipe_output_items ro ON r.id = ro.recipe_id
		INNER JOIN inventory.items i ON ro.item_id = i.id
		INNER JOIN inventory.classifier_items ci_type ON i.item_type_id = ci_type.id
		WHERE r.operation_class_code = 'trade_purchase'
		  AND r.is_active = true
		  AND ci_type.code = $1
	`

	args := []interface{}{itemCode}
	argIndex := 2

	// Add quality level filter if specified
	if qualityLevelCode != nil {
		query += " AND ro.fixed_quality_level_code = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *qualityLevelCode)
		argIndex++
	}

	// Add collection filter if specified
	if collectionCode != nil {
		if *collectionCode == "base" {
			// For base collection, match both NULL and 'base' values
			query += " AND (ro.fixed_collection_code IS NULL OR ro.fixed_collection_code = $" + fmt.Sprintf("%d", argIndex) + ")"
			args = append(args, *collectionCode)
		} else {
			// For non-base collections, match exact value
			query += " AND ro.fixed_collection_code = $" + fmt.Sprintf("%d", argIndex)
			args = append(args, *collectionCode)
		}
		argIndex++
	}

	s.logger.Debug("Executing recipe search query",
		"item_code", itemCode,
		"quality_level_code", qualityLevelCode,
		"collection_code", collectionCode)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to execute recipe search query",
			"error", err,
			"item_code", itemCode,
			"quality_level_code", qualityLevelCode,
			"collection_code", collectionCode)
		return nil, errors.Wrap(err, "failed to query recipes")
	}
	defer rows.Close()

	var recipes []service.Recipe
	for rows.Next() {
		var recipe service.Recipe
		var productionTimeSeconds int
		var isActive bool
		err := rows.Scan(
			&recipe.ID,
			&recipe.Code,
			&recipe.OperationClassCode,
			&productionTimeSeconds,
			&isActive,
			&recipe.FixedQualityLevel,
			&recipe.FixedCollectionCode,
		)
		if err != nil {
			s.logger.Error("Failed to scan recipe row",
				"error", err)
			return nil, errors.Wrap(err, "failed to scan recipe")
		}
		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error occurred while iterating rows",
			"error", err)
		return nil, errors.Wrap(err, "failed to iterate recipe rows")
	}

	s.logger.Debug("Recipe search completed",
		"item_code", itemCode,
		"quality_level_code", qualityLevelCode,
		"collection_code", collectionCode,
		"recipes_found", len(recipes))

	return recipes, nil
}

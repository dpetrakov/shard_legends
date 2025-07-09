package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
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

// GetSapphireShopRecipes returns recipes that can be purchased with sapphires only
func (s *RecipeStorage) GetSapphireShopRecipes(ctx context.Context) ([]service.SapphireShopRecipe, error) {
	query := `
		SELECT DISTINCT
			r.id,
			r.code,
			-- Input items (sapphires)
			ri.item_id as input_item_id,
			ci_input_type.code as input_item_code,
			ri.collection_code as input_collection_code,
			ri.quality_level_code as input_quality_level_code,
			ri.quantity as input_quantity,
			-- Output items
			ro.item_id as output_item_id,
			ci_output_type.code as output_item_code,
			ro.fixed_collection_code as output_collection_code,
			ro.fixed_quality_level_code as output_quality_level_code,
			ro.min_quantity as output_min_quantity,
			ro.max_quantity as output_max_quantity
		FROM production.recipes r
		INNER JOIN production.recipe_input_items ri ON r.id = ri.recipe_id
		INNER JOIN production.recipe_output_items ro ON r.id = ro.recipe_id
		INNER JOIN inventory.items i_input ON ri.item_id = i_input.id
		INNER JOIN inventory.items i_output ON ro.item_id = i_output.id
		INNER JOIN inventory.classifier_items ci_input_type ON i_input.item_type_id = ci_input_type.id
		INNER JOIN inventory.classifier_items ci_output_type ON i_output.item_type_id = ci_output_type.id
		WHERE r.operation_class_code = 'trade_purchase'
		  AND r.is_active = true
		  AND ci_input_type.code = 'sapphires'
		  AND r.id IN (
				-- Recipes that have only sapphires as input
				SELECT recipe_id
				FROM production.recipe_input_items ri2
				INNER JOIN inventory.items i2 ON ri2.item_id = i2.id
				INNER JOIN inventory.classifier_items ci2 ON i2.item_type_id = ci2.id
				GROUP BY recipe_id
				HAVING COUNT(DISTINCT ci2.code) = 1 AND MIN(ci2.code) = 'sapphires'
			  )
		ORDER BY r.code
	`

	s.logger.Debug("Executing sapphire shop recipes query")

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		s.logger.Error("Failed to execute sapphire shop recipes query", "error", err)
		return nil, errors.Wrap(err, "failed to query sapphire shop recipes")
	}
	defer rows.Close()

	recipesMap := make(map[string]*service.SapphireShopRecipe)

	for rows.Next() {
		var recipeID uuid.UUID
		var recipeCode string
		var inputItemID, outputItemID uuid.UUID
		var inputItemCode, outputItemCode string
		var inputCollectionCode, inputQualityLevelCode *string
		var outputCollectionCode, outputQualityLevelCode *string
		var inputQuantity, outputMinQuantity, outputMaxQuantity int

		err := rows.Scan(
			&recipeID,
			&recipeCode,
			&inputItemID,
			&inputItemCode,
			&inputCollectionCode,
			&inputQualityLevelCode,
			&inputQuantity,
			&outputItemID,
			&outputItemCode,
			&outputCollectionCode,
			&outputQualityLevelCode,
			&outputMinQuantity,
			&outputMaxQuantity,
		)
		if err != nil {
			s.logger.Error("Failed to scan sapphire shop recipe row", "error", err)
			return nil, errors.Wrap(err, "failed to scan sapphire shop recipe")
		}

		// Handle NULL values as "base" for collection codes
		if inputCollectionCode != nil && *inputCollectionCode == "" {
			inputCollectionCode = nil
		}
		if outputCollectionCode != nil && *outputCollectionCode == "" {
			outputCollectionCode = nil
		}

		// Get or create recipe
		recipe, exists := recipesMap[recipeCode]
		if !exists {
			recipe = &service.SapphireShopRecipe{
				RecipeID: recipeID,
				Code:     recipeCode,
				Input:    []service.SapphireShopItemInfo{},
				Output: service.SapphireShopItemInfo{
					ItemID:           outputItemID,
					Code:             outputItemCode,
					CollectionCode:   outputCollectionCode,
					QualityLevelCode: outputQualityLevelCode,
					MinQuantity:      &outputMinQuantity,
					MaxQuantity:      &outputMaxQuantity,
				},
			}
			recipesMap[recipeCode] = recipe
		}

		// Add input item if not already present
		inputItem := service.SapphireShopItemInfo{
			ItemID:           inputItemID,
			Code:             inputItemCode,
			CollectionCode:   inputCollectionCode,
			QualityLevelCode: inputQualityLevelCode,
			Quantity:         &inputQuantity,
		}

		// Check if this input item is already in the recipe
		alreadyExists := false
		for _, existingInput := range recipe.Input {
			if existingInput.ItemID == inputItem.ItemID {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			recipe.Input = append(recipe.Input, inputItem)
		}
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error occurred while iterating sapphire shop recipe rows", "error", err)
		return nil, errors.Wrap(err, "failed to iterate sapphire shop recipe rows")
	}

	// Convert map to slice
	recipes := make([]service.SapphireShopRecipe, 0, len(recipesMap))
	for _, recipe := range recipesMap {
		recipes = append(recipes, *recipe)
	}

	s.logger.Debug("Sapphire shop recipes query completed", "recipes_found", len(recipes))

	return recipes, nil
}

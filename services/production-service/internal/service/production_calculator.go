package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
	"go.uber.org/zap"
)

// ProductionCalculator реализует алгоритм предрасчета результатов производства
type ProductionCalculator struct {
	classifierRepo  storage.ClassifierRepository
	modifierService *ModifierService
	logger          *zap.Logger
	rng             *rand.Rand
}

// NewProductionCalculator создает новый экземпляр ProductionCalculator
func NewProductionCalculator(
	classifierRepo storage.ClassifierRepository,
	modifierService *ModifierService,
	logger *zap.Logger,
) *ProductionCalculator {
	return &ProductionCalculator{
		classifierRepo:  classifierRepo,
		modifierService: modifierService,
		logger:          logger,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CalculationContext содержит контекст для расчета производства
type CalculationContext struct {
	UserID              uuid.UUID
	Recipe              *models.ProductionRecipe
	Request             models.StartProductionRequest
	Modifiers           []Modifier
	ModificationResults []ModificationResult
}

// PrecalculateProduction выполняет полный предрасчет производства с применением модификаторов
func (c *ProductionCalculator) PrecalculateProduction(ctx context.Context, userID uuid.UUID, recipe *models.ProductionRecipe, request models.StartProductionRequest) (*CalculationContext, error) {
	calcCtx := &CalculationContext{
		UserID:  userID,
		Recipe:  recipe,
		Request: request,
	}

	// 1. Получаем все модификаторы
	modifiers, err := c.modifierService.GetAllModifiers(ctx, userID, request.Boosters)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifiers: %w", err)
	}
	calcCtx.Modifiers = modifiers

	// 2. Применяем модификаторы к времени производства
	timeResult := c.modifierService.ApplyProductionTimeModifiers(recipe.ProductionTimeSeconds, modifiers)
	calcCtx.ModificationResults = append(calcCtx.ModificationResults, timeResult)

	// 3. Применяем модификаторы к входным предметам (снижение затрат)
	costResult := c.modifierService.ApplyCostReductionModifiers(recipe.InputItems, modifiers)
	calcCtx.ModificationResults = append(calcCtx.ModificationResults, costResult)

	return calcCtx, nil
}

// CalculateOutputItems рассчитывает выходные предметы с учетом всех модификаторов
func (c *ProductionCalculator) CalculateOutputItems(ctx context.Context, calcCtx *CalculationContext) ([]models.TaskOutputItem, error) {
	var outputs []models.TaskOutputItem

	// Применяем количество исполнений
	for execution := 0; execution < calcCtx.Request.ExecutionCount; execution++ {
		executionOutputs, err := c.calculateSingleExecution(ctx, calcCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate execution %d: %w", execution+1, err)
		}
		outputs = append(outputs, executionOutputs...)
	}

	// Группируем одинаковые предметы
	return c.consolidateOutputs(outputs), nil
}

// calculateSingleExecution рассчитывает результат одного исполнения рецепта
func (c *ProductionCalculator) calculateSingleExecution(ctx context.Context, calcCtx *CalculationContext) ([]models.TaskOutputItem, error) {
	var outputs []models.TaskOutputItem

	// Группируем выходные предметы по группам
	outputGroups := make(map[string][]models.RecipeOutputItem)
	var ungroupedItems []models.RecipeOutputItem

	for _, output := range calcCtx.Recipe.OutputItems {
		if output.OutputGroup != nil && *output.OutputGroup != "" {
			outputGroups[*output.OutputGroup] = append(outputGroups[*output.OutputGroup], output)
		} else {
			ungroupedItems = append(ungroupedItems, output)
		}
	}

	// Обрабатываем альтернативные группы (взаимоисключающие)
	for groupName, groupItems := range outputGroups {
		selectedItem, err := c.selectFromGroup(groupItems, calcCtx.Modifiers)
		if err != nil {
			return nil, fmt.Errorf("failed to select from group %s: %w", groupName, err)
		}

		if selectedItem != nil {
			output, err := c.createOutputItem(ctx, *selectedItem, calcCtx)
			if err != nil {
				return nil, fmt.Errorf("failed to create output item from group %s: %w", groupName, err)
			}
			outputs = append(outputs, output)
		}
	}

	// Обрабатываем независимые предметы
	for _, item := range ungroupedItems {
		// Применяем модификаторы вероятности
		probResult := c.modifierService.ApplyProbabilityModifiers(item.ProbabilityPercent, calcCtx.Modifiers)
		finalProbability := probResult.ModifiedValue.(int)

		roll := c.rng.Intn(100)
		if roll < finalProbability {
			output, err := c.createOutputItem(ctx, item, calcCtx)
			if err != nil {
				return nil, fmt.Errorf("failed to create independent output item: %w", err)
			}
			outputs = append(outputs, output)
		}
	}

	return outputs, nil
}

// selectFromGroup выбирает предмет из альтернативной группы
func (c *ProductionCalculator) selectFromGroup(groupItems []models.RecipeOutputItem, modifiers []Modifier) (*models.RecipeOutputItem, error) {
	// Применяем модификаторы вероятности к каждому предмету в группе
	var modifiedItems []struct {
		item        models.RecipeOutputItem
		probability int
	}

	totalProbability := 0
	for _, item := range groupItems {
		probResult := c.modifierService.ApplyProbabilityModifiers(item.ProbabilityPercent, modifiers)
		finalProbability := probResult.ModifiedValue.(int)

		modifiedItems = append(modifiedItems, struct {
			item        models.RecipeOutputItem
			probability int
		}{
			item:        item,
			probability: finalProbability,
		})
		totalProbability += finalProbability
	}

	if totalProbability == 0 {
		return nil, nil // Ничего не выпало
	}

	// Выбираем предмет по накопленным вероятностям
	roll := c.rng.Intn(totalProbability)
	currentProb := 0

	for _, modItem := range modifiedItems {
		currentProb += modItem.probability
		if roll < currentProb {
			return &modItem.item, nil
		}
	}

	return nil, nil // Ничего не выпало (не должно происходить)
}

// createOutputItem создает выходной предмет задания с учетом количества и наследования
func (c *ProductionCalculator) createOutputItem(ctx context.Context, recipeOutput models.RecipeOutputItem, calcCtx *CalculationContext) (models.TaskOutputItem, error) {
	// Применяем модификаторы количества
	qtyResult := c.modifierService.ApplyQuantityModifiers(recipeOutput.MinQuantity, recipeOutput.MaxQuantity, calcCtx.Modifiers)
	modifiedQty := qtyResult.ModifiedValue.(map[string]int)

	// Определяем случайное количество в модифицированном диапазоне
	minQty := modifiedQty["min"]
	maxQty := modifiedQty["max"]

	quantity := minQty
	if maxQty > minQty {
		quantity = minQty + c.rng.Intn(maxQty-minQty+1)
	}

	output := models.TaskOutputItem{
		ItemID:   recipeOutput.ItemID,
		Quantity: quantity,
	}

	// Обрабатываем наследование и фиксированные значения коллекций
	err := c.processCollectionInheritance(ctx, &output, recipeOutput, calcCtx.Recipe.InputItems)
	if err != nil {
		return output, fmt.Errorf("failed to process collection inheritance: %w", err)
	}

	// Обрабатываем наследование и фиксированные значения качества
	err = c.processQualityInheritance(ctx, &output, recipeOutput, calcCtx.Recipe.InputItems)
	if err != nil {
		return output, fmt.Errorf("failed to process quality inheritance: %w", err)
	}

	return output, nil
}

// processCollectionInheritance обрабатывает наследование коллекций
func (c *ProductionCalculator) processCollectionInheritance(ctx context.Context, output *models.TaskOutputItem, recipeOutput models.RecipeOutputItem, inputItems []models.RecipeInputItem) error {
	// Фиксированная коллекция имеет приоритет
	if recipeOutput.FixedCollectionCode != nil {
		output.CollectionCode = recipeOutput.FixedCollectionCode

		// Конвертируем код в UUID
		collectionID, err := c.classifierRepo.ConvertCodeToUUID(ctx, "item_collections", *recipeOutput.FixedCollectionCode)
		if err != nil {
			c.logger.Warn("Failed to convert collection code to UUID",
				zap.String("code", *recipeOutput.FixedCollectionCode),
				zap.Error(err))
		} else if collectionID != nil {
			output.CollectionID = collectionID
		}
		return nil
	}

	// Наследование от входного предмета
	if recipeOutput.CollectionSourceInputIndex != nil &&
		*recipeOutput.CollectionSourceInputIndex < len(inputItems) {

		inputItem := inputItems[*recipeOutput.CollectionSourceInputIndex]
		output.CollectionCode = inputItem.CollectionCode
		output.CollectionID = inputItem.CollectionID
	}

	return nil
}

// processQualityInheritance обрабатывает наследование качества
func (c *ProductionCalculator) processQualityInheritance(ctx context.Context, output *models.TaskOutputItem, recipeOutput models.RecipeOutputItem, inputItems []models.RecipeInputItem) error {
	// Фиксированное качество имеет приоритет
	if recipeOutput.FixedQualityLevelCode != nil {
		output.QualityLevelCode = recipeOutput.FixedQualityLevelCode

		// Конвертируем код в UUID
		qualityID, err := c.classifierRepo.ConvertCodeToUUID(ctx, "quality_levels", *recipeOutput.FixedQualityLevelCode)
		if err != nil {
			c.logger.Warn("Failed to convert quality code to UUID",
				zap.String("code", *recipeOutput.FixedQualityLevelCode),
				zap.Error(err))
		} else if qualityID != nil {
			output.QualityLevelID = qualityID
		}
		return nil
	}

	// Наследование от входного предмета
	if recipeOutput.QualitySourceInputIndex != nil &&
		*recipeOutput.QualitySourceInputIndex < len(inputItems) {

		inputItem := inputItems[*recipeOutput.QualitySourceInputIndex]
		output.QualityLevelCode = inputItem.QualityLevelCode
		output.QualityLevelID = inputItem.QualityLevelID
	}

	return nil
}

// consolidateOutputs группирует одинаковые предметы в единые записи
func (c *ProductionCalculator) consolidateOutputs(outputs []models.TaskOutputItem) []models.TaskOutputItem {
	itemMap := make(map[string]*models.TaskOutputItem)

	for _, output := range outputs {
		// Создаем ключ для группировки
		key := c.createItemKey(output)

		if existing, exists := itemMap[key]; exists {
			// Объединяем количество
			existing.Quantity += output.Quantity
		} else {
			// Создаем копию для добавления в карту
			newOutput := output
			itemMap[key] = &newOutput
		}
	}

	// Преобразуем карту обратно в слайс
	consolidated := make([]models.TaskOutputItem, 0, len(itemMap))
	for _, item := range itemMap {
		consolidated = append(consolidated, *item)
	}

	return consolidated
}

// createItemKey создает уникальный ключ для предмета для группировки
func (c *ProductionCalculator) createItemKey(item models.TaskOutputItem) string {
	key := item.ItemID.String()

	if item.CollectionID != nil {
		key += "_coll_" + item.CollectionID.String()
	}

	if item.QualityLevelID != nil {
		key += "_qual_" + item.QualityLevelID.String()
	}

	return key
}

// GetModifiedProductionTime возвращает модифицированное время производства из контекста
func (c *ProductionCalculator) GetModifiedProductionTime(calcCtx *CalculationContext) int {
	for _, result := range calcCtx.ModificationResults {
		if timeValue, ok := result.ModifiedValue.(int); ok {
			// Проверяем, что это результат модификации времени
			if result.OriginalValue != nil {
				if _, isInt := result.OriginalValue.(int); isInt {
					return timeValue
				}
			}
		}
	}

	// Возвращаем базовое время, если модификация не найдена
	return calcCtx.Recipe.ProductionTimeSeconds
}

// GetModifiedInputItems возвращает модифицированные входные предметы из контекста
func (c *ProductionCalculator) GetModifiedInputItems(calcCtx *CalculationContext) []models.RecipeInputItem {
	for _, result := range calcCtx.ModificationResults {
		if items, ok := result.ModifiedValue.([]models.RecipeInputItem); ok {
			return items
		}
	}

	// Возвращаем базовые предметы, если модификация не найдена
	return calcCtx.Recipe.InputItems
}

package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/shard-legends/production-service/internal/models"
)

// Тест для CalculationContext структуры
func TestCalculationContext_Structure(t *testing.T) {
	userID := uuid.New()
	recipe := &models.ProductionRecipe{
		ID:   uuid.New(),
		Name: "Test Recipe",
	}
	request := models.StartProductionRequest{
		ExecutionCount: 1,
	}

	ctx := &CalculationContext{
		UserID:  userID,
		Recipe:  recipe,
		Request: request,
	}

	assert.Equal(t, userID, ctx.UserID)
	assert.Equal(t, recipe, ctx.Recipe)
	assert.Equal(t, request, ctx.Request)
	assert.Empty(t, ctx.Modifiers)
	assert.Empty(t, ctx.ModificationResults)
}

// Тест для GetModifiedProductionTime с базовым временем
func TestProductionCalculator_GetModifiedProductionTime_BaseTime(t *testing.T) {
	calculator := NewProductionCalculator(nil, &ModifierService{}, nil)

	recipe := &models.ProductionRecipe{
		ProductionTimeSeconds: 1800,
	}

	calcCtx := &CalculationContext{
		Recipe: recipe,
		ModificationResults: []ModificationResult{
			{
				OriginalValue: "some_string", // Не int, поэтому не будет считаться временем
				ModifiedValue: 4,
				AppliedModifiers: []AppliedModifierInfo{},
			},
		},
	}

	result := calculator.GetModifiedProductionTime(calcCtx)
	assert.Equal(t, 1800, result) // Должен вернуть базовое время
}

// Тест для GetModifiedInputItems с базовыми предметами
func TestProductionCalculator_GetModifiedInputItems_BaseItems(t *testing.T) {
	calculator := NewProductionCalculator(nil, &ModifierService{}, nil)

	originalItems := []models.RecipeInputItem{
		{
			ItemID:   uuid.New(),
			Quantity: 10,
		},
	}

	calcCtx := &CalculationContext{
		Recipe: &models.ProductionRecipe{
			InputItems: originalItems,
		},
		ModificationResults: []ModificationResult{
			{
				OriginalValue: 3600,
				ModifiedValue: 2880,
				AppliedModifiers: []AppliedModifierInfo{},
			},
		},
	}

	result := calculator.GetModifiedInputItems(calcCtx)
	assert.Equal(t, originalItems, result)
}

// Тест для createItemKey
func TestProductionCalculator_CreateItemKey(t *testing.T) {
	calculator := NewProductionCalculator(nil, &ModifierService{}, nil)

	itemID := uuid.New()
	collectionID := uuid.New()
	qualityID := uuid.New()

	tests := []struct {
		name     string
		item     models.TaskOutputItem
		expected string
	}{
		{
			name: "item only",
			item: models.TaskOutputItem{
				ItemID: itemID,
			},
			expected: itemID.String(),
		},
		{
			name: "item with collection",
			item: models.TaskOutputItem{
				ItemID:       itemID,
				CollectionID: &collectionID,
			},
			expected: itemID.String() + "_coll_" + collectionID.String(),
		},
		{
			name: "item with collection and quality",
			item: models.TaskOutputItem{
				ItemID:         itemID,
				CollectionID:   &collectionID,
				QualityLevelID: &qualityID,
			},
			expected: itemID.String() + "_coll_" + collectionID.String() + "_qual_" + qualityID.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.createItemKey(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Тест для consolidateOutputs с базовым случаем
func TestProductionCalculator_ConsolidateOutputs_Basic(t *testing.T) {
	calculator := NewProductionCalculator(nil, &ModifierService{}, nil)

	itemID := uuid.New()
	outputs := []models.TaskOutputItem{
		{
			ItemID:   itemID,
			Quantity: 5,
		},
		{
			ItemID:   itemID,
			Quantity: 3,
		},
	}

	result := calculator.consolidateOutputs(outputs)

	assert.Len(t, result, 1)
	assert.Equal(t, itemID, result[0].ItemID)
	assert.Equal(t, 8, result[0].Quantity) // 5 + 3
}
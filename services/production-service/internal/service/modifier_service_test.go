package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserClient мок для UserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUserProductionSlots(ctx context.Context, userID uuid.UUID) (*models.UserProductionSlots, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.UserProductionSlots), args.Error(1)
}

func (m *MockUserClient) GetUserProductionModifiers(ctx context.Context, userID uuid.UUID) (*models.UserProductionModifiers, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.UserProductionModifiers), args.Error(1)
}

func TestModifierService_ApplyProductionTimeModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	tests := []struct {
		name         string
		baseTime     int
		modifiers    []Modifier
		expectedTime int
	}{
		{
			name:         "No modifiers",
			baseTime:     100,
			modifiers:    []Modifier{},
			expectedTime: 100,
		},
		{
			name:     "Single speed modifier",
			baseTime: 100,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.2, // 20% speed bonus
				},
			},
			expectedTime: 80, // 100 * (1 - 0.2) = 80
		},
		{
			name:     "Multiple speed modifiers",
			baseTime: 100,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.2, // 20% speed bonus
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.1, // 10% speed bonus
				},
			},
			expectedTime: 70, // 100 * (1 - 0.3) = 70
		},
		{
			name:     "Mixed modifiers (only speed should apply)",
			baseTime: 100,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.2,
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeQuantity, // Should be ignored
					Value: 0.5,
				},
			},
			expectedTime: 80,
		},
		{
			name:     "Very high speed bonus (minimum 1 second)",
			baseTime: 10,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.99, // 99% speed bonus
				},
			},
			expectedTime: 1, // Should not go below 1
		},
		{
			name:     "Instant task (0 seconds) remains instant",
			baseTime: 0,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed,
					Value: 0.5, // 50% speed bonus
				},
			},
			expectedTime: 0, // Should remain 0 for instant tasks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyProductionTimeModifiers(tt.baseTime, tt.modifiers)

			assert.Equal(t, tt.baseTime, result.OriginalValue)
			assert.Equal(t, tt.expectedTime, result.ModifiedValue)

			// Проверяем, что в результате есть информация о примененных модификаторах
			speedModifierCount := 0
			for _, modifier := range tt.modifiers {
				if modifier.Type == ModifierTypeSpeed {
					speedModifierCount++
				}
			}
			assert.Len(t, result.AppliedModifiers, speedModifierCount)
		})
	}
}

func TestModifierService_ApplyQuantityModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	tests := []struct {
		name        string
		minQty      int
		maxQty      int
		modifiers   []Modifier
		expectedMin int
		expectedMax int
	}{
		{
			name:        "No modifiers",
			minQty:      5,
			maxQty:      10,
			modifiers:   []Modifier{},
			expectedMin: 5,
			expectedMax: 10,
		},
		{
			name:   "Single quantity modifier",
			minQty: 5,
			maxQty: 10,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeQuantity,
					Value: 0.5, // 50% quantity bonus
				},
			},
			expectedMin: 7,  // 5 * 1.5 = 7.5 -> 7
			expectedMax: 15, // 10 * 1.5 = 15
		},
		{
			name:   "Multiple quantity modifiers",
			minQty: 4,
			maxQty: 8,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeQuantity,
					Value: 0.25, // 25% bonus
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeQuantity,
					Value: 0.25, // 25% bonus
				},
			},
			expectedMin: 6,  // 4 * 1.5 = 6
			expectedMax: 12, // 8 * 1.5 = 12
		},
		{
			name:   "Mixed modifiers (only quantity should apply)",
			minQty: 5,
			maxQty: 10,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeQuantity,
					Value: 0.2,
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeSpeed, // Should be ignored
					Value: 0.5,
				},
			},
			expectedMin: 6,  // 5 * 1.2 = 6
			expectedMax: 12, // 10 * 1.2 = 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyQuantityModifiers(tt.minQty, tt.maxQty, tt.modifiers)

			originalValue := result.OriginalValue.(map[string]int)
			modifiedValue := result.ModifiedValue.(map[string]int)

			assert.Equal(t, tt.minQty, originalValue["min"])
			assert.Equal(t, tt.maxQty, originalValue["max"])
			assert.Equal(t, tt.expectedMin, modifiedValue["min"])
			assert.Equal(t, tt.expectedMax, modifiedValue["max"])
		})
	}
}

func TestModifierService_ApplyProbabilityModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	tests := []struct {
		name                string
		baseProbability     float64
		modifiers           []Modifier
		expectedProbability float64
	}{
		{
			name:                "No modifiers",
			baseProbability:     50,
			modifiers:           []Modifier{},
			expectedProbability: 50,
		},
		{
			name:            "Single probability modifier",
			baseProbability: 50,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeProbability,
					Value: 0.2, // 20% probability bonus
				},
			},
			expectedProbability: 60, // 50 * 1.2 = 60
		},
		{
			name:            "Multiple probability modifiers",
			baseProbability: 40,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeProbability,
					Value: 0.25, // 25% bonus
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeProbability,
					Value: 0.25, // 25% bonus
				},
			},
			expectedProbability: 60, // 40 * 1.5 = 60
		},
		{
			name:            "Probability cap at 100%",
			baseProbability: 80,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeProbability,
					Value: 0.5, // 50% bonus
				},
			},
			expectedProbability: 100, // 80 * 1.5 = 120 -> capped at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyProbabilityModifiers(tt.baseProbability, tt.modifiers)

			assert.Equal(t, tt.baseProbability, result.OriginalValue)
			assert.Equal(t, tt.expectedProbability, result.ModifiedValue)
		})
	}
}

func TestModifierService_ApplyCostReductionModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	itemID1 := uuid.New()
	itemID2 := uuid.New()

	inputItems := []models.RecipeInputItem{
		{
			ItemID:   itemID1,
			Quantity: 10,
		},
		{
			ItemID:   itemID2,
			Quantity: 5,
		},
	}

	tests := []struct {
		name               string
		inputItems         []models.RecipeInputItem
		modifiers          []Modifier
		expectedQuantities []int
	}{
		{
			name:               "No modifiers",
			inputItems:         inputItems,
			modifiers:          []Modifier{},
			expectedQuantities: []int{10, 5},
		},
		{
			name:       "Single cost reduction modifier",
			inputItems: inputItems,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeCostReduction,
					Value: 0.2, // 20% cost reduction
				},
			},
			expectedQuantities: []int{8, 4}, // 10*0.8=8, 5*0.8=4
		},
		{
			name:       "Multiple cost reduction modifiers",
			inputItems: inputItems,
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeCostReduction,
					Value: 0.1, // 10% reduction
				},
				{
					ID:    uuid.New(),
					Type:  ModifierTypeCostReduction,
					Value: 0.1, // 10% reduction
				},
			},
			expectedQuantities: []int{8, 4}, // 10*0.8=8, 5*0.8=4
		},
		{
			name:       "Cost reduction minimum 1",
			inputItems: []models.RecipeInputItem{{ItemID: itemID1, Quantity: 2}},
			modifiers: []Modifier{
				{
					ID:    uuid.New(),
					Type:  ModifierTypeCostReduction,
					Value: 0.9, // 90% reduction
				},
			},
			expectedQuantities: []int{1}, // 2*0.1=0.2 -> minimum 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyCostReductionModifiers(tt.inputItems, tt.modifiers)

			modifiedItems := result.ModifiedValue.([]models.RecipeInputItem)

			assert.Len(t, modifiedItems, len(tt.expectedQuantities))
			for i, expectedQty := range tt.expectedQuantities {
				assert.Equal(t, expectedQty, modifiedItems[i].Quantity)
				assert.Equal(t, tt.inputItems[i].ItemID, modifiedItems[i].ItemID)
			}
		})
	}
}

func TestModifierService_GetUserModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	userID := uuid.New()
	ctx := context.Background()

	// Настраиваем мок для возврата пользовательских модификаторов
	userModifiers := &models.UserProductionModifiers{
		UserID: userID,
		Modifiers: models.Modifiers{
			VIPStatus: models.VIPStatus{
				Level:                "gold",
				ProductionSpeedBonus: 0.15,
				QualityBonus:         0.1,
			},
			CharacterLevel: models.CharacterLevel{
				Level:         10,
				CraftingBonus: 0.05,
			},
			ClanBonuses: models.ClanBonuses{
				ProductionSpeed: 0.1,
			},
		},
	}

	mockUserClient.On("GetUserProductionModifiers", ctx, userID).Return(userModifiers, nil)

	// Тестируем получение модификаторов
	modifiers, err := service.getUserModifiers(ctx, userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, modifiers)

	// Проверяем, что создались правильные модификаторы
	var speedModifiers []Modifier
	var qualityModifiers []Modifier
	var probabilityModifiers []Modifier

	for _, modifier := range modifiers {
		switch modifier.Type {
		case ModifierTypeSpeed:
			speedModifiers = append(speedModifiers, modifier)
		case ModifierTypeQuality:
			qualityModifiers = append(qualityModifiers, modifier)
		case ModifierTypeProbability:
			probabilityModifiers = append(probabilityModifiers, modifier)
		}
	}

	// Должно быть 2 скоростных модификатора (VIP + Clan)
	assert.Len(t, speedModifiers, 2)
	// Должен быть 1 модификатор качества (VIP)
	assert.Len(t, qualityModifiers, 1)
	// Должен быть 1 модификатор вероятности (Character level)
	assert.Len(t, probabilityModifiers, 1)

	mockUserClient.AssertExpectations(t)
}

func TestModifierService_GetBoosterModifiers(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	ctx := context.Background()

	boosters := []models.BoosterItem{
		{
			ItemID:   uuid.New(),
			Quantity: 1,
		},
		{
			ItemID:   uuid.New(),
			Quantity: 2,
		},
	}

	modifiers, err := service.getBoosterModifiers(ctx, boosters)

	assert.NoError(t, err)
	assert.Len(t, modifiers, 2)

	// Проверяем, что все модификаторы имеют правильный тип и источник
	for i, modifier := range modifiers {
		assert.Equal(t, ModifierTypeSpeed, modifier.Type)
		assert.Equal(t, ModifierSourceBooster, modifier.Source)
		assert.Equal(t, 0.1, modifier.Value) // Базовый бонус 10%
		assert.Equal(t, &boosters[i].ItemID, modifier.ItemID)
	}
}

func TestModifierService_BuildAppliedModifiersForAudit(t *testing.T) {
	logger := zap.NewNop()
	mockUserClient := &MockUserClient{}
	service := NewModifierService(mockUserClient, logger)

	results := []ModificationResult{
		{
			OriginalValue: 100,
			ModifiedValue: 80,
			AppliedModifiers: []AppliedModifierInfo{
				{
					ModifierID:  uuid.New(),
					Type:        ModifierTypeSpeed,
					Source:      ModifierSourceUser,
					Value:       0.2,
					Description: "VIP speed bonus",
				},
			},
		},
		{
			OriginalValue: map[string]int{"min": 5, "max": 10},
			ModifiedValue: map[string]int{"min": 6, "max": 12},
			AppliedModifiers: []AppliedModifierInfo{
				{
					ModifierID:  uuid.New(),
					Type:        ModifierTypeQuantity,
					Source:      ModifierSourceBooster,
					Value:       0.2,
					Description: "Item quantity bonus",
				},
			},
		},
	}

	audit := service.BuildAppliedModifiersForAudit(results)

	assert.Len(t, audit, 2)
	assert.Contains(t, audit, "modification_0")
	assert.Contains(t, audit, "modification_1")

	// Проверяем структуру аудита
	mod0 := audit["modification_0"].(map[string]interface{})
	assert.Equal(t, 100, mod0["original_value"])
	assert.Equal(t, 80, mod0["modified_value"])

	appliedMods0 := mod0["applied_modifiers"].([]AppliedModifierInfo)
	assert.Len(t, appliedMods0, 1)
	assert.Equal(t, ModifierTypeSpeed, appliedMods0[0].Type)
}

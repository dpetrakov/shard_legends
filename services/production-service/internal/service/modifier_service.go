package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"go.uber.org/zap"
)

// ModifierType представляет тип модификатора
type ModifierType string

const (
	ModifierTypeSpeed         ModifierType = "speed"          // Ускорение производства
	ModifierTypeQuantity      ModifierType = "quantity"       // Увеличение количества
	ModifierTypeProbability   ModifierType = "probability"    // Повышение шансов
	ModifierTypeCostReduction ModifierType = "cost_reduction" // Снижение затрат
	ModifierTypeQuality       ModifierType = "quality"        // Повышение качества
)

// ModifierSource представляет источник модификатора
type ModifierSource string

const (
	ModifierSourceUser    ModifierSource = "user"    // Пользовательские модификаторы
	ModifierSourceBooster ModifierSource = "booster" // Ускорители-предметы
	ModifierSourceClan    ModifierSource = "clan"    // Клановые бонусы
	ModifierSourceEvent   ModifierSource = "event"   // Событийные модификаторы
	ModifierSourceServer  ModifierSource = "server"  // Серверные буферы
)

// Modifier представляет модификатор
type Modifier struct {
	ID           uuid.UUID      `json:"id"`
	Type         ModifierType   `json:"type"`
	Source       ModifierSource `json:"source"`
	Value        float64        `json:"value"`             // Значение модификатора (множитель или добавка)
	IsMultiplier bool           `json:"is_multiplier"`     // Является ли значение множителем
	ItemID       *uuid.UUID     `json:"item_id,omitempty"` // ID предмета-ускорителя
	Description  string         `json:"description"`
}

// AppliedModifierInfo представляет информацию о примененном модификаторе для аудита
type AppliedModifierInfo struct {
	ModifierID  uuid.UUID      `json:"modifier_id"`
	Type        ModifierType   `json:"type"`
	Source      ModifierSource `json:"source"`
	Value       float64        `json:"value"`
	Description string         `json:"description"`
	AppliedAt   string         `json:"applied_at"`
}

// ModificationResult представляет результат применения модификаторов
type ModificationResult struct {
	OriginalValue    interface{}           `json:"original_value"`
	ModifiedValue    interface{}           `json:"modified_value"`
	AppliedModifiers []AppliedModifierInfo `json:"applied_modifiers"`
}

// ModifierService реализует бизнес-логику для работы с модификаторами
type ModifierService struct {
	userClient UserClient
	logger     *zap.Logger
}

// NewModifierService создает новый экземпляр ModifierService
func NewModifierService(userClient UserClient, logger *zap.Logger) *ModifierService {
	return &ModifierService{
		userClient: userClient,
		logger:     logger,
	}
}

// GetAllModifiers собирает все доступные модификаторы для пользователя
func (s *ModifierService) GetAllModifiers(ctx context.Context, userID uuid.UUID, boosters []models.BoosterItem) ([]Modifier, error) {
	var modifiers []Modifier

	// 1. Получаем пользовательские модификаторы
	userModifiers, err := s.getUserModifiers(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user modifiers, using defaults", zap.Error(err))
		userModifiers = []Modifier{}
	}
	modifiers = append(modifiers, userModifiers...)

	// 2. Получаем модификаторы от ускорителей-предметов
	boosterModifiers, err := s.getBoosterModifiers(ctx, boosters)
	if err != nil {
		s.logger.Warn("Failed to get booster modifiers, skipping", zap.Error(err))
	} else {
		modifiers = append(modifiers, boosterModifiers...)
	}

	// 3. Получаем автоматические модификаторы (клановые бонусы)
	autoModifiers := s.getAutomaticModifiers(ctx, userID)
	modifiers = append(modifiers, autoModifiers...)

	// 4. Получаем событийные модификаторы (TBD - в будущих версиях)
	// eventModifiers := s.getEventModifiers(ctx)
	// modifiers = append(modifiers, eventModifiers...)

	return modifiers, nil
}

// ApplyProductionTimeModifiers применяет модификаторы скорости к времени производства
func (s *ModifierService) ApplyProductionTimeModifiers(baseTime int, modifiers []Modifier) ModificationResult {
	speedBonus := 0.0
	var appliedModifiers []AppliedModifierInfo

	for _, modifier := range modifiers {
		if modifier.Type == ModifierTypeSpeed {
			speedBonus += modifier.Value
			appliedModifiers = append(appliedModifiers, AppliedModifierInfo{
				ModifierID:  modifier.ID,
				Type:        modifier.Type,
				Source:      modifier.Source,
				Value:       modifier.Value,
				Description: modifier.Description,
				AppliedAt:   "production_time",
			})
		}
	}

	// Применяем скоростные бонусы
	finalTime := float64(baseTime) * (1 - speedBonus)
	if finalTime < 1 {
		finalTime = 1
	}

	return ModificationResult{
		OriginalValue:    baseTime,
		ModifiedValue:    int(finalTime),
		AppliedModifiers: appliedModifiers,
	}
}

// ApplyQuantityModifiers применяет модификаторы количества к диапазону выпадений
func (s *ModifierService) ApplyQuantityModifiers(minQty, maxQty int, modifiers []Modifier) ModificationResult {
	quantityBonus := 0.0
	var appliedModifiers []AppliedModifierInfo

	for _, modifier := range modifiers {
		if modifier.Type == ModifierTypeQuantity {
			quantityBonus += modifier.Value
			appliedModifiers = append(appliedModifiers, AppliedModifierInfo{
				ModifierID:  modifier.ID,
				Type:        modifier.Type,
				Source:      modifier.Source,
				Value:       modifier.Value,
				Description: modifier.Description,
				AppliedAt:   "quantity",
			})
		}
	}

	// Применяем бонусы количества
	finalMinQty := int(math.Max(1, float64(minQty)*(1+quantityBonus)))
	finalMaxQty := int(math.Max(1, float64(maxQty)*(1+quantityBonus)))

	return ModificationResult{
		OriginalValue:    map[string]int{"min": minQty, "max": maxQty},
		ModifiedValue:    map[string]int{"min": finalMinQty, "max": finalMaxQty},
		AppliedModifiers: appliedModifiers,
	}
}

// ApplyProbabilityModifiers применяет модификаторы вероятности к шансам выпадения
func (s *ModifierService) ApplyProbabilityModifiers(baseProbability int, modifiers []Modifier) ModificationResult {
	probabilityBonus := 0.0
	var appliedModifiers []AppliedModifierInfo

	for _, modifier := range modifiers {
		if modifier.Type == ModifierTypeProbability {
			probabilityBonus += modifier.Value
			appliedModifiers = append(appliedModifiers, AppliedModifierInfo{
				ModifierID:  modifier.ID,
				Type:        modifier.Type,
				Source:      modifier.Source,
				Value:       modifier.Value,
				Description: modifier.Description,
				AppliedAt:   "probability",
			})
		}
	}

	// Применяем бонусы вероятности
	finalProbability := int(math.Min(100, float64(baseProbability)*(1+probabilityBonus)))

	return ModificationResult{
		OriginalValue:    baseProbability,
		ModifiedValue:    finalProbability,
		AppliedModifiers: appliedModifiers,
	}
}

// ApplyCostReductionModifiers применяет модификаторы снижения затрат к входным предметам
func (s *ModifierService) ApplyCostReductionModifiers(inputItems []models.RecipeInputItem, modifiers []Modifier) ModificationResult {
	costReduction := 0.0
	var appliedModifiers []AppliedModifierInfo

	for _, modifier := range modifiers {
		if modifier.Type == ModifierTypeCostReduction {
			costReduction += modifier.Value
			appliedModifiers = append(appliedModifiers, AppliedModifierInfo{
				ModifierID:  modifier.ID,
				Type:        modifier.Type,
				Source:      modifier.Source,
				Value:       modifier.Value,
				Description: modifier.Description,
				AppliedAt:   "cost_reduction",
			})
		}
	}

	// Применяем снижение затрат
	modifiedItems := make([]models.RecipeInputItem, len(inputItems))
	for i, item := range inputItems {
		modifiedItems[i] = item
		finalQuantity := int(math.Max(1, float64(item.Quantity)*(1-costReduction)))
		modifiedItems[i].Quantity = finalQuantity
	}

	return ModificationResult{
		OriginalValue:    inputItems,
		ModifiedValue:    modifiedItems,
		AppliedModifiers: appliedModifiers,
	}
}

// getUserModifiers получает пользовательские модификаторы из User Service
func (s *ModifierService) getUserModifiers(ctx context.Context, userID uuid.UUID) ([]Modifier, error) {
	userMods, err := s.userClient.GetUserProductionModifiers(ctx, userID)
	if err != nil {
		return nil, err
	}

	var modifiers []Modifier

	// VIP модификаторы
	if userMods.Modifiers.VIPStatus.ProductionSpeedBonus > 0 {
		modifiers = append(modifiers, Modifier{
			ID:           uuid.New(),
			Type:         ModifierTypeSpeed,
			Source:       ModifierSourceUser,
			Value:        userMods.Modifiers.VIPStatus.ProductionSpeedBonus,
			IsMultiplier: false,
			Description:  fmt.Sprintf("VIP %s speed bonus", userMods.Modifiers.VIPStatus.Level),
		})
	}

	if userMods.Modifiers.VIPStatus.QualityBonus > 0 {
		modifiers = append(modifiers, Modifier{
			ID:           uuid.New(),
			Type:         ModifierTypeQuality,
			Source:       ModifierSourceUser,
			Value:        userMods.Modifiers.VIPStatus.QualityBonus,
			IsMultiplier: false,
			Description:  fmt.Sprintf("VIP %s quality bonus", userMods.Modifiers.VIPStatus.Level),
		})
	}

	// Модификаторы уровня персонажа
	if userMods.Modifiers.CharacterLevel.CraftingBonus > 0 {
		modifiers = append(modifiers, Modifier{
			ID:           uuid.New(),
			Type:         ModifierTypeProbability,
			Source:       ModifierSourceUser,
			Value:        userMods.Modifiers.CharacterLevel.CraftingBonus,
			IsMultiplier: false,
			Description:  fmt.Sprintf("Character level %d crafting bonus", userMods.Modifiers.CharacterLevel.Level),
		})
	}

	// Клановые бонусы
	if userMods.Modifiers.ClanBonuses.ProductionSpeed > 0 {
		modifiers = append(modifiers, Modifier{
			ID:           uuid.New(),
			Type:         ModifierTypeSpeed,
			Source:       ModifierSourceClan,
			Value:        userMods.Modifiers.ClanBonuses.ProductionSpeed,
			IsMultiplier: false,
			Description:  "Clan production speed bonus",
		})
	}

	return modifiers, nil
}

// getBoosterModifiers получает модификаторы от ускорителей-предметов
func (s *ModifierService) getBoosterModifiers(ctx context.Context, boosters []models.BoosterItem) ([]Modifier, error) {
	var modifiers []Modifier

	// TODO: Здесь должна быть логика получения модификаторов от предметов
	// Для этого нужна интеграция с inventory-service для получения свойств предметов
	// В текущей версии возвращаем базовые модификаторы

	for _, booster := range boosters {
		// Заглушка: предполагаем что все ускорители дают 10% бонус скорости
		modifiers = append(modifiers, Modifier{
			ID:           uuid.New(),
			Type:         ModifierTypeSpeed,
			Source:       ModifierSourceBooster,
			Value:        0.1, // 10% бонус скорости
			IsMultiplier: false,
			ItemID:       &booster.ItemID,
			Description:  fmt.Sprintf("Booster item speed bonus (x%d)", booster.Quantity),
		})
	}

	return modifiers, nil
}

// getAutomaticModifiers получает автоматические модификаторы (серверные буферы)
func (s *ModifierService) getAutomaticModifiers(ctx context.Context, userID uuid.UUID) []Modifier {
	var modifiers []Modifier

	// TODO: Здесь должна быть логика получения серверных буферов и событийных модификаторов
	// В текущей версии возвращаем пустой список

	// Пример серверного бафа (можно настроить через конфигурацию)
	// modifiers = append(modifiers, Modifier{
	//     ID:           uuid.New(),
	//     Type:         ModifierTypeSpeed,
	//     Source:       ModifierSourceServer,
	//     Value:        0.05, // 5% глобальный бонус скорости
	//     IsMultiplier: false,
	//     Description:  "Server-wide production boost",
	// })

	return modifiers
}

// BuildAppliedModifiersForAudit формирует JSON с примененными модификаторами для аудита
func (s *ModifierService) BuildAppliedModifiersForAudit(results []ModificationResult) models.AppliedModifiers {
	audit := make(models.AppliedModifiers)

	for i, result := range results {
		key := fmt.Sprintf("modification_%d", i)
		audit[key] = map[string]interface{}{
			"original_value":    result.OriginalValue,
			"modified_value":    result.ModifiedValue,
			"applied_modifiers": result.AppliedModifiers,
		}
	}

	return audit
}

// LogModifiersApplication логирует применение модификаторов для отладки
func (s *ModifierService) LogModifiersApplication(userID uuid.UUID, taskID uuid.UUID, results []ModificationResult) {
	for _, result := range results {
		if len(result.AppliedModifiers) > 0 {
			resultJSON, _ := json.Marshal(result)
			s.logger.Info("Applied modifiers",
				zap.String("userID", userID.String()),
				zap.String("taskID", taskID.String()),
				zap.String("modification_result", string(resultJSON)),
			)
		}
	}
}

package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProductionRecipe представляет производственный рецепт
type ProductionRecipe struct {
	ID                    uuid.UUID           `json:"id" db:"id"`
	Name                  string              `json:"name" db:"name"`
	OperationClassCode    string              `json:"operation_class_code" db:"operation_class_code"`
	Description           *string             `json:"description,omitempty" db:"description"`
	IsActive              bool                `json:"is_active" db:"is_active"`
	ProductionTimeSeconds int                 `json:"production_time_seconds" db:"production_time_seconds"`
	CreatedAt             time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at" db:"updated_at"`
	
	// Связанные данные (заполняется при необходимости)
	InputItems  []RecipeInputItem  `json:"input_items,omitempty"`
	OutputItems []RecipeOutputItem `json:"output_items,omitempty"`
	Limits      []RecipeLimit      `json:"limits,omitempty"`
}

// RecipeInputItem представляет входной предмет рецепта
type RecipeInputItem struct {
	RecipeID         uuid.UUID  `json:"recipe_id" db:"recipe_id"`
	ItemID           uuid.UUID  `json:"item_id" db:"item_id"`
	CollectionCode   *string    `json:"collection_code,omitempty" db:"collection_code"`
	QualityLevelCode *string    `json:"quality_level_code,omitempty" db:"quality_level_code"`
	Quantity         int        `json:"quantity" db:"quantity"`
	
	// Внутренние поля после преобразования кодов в UUID
	CollectionID   *uuid.UUID `json:"-" db:"collection_id"`
	QualityLevelID *uuid.UUID `json:"-" db:"quality_level_id"`
}

// RecipeOutputItem представляет выходной предмет рецепта
type RecipeOutputItem struct {
	RecipeID                  uuid.UUID  `json:"recipe_id" db:"recipe_id"`
	ItemID                    uuid.UUID  `json:"item_id" db:"item_id"`
	MinQuantity               int        `json:"min_quantity" db:"min_quantity"`
	MaxQuantity               int        `json:"max_quantity" db:"max_quantity"`
	ProbabilityPercent        int        `json:"probability_percent" db:"probability_percent"`
	OutputGroup               *string    `json:"output_group,omitempty" db:"output_group"`
	CollectionSourceInputIndex *int      `json:"collection_source_input_index,omitempty" db:"collection_source_input_index"`
	QualitySourceInputIndex   *int       `json:"quality_source_input_index,omitempty" db:"quality_source_input_index"`
	FixedCollectionCode       *string    `json:"fixed_collection_code,omitempty" db:"fixed_collection_code"`
	FixedQualityLevelCode     *string    `json:"fixed_quality_level_code,omitempty" db:"fixed_quality_level_code"`
	
	// Внутренние поля после преобразования кодов в UUID
	FixedCollectionID   *uuid.UUID `json:"-" db:"fixed_collection_id"`
	FixedQualityLevelID *uuid.UUID `json:"-" db:"fixed_quality_level_id"`
}

// RecipeLimit представляет лимит использования рецепта
type RecipeLimit struct {
	RecipeID      uuid.UUID  `json:"recipe_id" db:"recipe_id"`
	LimitType     string     `json:"limit_type" db:"limit_type"`
	LimitObject   string     `json:"limit_object" db:"limit_object"`
	TargetItemID  *uuid.UUID `json:"target_item_id,omitempty" db:"target_item_id"`
	LimitQuantity int        `json:"limit_quantity" db:"limit_quantity"`
}

// RecipeFilters представляет фильтры для поиска рецептов
type RecipeFilters struct {
	OperationClassCode *string `json:"operation_class_code,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
	UserID             *uuid.UUID `json:"user_id,omitempty"` // для проверки лимитов
}

// RecipeUsageStats представляет статистику использования рецепта пользователем
type RecipeUsageStats struct {
	RecipeID      uuid.UUID `json:"recipe_id" db:"recipe_id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	LimitType     string    `json:"limit_type" db:"limit_type"`
	LimitObject   string    `json:"limit_object" db:"limit_object"`
	TargetItemID  *uuid.UUID `json:"target_item_id,omitempty" db:"target_item_id"`
	CurrentUsage  int       `json:"current_usage" db:"current_usage"`
	LimitQuantity int       `json:"limit_quantity" db:"limit_quantity"`
	PeriodStart   time.Time `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time `json:"period_end" db:"period_end"`
}

// AppliedModifiers представляет примененные модификаторы (сохраняется как JSON)
type AppliedModifiers map[string]interface{}

// Value реализует driver.Valuer для AppliedModifiers
func (a AppliedModifiers) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan реализует sql.Scanner для AppliedModifiers
func (a *AppliedModifiers) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, a)
}

// ProductionTask представляет производственное задание
type ProductionTask struct {
	ID                    uuid.UUID         `json:"id" db:"id"`
	UserID                uuid.UUID         `json:"user_id" db:"user_id"`
	RecipeID              uuid.UUID         `json:"recipe_id" db:"recipe_id"`
	OperationClassCode    string            `json:"operation_class_code" db:"operation_class_code"`
	Status                string            `json:"status" db:"status"`
	ProductionTimeSeconds int               `json:"production_time_seconds" db:"production_time_seconds"`
	CreatedAt             time.Time         `json:"created_at" db:"created_at"`
	StartedAt             *time.Time        `json:"started_at,omitempty" db:"started_at"`
	CompletedAt           *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	AppliedModifiers      AppliedModifiers  `json:"applied_modifiers,omitempty" db:"applied_modifiers"`
	
	// Связанные данные
	OutputItems []TaskOutputItem `json:"output_items,omitempty"`
	Recipe      *ProductionRecipe `json:"recipe,omitempty"`
}

// TaskOutputItem представляет выходной предмет задания (предрасчитанный результат)
type TaskOutputItem struct {
	TaskID           uuid.UUID  `json:"task_id" db:"task_id"`
	ItemID           uuid.UUID  `json:"item_id" db:"item_id"`
	CollectionCode   *string    `json:"collection_code,omitempty" db:"collection_code"`
	QualityLevelCode *string    `json:"quality_level_code,omitempty" db:"quality_level_code"`
	Quantity         int        `json:"quantity" db:"quantity"`
	
	// Внутренние поля после преобразования кодов в UUID
	CollectionID   *uuid.UUID `json:"-" db:"collection_id"`
	QualityLevelID *uuid.UUID `json:"-" db:"quality_level_id"`
}

// Constants для статусов заданий
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusClaimed    = "claimed"
	TaskStatusCancelled  = "cancelled"
	TaskStatusFailed     = "failed"
)

// Constants для типов лимитов
const (
	LimitTypeTotal      = "total"
	LimitTypeDaily      = "daily"
	LimitTypeWeekly     = "weekly"
	LimitTypeMonthly    = "monthly"
	LimitTypeSeasonal   = "seasonal"
	LimitTypePerEvent   = "per_event"
)

// Constants для объектов лимитов
const (
	LimitObjectRecipeExecution = "recipe_execution"
	LimitObjectItemReward      = "item_reward"
)

// Constants для классов операций
const (
	OperationClassCrafting         = "crafting"
	OperationClassSmelting         = "smelting"
	OperationClassChestOpening     = "chest_opening"
	OperationClassResourceGathering = "resource_gathering"
	OperationClassSpecial          = "special"
)
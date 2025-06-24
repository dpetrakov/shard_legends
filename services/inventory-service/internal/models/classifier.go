package models

import (
	"time"

	"github.com/google/uuid"
)

// Classifier represents a classifier type in the system
type Classifier struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ClassifierItem represents an item within a classifier
type ClassifierItem struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ClassifierID uuid.UUID  `json:"classifier_id" db:"classifier_id"`
	Code         string     `json:"code" db:"code"`
	Description  *string    `json:"description,omitempty" db:"description"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Common classifier codes used across the system
const (
	ClassifierItemClass         = "item_class"
	ClassifierQualityLevel      = "quality_level"
	ClassifierCollection        = "collection"
	ClassifierInventorySection  = "inventory_section"
	ClassifierOperationType     = "operation_type"
	ClassifierResourceType      = "resource_type"
	ClassifierReagentType       = "reagent_type"
	ClassifierBoosterType       = "booster_type"
	ClassifierToolType          = "tool_type"
	ClassifierKeyType           = "key_type"
	ClassifierCurrencyType      = "currency_type"
	ClassifierToolQualityLevels = "tool_quality_levels"
	ClassifierKeyQualityLevels  = "key_quality_levels"
)

// Common classifier item codes
const (
	// Item classes
	ItemClassResources   = "resources"
	ItemClassReagents    = "reagents"
	ItemClassBoosters    = "boosters"
	ItemClassBlueprints  = "blueprints"
	ItemClassTools       = "tools"
	ItemClassKeys        = "keys"
	ItemClassCurrencies  = "currencies"

	// Quality levels
	QualityLevelWooden   = "wooden"
	QualityLevelStone    = "stone"
	QualityLevelMetal    = "metal"
	QualityLevelDiamond  = "diamond"
	QualityLevelSmall    = "small"
	QualityLevelMedium   = "medium"
	QualityLevelLarge    = "large"

	// Inventory sections
	SectionMain    = "main"
	SectionFactory = "factory"
	SectionTrade   = "trade"

	// Operation types
	OperationTypeChestReward        = "chest_reward"
	OperationTypeCraftResult        = "craft_result"
	OperationTypeTradeSale          = "trade_sale"
	OperationTypeTradePurchase      = "trade_purchase"
	OperationTypeAdminAdjustment    = "admin_adjustment"
	OperationTypeSystemReward       = "system_reward"
	OperationTypeSystemPenalty      = "system_penalty"
	OperationTypeDailyQuestReward   = "daily_quest_reward"
	OperationTypeFactoryReservation = "factory_reservation"
	OperationTypeFactoryReturn      = "factory_return"
	OperationTypeFactoryConsumption = "factory_consumption"
	OperationTypeTradeListing       = "trade_listing"
	OperationTypeTradeDelisting     = "trade_delisting"
)
package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a game item definition
type Item struct {
	ID                        uuid.UUID `json:"id" db:"id"`
	ItemClassID               uuid.UUID `json:"item_class_id" db:"item_class_id"`
	ItemTypeID                uuid.UUID `json:"item_type_id" db:"item_type_id"`
	QualityLevelsClassifierID uuid.UUID `json:"quality_levels_classifier_id" db:"quality_levels_classifier_id"`
	CollectionsClassifierID   uuid.UUID `json:"collections_classifier_id" db:"collections_classifier_id"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

// ItemImage represents an image for a specific item variant
type ItemImage struct {
	ItemID         uuid.UUID `json:"item_id" db:"item_id"`
	CollectionID   uuid.UUID `json:"collection_id" db:"collection_id"`
	QualityLevelID uuid.UUID `json:"quality_level_id" db:"quality_level_id"`
	ImageURL       string    `json:"image_url" db:"image_url"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ItemWithDetails represents an item with its classifier details loaded
type ItemWithDetails struct {
	Item
	ItemClass    string `json:"item_class" db:"item_class"`
	ItemType     string `json:"item_type" db:"item_type"`
}

// Common item type codes
const (
	// Resource types
	ResourceTypeStone   = "stone"
	ResourceTypeWood    = "wood"
	ResourceTypeOre     = "ore"
	ResourceTypeDiamond = "diamond"

	// Reagent types
	ReagentTypeAbrasive = "abrasive"
	ReagentTypeDisc     = "disc"
	ReagentTypeInductor = "inductor"
	ReagentTypePaste    = "paste"

	// Booster types
	BoosterTypeRepairTool      = "repair_tool"
	BoosterTypeSpeedProcessing = "speed_processing"
	BoosterTypeSpeedCrafting   = "speed_crafting"

	// Tool types
	ToolTypeShovel  = "shovel"
	ToolTypeSickle  = "sickle"
	ToolTypeAxe     = "axe"
	ToolTypePickaxe = "pickaxe"

	// Key types
	KeyTypeKey          = "key"
	KeyTypeBlueprintKey = "blueprint_key"

	// Currency types
	CurrencyTypeDiamonds = "diamonds"
)
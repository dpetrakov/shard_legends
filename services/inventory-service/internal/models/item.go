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
	ItemClass string `json:"item_class" db:"item_class"`
	ItemType  string `json:"item_type" db:"item_type"`
}

// Translation represents an i18n translation for any entity
type Translation struct {
	ID           uuid.UUID `json:"id" db:"id"`
	EntityType   string    `json:"entity_type" db:"entity_type"`
	EntityID     uuid.UUID `json:"entity_id" db:"entity_id"`
	FieldName    string    `json:"field_name" db:"field_name"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	Content      string    `json:"content" db:"content"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Language represents a supported language
type Language struct {
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ItemDetailsRequest represents a request for item details
type ItemDetailsRequest struct {
	Items []ItemDetailRequestItem `json:"items" validate:"required,min=1,max=100"`
}

// ItemDetailRequestItem represents a single item in the details request
type ItemDetailRequestItem struct {
	ItemID       uuid.UUID `json:"item_id" validate:"required"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
}

// ItemDetailsResponse represents the response with item details
type ItemDetailsResponse struct {
	Items []ItemDetailResponseItem `json:"items"`
}

// ItemDetailResponseItem represents a single item with its localized details
type ItemDetailResponseItem struct {
	ItemID       uuid.UUID `json:"item_id"`
	ItemClass    string    `json:"item_class"`
	ItemType     string    `json:"item_type"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
}

// I18n entity types
const (
	EntityTypeItem           = "item"
	EntityTypeClassifier     = "classifier"
	EntityTypeClassifierItem = "classifier_item"
	EntityTypeAchievement    = "achievement"
	EntityTypeRecipe         = "recipe"
	EntityTypeUserMessage    = "user_message"
)

// I18n field names
const (
	FieldNameName        = "name"
	FieldNameDescription = "description"
	FieldNameTooltip     = "tooltip"
	FieldNameTitle       = "title"
	FieldNameContent     = "content"
)

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

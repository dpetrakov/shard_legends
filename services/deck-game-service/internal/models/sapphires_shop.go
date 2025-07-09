package models

import "github.com/google/uuid"

// SapphiresShopRecipe represents a recipe for purchasing items with sapphires
type SapphiresShopRecipe struct {
	RecipeID uuid.UUID            `json:"recipe_id"`
	Code     string               `json:"code"`
	Input    []SapphiresShopItem  `json:"input"`
	Output   SapphiresShopItem    `json:"output"`
}

// SapphiresShopItem represents an item in the sapphires shop
type SapphiresShopItem struct {
	ItemID           uuid.UUID `json:"item_id"`
	Code             string    `json:"code"`
	CollectionCode   *string   `json:"collection_code,omitempty"`
	QualityLevelCode *string   `json:"quality_level_code,omitempty"`
	Quantity         *int      `json:"quantity,omitempty"`         // For input items
	MinQuantity      *int      `json:"min_quantity,omitempty"`      // For output items
	MaxQuantity      *int      `json:"max_quantity,omitempty"`      // For output items
}

package models

import (
	"errors"

	"github.com/google/uuid"
)

// BuyItemRequest represents the request to buy an item
type BuyItemRequest struct {
	RecipeID         *uuid.UUID `json:"recipe_id,omitempty"`
	ItemCode         *string    `json:"item_code,omitempty"`
	QualityLevelCode *string    `json:"quality_level_code,omitempty"`
	CollectionCode   *string    `json:"collection_code,omitempty"`
	Quantity         int        `json:"quantity"`
}

// BuyItemResponse represents the response from buying an item
type BuyItemResponse struct {
	Items []ItemInfo `json:"items"`
}


// Validate validates the BuyItemRequest
func (r *BuyItemRequest) Validate() error {
	// Check that either recipe_id or item_code is provided, but not both
	if r.RecipeID != nil && r.ItemCode != nil {
		return errors.New("only one of recipe_id or item_code should be provided")
	}

	if r.RecipeID == nil && r.ItemCode == nil {
		return errors.New("either recipe_id or item_code must be provided")
	}

	// If using item_code, it should not be empty
	if r.ItemCode != nil && *r.ItemCode == "" {
		return errors.New("item_code cannot be empty")
	}

	// Validate quantity
	if r.Quantity < 1 {
		return errors.New("quantity must be at least 1")
	}

	return nil
}

package models

import "fmt"

// OpenChestRequest represents the request for POST /deck/chest/open
type OpenChestRequest struct {
	ChestType    string `json:"chest_type" binding:"required,oneof=resource_chest reagent_chest booster_chest blueprint_chest"`
	QualityLevel string `json:"quality_level" binding:"required,oneof=small medium large"`
	Quantity     *int   `json:"quantity,omitempty" binding:"omitempty,min=1,max=100"`
	OpenAll      *bool  `json:"open_all,omitempty"`
}

// Validate performs custom validation for mutually exclusive quantity/open_all fields
func (r *OpenChestRequest) Validate() error {
	// Exactly one of quantity or open_all must be set
	if (r.Quantity == nil && r.OpenAll == nil) || (r.Quantity != nil && r.OpenAll != nil) {
		return fmt.Errorf("exactly one of 'quantity' or 'open_all' must be specified")
	}

	// If open_all is set, it must be true
	if r.OpenAll != nil && !*r.OpenAll {
		return fmt.Errorf("open_all must be true when specified")
	}

	return nil
}

// OpenChestResponse represents the response for POST /deck/chest/open
type OpenChestResponse struct {
	Items          []ItemInfo `json:"items"`
	QuantityOpened int        `json:"quantity_opened"`
}

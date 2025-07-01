package models

import (
	"time"
)

// StatusResponse represents the response for GET /deck/daily-chest/status
type StatusResponse struct {
	ExpectedCombo int        `json:"expected_combo,omitempty"`
	Finished      bool       `json:"finished"`
	CraftsDone    int        `json:"crafts_done"`
	LastRewardAt  *time.Time `json:"last_reward_at,omitempty"`
}

// ClaimRequest represents the request for POST /deck/daily-chest/claim
type ClaimRequest struct {
	Combo        int   `json:"combo" binding:"required,min=5,max=15"`
	ChestIndices []int `json:"chest_indices" binding:"required,min=1,dive,min=1,max=6"`
}

// ItemInfo represents detailed information about an item
type ItemInfo struct {
	ItemID       string  `json:"item_id"`
	ItemClass    string  `json:"item_class"`
	ItemType     string  `json:"item_type"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	Collection   *string `json:"collection,omitempty"`
	QualityLevel *string `json:"quality_level,omitempty"`
	Quantity     int     `json:"quantity"`
}

// ClaimResponse represents the response for POST /deck/daily-chest/claim
type ClaimResponse struct {
	Items             []ItemInfo `json:"items"`
	NextExpectedCombo int        `json:"next_expected_combo"`
	CraftsDone        int        `json:"crafts_done"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/deck-game-service/internal/models"
)

// DeckGameService defines the main business logic interface for deck game operations
type DeckGameService interface {
	// GetDailyChestStatus returns the current status of daily chest rewards for a user
	GetDailyChestStatus(ctx context.Context, userID uuid.UUID) (*models.StatusResponse, error)

	// ClaimDailyChest processes the daily chest claim request
	ClaimDailyChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.ClaimRequest) (*models.ClaimResponse, error)
}

// DailyChestRepository defines the interface for daily chest data access
type DailyChestRepository interface {
	// GetCompletedTasksCountToday returns the count of completed daily chest tasks for user today
	GetCompletedTasksCountToday(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (int, error)

	// GetLastRewardTime returns the time of the last completed daily chest task for user
	GetLastRewardTime(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (*time.Time, error)
}

// ProductionClient defines the interface for Production Service integration
type ProductionClient interface {
	// StartProduction starts a production task
	StartProduction(ctx context.Context, jwtToken string, userID uuid.UUID, recipeID uuid.UUID, executionCount int) (*ProductionStartResponse, error)

	// ClaimProduction claims the results of a completed production task
	ClaimProduction(ctx context.Context, jwtToken string, userID uuid.UUID, taskID uuid.UUID) (*ProductionClaimResponse, error)
}

// InventoryClient defines the interface for Inventory Service integration
type InventoryClient interface {
	// GetItemsDetails returns detailed information about items
	GetItemsDetails(ctx context.Context, jwtToken string, items []ItemDetailsRequest, lang string) (*ItemDetailsResponse, error)
}

// ProductionStartResponse represents the response from Production Service start endpoint
type ProductionStartResponse struct {
	TaskID uuid.UUID `json:"task_id"`
	Status string    `json:"status"`
}

// ProductionClaimResponse represents the response from Production Service claim endpoint
type ProductionClaimResponse struct {
	Success       bool             `json:"success"`
	ItemsReceived []TaskOutputItem `json:"items_received"`
}

// TaskOutputItem represents an item received from production
type TaskOutputItem struct {
	ItemID       uuid.UUID `json:"item_id"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
	Quantity     int       `json:"quantity"`
}

// ItemDetailsRequest represents a request for item details
type ItemDetailsRequest struct {
	ItemID       uuid.UUID `json:"item_id"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
}

// ItemDetailsResponse represents the response from Inventory Service items/details endpoint
type ItemDetailsResponse struct {
	Items []ItemDetails `json:"items"`
}

// ItemDetails represents detailed information about an item from Inventory Service
type ItemDetails struct {
	ItemID       uuid.UUID `json:"item_id"`
	ItemClass    string    `json:"item_class"`
	ItemType     string    `json:"item_type"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
}

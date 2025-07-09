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

	// OpenChest opens specified chests and returns the aggregated items received
	OpenChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.OpenChestRequest) (*models.OpenChestResponse, error)

	// BuyItem processes the buy item request
	BuyItem(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.BuyItemRequest) (*models.BuyItemResponse, error)

	// GetSapphiresShopItems returns list of items available for purchase with sapphires
	GetSapphiresShopItems(ctx context.Context) ([]models.SapphiresShopRecipe, error)
}

// DailyChestRepository defines the interface for daily chest data access
type DailyChestRepository interface {
	// GetCompletedTasksCountToday returns the count of completed daily chest tasks for user today
	GetCompletedTasksCountToday(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (int, error)

	// GetLastRewardTime returns the time of the last completed daily chest task for user
	GetLastRewardTime(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (*time.Time, error)
}

// RecipeRepository defines the interface for recipe data access
type RecipeRepository interface {
	// FindRecipesByOutputItem finds recipes by output item characteristics
	FindRecipesByOutputItem(ctx context.Context, itemCode string, qualityLevelCode, collectionCode *string) ([]Recipe, error)

	// GetSapphireShopRecipes returns recipes that can be purchased with sapphires only
	GetSapphireShopRecipes(ctx context.Context) ([]SapphireShopRecipe, error)
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

	// GetItemQuantity returns the quantity of a specific item in the user's inventory
	GetItemQuantity(ctx context.Context, jwtToken string, itemID uuid.UUID) (int, error)

	// GetInventory returns the full inventory for the authenticated user
	GetInventory(ctx context.Context, jwtToken string) ([]InventoryItem, error)
}

// InventoryItem represents a single item in the user's inventory
type InventoryItem struct {
	ItemID       uuid.UUID `json:"item_id"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
	Quantity     int       `json:"quantity"`
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
	Collection   *string   `json:"collection_code,omitempty"`
	QualityLevel *string   `json:"quality_level_code,omitempty"`
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

// Recipe represents a recipe from the database
type Recipe struct {
	ID                   uuid.UUID `json:"id"`
	Code                 string    `json:"code"`
	OperationClassCode   string    `json:"operation_class_code"`
	ProductionTimeSeconds int      `json:"production_time_seconds"`
	IsActive             bool      `json:"is_active"`
	FixedQualityLevel    *string   `json:"fixed_quality_level,omitempty"`
	FixedCollectionCode  *string   `json:"fixed_collection_code,omitempty"`
}

// SapphireShopRecipe represents a recipe that can be purchased with sapphires
type SapphireShopRecipe struct {
	RecipeID uuid.UUID               `json:"recipe_id"`
	Code     string                  `json:"code"`
	Input    []SapphireShopItemInfo  `json:"input"`
	Output   SapphireShopItemInfo    `json:"output"`
}

// SapphireShopItemInfo represents item information for sapphire shop
type SapphireShopItemInfo struct {
	ItemID           uuid.UUID `json:"item_id"`
	Code             string    `json:"code"`
	CollectionCode   *string   `json:"collection_code,omitempty"`
	QualityLevelCode *string   `json:"quality_level_code,omitempty"`
	Quantity         *int      `json:"quantity,omitempty"`         // For input items
	MinQuantity      *int      `json:"min_quantity,omitempty"`      // For output items
	MaxQuantity      *int      `json:"max_quantity,omitempty"`      // For output items
}

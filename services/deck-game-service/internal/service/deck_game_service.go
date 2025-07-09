package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/shard-legends/deck-game-service/internal/config"
	"github.com/shard-legends/deck-game-service/internal/models"
)

// Constants for daily chest business logic
const (
	BASE_COMBO       = 5
	MAX_COMBO        = 15
	MAX_DAILY_CRAFTS = 10
)

// deckGameService implements DeckGameService interface
type deckGameService struct {
	repo             DailyChestRepository
	recipeRepo       RecipeRepository
	productionClient ProductionClient
	inventoryClient  InventoryClient
	config           *config.Config
	logger           *slog.Logger
}

// NewDeckGameService creates a new deck game service
func NewDeckGameService(
	repo DailyChestRepository,
	recipeRepo RecipeRepository,
	productionClient ProductionClient,
	inventoryClient InventoryClient,
	cfg *config.Config,
	logger *slog.Logger,
) DeckGameService {
	return &deckGameService{
		repo:             repo,
		recipeRepo:       recipeRepo,
		productionClient: productionClient,
		inventoryClient:  inventoryClient,
		config:           cfg,
		logger:           logger,
	}
}

// GetDailyChestStatus returns the current status of daily chest rewards for a user
func (s *deckGameService) GetDailyChestStatus(ctx context.Context, userID uuid.UUID) (*models.StatusResponse, error) {
	// Parse recipe ID from config
	recipeID, err := uuid.Parse(s.config.DailyChestRecipeID)
	if err != nil {
		s.logger.Error("Invalid daily chest recipe ID in config", "recipe_id", s.config.DailyChestRecipeID, "error", err)
		return nil, errors.Wrap(err, "invalid recipe ID configuration")
	}

	// Get completed tasks count for today
	craftsCount, err := s.repo.GetCompletedTasksCountToday(ctx, userID, recipeID)
	if err != nil {
		s.logger.Error("Failed to get completed tasks count", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get completed tasks count")
	}

	// Get last reward time
	lastRewardAt, err := s.repo.GetLastRewardTime(ctx, userID, recipeID)
	if err != nil {
		s.logger.Error("Failed to get last reward time", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get last reward time")
	}

	// Check if user has reached daily limit
	finished := craftsCount >= MAX_DAILY_CRAFTS

	// Calculate expected combo: BASE_COMBO + crafts_count (5 + 0..9 = 5..14)
	// If finished, we don't include expected_combo in response
	var expectedCombo *int
	if !finished {
		combo := BASE_COMBO + craftsCount
		if combo > MAX_COMBO {
			combo = MAX_COMBO
		}
		expectedCombo = &combo
	}

	response := &models.StatusResponse{
		Finished:     finished,
		CraftsDone:   craftsCount,
		LastRewardAt: lastRewardAt,
	}

	// Only include expected_combo if not finished
	if expectedCombo != nil {
		response.ExpectedCombo = *expectedCombo
	}

	s.logger.Info("Daily chest status calculated",
		"user_id", userID,
		"crafts_done", craftsCount,
		"expected_combo", expectedCombo,
		"finished", finished,
		"last_reward_at", lastRewardAt)

	return response, nil
}

// ClaimDailyChest processes the daily chest claim request
func (s *deckGameService) ClaimDailyChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.ClaimRequest) (*models.ClaimResponse, error) {
	// Parse recipe ID from config
	recipeID, err := uuid.Parse(s.config.DailyChestRecipeID)
	if err != nil {
		s.logger.Error("Invalid daily chest recipe ID in config", "recipe_id", s.config.DailyChestRecipeID, "error", err)
		return nil, errors.Wrap(err, "invalid recipe ID configuration")
	}

	// Get current status
	craftsCount, err := s.repo.GetCompletedTasksCountToday(ctx, userID, recipeID)
	if err != nil {
		s.logger.Error("Failed to get completed tasks count", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get completed tasks count")
	}

	// Check if user has reached daily limit
	if craftsCount >= MAX_DAILY_CRAFTS {
		s.logger.Warn("Daily limit reached", "user_id", userID, "crafts_count", craftsCount)
		return nil, fmt.Errorf("daily_finished")
	}

	// Calculate expected combo
	expectedCombo := BASE_COMBO + craftsCount
	if expectedCombo > MAX_COMBO {
		expectedCombo = MAX_COMBO
	}

	// Validate combo
	if request.Combo < expectedCombo {
		s.logger.Warn("Invalid combo",
			"user_id", userID,
			"provided_combo", request.Combo,
			"expected_combo", expectedCombo)
		return nil, fmt.Errorf("invalid_combo")
	}

	// Check cooldown
	lastRewardAt, err := s.repo.GetLastRewardTime(ctx, userID, recipeID)
	if err != nil {
		s.logger.Error("Failed to get last reward time", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get last reward time")
	}

	if lastRewardAt != nil {
		cooldownDuration := time.Duration(s.config.CooldownSec) * time.Second
		if time.Since(*lastRewardAt) < cooldownDuration {
			s.logger.Warn("Cooldown violation",
				"user_id", userID,
				"last_reward_at", lastRewardAt,
				"cooldown_sec", s.config.CooldownSec)
			return nil, fmt.Errorf("daily_finished")
		}
	}

	s.logger.Info("Starting daily chest claim process",
		"user_id", userID,
		"combo", request.Combo,
		"chest_indices", request.ChestIndices,
		"expected_combo", expectedCombo)

	// Start production task
	startResp, err := s.productionClient.StartProduction(ctx, jwtToken, userID, recipeID, 1)
	if err != nil {
		s.logger.Error("Failed to start production", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to start production")
	}

	// Claim production results immediately
	claimResp, err := s.productionClient.ClaimProduction(ctx, jwtToken, userID, startResp.TaskID)
	if err != nil {
		s.logger.Error("Failed to claim production", "user_id", userID, "task_id", startResp.TaskID, "error", err)
		return nil, errors.Wrap(err, "failed to claim production")
	}

	// Get detailed item information
	itemDetailsRequests := make([]ItemDetailsRequest, len(claimResp.ItemsReceived))
	for i, item := range claimResp.ItemsReceived {
		s.logger.Info("Processing item from production",
			"user_id", userID,
			"item_id", item.ItemID,
			"collection", item.Collection,
			"quality_level", item.QualityLevel,
			"quantity", item.Quantity)

		itemDetailsRequests[i] = ItemDetailsRequest{
			ItemID:       item.ItemID,
			Collection:   item.Collection,
			QualityLevel: item.QualityLevel,
		}
	}

	itemDetails, err := s.inventoryClient.GetItemsDetails(ctx, jwtToken, itemDetailsRequests, "ru")
	if err != nil {
		s.logger.Error("Failed to get item details", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get item details")
	}

	// Build response items
	responseItems := make([]models.ItemInfo, len(claimResp.ItemsReceived))
	for i, receivedItem := range claimResp.ItemsReceived {
		// Find corresponding item details
		var details *ItemDetails
		for _, detail := range itemDetails.Items {
			if detail.ItemID == receivedItem.ItemID {
				details = &detail
				break
			}
		}

		if details == nil {
			s.logger.Error("Item details not found", "item_id", receivedItem.ItemID)
			return nil, fmt.Errorf("item details not found for item %s", receivedItem.ItemID)
		}

		s.logger.Info("Item details retrieved",
			"user_id", userID,
			"item_id", details.ItemID,
			"details_collection", details.Collection,
			"details_quality_level", details.QualityLevel,
			"filtered_collection", filterBaseValue(details.Collection),
			"filtered_quality_level", filterBaseValue(details.QualityLevel))

		responseItems[i] = models.ItemInfo{
			ItemID:       details.ItemID.String(),
			ItemClass:    details.ItemClass,
			ItemType:     details.ItemType,
			Name:         details.Name,
			Description:  details.Description,
			ImageURL:     details.ImageURL,
			Collection:   filterBaseValue(receivedItem.Collection),
			QualityLevel: filterBaseValue(receivedItem.QualityLevel),
			Quantity:     receivedItem.Quantity,
		}
	}

	// Calculate next expected combo
	nextCraftsCount := craftsCount + 1
	nextExpectedCombo := BASE_COMBO + nextCraftsCount
	if nextExpectedCombo > MAX_COMBO {
		nextExpectedCombo = MAX_COMBO
	}

	response := &models.ClaimResponse{
		Items:             responseItems,
		NextExpectedCombo: nextExpectedCombo,
		CraftsDone:        nextCraftsCount,
	}

	s.logger.Info("Daily chest claimed successfully",
		"user_id", userID,
		"task_id", startResp.TaskID,
		"items_count", len(responseItems),
		"next_expected_combo", nextExpectedCombo)

	return response, nil
}

// OpenChest opens specified chests and returns the aggregated items received
func (s *deckGameService) OpenChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.OpenChestRequest) (*models.OpenChestResponse, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		s.logger.Error("Invalid open chest request", "user_id", userID, "error", err)
		return nil, fmt.Errorf("invalid_input: %w", err)
	}

	// Get recipe ID for this chest type and quality
	recipeID, err := GetRecipeIDForChest(request.ChestType, request.QualityLevel)
	if err != nil {
		s.logger.Error("Recipe not found for chest",
			"user_id", userID,
			"chest_type", request.ChestType,
			"quality_level", request.QualityLevel,
			"error", err)
		return nil, fmt.Errorf("recipe_not_found: %s", err.Error())
	}

	// Determine execution count
	var executionCount int
	if request.OpenAll != nil && *request.OpenAll {
		// Get chest item ID to check inventory
		chestItemID, err := GetItemIDForChest(request.ChestType, request.QualityLevel)
		if err != nil {
			s.logger.Error("Failed to get chest item ID", "user_id", userID, "error", err)
			return nil, fmt.Errorf("invalid_input: %s", err.Error())
		}

		// Fetch full inventory and count available chests
		inventoryItems, err := s.inventoryClient.GetInventory(ctx, jwtToken)
		if err != nil {
			s.logger.Error("Failed to fetch inventory for open_all", "user_id", userID, "error", err)
			return nil, errors.Wrap(err, "failed to fetch inventory")
		}

		var quantity int
		for _, item := range inventoryItems {
			if item.ItemID == chestItemID {
				quantity = item.Quantity
				break
			}
		}

		if quantity == 0 {
			s.logger.Warn("No chests available for open_all", "user_id", userID, "chest_type", request.ChestType, "quality_level", request.QualityLevel)
			return nil, fmt.Errorf("insufficient_chests")
		}

		executionCount = quantity
	} else {
		executionCount = *request.Quantity
	}

	s.logger.Info("Starting chest opening process",
		"user_id", userID,
		"chest_type", request.ChestType,
		"quality_level", request.QualityLevel,
		"execution_count", executionCount,
		"recipe_id", recipeID)

	// Start production task
	startResp, err := s.productionClient.StartProduction(ctx, jwtToken, userID, recipeID, executionCount)
	if err != nil {
		// Check if it's insufficient items error
		if err.Error() == "insufficient_items" {
			s.logger.Warn("Insufficient chests for opening", "user_id", userID, "chest_type", request.ChestType, "quality_level", request.QualityLevel)
			return nil, fmt.Errorf("insufficient_chests")
		}

		s.logger.Error("Failed to start chest opening production", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to start production")
	}

	// Claim production results immediately
	claimResp, err := s.productionClient.ClaimProduction(ctx, jwtToken, userID, startResp.TaskID)
	if err != nil {
		s.logger.Error("Failed to claim chest opening production", "user_id", userID, "task_id", startResp.TaskID, "error", err)
		return nil, errors.Wrap(err, "failed to claim production")
	}

	// Get detailed item information
	itemDetailsRequests := make([]ItemDetailsRequest, len(claimResp.ItemsReceived))
	for i, item := range claimResp.ItemsReceived {
		s.logger.Info("Processing received item",
			"user_id", userID,
			"item_id", item.ItemID,
			"collection", item.Collection,
			"quality_level", item.QualityLevel,
			"quantity", item.Quantity)

		itemDetailsRequests[i] = ItemDetailsRequest{
			ItemID:       item.ItemID,
			Collection:   item.Collection,
			QualityLevel: item.QualityLevel,
		}
	}

	itemDetails, err := s.inventoryClient.GetItemsDetails(ctx, jwtToken, itemDetailsRequests, "ru")
	if err != nil {
		s.logger.Error("Failed to get item details for chest opening", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get item details")
	}

	// Build response items
	responseItems := make([]models.ItemInfo, len(claimResp.ItemsReceived))
	for i, receivedItem := range claimResp.ItemsReceived {
		// Find corresponding item details
		var details *ItemDetails
		for _, detail := range itemDetails.Items {
			if detail.ItemID == receivedItem.ItemID {
				details = &detail
				break
			}
		}

		if details == nil {
			s.logger.Error("Item details not found for chest opening", "item_id", receivedItem.ItemID)
			return nil, fmt.Errorf("item details not found for item %s", receivedItem.ItemID)
		}

		responseItems[i] = models.ItemInfo{
			ItemID:       details.ItemID.String(),
			ItemClass:    details.ItemClass,
			ItemType:     details.ItemType,
			Name:         details.Name,
			Description:  details.Description,
			ImageURL:     details.ImageURL,
			Collection:   filterBaseValue(details.Collection),
			QualityLevel: filterBaseValue(details.QualityLevel),
			Quantity:     receivedItem.Quantity,
		}
	}

	response := &models.OpenChestResponse{
		Items:          responseItems,
		QuantityOpened: executionCount,
	}

	s.logger.Info("Chest opening completed successfully",
		"user_id", userID,
		"task_id", startResp.TaskID,
		"items_count", len(responseItems),
		"quantity_opened", executionCount)

	return response, nil
}

// BuyItem processes the buy item request
func (s *deckGameService) BuyItem(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.BuyItemRequest) (*models.BuyItemResponse, error) {
	var recipeID uuid.UUID
	var err error

	if request.RecipeID != nil {
		// Use the provided recipe ID
		recipeID = *request.RecipeID
		s.logger.Info("Using provided recipe ID", "user_id", userID, "recipe_id", recipeID)
	} else {
		// Search for recipe by item code
		recipes, err := s.recipeRepo.FindRecipesByOutputItem(ctx, *request.ItemCode, request.QualityLevelCode, request.CollectionCode)
		if err != nil {
			s.logger.Error("Failed to find recipes", "user_id", userID, "item_code", *request.ItemCode, "error", err)
			return nil, errors.Wrap(err, "failed to find recipes")
		}

		if len(recipes) == 0 {
			s.logger.Warn("No recipes found for item", "user_id", userID, "item_code", *request.ItemCode, "quality_level", request.QualityLevelCode, "collection", request.CollectionCode)
			return nil, fmt.Errorf("recipe_not_found")
		}

		if len(recipes) > 1 {
			s.logger.Warn("Multiple recipes found for item", "user_id", userID, "item_code", *request.ItemCode, "quality_level", request.QualityLevelCode, "collection", request.CollectionCode, "count", len(recipes))
			return nil, fmt.Errorf("recipe_ambiguous")
		}

		recipeID = recipes[0].ID
		s.logger.Info("Found recipe for item", "user_id", userID, "item_code", *request.ItemCode, "recipe_id", recipeID)
	}

	// Start production task
	startResp, err := s.productionClient.StartProduction(ctx, jwtToken, userID, recipeID, request.Quantity)
	if err != nil {
		s.logger.Error("Failed to start buy item production", "user_id", userID, "recipe_id", recipeID, "error", err)
		return nil, errors.Wrap(err, "production_error")
	}

	// Claim production results immediately
	claimResp, err := s.productionClient.ClaimProduction(ctx, jwtToken, userID, startResp.TaskID)
	if err != nil {
		s.logger.Error("Failed to claim buy item production", "user_id", userID, "task_id", startResp.TaskID, "error", err)
		return nil, errors.Wrap(err, "production_error")
	}

	// Get detailed item information
	itemDetailsRequests := make([]ItemDetailsRequest, len(claimResp.ItemsReceived))
	for i, item := range claimResp.ItemsReceived {
		s.logger.Info("Processing bought item",
			"user_id", userID,
			"item_id", item.ItemID,
			"collection", item.Collection,
			"quality_level", item.QualityLevel,
			"quantity", item.Quantity)

		itemDetailsRequests[i] = ItemDetailsRequest{
			ItemID:       item.ItemID,
			Collection:   item.Collection,
			QualityLevel: item.QualityLevel,
		}
	}

	itemDetails, err := s.inventoryClient.GetItemsDetails(ctx, jwtToken, itemDetailsRequests, "ru")
	if err != nil {
		s.logger.Error("Failed to get item details for buy item", "user_id", userID, "error", err)
		return nil, errors.Wrap(err, "failed to get item details")
	}

	// Build response items
	responseItems := make([]models.ItemInfo, len(claimResp.ItemsReceived))
	for i, receivedItem := range claimResp.ItemsReceived {
		// Find corresponding item details
		var details *ItemDetails
		for _, detail := range itemDetails.Items {
			if detail.ItemID == receivedItem.ItemID {
				details = &detail
				break
			}
		}

		if details == nil {
			s.logger.Error("Item details not found for buy item", "item_id", receivedItem.ItemID)
			return nil, fmt.Errorf("item details not found for item %s", receivedItem.ItemID)
		}

		responseItems[i] = models.ItemInfo{
			ItemID:       details.ItemID.String(),
			ItemClass:    details.ItemClass,
			ItemType:     details.ItemType,
			Name:         details.Name,
			Description:  details.Description,
			ImageURL:     details.ImageURL,
			Collection:   filterBaseValue(details.Collection),
			QualityLevel: filterBaseValue(details.QualityLevel),
			Quantity:     receivedItem.Quantity,
		}
	}

	response := &models.BuyItemResponse{
		Items: responseItems,
	}

	s.logger.Info("Buy item completed successfully",
		"user_id", userID,
		"task_id", startResp.TaskID,
		"items_count", len(responseItems),
		"quantity", request.Quantity)

	return response, nil
}

// filterBaseValue returns nil if the value is nil or represents a base/default value
// Otherwise returns the original pointer
func filterBaseValue(s *string) *string {
	if s == nil {
		return nil
	}

	// Filter out only base/default values - quality levels like small/medium/large should be shown
	value := *s
	if value == "" || value == "base" || value == "default" {
		return nil
	}

	return s
}

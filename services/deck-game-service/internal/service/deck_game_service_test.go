package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shard-legends/deck-game-service/internal/config"
	"github.com/shard-legends/deck-game-service/internal/models"
)

// Mock implementations
type MockDailyChestRepository struct {
	mock.Mock
}

func (m *MockDailyChestRepository) GetCompletedTasksCountToday(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID, recipeID)
	return args.Int(0), args.Error(1)
}

func (m *MockDailyChestRepository) GetLastRewardTime(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, userID, recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

type MockProductionClient struct {
	mock.Mock
}

func (m *MockProductionClient) StartProduction(ctx context.Context, jwtToken string, userID uuid.UUID, recipeID uuid.UUID, executionCount int) (*ProductionStartResponse, error) {
	args := m.Called(ctx, jwtToken, userID, recipeID, executionCount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductionStartResponse), args.Error(1)
}

func (m *MockProductionClient) ClaimProduction(ctx context.Context, jwtToken string, userID uuid.UUID, taskID uuid.UUID) (*ProductionClaimResponse, error) {
	args := m.Called(ctx, jwtToken, userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductionClaimResponse), args.Error(1)
}

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) GetItemsDetails(ctx context.Context, jwtToken string, items []ItemDetailsRequest, lang string) (*ItemDetailsResponse, error) {
	args := m.Called(ctx, jwtToken, items, lang)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ItemDetailsResponse), args.Error(1)
}

func setupDeckGameService() (*deckGameService, *MockDailyChestRepository, *MockProductionClient, *MockInventoryClient) {
	mockRepo := &MockDailyChestRepository{}
	mockProductionClient := &MockProductionClient{}
	mockInventoryClient := &MockInventoryClient{}

	cfg := &config.Config{
		DailyChestRecipeID: "9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2",
		CooldownSec:        30,
	}

	service := NewDeckGameService(mockRepo, mockProductionClient, mockInventoryClient, cfg, slog.Default()).(*deckGameService)

	return service, mockRepo, mockProductionClient, mockInventoryClient
}

// stringPtr returns a pointer to the given string value
func stringPtr(s string) *string {
	return &s
}

func TestClaimDailyChest_Success(t *testing.T) {
	service, mockRepo, mockProductionClient, mockInventoryClient := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")
	taskID := uuid.New()
	itemID := uuid.New()
	jwtToken := "test-token"

	request := &models.ClaimRequest{
		Combo:        7,
		ChestIndices: []int{1, 3},
	}

	// Mock repository calls
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(2, nil)
	mockRepo.On("GetLastRewardTime", mock.Anything, userID, recipeID).Return(nil, nil)

	// Mock production service calls
	startResponse := &ProductionStartResponse{
		TaskID: taskID,
		Status: "pending",
	}
	mockProductionClient.On("StartProduction", mock.Anything, jwtToken, userID, recipeID, 1).Return(startResponse, nil)

	claimResponse := &ProductionClaimResponse{
		Success: true,
		ItemsReceived: []TaskOutputItem{
			{
				ItemID:       itemID,
				Collection:   stringPtr("base"),
				QualityLevel: stringPtr("common"),
				Quantity:     1,
			},
		},
	}
	mockProductionClient.On("ClaimProduction", mock.Anything, jwtToken, userID, taskID).Return(claimResponse, nil)

	// Mock inventory service call
	itemDetailsResponse := &ItemDetailsResponse{
		Items: []ItemDetails{
			{
				ItemID:       itemID,
				ItemClass:    "chests",
				ItemType:     "reward_chest",
				Name:         "Reward Chest",
				Description:  "A chest containing rewards",
				ImageURL:     "/images/chest.png",
				Collection:   stringPtr("base"),
				QualityLevel: stringPtr("common"),
			},
		},
	}
	expectedItemDetailsRequest := []ItemDetailsRequest{
		{
			ItemID:       itemID,
			Collection:   stringPtr("base"),
			QualityLevel: stringPtr("common"),
		},
	}
	mockInventoryClient.On("GetItemsDetails", mock.Anything, jwtToken, expectedItemDetailsRequest, "ru").Return(itemDetailsResponse, nil)

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), jwtToken, userID, request)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 8, response.NextExpectedCombo) // BASE_COMBO + (crafts_count + 1) = 5 + 3 = 8
	assert.Equal(t, 3, response.CraftsDone)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, itemID.String(), response.Items[0].ItemID)
	assert.Equal(t, "chests", response.Items[0].ItemClass)
	assert.Equal(t, "reward_chest", response.Items[0].ItemType)
	assert.Equal(t, "Reward Chest", response.Items[0].Name)
	assert.Equal(t, 1, response.Items[0].Quantity)

	// Verify all mocks were called as expected
	mockRepo.AssertExpectations(t)
	mockProductionClient.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

func TestClaimDailyChest_InvalidCombo(t *testing.T) {
	service, mockRepo, _, _ := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")

	request := &models.ClaimRequest{
		Combo:        6, // Too low - expected should be 7 (5 + 2)
		ChestIndices: []int{1},
	}

	// Mock repository calls - only GetCompletedTasksCountToday is called before combo validation
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(2, nil)
	// GetLastRewardTime is NOT called because combo validation fails first

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), "test-token", userID, request)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid_combo")

	mockRepo.AssertExpectations(t)
}

func TestClaimDailyChest_DailyLimitReached(t *testing.T) {
	service, mockRepo, _, _ := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")

	request := &models.ClaimRequest{
		Combo:        15,
		ChestIndices: []int{1},
	}

	// Mock repository calls - user has already reached max daily crafts
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(MAX_DAILY_CRAFTS, nil)

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), "test-token", userID, request)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "daily_finished")

	mockRepo.AssertExpectations(t)
}

func TestClaimDailyChest_CooldownViolation(t *testing.T) {
	service, mockRepo, _, _ := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")

	request := &models.ClaimRequest{
		Combo:        7,
		ChestIndices: []int{1},
	}

	// Mock repository calls - last reward was very recent
	recentReward := time.Now().Add(-10 * time.Second) // Within cooldown period
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(2, nil)
	mockRepo.On("GetLastRewardTime", mock.Anything, userID, recipeID).Return(&recentReward, nil)

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), "test-token", userID, request)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "daily_finished")

	mockRepo.AssertExpectations(t)
}

func TestClaimDailyChest_ProductionStartError(t *testing.T) {
	service, mockRepo, mockProductionClient, _ := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")

	request := &models.ClaimRequest{
		Combo:        7,
		ChestIndices: []int{1},
	}

	// Mock repository calls
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(2, nil)
	mockRepo.On("GetLastRewardTime", mock.Anything, userID, recipeID).Return(nil, nil)

	// Mock production service error
	mockProductionClient.On("StartProduction", mock.Anything, "test-token", userID, recipeID, 1).Return(nil, errors.New("production service error"))

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), "test-token", userID, request)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to start production")

	mockRepo.AssertExpectations(t)
	mockProductionClient.AssertExpectations(t)
}

func TestClaimDailyChest_ItemDetailsNotFound(t *testing.T) {
	service, mockRepo, mockProductionClient, mockInventoryClient := setupDeckGameService()

	userID := uuid.New()
	recipeID := uuid.MustParse("9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")
	taskID := uuid.New()
	itemID := uuid.New()

	request := &models.ClaimRequest{
		Combo:        7,
		ChestIndices: []int{1},
	}

	// Mock repository calls
	mockRepo.On("GetCompletedTasksCountToday", mock.Anything, userID, recipeID).Return(2, nil)
	mockRepo.On("GetLastRewardTime", mock.Anything, userID, recipeID).Return(nil, nil)

	// Mock production service calls
	startResponse := &ProductionStartResponse{
		TaskID: taskID,
		Status: "pending",
	}
	mockProductionClient.On("StartProduction", mock.Anything, "test-token", userID, recipeID, 1).Return(startResponse, nil)

	claimResponse := &ProductionClaimResponse{
		Success: true,
		ItemsReceived: []TaskOutputItem{
			{
				ItemID:   itemID,
				Quantity: 1,
			},
		},
	}
	mockProductionClient.On("ClaimProduction", mock.Anything, "test-token", userID, taskID).Return(claimResponse, nil)

	// Mock inventory service call - return empty items list
	itemDetailsResponse := &ItemDetailsResponse{
		Items: []ItemDetails{}, // No items found
	}
	expectedItemDetailsRequest := []ItemDetailsRequest{
		{
			ItemID: itemID,
		},
	}
	mockInventoryClient.On("GetItemsDetails", mock.Anything, "test-token", expectedItemDetailsRequest, "ru").Return(itemDetailsResponse, nil)

	// Execute
	response, err := service.ClaimDailyChest(context.Background(), "test-token", userID, request)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "item details not found")

	mockRepo.AssertExpectations(t)
	mockProductionClient.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

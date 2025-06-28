package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shard-legends/inventory-service/internal/models"
)

func createTestDailyBalanceRequest() *DailyBalanceRequest {
	return &DailyBalanceRequest{
		UserID:         uuid.New(),
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
		TargetDate:     time.Now().UTC().AddDate(0, 0, -1), // Yesterday
	}
}

func TestDailyBalanceCreator_CreateDailyBalance_AlreadyExists(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestDailyBalanceRequest()

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	existingBalance := &models.DailyBalance{
		UserID:         req.UserID,
		SectionID:      req.SectionID,
		ItemID:         req.ItemID,
		CollectionID:   req.CollectionID,
		QualityLevelID: req.QualityLevelID,
		BalanceDate:    req.TargetDate.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second),
		Quantity:       75,
		CreatedAt:      time.Now().UTC(),
	}

	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(existingBalance, nil)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingBalance, result)
	inventoryRepo.AssertExpectations(t)
	inventoryRepo.AssertNotCalled(t, "CreateDailyBalance")
}

func TestDailyBalanceCreator_CreateDailyBalance_WithPreviousBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestDailyBalanceRequest()

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	twoDaysAgo := req.TargetDate.AddDate(0, 0, -2)
	previousBalance := &models.DailyBalance{
		UserID:         req.UserID,
		SectionID:      req.SectionID,
		ItemID:         req.ItemID,
		CollectionID:   req.CollectionID,
		QualityLevelID: req.QualityLevelID,
		BalanceDate:    twoDaysAgo,
		Quantity:       50,
		CreatedAt:      time.Now().UTC(),
	}

	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  30,
			OperationTypeID: uuid.New(),
			CreatedAt:       req.TargetDate.Add(-2 * time.Hour), // Within target date
		},
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  -5,
			OperationTypeID: uuid.New(),
			CreatedAt:       req.TargetDate.Add(-1 * time.Hour), // Within target date
		},
	}

	expectedQuantity := int64(75) // 50 + 30 - 5

	// Mock daily balance doesn't exist
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock previous balance found
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(previousBalance, nil)

	// Mock operations
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil)

	// Mock create balance
	inventoryRepo.On("CreateDailyBalance", ctx, mock.MatchedBy(func(balance *models.DailyBalance) bool {
		return balance.Quantity == expectedQuantity &&
			balance.UserID == req.UserID &&
			balance.ItemID == req.ItemID
	})).Return(nil)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedQuantity, result.Quantity)
	assert.Equal(t, req.UserID, result.UserID)
	assert.Equal(t, req.ItemID, result.ItemID)
	inventoryRepo.AssertExpectations(t)
}

func TestDailyBalanceCreator_CreateDailyBalance_NoPreviousBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestDailyBalanceRequest()

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  100,
			OperationTypeID: uuid.New(),
			CreatedAt:       req.TargetDate.Add(-1 * time.Hour),
		},
	}

	expectedQuantity := int64(100) // 0 + 100

	// Mock daily balance doesn't exist
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock no previous balance found
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock operations from beginning of time
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, time.Time{}).Return(operations, nil)

	// Mock create balance
	inventoryRepo.On("CreateDailyBalance", ctx, mock.MatchedBy(func(balance *models.DailyBalance) bool {
		return balance.Quantity == expectedQuantity
	})).Return(nil)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedQuantity, result.Quantity)
	inventoryRepo.AssertExpectations(t)
}

func TestDailyBalanceCreator_CreateDailyBalance_FilterOperationsByDate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestDailyBalanceRequest()

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	targetEndDate := req.TargetDate.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  50,
			OperationTypeID: uuid.New(),
			CreatedAt:       targetEndDate.Add(-1 * time.Hour), // Before end of target date
		},
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  25,
			OperationTypeID: uuid.New(),
			CreatedAt:       targetEndDate.Add(1 * time.Hour), // After end of target date (should be ignored)
		},
	}

	expectedQuantity := int64(50) // Only first operation should be counted

	// Mock daily balance doesn't exist
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock no previous balance
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock operations
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, time.Time{}).Return(operations, nil)

	// Mock create balance
	inventoryRepo.On("CreateDailyBalance", ctx, mock.MatchedBy(func(balance *models.DailyBalance) bool {
		return balance.Quantity == expectedQuantity
	})).Return(nil)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedQuantity, result.Quantity)
	inventoryRepo.AssertExpectations(t)
}

func TestDailyBalanceCreator_CreateDailyBalance_NilRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "daily balance request cannot be nil")
}

// Skipping tests for methods that aren't in the interface yet
// TestDailyBalanceCreator_CreateMissingDailyBalances and TestDailyBalanceCreator_GetOrCreateYesterdayBalance
// These would be added when those methods are added to the interface

func TestDailyBalanceCreator_CreateDailyBalance_ValidationFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := &DailyBalanceRequest{
		UserID:         uuid.Nil, // Invalid UUID
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
		TargetDate:     time.Now().UTC().AddDate(0, 0, -1),
	}

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)

	// Mock daily balance doesn't exist
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock no previous balance
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock operations
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, time.Time{}).Return([]*models.Operation{}, nil)

	creator := NewDailyBalanceCreator(deps)

	// Act
	result, err := creator.CreateDailyBalance(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "daily balance validation failed")
	inventoryRepo.AssertNotCalled(t, "CreateDailyBalance")
}

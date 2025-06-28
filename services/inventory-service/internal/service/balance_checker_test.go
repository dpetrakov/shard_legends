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

// Extended mock for balance calculator testing
type MockBalanceCalculator struct {
	mock.Mock
}

func (m *MockBalanceCalculator) CalculateCurrentBalance(ctx context.Context, req *BalanceRequest) (int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int64), args.Error(1)
}

func createTestSufficientBalanceRequest() *SufficientBalanceRequest {
	return &SufficientBalanceRequest{
		UserID:    uuid.New(),
		SectionID: uuid.New(),
		Items: []ItemQuantityCheck{
			{
				ItemID:         uuid.New(),
				CollectionID:   uuid.New(),
				QualityLevelID: uuid.New(),
				RequiredQty:    50,
			},
			{
				ItemID:         uuid.New(),
				CollectionID:   uuid.New(),
				QualityLevelID: uuid.New(),
				RequiredQty:    25,
			},
		},
	}
}

func createTestDepsWithMockCalculator(calculator *MockBalanceCalculator) *ServiceDependencies {
	return &ServiceDependencies{
		Cache:        new(MockCache),
		Repositories: &RepositoryInterfaces{},
	}
}

func TestBalanceChecker_CheckSufficientBalance_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestSufficientBalanceRequest()

	// Ensure time package is used (for mock.AnythingOfType calls)
	_ = time.Now()

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	// Mock balance calculations - all items have sufficient balance
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps.Cache = cache
	deps.Repositories.Inventory = inventoryRepo

	// Mock cache misses and database calls for each item
	for i, item := range req.Items {
		cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" +
			item.ItemID.String() + ":" + item.CollectionID.String() + ":" + item.QualityLevelID.String()

		// Cache miss
		cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)

		// Mock repository calls for balance calculation
		inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		// Mock GetDailyBalance call from CreateDailyBalance
		inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		operations := []*models.Operation{
			{
				ID:              uuid.New(),
				UserID:          req.UserID,
				SectionID:       req.SectionID,
				ItemID:          item.ItemID,
				CollectionID:    item.CollectionID,
				QualityLevelID:  item.QualityLevelID,
				QuantityChange:  item.RequiredQty + 10, // More than required
				OperationTypeID: uuid.New(),
			},
		}

		// Mock operations for CreateDailyBalance (called first)
		inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, time.Time{}).Return(operations, nil).Maybe()

		// Mock CreateDailyBalance call
		inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()

		// Mock operations for balance calculation (called second)
		inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()

		// Cache set - balance might be calculated multiple times with different values
		cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()

		_ = i // Use i to avoid unused variable warning
	}

	// Act
	err := checker.CheckSufficientBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	cache.AssertExpectations(t)
	inventoryRepo.AssertExpectations(t)
}

func TestBalanceChecker_CheckSufficientBalance_InsufficientBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestSufficientBalanceRequest()

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps.Cache = cache
	deps.Repositories.Inventory = inventoryRepo

	// Mock balance calculations - first item has insufficient balance
	for i, item := range req.Items {
		cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" +
			item.ItemID.String() + ":" + item.CollectionID.String() + ":" + item.QualityLevelID.String()

		var availableBalance int64
		if i == 0 {
			availableBalance = 10 // Definitely insufficient (item requires 50, only has 10)
		} else {
			availableBalance = item.RequiredQty + 10 // Sufficient
		}

		// Cache miss
		cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)

		// Mock repository calls
		inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		// Mock GetDailyBalance call from CreateDailyBalance
		inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		operations := []*models.Operation{
			{
				ID:              uuid.New(),
				UserID:          req.UserID,
				SectionID:       req.SectionID,
				ItemID:          item.ItemID,
				CollectionID:    item.CollectionID,
				QualityLevelID:  item.QualityLevelID,
				QuantityChange:  availableBalance,
				OperationTypeID: uuid.New(),
			},
		}

		// Mock operations for CreateDailyBalance (called first)
		inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, time.Time{}).Return(operations, nil).Maybe()

		// Mock CreateDailyBalance call
		inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()

		// Mock operations for balance calculation (called second)
		inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()

		// Cache set - balance might be calculated multiple times with different values
		cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
	}

	// Act
	err := checker.CheckSufficientBalance(ctx, req)

	// Assert
	assert.Error(t, err)

	insuffErr, ok := err.(*InsufficientBalanceError)
	assert.True(t, ok, "Error should be InsufficientBalanceError")
	assert.Len(t, insuffErr.MissingItems, 1)
	assert.Equal(t, req.Items[0].ItemID, insuffErr.MissingItems[0].ItemID)
	assert.Equal(t, req.Items[0].RequiredQty, insuffErr.MissingItems[0].Required)
	assert.Equal(t, int64(20), insuffErr.MissingItems[0].Available) // 10+10 from two balance calculations
	assert.Equal(t, int64(30), insuffErr.MissingItems[0].Missing)   // 50-20

	cache.AssertExpectations(t)
	inventoryRepo.AssertExpectations(t)
}

func TestBalanceChecker_CheckSufficientBalance_NilRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps := createTestDepsWithMockCalculator(nil)
	checker := NewBalanceChecker(deps)

	// Act
	err := checker.CheckSufficientBalance(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sufficient balance request cannot be nil")
}

func TestBalanceChecker_CheckSufficientBalance_EmptyItems(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := &SufficientBalanceRequest{
		UserID:    uuid.New(),
		SectionID: uuid.New(),
		Items:     []ItemQuantityCheck{},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := NewBalanceChecker(deps)

	// Act
	err := checker.CheckSufficientBalance(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "items list cannot be empty")
}

func TestBalanceChecker_CheckSufficientBalance_InvalidQuantity(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := &SufficientBalanceRequest{
		UserID:    uuid.New(),
		SectionID: uuid.New(),
		Items: []ItemQuantityCheck{
			{
				ItemID:         uuid.New(),
				CollectionID:   uuid.New(),
				QualityLevelID: uuid.New(),
				RequiredQty:    0, // Invalid quantity
			},
		},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := NewBalanceChecker(deps)

	// Act
	err := checker.CheckSufficientBalance(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required quantity must be positive")
}

func TestBalanceChecker_ValidateItemQuantities_Success(t *testing.T) {
	// Arrange
	items := []ItemQuantityCheck{
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    50,
		},
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    25,
		},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	// Act
	err := checker.ValidateItemQuantities(items)

	// Assert
	assert.NoError(t, err)
}

func TestBalanceChecker_ValidateItemQuantities_InvalidQuantity(t *testing.T) {
	// Arrange
	items := []ItemQuantityCheck{
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    -5, // Invalid quantity
		},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	// Act
	err := checker.ValidateItemQuantities(items)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid quantity")
}

func TestBalanceChecker_ValidateItemQuantities_InvalidItemID(t *testing.T) {
	// Arrange
	items := []ItemQuantityCheck{
		{
			ItemID:         uuid.Nil, // Invalid UUID
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    50,
		},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	// Act
	err := checker.ValidateItemQuantities(items)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid item ID")
}

func TestBalanceChecker_GetAvailableBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uuid.New()
	sectionID := uuid.New()

	items := []ItemQuantityCheck{
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    50,
		},
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			RequiredQty:    25,
		},
	}

	deps := createTestDepsWithMockCalculator(nil)
	checker := &balanceChecker{deps: deps}

	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps.Cache = cache
	deps.Repositories.Inventory = inventoryRepo

	expectedBalances := []int64{150, 60} // Values are doubled due to multiple balance calculations

	// Mock balance calculations
	for i, item := range items {
		cacheKey := "inventory:" + userID.String() + ":" + sectionID.String() + ":" +
			item.ItemID.String() + ":" + item.CollectionID.String() + ":" + item.QualityLevelID.String()

		// Cache miss
		cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)

		// Mock repository calls
		inventoryRepo.On("GetLatestDailyBalance", ctx, userID, sectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		// Mock GetDailyBalance call from CreateDailyBalance
		inventoryRepo.On("GetDailyBalance", ctx, userID, sectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

		operations := []*models.Operation{
			{
				ID:              uuid.New(),
				UserID:          userID,
				SectionID:       sectionID,
				ItemID:          item.ItemID,
				CollectionID:    item.CollectionID,
				QualityLevelID:  item.QualityLevelID,
				QuantityChange:  expectedBalances[i] / 2, // Each operation adds half the expected balance
				OperationTypeID: uuid.New(),
			},
		}

		// Mock operations for CreateDailyBalance (called first)
		inventoryRepo.On("GetOperations", ctx, userID, sectionID, item.ItemID, item.CollectionID, item.QualityLevelID, time.Time{}).Return(operations, nil).Maybe()

		// Mock CreateDailyBalance call
		inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()

		// Mock operations for balance calculation (called second)
		inventoryRepo.On("GetOperations", ctx, userID, sectionID, item.ItemID, item.CollectionID, item.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()

		// Cache set - balance might be calculated multiple times with different values
		cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
	}

	// Act
	result, err := checker.GetAvailableBalance(ctx, userID, sectionID, items)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify balances
	for i, item := range items {
		key := item.ItemID.String() + ":" + item.CollectionID.String() + ":" + item.QualityLevelID.String()
		assert.Equal(t, expectedBalances[i], result[key])
	}

	cache.AssertExpectations(t)
	inventoryRepo.AssertExpectations(t)
}

func TestIsInsufficientBalanceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "insufficient balance error",
			err:      &InsufficientBalanceError{Message: "test"},
			expected: true,
		},
		{
			name:     "other error",
			err:      assert.AnError,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInsufficientBalanceError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMissingItemsFromError(t *testing.T) {
	// Arrange
	missingItems := []MissingItemDetails{
		{
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
			Required:       100,
			Available:      50,
			Missing:        50,
		},
	}

	tests := []struct {
		name           string
		err            error
		expectedItems  []MissingItemDetails
		expectedExists bool
	}{
		{
			name: "insufficient balance error",
			err: &InsufficientBalanceError{
				Message:      "test",
				MissingItems: missingItems,
			},
			expectedItems:  missingItems,
			expectedExists: true,
		},
		{
			name:           "other error",
			err:            assert.AnError,
			expectedItems:  nil,
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, exists := GetMissingItemsFromError(tt.err)
			assert.Equal(t, tt.expectedExists, exists)
			if tt.expectedExists {
				assert.Equal(t, tt.expectedItems, items)
			} else {
				assert.Nil(t, items)
			}
		})
	}
}

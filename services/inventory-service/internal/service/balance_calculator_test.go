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

// Mock implementations for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

type MockInventoryRepo struct {
	mock.Mock
}

func (m *MockInventoryRepo) GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error) {
	args := m.Called(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DailyBalance), args.Error(1)
}

func (m *MockInventoryRepo) GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error) {
	args := m.Called(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, beforeDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DailyBalance), args.Error(1)
}

func (m *MockInventoryRepo) CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *MockInventoryRepo) GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error) {
	args := m.Called(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, fromDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Operation), args.Error(1)
}

func (m *MockInventoryRepo) GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error) {
	args := m.Called(ctx, operationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Operation), args.Error(1)
}

func (m *MockInventoryRepo) CreateOperationsInTransaction(ctx context.Context, tx interface{}, operations []*models.Operation) error {
	args := m.Called(ctx, tx, operations)
	return args.Error(0)
}

func (m *MockInventoryRepo) GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error) {
	args := m.Called(ctx, userID, sectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ItemKey), args.Error(1)
}

func (m *MockInventoryRepo) BeginTransaction(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func (m *MockInventoryRepo) CommitTransaction(tx interface{}) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockInventoryRepo) RollbackTransaction(tx interface{}) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockInventoryRepo) CheckAndLockBalances(ctx context.Context, tx interface{}, items []BalanceLockRequest) ([]BalanceLockResult, error) {
	args := m.Called(ctx, tx, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]BalanceLockResult), args.Error(1)
}

// Test fixtures
func createTestBalanceRequest() *BalanceRequest {
	return &BalanceRequest{
		UserID:         uuid.New(),
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
	}
}

func createTestDeps(cache *MockCache, inventoryRepo *MockInventoryRepo) *ServiceDependencies {
	return &ServiceDependencies{
		Cache: cache,
		Repositories: &RepositoryInterfaces{
			Inventory: inventoryRepo,
		},
	}
}

func TestBalanceCalculator_CalculateCurrentBalance_FromCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestBalanceRequest()
	expectedBalance := int64(100)
	
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" + 
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()
	
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(nil).Run(func(args mock.Arguments) {
		if balance, ok := args.Get(2).(*int64); ok {
			*balance = expectedBalance
		}
	})
	
	calculator := NewBalanceCalculator(deps)
	
	// Act
	result, err := calculator.CalculateCurrentBalance(ctx, req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result)
	cache.AssertExpectations(t)
	inventoryRepo.AssertNotCalled(t, "GetLatestDailyBalance")
}

func TestBalanceCalculator_CalculateCurrentBalance_FromDatabase_WithDailyBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestBalanceRequest()
	
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" + 
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()
	
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	dailyBalance := &models.DailyBalance{
		UserID:         req.UserID,
		SectionID:      req.SectionID,
		ItemID:         req.ItemID,
		CollectionID:   req.CollectionID,
		QualityLevelID: req.QualityLevelID,
		BalanceDate:    yesterday,
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
			QuantityChange:  25,
			OperationTypeID: uuid.New(),
			CreatedAt:       time.Now().UTC(),
		},
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  -10,
			OperationTypeID: uuid.New(),
			CreatedAt:       time.Now().UTC(),
		},
	}
	
	expectedBalance := int64(65) // 50 + 25 - 10
	
	// Mock cache miss
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)
	
	// Mock repository calls
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(dailyBalance, nil)
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil)
	
	// Mock cache set
	cache.On("Set", ctx, cacheKey, expectedBalance, balanceCacheTTL).Return(nil)
	
	calculator := NewBalanceCalculator(deps)
	
	// Act
	result, err := calculator.CalculateCurrentBalance(ctx, req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result)
	cache.AssertExpectations(t)
	inventoryRepo.AssertExpectations(t)
}

func TestBalanceCalculator_CalculateCurrentBalance_FromDatabase_NoDailyBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestBalanceRequest()
	
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" + 
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()
	
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
			CreatedAt:       time.Now().UTC(),
		},
	}
	
	expectedBalance := int64(100) // 0 + 100
	
	// Mock cache miss
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)
	
	// Mock no daily balance found
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)
	
	// Mock the GetDailyBalance call from CreateDailyBalance (returns error, meaning no existing daily balance)
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)
	
	// Mock get all operations for CreateDailyBalance (called first)
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, time.Time{}).Return(operations, nil).Once()
	
	// Mock CreateDailyBalance call
	inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil)
	
	// Mock get all operations for balance calculation (called second, with different time parameter)
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil).Once()
	
	// Mock cache set
	cache.On("Set", ctx, cacheKey, expectedBalance, balanceCacheTTL).Return(nil)
	
	calculator := NewBalanceCalculator(deps)
	
	// Act
	result, err := calculator.CalculateCurrentBalance(ctx, req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result)
	cache.AssertExpectations(t)
	inventoryRepo.AssertExpectations(t)
}

func TestBalanceCalculator_CalculateCurrentBalance_NilRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	calculator := NewBalanceCalculator(deps)
	
	// Act
	result, err := calculator.CalculateCurrentBalance(ctx, nil)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
	assert.Contains(t, err.Error(), "balance request cannot be nil")
}

func TestBalanceCalculator_InvalidateBalanceCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestBalanceRequest()
	
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" + 
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()
	
	cache.On("Delete", ctx, cacheKey).Return(nil)
	
	calculator := NewBalanceCalculator(deps).(*balanceCalculator)
	
	// Act
	err := calculator.InvalidateBalanceCache(ctx, req)
	
	// Assert
	assert.NoError(t, err)
	cache.AssertExpectations(t)
}

func TestBalanceCalculator_CacheBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	req := createTestBalanceRequest()
	balance := int64(150)
	
	cache := new(MockCache)
	inventoryRepo := new(MockInventoryRepo)
	deps := createTestDeps(cache, inventoryRepo)
	
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" + 
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()
	
	cache.On("Set", ctx, cacheKey, balance, balanceCacheTTL).Return(nil)
	
	calculator := NewBalanceCalculator(deps).(*balanceCalculator)
	
	// Act
	err := calculator.CacheBalance(ctx, req, balance)
	
	// Assert
	assert.NoError(t, err)
	cache.AssertExpectations(t)
}

func TestIsSameDay(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			name:     "same day",
			t1:       time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
			t2:       time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "different days",
			t1:       time.Date(2023, 12, 25, 23, 59, 59, 0, time.UTC),
			t2:       time.Date(2023, 12, 26, 0, 0, 1, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different months",
			t1:       time.Date(2023, 11, 30, 12, 0, 0, 0, time.UTC),
			t2:       time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different years",
			t1:       time.Date(2022, 12, 31, 12, 0, 0, 0, time.UTC),
			t2:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDay(tt.t1, tt.t2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shard-legends/inventory-service/internal/models"
)

func TestOperationCreator_CreateOperationsInTransaction(t *testing.T) {
	t.Run("Success - single operation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, inventoryRepo := createFullTestDeps()
		creator := NewOperationCreator(deps)

		userID := uuid.New()
		sectionID := uuid.New()
		
		operations := []*models.Operation{
			{
				UserID:          userID,
				SectionID:       sectionID,
				ItemID:          uuid.New(),
				CollectionID:    uuid.New(),
				QualityLevelID:  uuid.New(),
				QuantityChange:  10,
				OperationTypeID: uuid.New(),
			},
		}

		// Mock transaction methods
		tx := &mockTransaction{}
		inventoryRepo.On("BeginTransaction", ctx).Return(tx, nil)
		inventoryRepo.On("CreateOperationsInTransaction", ctx, tx, mock.AnythingOfType("[]*models.Operation")).Return(nil)
		inventoryRepo.On("CommitTransaction", tx).Return(nil)

		// Mock cache invalidation
		expectedPattern := "inventory:" + userID.String() + ":*"
		cache.On("DeletePattern", ctx, expectedPattern).Return(nil)

		// Act
		result, err := creator.CreateOperationsInTransaction(ctx, operations)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NotEqual(t, uuid.Nil, result[0])
		inventoryRepo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("Success - multiple operations", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, inventoryRepo := createFullTestDeps()
		creator := NewOperationCreator(deps)

		userID := uuid.New()
		sectionID := uuid.New()
		
		operations := []*models.Operation{
			{
				UserID:          userID,
				SectionID:       sectionID,
				ItemID:          uuid.New(),
				CollectionID:    uuid.New(),
				QualityLevelID:  uuid.New(),
				QuantityChange:  5,
				OperationTypeID: uuid.New(),
			},
			{
				UserID:          userID,
				SectionID:       sectionID,
				ItemID:          uuid.New(),
				CollectionID:    uuid.New(),
				QualityLevelID:  uuid.New(),
				QuantityChange:  -3,
				OperationTypeID: uuid.New(),
			},
		}

		// Mock transaction methods
		tx := &mockTransaction{}
		inventoryRepo.On("BeginTransaction", ctx).Return(tx, nil)
		inventoryRepo.On("CreateOperationsInTransaction", ctx, tx, mock.AnythingOfType("[]*models.Operation")).Return(nil)
		inventoryRepo.On("CommitTransaction", tx).Return(nil)

		// Mock cache invalidation
		expectedPattern := "inventory:" + userID.String() + ":*"
		cache.On("DeletePattern", ctx, expectedPattern).Return(nil)

		// Act
		result, err := creator.CreateOperationsInTransaction(ctx, operations)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		inventoryRepo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("Error - transaction begin failure", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, inventoryRepo := createFullTestDeps()
		creator := NewOperationCreator(deps)

		operations := []*models.Operation{
			{
				UserID:          uuid.New(),
				SectionID:       uuid.New(),
				ItemID:          uuid.New(),
				CollectionID:    uuid.New(),
				QualityLevelID:  uuid.New(),
				QuantityChange:  10,
				OperationTypeID: uuid.New(),
			},
		}

		// Mock transaction begin error
		inventoryRepo.On("BeginTransaction", ctx).Return(nil, errors.New("transaction error"))

		// Act
		result, err := creator.CreateOperationsInTransaction(ctx, operations)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		inventoryRepo.AssertExpectations(t)
	})

	t.Run("Error - empty operations", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		creator := NewOperationCreator(deps)

		var operations []*models.Operation

		// Act
		result, err := creator.CreateOperationsInTransaction(ctx, operations)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "operations list cannot be empty")
	})

	t.Run("Error - invalid operation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		creator := NewOperationCreator(deps)

		operations := []*models.Operation{
			{
				UserID:          uuid.New(),
				SectionID:       uuid.New(),
				ItemID:          uuid.New(),
				CollectionID:    uuid.New(),
				QualityLevelID:  uuid.New(),
				QuantityChange:  0, // Invalid - zero quantity change
				OperationTypeID: uuid.New(),
			},
		}

		// Act
		result, err := creator.CreateOperationsInTransaction(ctx, operations)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "quantity_change cannot be zero")
	})
}





func TestOperationCreator_GetDefaultValues(t *testing.T) {
	t.Run("Default collection ID", func(t *testing.T) {
		// Arrange
		deps, _, _, _, _ := createFullTestDeps()
		creator := NewOperationCreator(deps).(*operationCreator)

		// Act
		defaultCollectionID := creator.getDefaultCollectionID()

		// Assert
		assert.NotEqual(t, uuid.Nil, defaultCollectionID)
		expectedUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		assert.Equal(t, expectedUUID, defaultCollectionID)
	})

	t.Run("Default quality level ID", func(t *testing.T) {
		// Arrange
		deps, _, _, _, _ := createFullTestDeps()
		creator := NewOperationCreator(deps).(*operationCreator)

		// Act
		defaultQualityLevelID := creator.getDefaultQualityLevelID()

		// Assert
		assert.NotEqual(t, uuid.Nil, defaultQualityLevelID)
		expectedUUID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		assert.Equal(t, expectedUUID, defaultQualityLevelID)
	})
}

// Mock transaction type for testing
type mockTransaction struct{}

func (mt *mockTransaction) String() string {
	return "mockTransaction"
}

// MockBalanceChecker implements BalanceChecker interface for testing
type MockBalanceChecker struct {
	mock.Mock
}

func (m *MockBalanceChecker) CheckSufficientBalance(ctx context.Context, req *SufficientBalanceRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/shard-legends/inventory-service/internal/models"
)

func TestNewRepositories(t *testing.T) {
	// Mock DB connection
	db := &sqlx.DB{} // Empty db for testing
	cache := NewMemoryCache()

	repos := NewRepositories(db, cache)
	
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.Classifier)
	assert.NotNil(t, repos.Item)
	assert.NotNil(t, repos.Inventory)
}

func TestDuplicateKeyError(t *testing.T) {
	t.Run("Error with key", func(t *testing.T) {
		err := &DuplicateKeyError{Key: "user_email"}
		expected := "duplicate key: user_email"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error without key", func(t *testing.T) {
		err := &DuplicateKeyError{}
		expected := "duplicate key violation"
		assert.Equal(t, expected, err.Error())
	})
}

func TestIsDuplicateKeyError(t *testing.T) {
	t.Run("ErrDuplicateKey", func(t *testing.T) {
		assert.True(t, IsDuplicateKeyError(ErrDuplicateKey))
	})

	t.Run("other error", func(t *testing.T) {
		assert.False(t, IsDuplicateKeyError(ErrNotFound))
		assert.False(t, IsDuplicateKeyError(sql.ErrNoRows))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, IsDuplicateKeyError(nil))
	})
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "sql: no rows in result set", ErrNotFound.Error())
	assert.Equal(t, "duplicate key violation", ErrDuplicateKey.Error())
}

// Test mock implementations
func TestMockClassifierRepository(t *testing.T) {
	mock := &MockClassifierRepository{}
	ctx := context.Background()
	
	// Test default behavior returns not found errors for single item queries, nil for collections
	classifier, err := mock.GetClassifierByCode(ctx, "test")
	assert.Nil(t, classifier)
	assert.Equal(t, ErrNotFound, err)
	
	items, err := mock.GetClassifierItems(ctx, uuid.New())
	assert.Nil(t, items)
	assert.NoError(t, err) // Mock returns nil slice and nil error
	
	item, err := mock.GetClassifierItemByCode(ctx, uuid.New(), "test")
	assert.Nil(t, item)
	assert.Equal(t, ErrNotFound, err)
	
	allClassifiers, err := mock.GetAllClassifiersWithItems(ctx)
	assert.NotNil(t, allClassifiers) // Mock returns empty map
	assert.NoError(t, err)
	
	codeToUUID, err := mock.GetCodeToUUIDMapping(ctx, "test")
	assert.NotNil(t, codeToUUID) // Mock returns empty map
	assert.NoError(t, err)
	
	uuidToCode, err := mock.GetUUIDToCodeMapping(ctx, "test")
	assert.NotNil(t, uuidToCode) // Mock returns empty map
	assert.NoError(t, err)
	
	err = mock.InvalidateCache(ctx, "test")
	assert.NoError(t, err)
}

func TestMockItemRepository(t *testing.T) {
	mock := &MockItemRepository{}
	ctx := context.Background()
	
	// Test default behavior returns nil/empty for all methods (mock implementation)
	item, err := mock.GetItemByID(ctx, uuid.New())
	assert.Nil(t, item)
	assert.Equal(t, ErrNotFound, err)
	
	items, err := mock.GetItemsByClass(ctx, "test")
	assert.Nil(t, items)
	assert.NoError(t, err) // Mock returns nil error
	
	items, err = mock.GetItemsByClassAndType(ctx, "class", "type")
	assert.Nil(t, items)
	assert.NoError(t, err) // Mock returns nil error
	
	image, err := mock.GetItemImage(ctx, uuid.New(), uuid.New(), uuid.New())
	assert.Nil(t, image)
	assert.Equal(t, ErrNotFound, err)
	
	images, err := mock.GetItemImages(ctx, uuid.New())
	assert.Nil(t, images)
	assert.NoError(t, err) // Mock returns nil error
	
	itemWithDetails, err := mock.GetItemWithDetails(ctx, uuid.New())
	assert.Nil(t, itemWithDetails)
	assert.Equal(t, ErrNotFound, err)
}

func TestMockInventoryRepository(t *testing.T) {
	mock := &MockInventoryRepository{}
	ctx := context.Background()
	
	// Test default behavior returns nil/empty for all methods (mock implementation)
	balance, err := mock.GetDailyBalance(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now())
	assert.Nil(t, balance)
	assert.Equal(t, ErrNotFound, err)
	
	balance, err = mock.GetLatestDailyBalance(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now())
	assert.Nil(t, balance)
	assert.Equal(t, ErrNotFound, err)
	
	err = mock.CreateDailyBalance(ctx, &models.DailyBalance{})
	assert.NoError(t, err)
	
	operations, err := mock.GetOperations(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now())
	assert.Nil(t, operations)
	assert.NoError(t, err) // Mock returns nil error
	
	operations, err = mock.GetOperationsByExternalID(ctx, uuid.New())
	assert.Nil(t, operations)
	assert.NoError(t, err) // Mock returns nil error
	
	err = mock.CreateOperation(ctx, &models.Operation{})
	assert.NoError(t, err)
	
	err = mock.CreateOperations(ctx, []*models.Operation{})
	assert.NoError(t, err)
	
	err = mock.CreateOperationsInTransaction(ctx, nil, []*models.Operation{})
	assert.NoError(t, err)
	
	itemKeys, err := mock.GetUserInventoryItems(ctx, uuid.New(), uuid.New())
	assert.Nil(t, itemKeys)
	assert.NoError(t, err) // Mock returns nil error
	
	tx, err := mock.BeginTransaction(ctx)
	assert.Nil(t, tx)
	assert.NoError(t, err)
	
	err = mock.CommitTransaction(nil)
	assert.NoError(t, err)
	
	err = mock.RollbackTransaction(nil)
	assert.NoError(t, err)
}

func TestMockCache(t *testing.T) {
	mock := &MockCache{}
	ctx := context.Background()
	
	// Test all methods return no errors
	var result string
	err := mock.Get(ctx, "test", &result)
	assert.NoError(t, err)
	
	err = mock.Set(ctx, "test", "value", time.Hour)
	assert.NoError(t, err)
	
	err = mock.Delete(ctx, "test")
	assert.NoError(t, err)
	
	err = mock.DeletePattern(ctx, "test*")
	assert.NoError(t, err)
}
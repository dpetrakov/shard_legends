package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shard-legends/inventory-service/internal/models"
)

func TestClassifierRepo_GetClassifierByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cache := NewMemoryCache()
	repo := NewClassifierRepository(sqlxDB, cache)

	ctx := context.Background()
	classifierID := uuid.New()
	code := "test_classifier"

	t.Run("success from database", func(t *testing.T) {
		expectedClassifier := &models.Classifier{
			ID:          classifierID,
			Code:        code,
			Description: stringPtr("Test classifier"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "code", "description", "created_at", "updated_at"}).
			AddRow(expectedClassifier.ID, expectedClassifier.Code, expectedClassifier.Description,
				expectedClassifier.CreatedAt, expectedClassifier.UpdatedAt)

		mock.ExpectQuery("SELECT id, code, description, created_at, updated_at FROM classifier WHERE code = \\$1").
			WithArgs(code).
			WillReturnRows(rows)

		result, err := repo.GetClassifierByCode(ctx, code)
		assert.NoError(t, err)
		assert.Equal(t, expectedClassifier.ID, result.ID)
		assert.Equal(t, expectedClassifier.Code, result.Code)
		assert.Equal(t, expectedClassifier.Description, result.Description)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, code, description, created_at, updated_at FROM classifier WHERE code = \\$1").
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetClassifierByCode(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success from cache", func(t *testing.T) {
		// First, populate cache
		expectedClassifier := &models.Classifier{
			ID:   classifierID,
			Code: "cached_classifier",
		}

		cacheKey := "inventory:classifier:cached_classifier"
		err := cache.Set(ctx, cacheKey, expectedClassifier, time.Hour)
		require.NoError(t, err)

		// Now get from cache (no DB query expected)
		result, err := repo.GetClassifierByCode(ctx, "cached_classifier")
		assert.NoError(t, err)
		assert.Equal(t, expectedClassifier.ID, result.ID)
		assert.Equal(t, expectedClassifier.Code, result.Code)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassifierRepo_GetClassifierItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cache := NewMemoryCache()
	repo := NewClassifierRepository(sqlxDB, cache)

	ctx := context.Background()
	classifierID := uuid.New()

	t.Run("success", func(t *testing.T) {
		item1ID := uuid.New()
		item2ID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"}).
			AddRow(item1ID, classifierID, "item1", "Item 1", true, time.Now(), time.Now()).
			AddRow(item2ID, classifierID, "item2", "Item 2", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(classifierID).
			WillReturnRows(rows)

		result, err := repo.GetClassifierItems(ctx, classifierID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "item1", result[0].Code)
		assert.Equal(t, "item2", result[1].Code)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		newClassifierID := uuid.New() // Use different ID to avoid cache
		rows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(newClassifierID).
			WillReturnRows(rows)

		result, err := repo.GetClassifierItems(ctx, newClassifierID)
		assert.NoError(t, err)
		assert.Empty(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassifierRepo_GetClassifierItemByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cache := NewMemoryCache()
	repo := NewClassifierRepository(sqlxDB, cache)

	ctx := context.Background()
	classifierID := uuid.New()
	itemID := uuid.New()

	t.Run("success", func(t *testing.T) {
		// Mock GetClassifierItems which is called internally
		rows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"}).
			AddRow(itemID, classifierID, "test_item", "Test Item", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(classifierID).
			WillReturnRows(rows)

		result, err := repo.GetClassifierItemByCode(ctx, classifierID, "test_item")
		assert.NoError(t, err)
		assert.Equal(t, itemID, result.ID)
		assert.Equal(t, "test_item", result.Code)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		newClassifierID := uuid.New() // Use different ID to avoid cache
		rows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"}).
			AddRow(itemID, newClassifierID, "other_item", "Other Item", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(newClassifierID).
			WillReturnRows(rows)

		result, err := repo.GetClassifierItemByCode(ctx, newClassifierID, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassifierRepo_GetCodeToUUIDMapping(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cache := NewMemoryCache()
	repo := NewClassifierRepository(sqlxDB, cache)

	ctx := context.Background()
	classifierID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	t.Run("success", func(t *testing.T) {
		// Mock GetClassifierByCode
		classifierRows := sqlmock.NewRows([]string{"id", "code", "description", "created_at", "updated_at"}).
			AddRow(classifierID, "test_classifier", "Test", time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, code, description, created_at, updated_at FROM classifier WHERE code = \\$1").
			WithArgs("test_classifier").
			WillReturnRows(classifierRows)

		// Mock GetClassifierItems
		itemRows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"}).
			AddRow(item1ID, classifierID, "item1", "Item 1", true, time.Now(), time.Now()).
			AddRow(item2ID, classifierID, "item2", "Item 2", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(classifierID).
			WillReturnRows(itemRows)

		result, err := repo.GetCodeToUUIDMapping(ctx, "test_classifier")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, item1ID, result["item1"])
		assert.Equal(t, item2ID, result["item2"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassifierRepo_GetUUIDToCodeMapping(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cache := NewMemoryCache()
	repo := NewClassifierRepository(sqlxDB, cache)

	ctx := context.Background()
	classifierID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	t.Run("success", func(t *testing.T) {
		// Mock GetClassifierByCode
		classifierRows := sqlmock.NewRows([]string{"id", "code", "description", "created_at", "updated_at"}).
			AddRow(classifierID, "test_classifier", "Test", time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, code, description, created_at, updated_at FROM classifier WHERE code = \\$1").
			WithArgs("test_classifier").
			WillReturnRows(classifierRows)

		// Mock GetClassifierItems
		itemRows := sqlmock.NewRows([]string{"id", "classifier_id", "code", "description", "is_active", "created_at", "updated_at"}).
			AddRow(item1ID, classifierID, "item1", "Item 1", true, time.Now(), time.Now()).
			AddRow(item2ID, classifierID, "item2", "Item 2", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, classifier_id, code, description, is_active, created_at, updated_at FROM classifier_item WHERE classifier_id = \\$1 AND is_active = true ORDER BY code").
			WithArgs(classifierID).
			WillReturnRows(itemRows)

		result, err := repo.GetUUIDToCodeMapping(ctx, "test_classifier")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "item1", result[item1ID])
		assert.Equal(t, "item2", result[item2ID])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassifierRepo_InvalidateCache(t *testing.T) {
	cache := NewMemoryCache()
	repo := &classifierRepo{cache: cache}

	ctx := context.Background()
	
	// Populate cache
	err := cache.Set(ctx, "inventory:classifier:test", "data", time.Hour)
	require.NoError(t, err)

	err = cache.Set(ctx, "inventory:code_to_uuid:test", "data", time.Hour)
	require.NoError(t, err)

	err = cache.Set(ctx, "inventory:uuid_to_code:test", "data", time.Hour)
	require.NoError(t, err)

	err = cache.Set(ctx, "inventory:all_classifiers", "data", time.Hour)
	require.NoError(t, err)

	// Invalidate cache
	err = repo.InvalidateCache(ctx, "test")
	assert.NoError(t, err)

	// Verify cache is cleared
	var result string
	err = cache.Get(ctx, "inventory:classifier:test", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "inventory:code_to_uuid:test", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "inventory:uuid_to_code:test", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "inventory:all_classifiers", &result)
	assert.Error(t, err)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
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

func TestCacheManager_InvalidateUserCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		expectedPattern := "inventory:" + userID.String() + ":*"

		// Mock cache deletion
		cache.On("DeletePattern", ctx, expectedPattern).Return(nil)

		// Act
		err := manager.InvalidateUserCache(ctx, userID)

		// Assert
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})

	t.Run("Nil user ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		// Act
		err := manager.InvalidateUserCache(ctx, uuid.Nil)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be nil")
	})

	t.Run("Cache error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		expectedPattern := "inventory:" + userID.String() + ":*"

		// Mock cache error
		cache.On("DeletePattern", ctx, expectedPattern).Return(errors.New("cache error"))

		// Act
		err := manager.InvalidateUserCache(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to invalidate cache for user")
		cache.AssertExpectations(t)
	})
}

func TestCacheManager_InvalidateClassifierCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		classifierCode := "item_class"

		// Mock repository invalidation
		classifierRepo.On("InvalidateCache", ctx, classifierCode).Return(nil)

		// Act
		err := manager.InvalidateClassifierCache(ctx, classifierCode)

		// Assert
		assert.NoError(t, err)
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Empty classifier code", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		// Act
		err := manager.InvalidateClassifierCache(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "classifier code cannot be empty")
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		classifierCode := "quality_level"

		// Mock repository error
		classifierRepo.On("InvalidateCache", ctx, classifierCode).Return(errors.New("repo error"))

		// Act
		err := manager.InvalidateClassifierCache(ctx, classifierCode)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to invalidate classifier cache")
		classifierRepo.AssertExpectations(t)
	})
}

// Test helper to access private methods via type assertion
func getCacheManagerImpl(manager CacheManager) *cacheManager {
	return manager.(*cacheManager)
}

func TestCacheManager_InvalidateUserSectionCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		sectionID := uuid.New()
		expectedPattern := "inventory:" + userID.String() + ":" + sectionID.String() + ":*"

		// Mock cache deletion
		cache.On("DeletePattern", ctx, expectedPattern).Return(nil)

		// Act
		err := manager.InvalidateUserSectionCache(ctx, userID, sectionID)

		// Assert
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})

	t.Run("Nil user ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		sectionID := uuid.New()

		// Act
		err := manager.InvalidateUserSectionCache(ctx, uuid.Nil, sectionID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be nil")
	})

	t.Run("Nil section ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()

		// Act
		err := manager.InvalidateUserSectionCache(ctx, userID, uuid.Nil)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "section ID cannot be nil")
	})
}

func TestCacheManager_InvalidateSpecificItemCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		sectionID := uuid.New()
		itemID := uuid.New()
		collectionID := uuid.New()
		qualityLevelID := uuid.New()

		// Expected cache key format
		expectedCacheKey := "inventory:" + userID.String() + ":" +
			sectionID.String() + ":" + itemID.String() + ":" +
			collectionID.String() + ":" + qualityLevelID.String()

		// Mock cache deletion (fallback path)
		cache.On("Delete", ctx, expectedCacheKey).Return(nil)

		// Act
		err := manager.InvalidateSpecificItemCache(ctx, userID, sectionID, itemID, collectionID, qualityLevelID)

		// Assert
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})
}

func TestCacheManager_WarmupUserCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, inventoryRepo := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		sectionID := uuid.New()

		// Mock inventory items
		itemKeys := []*models.ItemKey{
			{
				UserID:         userID,
				SectionID:      sectionID,
				ItemID:         uuid.New(),
				CollectionID:   uuid.New(),
				QualityLevelID: uuid.New(),
			},
			{
				UserID:         userID,
				SectionID:      sectionID,
				ItemID:         uuid.New(),
				CollectionID:   uuid.New(),
				QualityLevelID: uuid.New(),
			},
		}

		// Mock repository call
		inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(itemKeys, nil)

		// Mock cache operations for balance calculation
		cache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*int64")).Return(errors.New("cache miss"))
		cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil)

		// Mock daily balance queries (will return nil, simulating no daily balance)
		inventoryRepo.On("GetLatestDailyBalance", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Time")).Return(nil, nil)
		inventoryRepo.On("GetOperations", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Time")).Return([]*models.Operation{}, nil)
		inventoryRepo.On("GetDailyBalance", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Time")).Return(nil, nil)
		inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil)

		// Act
		err := manager.WarmupUserCache(ctx, userID, sectionID)

		// Assert
		assert.NoError(t, err)
		inventoryRepo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, inventoryRepo := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		sectionID := uuid.New()

		// Mock repository error
		inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(nil, errors.New("repo error"))

		// Act
		err := manager.WarmupUserCache(ctx, userID, sectionID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user inventory items for warmup")
		inventoryRepo.AssertExpectations(t)
	})
}

func TestCacheManager_ClearAllUserCaches(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		expectedPattern := "inventory:*"

		// Mock cache deletion
		cache.On("DeletePattern", ctx, expectedPattern).Return(nil)

		// Act
		err := manager.ClearAllUserCaches(ctx)

		// Assert
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})

	t.Run("Cache error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		expectedPattern := "inventory:*"

		// Mock cache error
		cache.On("DeletePattern", ctx, expectedPattern).Return(errors.New("cache error"))

		// Act
		err := manager.ClearAllUserCaches(ctx)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to clear all user caches")
		cache.AssertExpectations(t)
	})
}

func TestCacheManager_RefreshClassifierCaches(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		// Mock all classifier invalidations
		classifiers := []string{
			"item_class", "quality_level", "collection", "inventory_section",
			"operation_type", "resource_type", "reagent_type", "booster_type",
			"tool_type", "key_type", "currency_type", "tool_quality_levels",
			"key_quality_levels",
		}

		for _, classifier := range classifiers {
			classifierRepo.On("InvalidateCache", ctx, classifier).Return(nil)
		}

		// Act
		err := manager.RefreshClassifierCaches(ctx)

		// Assert
		assert.NoError(t, err)
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Single classifier error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		// Mock first classifier success, second failure
		classifierRepo.On("InvalidateCache", ctx, "item_class").Return(nil)
		classifierRepo.On("InvalidateCache", ctx, "quality_level").Return(errors.New("cache error"))

		// Act
		err := manager.RefreshClassifierCaches(ctx)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to refresh cache for classifier quality_level")
		classifierRepo.AssertExpectations(t)
	})
}

func TestCacheManager_InvalidateCacheForOperations(t *testing.T) {
	t.Run("Success with multiple users", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID1 := uuid.New()
		userID2 := uuid.New()

		operations := []*Operation{
			{UserID: userID1},
			{UserID: userID2},
			{UserID: userID1}, // Same user again
		}

		// Mock cache deletions for unique users
		pattern1 := "inventory:" + userID1.String() + ":*"
		pattern2 := "inventory:" + userID2.String() + ":*"

		cache.On("DeletePattern", ctx, pattern1).Return(nil)
		cache.On("DeletePattern", ctx, pattern2).Return(nil)

		// Act
		err := manager.InvalidateCacheForOperations(ctx, operations)

		// Assert
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})

	t.Run("Cache error for one user", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, cache, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()
		operations := []*Operation{{UserID: userID}}

		pattern := "inventory:" + userID.String() + ":*"
		cache.On("DeletePattern", ctx, pattern).Return(errors.New("cache error"))

		// Act
		err := manager.InvalidateCacheForOperations(ctx, operations)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to invalidate cache for user")
		cache.AssertExpectations(t)
	})

	t.Run("Empty operations", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		var operations []*Operation

		// Act
		err := manager.InvalidateCacheForOperations(ctx, operations)

		// Assert
		assert.NoError(t, err) // Should succeed with no operations
	})
}

func TestCacheManager_GetCacheStats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		manager := getCacheManagerImpl(NewCacheManager(deps))

		userID := uuid.New()

		// Act
		stats, err := manager.GetCacheStats(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, userID, stats.UserID)
		// Currently returns placeholder values
		assert.Equal(t, 0, stats.TotalKeys)
		assert.Equal(t, 0.0, stats.HitRate)
		assert.Equal(t, 0.0, stats.MissRate)
	})
}

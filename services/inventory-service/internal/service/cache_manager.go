package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	// Cache key patterns for different types of data
	userCachePattern       = "inventory:%s:*"
	classifierCachePattern = "inventory:classifier:%s*"
	balanceCachePattern    = "inventory:%s:%s:*"
)

// cacheManager implements CacheManager interface
type cacheManager struct {
	deps *ServiceDependencies
}

// NewCacheManager creates a new cache manager
func NewCacheManager(deps *ServiceDependencies) CacheManager {
	return &cacheManager{
		deps: deps,
	}
}

// InvalidateUserCache invalidates all cache entries for a specific user
func (cm *cacheManager) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}

	// Create pattern to match all user-related cache keys
	pattern := fmt.Sprintf(userCachePattern, userID.String())

	// Delete all matching keys
	err := cm.deps.Cache.DeletePattern(ctx, pattern)
	if err != nil {
		return errors.Wrapf(err, "failed to invalidate cache for user %s", userID.String())
	}

	return nil
}

// InvalidateClassifierCache invalidates cache for a specific classifier
func (cm *cacheManager) InvalidateClassifierCache(ctx context.Context, classifierCode string) error {
	if classifierCode == "" {
		return errors.New("classifier code cannot be empty")
	}

	// Use the repository's invalidation method which handles all classifier-related caches
	err := cm.deps.Repositories.Classifier.InvalidateCache(ctx, classifierCode)
	if err != nil {
		return errors.Wrapf(err, "failed to invalidate classifier cache for %s", classifierCode)
	}

	return nil
}

// InvalidateUserSectionCache invalidates cache for a specific user and section
func (cm *cacheManager) InvalidateUserSectionCache(ctx context.Context, userID, sectionID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}
	if sectionID == uuid.Nil {
		return errors.New("section ID cannot be nil")
	}

	// Create pattern to match user and section specific cache keys
	pattern := fmt.Sprintf(balanceCachePattern, userID.String(), sectionID.String())

	// Delete all matching keys
	err := cm.deps.Cache.DeletePattern(ctx, pattern)
	if err != nil {
		return errors.Wrapf(err, "failed to invalidate cache for user %s section %s", userID.String(), sectionID.String())
	}

	return nil
}

// InvalidateSpecificItemCache invalidates cache for a specific item balance
func (cm *cacheManager) InvalidateSpecificItemCache(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID) error {
	// Use the balance calculator's invalidation method
	calculator := NewBalanceCalculator(cm.deps)

	req := &BalanceRequest{
		UserID:         userID,
		SectionID:      sectionID,
		ItemID:         itemID,
		CollectionID:   collectionID,
		QualityLevelID: qualityLevelID,
	}

	// This is a type assertion that might fail - we need to check if balance calculator has this method
	if calc, ok := calculator.(*balanceCalculator); ok {
		return calc.InvalidateBalanceCache(ctx, req)
	}

	// Fallback: delete specific cache key manually
	cacheKey := fmt.Sprintf(balanceCacheKeyFormat,
		userID.String(),
		sectionID.String(),
		itemID.String(),
		collectionID.String(),
		qualityLevelID.String(),
	)

	return cm.deps.Cache.Delete(ctx, cacheKey)
}

// WarmupUserCache pre-loads frequently accessed data for a user
func (cm *cacheManager) WarmupUserCache(ctx context.Context, userID, sectionID uuid.UUID) error {
	// Get all items for the user in the section
	itemKeys, err := cm.deps.Repositories.Inventory.GetUserInventoryItems(ctx, userID, sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to get user inventory items for warmup")
	}

	// Pre-calculate and cache balances for all items
	calculator := NewBalanceCalculator(cm.deps)
	for _, itemKey := range itemKeys {
		req := &BalanceRequest{
			UserID:         itemKey.UserID,
			SectionID:      itemKey.SectionID,
			ItemID:         itemKey.ItemID,
			CollectionID:   itemKey.CollectionID,
			QualityLevelID: itemKey.QualityLevelID,
		}

		// Calculate balance (this will cache it)
		_, err := calculator.CalculateCurrentBalance(ctx, req)
		if err != nil {
			// Log error but continue with other items
			continue
		}
	}

	return nil
}

// GetCacheStats returns statistics about cache usage
func (cm *cacheManager) GetCacheStats(ctx context.Context, userID uuid.UUID) (*CacheStats, error) {
	// This would require cache implementation to support stats
	// For now, return a placeholder
	return &CacheStats{
		UserID:    userID,
		TotalKeys: 0, // Would need to count keys with pattern
		HitRate:   0.0,
		MissRate:  0.0,
	}, nil
}

// ClearAllUserCaches clears all user-related caches (admin operation)
func (cm *cacheManager) ClearAllUserCaches(ctx context.Context) error {
	pattern := "inventory:*"

	err := cm.deps.Cache.DeletePattern(ctx, pattern)
	if err != nil {
		return errors.Wrap(err, "failed to clear all user caches")
	}

	return nil
}

// RefreshClassifierCaches refreshes all classifier-related caches
func (cm *cacheManager) RefreshClassifierCaches(ctx context.Context) error {
	// List of all classifiers that should be refreshed
	classifiers := []string{
		"item_class",
		"quality_level",
		"collection",
		"inventory_section",
		"operation_type",
		"resource_type",
		"reagent_type",
		"booster_type",
		"tool_type",
		"key_type",
		"currency_type",
		"tool_quality_levels",
		"key_quality_levels",
	}

	for _, classifier := range classifiers {
		err := cm.InvalidateClassifierCache(ctx, classifier)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh cache for classifier %s", classifier)
		}
	}

	return nil
}

// CacheStats represents cache statistics for a user
type CacheStats struct {
	UserID    uuid.UUID `json:"user_id"`
	TotalKeys int       `json:"total_keys"`
	HitRate   float64   `json:"hit_rate"`
	MissRate  float64   `json:"miss_rate"`
}

// InvalidateCacheForOperations invalidates cache for all users affected by operations
func (cm *cacheManager) InvalidateCacheForOperations(ctx context.Context, operations []*Operation) error {
	// Collect unique user IDs
	userIDs := make(map[uuid.UUID]bool)
	for _, op := range operations {
		userIDs[op.UserID] = true
	}

	// Invalidate cache for each user
	for userID := range userIDs {
		if err := cm.InvalidateUserCache(ctx, userID); err != nil {
			return errors.Wrapf(err, "failed to invalidate cache for user %s", userID.String())
		}
	}

	return nil
}

// Helper type for operations (should reference models.Operation in practice)
type Operation struct {
	UserID uuid.UUID
	// ... other fields
}

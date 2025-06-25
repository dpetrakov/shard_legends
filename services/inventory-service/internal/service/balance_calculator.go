package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	// Cache TTL for balance calculations (1 hour)
	balanceCacheTTL = 1 * time.Hour

	// Cache key format for balance
	balanceCacheKeyFormat = "inventory:%s:%s:%s:%s:%s"
)

// balanceCalculator implements BalanceCalculator interface
type balanceCalculator struct {
	deps *ServiceDependencies
}

// NewBalanceCalculator creates a new balance calculator
func NewBalanceCalculator(deps *ServiceDependencies) BalanceCalculator {
	return &balanceCalculator{
		deps: deps,
	}
}

// CalculateCurrentBalance calculates the current balance for a specific item
func (bc *balanceCalculator) CalculateCurrentBalance(ctx context.Context, req *BalanceRequest) (int64, error) {
	start := time.Now()
	if req == nil {
		return 0, errors.New("balance request cannot be nil")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf(balanceCacheKeyFormat,
		req.UserID.String(),
		req.SectionID.String(),
		req.ItemID.String(),
		req.CollectionID.String(),
		req.QualityLevelID.String(),
	)

	// Try to get from cache first
	var cachedBalance int64
	if err := bc.deps.Cache.Get(ctx, cacheKey, &cachedBalance); err == nil {
		if bc.deps.Metrics != nil {
			bc.deps.Metrics.RecordCacheHit("balance")
		}
		return cachedBalance, nil
	}

	if bc.deps.Metrics != nil {
		bc.deps.Metrics.RecordCacheMiss("balance")
	}

	// Calculate balance from database
	balance, err := bc.calculateFromDatabase(ctx, req)
	if err != nil {
		if bc.deps.Metrics != nil {
			bc.deps.Metrics.RecordBalanceCalculation(req.SectionID.String(), time.Since(start), "error")
		}
		return 0, errors.Wrap(err, "failed to calculate balance from database")
	}

	// Cache the result (ignore cache errors for graceful degradation)
	_ = bc.deps.Cache.Set(ctx, cacheKey, balance, balanceCacheTTL)

	if bc.deps.Metrics != nil {
		bc.deps.Metrics.RecordBalanceCalculation(req.SectionID.String(), time.Since(start), "success")
	}

	return balance, nil
}

// calculateFromDatabase calculates balance from database data
func (bc *balanceCalculator) calculateFromDatabase(ctx context.Context, req *BalanceRequest) (int64, error) {
	// Find the latest daily balance
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	latestBalance, err := bc.deps.Repositories.Inventory.GetLatestDailyBalance(
		ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, today,
	)

	var baseBalance int64 = 0
	var fromDate time.Time

	if err != nil || latestBalance == nil {
		// No daily balance found, start from zero and calculate from beginning of time
		fromDate = time.Time{} // This will get all operations
		baseBalance = 0
	} else {
		// Use the daily balance as base
		baseBalance = latestBalance.Quantity
		// Get operations from the day after the balance date
		fromDate = latestBalance.BalanceDate.Add(24 * time.Hour)
	}

	// If no daily balance exists for yesterday, create one lazily
	if latestBalance == nil || !isSameDay(latestBalance.BalanceDate, yesterday) {
		// We need to create yesterday's daily balance
		dailyBalanceReq := &DailyBalanceRequest{
			UserID:         req.UserID,
			SectionID:      req.SectionID,
			ItemID:         req.ItemID,
			CollectionID:   req.CollectionID,
			QualityLevelID: req.QualityLevelID,
			TargetDate:     yesterday,
		}

		creator := NewDailyBalanceCreator(bc.deps)
		createdBalance, err := creator.CreateDailyBalance(ctx, dailyBalanceReq)
		if err != nil {
			return 0, errors.Wrap(err, "failed to create daily balance")
		}

		// Use the created balance as base and get today's operations
		baseBalance = createdBalance.Quantity
		fromDate = today // Start of today
	}

	// Get all operations from the from date
	operations, err := bc.deps.Repositories.Inventory.GetOperations(
		ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, fromDate,
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get operations")
	}

	// Calculate current balance: base + sum of operations
	currentBalance := baseBalance
	for _, op := range operations {
		currentBalance += op.QuantityChange
	}

	return currentBalance, nil
}

// isSameDay checks if two times are on the same day (UTC)
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.UTC().Date()
	y2, m2, d2 := t2.UTC().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// InvalidateBalanceCache invalidates the balance cache for a specific item
func (bc *balanceCalculator) InvalidateBalanceCache(ctx context.Context, req *BalanceRequest) error {
	cacheKey := fmt.Sprintf(balanceCacheKeyFormat,
		req.UserID.String(),
		req.SectionID.String(),
		req.ItemID.String(),
		req.CollectionID.String(),
		req.QualityLevelID.String(),
	)

	return bc.deps.Cache.Delete(ctx, cacheKey)
}

// CacheBalance manually caches a balance value
func (bc *balanceCalculator) CacheBalance(ctx context.Context, req *BalanceRequest, balance int64) error {
	cacheKey := fmt.Sprintf(balanceCacheKeyFormat,
		req.UserID.String(),
		req.SectionID.String(),
		req.ItemID.String(),
		req.CollectionID.String(),
		req.QualityLevelID.String(),
	)

	return bc.deps.Cache.Set(ctx, cacheKey, balance, balanceCacheTTL)
}

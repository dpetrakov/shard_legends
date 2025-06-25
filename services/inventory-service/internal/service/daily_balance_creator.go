package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/shard-legends/inventory-service/internal/models"
)

// dailyBalanceCreator implements DailyBalanceCreator interface
type dailyBalanceCreator struct {
	deps *ServiceDependencies
}

// NewDailyBalanceCreator creates a new daily balance creator
func NewDailyBalanceCreator(deps *ServiceDependencies) DailyBalanceCreator {
	return &dailyBalanceCreator{
		deps: deps,
	}
}

// CreateDailyBalance creates a daily balance for a specific target date (usually yesterday)
func (dbc *dailyBalanceCreator) CreateDailyBalance(ctx context.Context, req *DailyBalanceRequest) (*models.DailyBalance, error) {
	if req == nil {
		return nil, errors.New("daily balance request cannot be nil")
	}

	// Normalize target date to beginning of day UTC
	targetDateStart := req.TargetDate.UTC().Truncate(24 * time.Hour)
	// End of target date (for filtering operations)
	targetDateEnd := targetDateStart.Add(24*time.Hour - time.Second)

	// Check if daily balance already exists for target date
	existingBalance, err := dbc.deps.Repositories.Inventory.GetDailyBalance(
		ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, targetDateStart,
	)
	if err == nil && existingBalance != nil {
		// Balance already exists, return it
		return existingBalance, nil
	}

	// Find the last existing daily balance before target date
	previousBalance, err := dbc.deps.Repositories.Inventory.GetLatestDailyBalance(
		ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, targetDateStart,
	)

	var baseBalance int64 = 0
	var fromDate time.Time

	if err != nil || previousBalance == nil {
		// No previous balance found, start from zero and beginning of time
		fromDate = time.Time{}
	} else {
		// Use previous balance as base
		baseBalance = previousBalance.Quantity
		// Calculate operations from the day after the previous balance
		fromDate = previousBalance.BalanceDate.Add(24 * time.Hour).Truncate(24 * time.Hour)
	}

	// Get all operations from fromDate
	operations, err := dbc.deps.Repositories.Inventory.GetOperations(
		ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, fromDate,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operations for daily balance calculation")
	}

	// Calculate the balance for target date by filtering operations
	newBalance := baseBalance
	for _, op := range operations {
		// Only include operations up to end of target date
		if (op.CreatedAt.After(fromDate) || op.CreatedAt.Equal(fromDate)) && 
		   (op.CreatedAt.Before(targetDateEnd) || op.CreatedAt.Equal(targetDateEnd)) {
			newBalance += op.QuantityChange
		}
	}

	// Create the daily balance record
	dailyBalance := &models.DailyBalance{
		UserID:         req.UserID,
		SectionID:      req.SectionID,
		ItemID:         req.ItemID,
		CollectionID:   req.CollectionID,
		QualityLevelID: req.QualityLevelID,
		BalanceDate:    targetDateStart, // Use start of day, not end
		Quantity:       newBalance,
		CreatedAt:      time.Now().UTC(),
	}

	// Validate the daily balance
	if err := models.ValidateDailyBalance(dailyBalance); err != nil {
		return nil, errors.Wrap(err, "daily balance validation failed")
	}

	// Save to database
	err = dbc.deps.Repositories.Inventory.CreateDailyBalance(ctx, dailyBalance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create daily balance in database")
	}

	return dailyBalance, nil
}

// CreateMissingDailyBalances creates missing daily balances for a range of dates
func (dbc *dailyBalanceCreator) CreateMissingDailyBalances(ctx context.Context, req *DailyBalanceRequest, endDate time.Time) ([]*models.DailyBalance, error) {
	var createdBalances []*models.DailyBalance

	currentDate := req.TargetDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		balanceReq := &DailyBalanceRequest{
			UserID:         req.UserID,
			SectionID:      req.SectionID,
			ItemID:         req.ItemID,
			CollectionID:   req.CollectionID,
			QualityLevelID: req.QualityLevelID,
			TargetDate:     currentDate,
		}

		balance, err := dbc.CreateDailyBalance(ctx, balanceReq)
		if err != nil {
			return createdBalances, errors.Wrapf(err, "failed to create daily balance for date %s", currentDate.Format("2006-01-02"))
		}

		createdBalances = append(createdBalances, balance)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return createdBalances, nil
}

// GetOrCreateYesterdayBalance gets or creates yesterday's daily balance
func (dbc *dailyBalanceCreator) GetOrCreateYesterdayBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID) (*models.DailyBalance, error) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)

	req := &DailyBalanceRequest{
		UserID:         userID,
		SectionID:      sectionID,
		ItemID:         itemID,
		CollectionID:   collectionID,
		QualityLevelID: qualityLevelID,
		TargetDate:     yesterday,
	}

	return dbc.CreateDailyBalance(ctx, req)
}

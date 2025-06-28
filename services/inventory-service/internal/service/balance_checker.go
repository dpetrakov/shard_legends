package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	pkgerrors "github.com/pkg/errors"

	"github.com/shard-legends/inventory-service/internal/models"
)

// InsufficientBalanceError represents an error when there are insufficient items
type InsufficientBalanceError struct {
	Message      string               `json:"message"`
	MissingItems []MissingItemDetails `json:"missing_items"`
}

func (e *InsufficientBalanceError) Error() string {
	return e.Message
}

// balanceChecker implements BalanceChecker interface
type balanceChecker struct {
	deps *ServiceDependencies
}

// NewBalanceChecker creates a new balance checker
func NewBalanceChecker(deps *ServiceDependencies) BalanceChecker {
	return &balanceChecker{
		deps: deps,
	}
}

// CheckSufficientBalance checks if user has sufficient balance for requested items
func (bc *balanceChecker) CheckSufficientBalance(ctx context.Context, req *SufficientBalanceRequest) error {
	if req == nil {
		return pkgerrors.New("sufficient balance request cannot be nil")
	}

	if len(req.Items) == 0 {
		return pkgerrors.New("items list cannot be empty")
	}

	var missingItems []MissingItemDetails
	calculator := NewBalanceCalculator(bc.deps)

	// Check each item
	for _, item := range req.Items {
		if item.RequiredQty <= 0 {
			return pkgerrors.Errorf("required quantity must be positive for item %s", item.ItemID.String())
		}

		// Calculate current balance for this item
		balanceReq := &BalanceRequest{
			UserID:         req.UserID,
			SectionID:      req.SectionID,
			ItemID:         item.ItemID,
			CollectionID:   item.CollectionID,
			QualityLevelID: item.QualityLevelID,
		}

		currentBalance, err := calculator.CalculateCurrentBalance(ctx, balanceReq)
		if err != nil {
			return pkgerrors.Wrapf(err, "failed to calculate balance for item %s", item.ItemID.String())
		}

		// Check if balance is sufficient
		if currentBalance < item.RequiredQty {
			missing := item.RequiredQty - currentBalance
			missingItems = append(missingItems, MissingItemDetails{
				ItemID:         item.ItemID,
				CollectionID:   item.CollectionID,
				QualityLevelID: item.QualityLevelID,
				Required:       item.RequiredQty,
				Available:      currentBalance,
				Missing:        missing,
			})
		}
	}

	// If there are missing items, return error with details
	if len(missingItems) > 0 {
		return &InsufficientBalanceError{
			Message:      fmt.Sprintf("Insufficient balance for %d item(s)", len(missingItems)),
			MissingItems: missingItems,
		}
	}

	return nil
}

// CheckSufficientBalanceForItems is a convenience method that takes item quantity requests
func (bc *balanceChecker) CheckSufficientBalanceForItems(ctx context.Context, userID, sectionID uuid.UUID, items []models.ItemQuantityRequest) error {
	if len(items) == 0 {
		return pkgerrors.New("items list cannot be empty")
	}

	// Convert to ItemQuantityCheck format
	checkItems := make([]ItemQuantityCheck, len(items))
	for i, item := range items {
		// For this conversion, we assume that codes have already been converted to UUIDs
		// In a real implementation, you'd need to convert codes to UUIDs first
		checkItems[i] = ItemQuantityCheck{
			ItemID: item.ItemID,
			// CollectionID and QualityLevelID would need to be obtained from code conversion
			RequiredQty: item.Quantity,
		}
	}

	req := &SufficientBalanceRequest{
		UserID:    userID,
		SectionID: sectionID,
		Items:     checkItems,
	}

	return bc.CheckSufficientBalance(ctx, req)
}

// GetAvailableBalance returns the available balance for specific items
func (bc *balanceChecker) GetAvailableBalance(ctx context.Context, userID, sectionID uuid.UUID, items []ItemQuantityCheck) (map[string]int64, error) {
	result := make(map[string]int64)
	calculator := NewBalanceCalculator(bc.deps)

	for _, item := range items {
		balanceReq := &BalanceRequest{
			UserID:         userID,
			SectionID:      sectionID,
			ItemID:         item.ItemID,
			CollectionID:   item.CollectionID,
			QualityLevelID: item.QualityLevelID,
		}

		balance, err := calculator.CalculateCurrentBalance(ctx, balanceReq)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "failed to get balance for item %s", item.ItemID.String())
		}

		// Create a unique key for this item combination
		key := fmt.Sprintf("%s:%s:%s",
			item.ItemID.String(),
			item.CollectionID.String(),
			item.QualityLevelID.String())

		result[key] = balance
	}

	return result, nil
}

// ValidateItemQuantities validates that all quantities are positive
func (bc *balanceChecker) ValidateItemQuantities(items []ItemQuantityCheck) error {
	for i, item := range items {
		if item.RequiredQty <= 0 {
			return pkgerrors.Errorf("item at index %d has invalid quantity: %d", i, item.RequiredQty)
		}

		if item.ItemID == uuid.Nil {
			return pkgerrors.Errorf("item at index %d has invalid item ID", i)
		}

		if item.CollectionID == uuid.Nil {
			return pkgerrors.Errorf("item at index %d has invalid collection ID", i)
		}

		if item.QualityLevelID == uuid.Nil {
			return pkgerrors.Errorf("item at index %d has invalid quality level ID", i)
		}
	}

	return nil
}

// IsInsufficientBalanceError checks if an error is an insufficient balance error
func IsInsufficientBalanceError(err error) bool {
	var insuffErr *InsufficientBalanceError
	return errors.As(err, &insuffErr)
}

// GetMissingItemsFromError extracts missing items from an insufficient balance error
func GetMissingItemsFromError(err error) ([]MissingItemDetails, bool) {
	var insuffErr *InsufficientBalanceError
	if errors.As(err, &insuffErr) {
		return insuffErr.MissingItems, true
	}
	return nil, false
}

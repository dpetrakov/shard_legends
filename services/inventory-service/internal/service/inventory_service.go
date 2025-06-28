package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	
	"github.com/shard-legends/inventory-service/internal/models"
)

// inventoryService implements InventoryService interface
type inventoryService struct {
	deps *ServiceDependencies
	
	// Component services
	balanceCalculator   BalanceCalculator
	dailyBalanceCreator DailyBalanceCreator
	codeConverter       CodeConverter
	balanceChecker      BalanceChecker
	operationCreator    OperationCreator
	cacheManager        CacheManager
}

// NewInventoryService creates a new inventory service with all components
func NewInventoryService(deps *ServiceDependencies) InventoryService {
	return &inventoryService{
		deps:                deps,
		balanceCalculator:   NewBalanceCalculator(deps),
		dailyBalanceCreator: NewDailyBalanceCreator(deps),
		codeConverter:       NewCodeConverter(deps),
		balanceChecker:      NewBalanceChecker(deps),
		operationCreator:    NewOperationCreator(deps),
		cacheManager:        NewCacheManager(deps),
	}
}

// Core operations

// CalculateCurrentBalance calculates the current balance for a specific item
func (is *inventoryService) CalculateCurrentBalance(ctx context.Context, req *BalanceRequest) (int64, error) {
	return is.balanceCalculator.CalculateCurrentBalance(ctx, req)
}

// CreateDailyBalance creates a daily balance record for a specific date
func (is *inventoryService) CreateDailyBalance(ctx context.Context, req *DailyBalanceRequest) (*models.DailyBalance, error) {
	return is.dailyBalanceCreator.CreateDailyBalance(ctx, req)
}

// CheckSufficientBalance checks if user has sufficient balance for requested items
func (is *inventoryService) CheckSufficientBalance(ctx context.Context, req *SufficientBalanceRequest) error {
	return is.balanceChecker.CheckSufficientBalance(ctx, req)
}

// CreateOperationsInTransaction creates multiple operations within a database transaction
func (is *inventoryService) CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error) {
	return is.operationCreator.CreateOperationsInTransaction(ctx, operations)
}

// Utility operations

// ConvertClassifierCodes converts between classifier codes and UUIDs
func (is *inventoryService) ConvertClassifierCodes(ctx context.Context, req *CodeConversionRequest) (*CodeConversionResponse, error) {
	return is.codeConverter.ConvertClassifierCodes(ctx, req)
}

// InvalidateUserCache invalidates all cache entries for a specific user
func (is *inventoryService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	return is.cacheManager.InvalidateUserCache(ctx, userID)
}

// High-level business operations

// GetUserInventory returns all inventory items for a user in a specific section
// D-15: Оптимизированная версия использует GetUserInventoryOptimized для устранения N+1 запросов
func (is *inventoryService) GetUserInventory(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error) {
	// D-15: Используем оптимизированный метод repository, который устраняет N+1 запросы
	// Вместо множественных отдельных запросов выполняется один JOIN запрос
	result, err := is.deps.Repositories.Inventory.GetUserInventoryOptimized(ctx, userID, sectionID)
	if err != nil {
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("get_inventory", sectionID.String(), "error")
		}
		return nil, errors.Wrap(err, "failed to get user inventory with optimized method")
	}

	// Record metrics for successful operation
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("get_inventory", sectionID.String(), "success")
		is.deps.Metrics.RecordItemsPerInventory(sectionID.String(), len(result))
	}

	return result, nil
}

// GetUserInventoryLegacy - оригинальная версия с N+1 запросами (для сравнения производительности)
// DEPRECATED: D-15 - используйте GetUserInventory вместо этого метода
func (is *inventoryService) GetUserInventoryLegacy(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error) {
	// Get all item keys for the user
	itemKeys, err := is.deps.Repositories.Inventory.GetUserInventoryItems(ctx, userID, sectionID)
	if err != nil {
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("get_inventory_legacy", sectionID.String(), "error")
		}
		return nil, errors.Wrap(err, "failed to get user inventory items")
	}

	var result []*models.InventoryItemResponse

	// Calculate balance for each item (N+1 problem here)
	for _, itemKey := range itemKeys {
		// Get current balance
		balanceReq := &BalanceRequest{
			UserID:         itemKey.UserID,
			SectionID:      itemKey.SectionID,
			ItemID:         itemKey.ItemID,
			CollectionID:   itemKey.CollectionID,
			QualityLevelID: itemKey.QualityLevelID,
		}

		balance, err := is.balanceCalculator.CalculateCurrentBalance(ctx, balanceReq)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to calculate balance for item %s", itemKey.ItemID.String())
		}

		// Skip items with zero balance
		if balance <= 0 {
			continue
		}

		// Get item details with classifier codes (N+1 problem here)
		itemDetails, err := is.deps.Repositories.Item.GetItemWithDetails(ctx, itemKey.ItemID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get item details for %s", itemKey.ItemID.String())
		}

		// Convert UUIDs to codes for response (N+1 problem here)
		collectionCode, err := is.getCodeFromUUID(ctx, models.ClassifierCollection, itemKey.CollectionID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get collection code")
		}

		qualityCode, err := is.getCodeFromUUID(ctx, models.ClassifierQualityLevel, itemKey.QualityLevelID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get quality level code")
		}

		// Create response item
		item := &models.InventoryItemResponse{
			ItemID:       itemKey.ItemID,
			ItemClass:    itemDetails.ItemClass,
			ItemType:     itemDetails.ItemType,
			Collection:   &collectionCode,
			QualityLevel: &qualityCode,
			Quantity:     balance,
		}

		result = append(result, item)
	}

	// Record metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("get_inventory_legacy", sectionID.String(), "success")
		is.deps.Metrics.RecordItemsPerInventory(sectionID.String(), len(result))
	}

	return result, nil
}


// AddItems adds items to user's inventory
func (is *inventoryService) AddItems(ctx context.Context, req *models.AddItemsRequest) ([]uuid.UUID, error) {
	if err := models.ValidateAddItemsRequest(req); err != nil {
		return nil, errors.Wrap(err, "add items request validation failed")
	}

	// Get section ID
	sectionID, err := is.getSectionID(ctx, req.Section)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get section ID")
	}

	// Get operation type ID
	operationTypeID, err := is.getOperationTypeID(ctx, req.OperationType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operation type ID")
	}

	// Create operations for each item
	var operations []*models.Operation
	for _, item := range req.Items {
		// Convert codes to UUIDs
		collectionID, qualityLevelID, err := is.convertItemCodes(ctx, item.Collection, item.QualityLevel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert item codes")
		}

		op := &models.Operation{
			UserID:          req.UserID,
			SectionID:       sectionID,
			ItemID:          item.ItemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  item.Quantity, // Positive for adding
			OperationTypeID: operationTypeID,
			OperationID:     &req.OperationID,
			Comment:         req.Comment,
		}

		operations = append(operations, op)
	}

	return is.operationCreator.CreateOperationsInTransaction(ctx, operations)
}

// Helper methods

// convertToBalanceCheck converts ItemQuantityRequest to ItemQuantityCheck
func (is *inventoryService) convertToBalanceCheck(ctx context.Context, userID uuid.UUID, items []models.ItemQuantityRequest) ([]ItemQuantityCheck, error) {
	var result []ItemQuantityCheck

	for _, item := range items {
		// Convert codes to UUIDs
		collectionID, qualityLevelID, err := is.convertItemCodes(ctx, item.Collection, item.QualityLevel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert item codes")
		}

		check := ItemQuantityCheck{
			ItemID:         item.ItemID,
			CollectionID:   collectionID,
			QualityLevelID: qualityLevelID,
			RequiredQty:    item.Quantity,
		}

		result = append(result, check)
	}

	return result, nil
}

// convertItemCodes converts optional collection and quality level codes to UUIDs
func (is *inventoryService) convertItemCodes(ctx context.Context, collection, qualityLevel *string) (uuid.UUID, uuid.UUID, error) {
	// Get default UUIDs (these should be configurable)
	defaultCollection := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default collection
	defaultQuality := uuid.MustParse("00000000-0000-0000-0000-000000000002")    // Default quality

	collectionID := defaultCollection
	qualityLevelID := defaultQuality

	// Convert collection if provided
	if collection != nil && *collection != "" {
		mapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
		if err != nil {
			return uuid.Nil, uuid.Nil, errors.Wrap(err, "failed to get collection mapping")
		}

		if id, found := mapping[*collection]; found {
			collectionID = id
		} else {
			return uuid.Nil, uuid.Nil, errors.Errorf("unknown collection code: %s", *collection)
		}
	}

	// Convert quality level if provided
	if qualityLevel != nil && *qualityLevel != "" {
		mapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
		if err != nil {
			return uuid.Nil, uuid.Nil, errors.Wrap(err, "failed to get quality level mapping")
		}

		if id, found := mapping[*qualityLevel]; found {
			qualityLevelID = id
		} else {
			return uuid.Nil, uuid.Nil, errors.Errorf("unknown quality level code: %s", *qualityLevel)
		}
	}

	return collectionID, qualityLevelID, nil
}

// getSectionID gets section UUID from code
func (is *inventoryService) getSectionID(ctx context.Context, sectionCode string) (uuid.UUID, error) {
	mapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get section mapping")
	}

	if id, found := mapping[sectionCode]; found {
		return id, nil
	}

	return uuid.Nil, errors.Errorf("unknown section code: %s", sectionCode)
}

// getOperationTypeID gets operation type UUID from code
func (is *inventoryService) getOperationTypeID(ctx context.Context, operationCode string) (uuid.UUID, error) {
	mapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get operation type mapping")
	}

	if id, found := mapping[operationCode]; found {
		return id, nil
	}

	return uuid.Nil, errors.Errorf("unknown operation type code: %s", operationCode)
}

// getCodeFromUUID gets classifier code from UUID
func (is *inventoryService) getCodeFromUUID(ctx context.Context, classifierType string, id uuid.UUID) (string, error) {
	mapping, err := is.deps.Repositories.Classifier.GetUUIDToCodeMapping(ctx, classifierType)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get UUID mapping for %s", classifierType)
	}

	if code, found := mapping[id]; found {
		return code, nil
	}

	return "", errors.Errorf("unknown UUID %s for classifier %s", id.String(), classifierType)
}

// AdjustInventory performs administrative inventory adjustments
func (is *inventoryService) AdjustInventory(ctx context.Context, req *models.AdjustInventoryRequest) (*models.AdjustInventoryResponse, error) {
	if err := models.ValidateAdjustInventoryRequest(req); err != nil {
		return nil, errors.Wrap(err, "adjust inventory request validation failed")
	}

	// Convert section code to UUID
	sectionMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get section mapping")
	}
	sectionID, exists := sectionMapping[req.Section]
	if !exists {
		return nil, errors.Errorf("unknown section code: %s", req.Section)
	}

	// Get admin adjustment operation type UUID
	operationTypeMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operation type mapping")
	}
	operationTypeID, exists := operationTypeMapping[models.OperationTypeAdminAdjustment]
	if !exists {
		return nil, errors.Errorf("unknown operation type: %s", models.OperationTypeAdminAdjustment)
	}

	// Create operations for each item adjustment
	var operations []*models.Operation
	operationID := uuid.New() // Single operation ID for the whole adjustment

	for _, itemReq := range req.Items {
		// Convert collection and quality codes to UUIDs if provided
		collectionID := uuid.Nil
		qualityLevelID := uuid.Nil

		if itemReq.Collection != nil && *itemReq.Collection != "" {
			collectionMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get collection mapping")
			}
			if id, exists := collectionMapping[*itemReq.Collection]; exists {
				collectionID = id
			} else {
				return nil, errors.Errorf("unknown collection code: %s", *itemReq.Collection)
			}
		}

		if itemReq.QualityLevel != nil && *itemReq.QualityLevel != "" {
			qualityMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get quality level mapping")
			}
			if id, exists := qualityMapping[*itemReq.QualityLevel]; exists {
				qualityLevelID = id
			} else {
				return nil, errors.Errorf("unknown quality level code: %s", *itemReq.QualityLevel)
			}
		}

		// Create operation
		operation := &models.Operation{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       sectionID,
			ItemID:          itemReq.ItemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  itemReq.QuantityChange,
			OperationTypeID: operationTypeID,
			OperationID:     &operationID,
			Comment:         &req.Reason,
			CreatedAt:       time.Now().UTC(),
		}

		operations = append(operations, operation)
	}

	// Create operations in transaction
	operationIDs, err := is.CreateOperationsInTransaction(ctx, operations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create adjustment operations")
	}

	// Calculate final balances for each adjusted item
	var finalBalances []models.InventoryItemResponse
	for _, itemReq := range req.Items {
		// Get item details
		itemDetails, err := is.deps.Repositories.Item.GetItemWithDetails(ctx, itemReq.ItemID)
		if err != nil {
			// Log error but continue with other items
			continue
		}

		// Convert collection and quality UUIDs back to codes
		var collectionCode, qualityCode *string

		if itemReq.Collection != nil && *itemReq.Collection != "" {
			collectionCode = itemReq.Collection
		}

		if itemReq.QualityLevel != nil && *itemReq.QualityLevel != "" {
			qualityCode = itemReq.QualityLevel
		}

		// Calculate current balance after adjustment
		collectionID := uuid.Nil
		qualityLevelID := uuid.Nil

		if collectionCode != nil {
			if collectionMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection); err == nil {
				if id, exists := collectionMapping[*collectionCode]; exists {
					collectionID = id
				}
			}
		}

		if qualityCode != nil {
			if qualityMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel); err == nil {
				if id, exists := qualityMapping[*qualityCode]; exists {
					qualityLevelID = id
				}
			}
		}

		balanceReq := &BalanceRequest{
			UserID:         req.UserID,
			SectionID:      sectionID,
			ItemID:         itemReq.ItemID,
			CollectionID:   collectionID,
			QualityLevelID: qualityLevelID,
		}

		currentBalance, err := is.balanceCalculator.CalculateCurrentBalance(ctx, balanceReq)
		if err != nil {
			currentBalance = 0 // Default to 0 if calculation fails
		}

		finalBalance := models.InventoryItemResponse{
			ItemID:       itemReq.ItemID,
			ItemClass:    itemDetails.ItemClass,
			ItemType:     itemDetails.ItemType,
			Collection:   collectionCode,
			QualityLevel: qualityCode,
			Quantity:     currentBalance,
		}

		finalBalances = append(finalBalances, finalBalance)
	}

	// Record metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("adjust_inventory", req.Section, "success")
	}

	return &models.AdjustInventoryResponse{
		Success:       true,
		OperationIDs:  operationIDs,
		FinalBalances: finalBalances,
	}, nil
}

// Reservation operations

// ReserveItems reserves items for factory production
func (is *inventoryService) ReserveItems(ctx context.Context, req *models.ReserveItemsRequest) ([]uuid.UUID, error) {
	if req == nil {
		return nil, errors.New("reserve items request cannot be nil")
	}

	// Use operation creator to handle the reservation logic
	operationIDs, err := is.operationCreator.CreateReservationOperations(ctx, req)
	if err != nil {
		// Record failure metrics
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("reserve_items", "main", "failure")
		}
		return nil, errors.Wrap(err, "failed to create reservation operations")
	}

	// Record success metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("reserve_items", "main", "success")
	}

	return operationIDs, nil
}

// ReturnReservedItems returns reserved items back to main inventory
func (is *inventoryService) ReturnReservedItems(ctx context.Context, req *models.ReturnReserveRequest) error {
	if req == nil {
		return errors.New("return reserve request cannot be nil")
	}

	// Use operation creator to handle the return logic
	err := is.operationCreator.CreateReturnOperations(ctx, req)
	if err != nil {
		// Record failure metrics
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("return_reserved", "factory", "failure")
		}
		return errors.Wrap(err, "failed to create return operations")
	}

	// Record success metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("return_reserved", "factory", "success")
	}

	return nil
}

// ConsumeReservedItems consumes reserved items (destroys them)
func (is *inventoryService) ConsumeReservedItems(ctx context.Context, req *models.ConsumeReserveRequest) error {
	if req == nil {
		return errors.New("consume reserve request cannot be nil")
	}

	// Use operation creator to handle the consumption logic
	err := is.operationCreator.CreateConsumptionOperations(ctx, req)
	if err != nil {
		// Record failure metrics
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("consume_reserved", "factory", "failure")
		}
		return errors.Wrap(err, "failed to create consumption operations")
	}

	// Record success metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("consume_reserved", "factory", "success")
	}

	return nil
}
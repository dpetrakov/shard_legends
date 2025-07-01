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
	// Get default collection UUID by looking up "base" code
	collectionMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.Wrap(err, "failed to get collection mapping for defaults")
	}

	defaultCollection, found := collectionMapping["base"]
	if !found {
		return uuid.Nil, uuid.Nil, errors.New("base collection not found in classifier mapping")
	}

	// Get default quality level UUID by looking up "base" code
	qualityMapping, err := is.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.Wrap(err, "failed to get quality level mapping for defaults")
	}

	defaultQuality, found := qualityMapping["base"]
	if !found {
		return uuid.Nil, uuid.Nil, errors.New("base quality level not found in classifier mapping")
	}

	collectionID := defaultCollection
	qualityLevelID := defaultQuality

	// Convert collection if provided
	if collection != nil && *collection != "" {
		if id, found := collectionMapping[*collection]; found {
			collectionID = id
		} else {
			return uuid.Nil, uuid.Nil, errors.Errorf("unknown collection code: %s", *collection)
		}
	}

	// Convert quality level if provided
	if qualityLevel != nil && *qualityLevel != "" {
		if id, found := qualityMapping[*qualityLevel]; found {
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

// GetItemsDetails gets localized item details with images
func (is *inventoryService) GetItemsDetails(ctx context.Context, req *models.ItemDetailsRequest, languageCode string) (*models.ItemDetailsResponse, error) {
	if req == nil {
		return nil, errors.New("item details request cannot be nil")
	}

	if len(req.Items) == 0 {
		return &models.ItemDetailsResponse{Items: []models.ItemDetailResponseItem{}}, nil
	}

	// Extract item IDs for batch operations
	itemIDs := make([]uuid.UUID, len(req.Items))
	for i, item := range req.Items {
		itemIDs[i] = item.ItemID
	}

	// Get items batch
	itemsMap, err := is.deps.Repositories.Item.GetItemsBatch(ctx, itemIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get items batch")
	}

	// Get translations for requested language
	translationsMap, err := is.deps.Repositories.Item.GetTranslationsBatch(ctx, models.EntityTypeItem, itemIDs, languageCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get translations batch")
	}

	// Get default language for fallback
	defaultLang, err := is.deps.Repositories.Item.GetDefaultLanguage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default language")
	}

	// Get fallback translations if needed
	var fallbackTranslations map[uuid.UUID]map[string]string
	if defaultLang != nil && defaultLang.Code != languageCode {
		fallbackTranslations, err = is.deps.Repositories.Item.GetTranslationsBatch(ctx, models.EntityTypeItem, itemIDs, defaultLang.Code)
		if err != nil {
			// Log error but continue without fallback
			fallbackTranslations = make(map[uuid.UUID]map[string]string)
		}
	} else {
		fallbackTranslations = make(map[uuid.UUID]map[string]string)
	}

	// Get images batch (simplified for now)
	imagesMap, err := is.deps.Repositories.Item.GetItemImagesBatch(ctx, req.Items)
	if err != nil {
		// Log error but continue without images
		imagesMap = make(map[string]string)
	}

	// Build response
	response := &models.ItemDetailsResponse{
		Items: make([]models.ItemDetailResponseItem, 0, len(req.Items)),
	}

	for _, requestItem := range req.Items {
		item, exists := itemsMap[requestItem.ItemID]
		if !exists {
			// Skip items that don't exist
			continue
		}

		// Get translations with fallback
		translations := translationsMap[requestItem.ItemID]
		if translations == nil {
			translations = make(map[string]string)
		}

		name := translations[models.FieldNameName]
		description := translations[models.FieldNameDescription]

		// Apply fallback if translations are missing
		if name == "" {
			if fallbackTranslations[requestItem.ItemID] != nil {
				name = fallbackTranslations[requestItem.ItemID][models.FieldNameName]
			}
		}
		if description == "" {
			if fallbackTranslations[requestItem.ItemID] != nil {
				description = fallbackTranslations[requestItem.ItemID][models.FieldNameDescription]
			}
		}

		// Create image key for lookup (simplified approach)
		imageKey := requestItem.ItemID.String()
		if requestItem.Collection != nil {
			imageKey += "_" + *requestItem.Collection
		}
		if requestItem.QualityLevel != nil {
			imageKey += "_" + *requestItem.QualityLevel
		}

		imageURL := imagesMap[imageKey]
		if imageURL == "" {
			// Default image fallback
			imageURL = "/images/items/default.png"
		}

		responseItem := models.ItemDetailResponseItem{
			ItemID:       requestItem.ItemID,
			ItemClass:    item.ItemClass,
			ItemType:     item.ItemType,
			Name:         name,
			Description:  description,
			ImageURL:     imageURL,
			Collection:   requestItem.Collection,
			QualityLevel: requestItem.QualityLevel,
		}

		response.Items = append(response.Items, responseItem)
	}

	// Record metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("get_items_details", languageCode, "success")
	}

	return response, nil
}

func (is *inventoryService) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	return is.deps.Repositories.Item.GetDefaultLanguage(ctx)
}

// GetReservationStatus checks the status of a reservation by operation ID
func (is *inventoryService) GetReservationStatus(ctx context.Context, operationID uuid.UUID) (*models.ReservationStatusResponse, error) {
	// Get all operations for this reservation operation ID
	operations, err := is.deps.Repositories.Inventory.GetOperationsByExternalID(ctx, operationID)
	if err != nil {
		// Record failure metrics
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("get_reservation_status", "factory", "error")
		}
		return nil, errors.Wrap(err, "failed to get operations by external ID")
	}

	// If no operations found, return reservation not found
	if len(operations) == 0 {
		response := &models.ReservationStatusResponse{
			ReservationExists: false,
			Error:             stringPtr("Reservation not found"),
		}

		// Record metrics
		if is.deps.Metrics != nil {
			is.deps.Metrics.RecordInventoryOperation("get_reservation_status", "factory", "not_found")
		}

		return response, nil
	}

	// Filter operations to get only factory reservation operations
	var reservationOps []*models.Operation
	var userID uuid.UUID
	var reservationDate *string

	for _, op := range operations {
		// Get operation type code to determine if it's a reservation
		operationTypeCode, err := is.getOperationTypeCode(ctx, op.OperationTypeID)
		if err != nil {
			continue // Skip operations we can't identify
		}

		if operationTypeCode == models.OperationTypeFactoryReservation {
			reservationOps = append(reservationOps, op)
			if userID == uuid.Nil {
				userID = op.UserID
			}
			if reservationDate == nil {
				dateStr := op.CreatedAt.Format("2006-01-02T15:04:05Z")
				reservationDate = &dateStr
			}
		}
	}

	// If no reservation operations found, return not found
	if len(reservationOps) == 0 {
		response := &models.ReservationStatusResponse{
			ReservationExists: false,
			Error:             stringPtr("Reservation not found"),
		}
		return response, nil
	}

	// Determine reservation status by checking for consumption or return operations
	status := "active" // Default status

	// Check for consumption operations (factory_consumption)
	for _, op := range operations {
		operationTypeCode, err := is.getOperationTypeCode(ctx, op.OperationTypeID)
		if err != nil {
			continue
		}
		if operationTypeCode == models.OperationTypeFactoryConsumption {
			status = "consumed"
			break
		}
	}

	// Check for return operations (factory_return) - only if not consumed
	if status == "active" {
		for _, op := range operations {
			operationTypeCode, err := is.getOperationTypeCode(ctx, op.OperationTypeID)
			if err != nil {
				continue
			}
			if operationTypeCode == models.OperationTypeFactoryReturn {
				status = "returned"
				break
			}
		}
	}

	// Group reservation operations by item to calculate totals
	itemQuantities := make(map[string]int64)
	itemDetails := make(map[string]models.ReservationItemResponse)

	for _, op := range reservationOps {
		// Only count positive quantity changes (items going into factory inventory)
		if op.QuantityChange <= 0 {
			continue
		}

		// Get section code to verify this is moving items TO factory
		sectionCode, err := is.getSectionCode(ctx, op.SectionID)
		if err != nil || sectionCode != "factory" {
			continue // Skip if not factory section
		}

		// Get item details for response
		item, err := is.deps.Repositories.Item.GetItemWithDetails(ctx, op.ItemID)
		if err != nil {
			continue // Skip items we can't get details for
		}

		// Get collection and quality level codes
		collectionCode, err := is.getCodeFromUUID(ctx, models.ClassifierCollection, op.CollectionID)
		if err != nil {
			collectionCode = "basic" // Default fallback
		}

		qualityLevelCode, err := is.getCodeFromUUID(ctx, models.ClassifierQualityLevel, op.QualityLevelID)
		if err != nil {
			qualityLevelCode = "common" // Default fallback
		}

		// Create unique key for this item combination
		itemKey := op.ItemID.String() + "_" + collectionCode + "_" + qualityLevelCode

		// Add to quantities
		itemQuantities[itemKey] += op.QuantityChange

		// Store item details (will be overwritten for same items, but that's fine)
		itemDetails[itemKey] = models.ReservationItemResponse{
			ItemCode:         item.ItemType,
			CollectionCode:   collectionCode,
			QualityLevelCode: qualityLevelCode,
			Quantity:         0, // Will be filled below
		}
	}

	// Build reserved items list
	var reservedItems []models.ReservationItemResponse
	for itemKey, quantity := range itemQuantities {
		if detail, exists := itemDetails[itemKey]; exists {
			detail.Quantity = quantity
			reservedItems = append(reservedItems, detail)
		}
	}

	// Create successful response
	response := &models.ReservationStatusResponse{
		ReservationExists: true,
		OperationID:       &operationID,
		UserID:            &userID,
		ReservedItems:     reservedItems,
		ReservationDate:   reservationDate,
		Status:            &status,
	}

	// Record success metrics
	if is.deps.Metrics != nil {
		is.deps.Metrics.RecordInventoryOperation("get_reservation_status", "factory", "success")
	}

	return response, nil
}

// Helper functions

// getOperationTypeCode gets operation type code from UUID
func (is *inventoryService) getOperationTypeCode(ctx context.Context, operationTypeID uuid.UUID) (string, error) {
	mapping, err := is.deps.Repositories.Classifier.GetUUIDToCodeMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return "", errors.Wrap(err, "failed to get operation type mapping")
	}

	if code, found := mapping[operationTypeID]; found {
		return code, nil
	}

	return "", errors.Errorf("unknown operation type UUID: %s", operationTypeID.String())
}

// getSectionCode gets section code from UUID
func (is *inventoryService) getSectionCode(ctx context.Context, sectionID uuid.UUID) (string, error) {
	mapping, err := is.deps.Repositories.Classifier.GetUUIDToCodeMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return "", errors.Wrap(err, "failed to get section mapping")
	}

	if code, found := mapping[sectionID]; found {
		return code, nil
	}

	return "", errors.Errorf("unknown section UUID: %s", sectionID.String())
}

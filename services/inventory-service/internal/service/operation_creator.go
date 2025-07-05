package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	internalerrors "github.com/shard-legends/inventory-service/internal/errors"
	"github.com/shard-legends/inventory-service/internal/models"
)

// operationCreator implements OperationCreator interface
type operationCreator struct {
	deps *ServiceDependencies
}

// NewOperationCreator creates a new operation creator
func NewOperationCreator(deps *ServiceDependencies) OperationCreator {
	return &operationCreator{
		deps: deps,
	}
}

// CreateOperationsInTransaction creates multiple operations within a database transaction
func (oc *operationCreator) CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error) {
	start := time.Now()

	if len(operations) == 0 {
		return nil, errors.New("operations list cannot be empty")
	}

	// Validate all operations first
	for i, op := range operations {
		if err := oc.validateOperation(op); err != nil {
			return nil, errors.Wrapf(err, "operation at index %d is invalid", i)
		}
	}

	// Begin transaction
	tx, err := oc.deps.Repositories.Inventory.BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}

	// Ensure transaction is rolled back on any error
	defer func() {
		if err != nil {
			_ = oc.deps.Repositories.Inventory.RollbackTransaction(tx)
		}
	}()

	// Prepare operations for insertion
	var operationIDs []uuid.UUID
	now := time.Now().UTC()

	for _, op := range operations {
		// Set ID if not already set
		if op.ID == uuid.Nil {
			op.ID = uuid.New()
		}

		// Set creation time
		op.CreatedAt = now

		operationIDs = append(operationIDs, op.ID)
	}

	// Create operations in transaction
	err = oc.deps.Repositories.Inventory.CreateOperationsInTransaction(ctx, tx, operations)
	if err != nil {
		// Handle database errors appropriately
		if dbErr := internalerrors.HandleDatabaseError(err, "create_operations_in_transaction"); dbErr != nil {
			return nil, dbErr
		}
		return nil, errors.Wrap(err, "failed to create operations in transaction")
	}

	// Commit transaction
	err = oc.deps.Repositories.Inventory.CommitTransaction(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Invalidate cache for affected users after successful transaction
	err = oc.invalidateCacheForOperations(ctx, operations)
	if err != nil {
		// Log the error but don't fail the operation since data is already committed
		// This is critical for cache consistency!
		fmt.Printf("CRITICAL: Failed to invalidate cache after operations: %v\n", err)
	}

	// Record metrics
	if oc.deps.Metrics != nil {
		// Determine operation type from first operation
		operationType := "mixed"
		if len(operations) > 0 && operations[0].OperationTypeID != uuid.Nil {
			operationType = operations[0].OperationTypeID.String()
		}
		oc.deps.Metrics.RecordTransactionMetrics(operationType, len(operations), time.Since(start))
	}

	return operationIDs, nil
}

// CreateOperationBatch creates a batch of operations with the same external operation ID
func (oc *operationCreator) CreateOperationBatch(ctx context.Context, batch *OperationBatch) ([]uuid.UUID, error) {
	if batch == nil {
		return nil, errors.New("operation batch cannot be nil")
	}

	if len(batch.Operations) == 0 {
		return nil, errors.New("operation batch cannot be empty")
	}

	// Set the external operation ID for all operations
	for _, op := range batch.Operations {
		op.OperationID = &batch.ExternalID
		op.UserID = batch.UserID
	}

	return oc.CreateOperationsInTransaction(ctx, batch.Operations)
}

// validateOperation validates a single operation before creation
func (oc *operationCreator) validateOperation(op *models.Operation) error {
	if op == nil {
		return errors.New("operation cannot be nil")
	}

	// Use the model's validation
	if err := models.ValidateOperation(op); err != nil {
		return err
	}

	// Additional business logic validation
	if op.QuantityChange == 0 {
		return errors.New("quantity change cannot be zero")
	}

	return nil
}

// invalidateCacheForOperations invalidates cache for all users affected by the operations
func (oc *operationCreator) invalidateCacheForOperations(ctx context.Context, operations []*models.Operation) error {
	// Collect unique user IDs
	userIDs := make(map[uuid.UUID]bool)
	for _, op := range operations {
		userIDs[op.UserID] = true
	}

	// Invalidate cache for each user
	cacheManager := NewCacheManager(oc.deps)
	for userID := range userIDs {
		if err := cacheManager.InvalidateUserCache(ctx, userID); err != nil {
			return errors.Wrapf(err, "failed to invalidate cache for user %s", userID.String())
		}
	}

	return nil
}

// CreateReservationOperations creates operations for item reservation
// According to spec: creates 2 operations per item (debit from main, credit to factory)
func (oc *operationCreator) CreateReservationOperations(ctx context.Context, req *models.ReserveItemsRequest) ([]uuid.UUID, error) {
	if req == nil {
		return nil, errors.New("reserve items request cannot be nil")
	}

	if err := models.ValidateReserveItemsRequest(req); err != nil {
		return nil, errors.Wrap(err, "reserve items request validation failed")
	}

	// Check if operation_id is already used (prevent duplicate reservations)
	existingOps, err := oc.deps.Repositories.Inventory.GetOperationsByExternalID(ctx, req.OperationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check existing operations")
	}
	if len(existingOps) > 0 {
		return nil, errors.Errorf("operation_id %s is already used", req.OperationID.String())
	}

	// Convert codes to UUIDs for each item
	for i := range req.Items {
		// For now, we'll handle code conversion manually since the current interface is different
		// TODO: Create a proper method for converting ItemQuantityRequest codes
		if req.Items[i].Collection != nil && *req.Items[i].Collection != "" {
			// Convert collection code to UUID
			mapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get collection mapping for item %d", i)
			}
			if collectionUUID, found := mapping[*req.Items[i].Collection]; found {
				req.Items[i].CollectionID = &collectionUUID
			} else {
				return nil, errors.Errorf("unknown collection code: %s for item %d", *req.Items[i].Collection, i)
			}
		}

		if req.Items[i].QualityLevel != nil && *req.Items[i].QualityLevel != "" {
			// Convert quality level code to UUID
			mapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get quality level mapping for item %d", i)
			}
			if qualityUUID, found := mapping[*req.Items[i].QualityLevel]; found {
				req.Items[i].QualityLevelID = &qualityUUID
			} else {
				return nil, errors.Errorf("unknown quality level code: %s for item %d", *req.Items[i].QualityLevel, i)
			}
		}
	}

	// Get section IDs
	sectionMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get section mapping")
	}

	mainSectionID, found := sectionMapping["main"]
	if !found {
		return nil, errors.New("main inventory section not found")
	}

	factorySectionID, found := sectionMapping["factory"]
	if !found {
		return nil, errors.New("factory inventory section not found")
	}

	// Begin transaction for atomic balance checking and reservation
	tx, err := oc.deps.Repositories.Inventory.BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin reservation transaction")
	}

	// Гарантируем завершение транзакции: откат, если Commit не выполнен
	committed := false
	defer func() {
		if !committed {
			_ = oc.deps.Repositories.Inventory.RollbackTransaction(tx)
		}
	}()

	// Convert to BalanceLockRequest format for atomic balance checking
	var lockRequests []BalanceLockRequest
	for _, item := range req.Items {
		collectionID, err := oc.getDefaultCollectionID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default collection ID")
		}
		if item.CollectionID != nil {
			collectionID = *item.CollectionID
		}

		qualityLevelID, err := oc.getDefaultQualityLevelID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default quality level ID")
		}
		if item.QualityLevelID != nil {
			qualityLevelID = *item.QualityLevelID
		}

		lockRequests = append(lockRequests, BalanceLockRequest{
			UserID:         req.UserID,
			SectionID:      mainSectionID,
			ItemID:         item.ItemID,
			CollectionID:   collectionID,
			QualityLevelID: qualityLevelID,
			RequiredQty:    item.Quantity,
		})
	}

	// Atomically check and lock balances using SELECT ... FOR UPDATE
	lockResults, err := oc.deps.Repositories.Inventory.CheckAndLockBalances(ctx, tx, lockRequests)
	if err != nil {
		// Handle database errors appropriately
		if dbErr := internalerrors.HandleDatabaseError(err, "check_and_lock_balances"); dbErr != nil {
			return nil, dbErr
		}
		return nil, errors.Wrap(err, "failed to check and lock balances")
	}

	// Check if all items have sufficient balance
	var insufficientItems []string
	for _, result := range lockResults {
		if result.Error != nil {
			return nil, errors.Wrapf(result.Error, "failed to check balance for item %s", result.ItemID.String())
		}

		if !result.Sufficient {
			insufficientItems = append(insufficientItems,
				fmt.Sprintf("item %s: required %d, available %d",
					result.ItemID.String(), result.RequiredQty, result.AvailableQty))
		}
	}

	if len(insufficientItems) > 0 {
		// Create structured insufficient items error
		missingItems := make([]models.MissingItem, 0, len(lockResults))
		for _, result := range lockResults {
			if !result.Sufficient {
				missingItems = append(missingItems, models.MissingItem{
					ItemID:    result.ItemID,
					Required:  result.RequiredQty,
					Available: result.AvailableQty,
				})
			}
		}

		return nil, &models.InsufficientItemsError{
			ErrorCode:    "insufficient_items",
			Message:      fmt.Sprintf("Insufficient items for reservation: %s", strings.Join(insufficientItems, "; ")),
			MissingItems: missingItems,
		}
	}

	// Get operation type ID for reservation
	operationMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operation type mapping")
	}

	reservationTypeID, found := operationMapping[models.OperationTypeFactoryReservation]
	if !found {
		return nil, errors.New("reservation operation type not found")
	}

	// Create paired operations for each item (2 operations per item)
	var operations []*models.Operation
	for _, item := range req.Items {
		// Get collection and quality level IDs, using defaults if not provided
		collectionID, err := oc.getDefaultCollectionID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default collection ID")
		}
		if item.CollectionID != nil {
			collectionID = *item.CollectionID
		}

		qualityLevelID, err := oc.getDefaultQualityLevelID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default quality level ID")
		}
		if item.QualityLevelID != nil {
			qualityLevelID = *item.QualityLevelID
		}

		// Operation 1: Debit from main inventory
		debitOp := &models.Operation{
			UserID:          req.UserID,
			SectionID:       mainSectionID,
			ItemID:          item.ItemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  -item.Quantity, // Negative for debit
			OperationTypeID: reservationTypeID,
			OperationID:     &req.OperationID,
			Comment:         stringPtr("Factory reservation - debit from main"),
		}

		// Operation 2: Credit to factory inventory
		creditOp := &models.Operation{
			UserID:          req.UserID,
			SectionID:       factorySectionID,
			ItemID:          item.ItemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  item.Quantity, // Positive for credit
			OperationTypeID: reservationTypeID,
			OperationID:     &req.OperationID,
			Comment:         stringPtr("Factory reservation - credit to factory"),
		}

		operations = append(operations, debitOp, creditOp)
	}

	// Create operations within the same transaction
	err = oc.deps.Repositories.Inventory.CreateOperationsInTransaction(ctx, tx, operations)
	if err != nil {
		// Handle database errors appropriately
		if dbErr := internalerrors.HandleDatabaseError(err, "create_reservation_operations"); dbErr != nil {
			return nil, dbErr
		}
		return nil, errors.Wrap(err, "failed to create reservation operations")
	}

	// Commit transaction
	err = oc.deps.Repositories.Inventory.CommitTransaction(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit reservation transaction")
	}

	// Помечаем успешный коммит, чтобы defer не делал Rollback
	committed = true

	// Invalidate cache for affected users after successful transaction
	err = oc.invalidateCacheForOperations(ctx, operations)
	if err != nil {
		// Log the error but don't fail the operation since data is already committed
		fmt.Printf("CRITICAL: Failed to invalidate cache after reservation: %v\n", err)
	}

	// Extract operation IDs for return
	var operationIDs []uuid.UUID
	for _, op := range operations {
		operationIDs = append(operationIDs, op.ID)
	}

	return operationIDs, nil
}

// CreateReturnOperations creates operations for returning reserved items
func (oc *operationCreator) CreateReturnOperations(ctx context.Context, req *models.ReturnReserveRequest) error {
	if req == nil {
		return errors.New("return reserve request cannot be nil")
	}

	// Find the original reservation operations
	reservedOps, err := oc.deps.Repositories.Inventory.GetOperationsByExternalID(ctx, req.OperationID)
	if err != nil {
		return errors.Wrap(err, "failed to get reserved operations")
	}

	if len(reservedOps) == 0 {
		return errors.New("no reservation found for the given operation ID")
	}

	// Get operation type ID for return and validation
	operationMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return errors.Wrap(err, "failed to get operation type mapping")
	}

	returnTypeID, found := operationMapping[models.OperationTypeFactoryReturn]
	if !found {
		return errors.New("return operation type not found")
	}

	consumeTypeID, consumeFound := operationMapping[models.OperationTypeFactoryConsumption]

	// Check if this reservation has already been returned or consumed
	for _, op := range reservedOps {
		if op.OperationTypeID == returnTypeID ||
			(consumeFound && op.OperationTypeID == consumeTypeID) {
			return errors.Errorf("operation_id %s has already been returned or consumed", req.OperationID.String())
		}
	}

	// Get section IDs for proper return operations
	sectionMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return errors.Wrap(err, "failed to get section mapping")
	}

	mainSectionID, found := sectionMapping["main"]
	if !found {
		return errors.New("main inventory section not found")
	}

	factorySectionID, found := sectionMapping["factory"]
	if !found {
		return errors.New("factory inventory section not found")
	}

	// Create paired return operations for each reserved item (2 operations per item)
	// According to spec: debit from factory, credit to main
	var returnOperations []*models.Operation
	for _, reservedOp := range reservedOps {
		// Only process factory credit operations from original reservation (skip main debits)
		if reservedOp.SectionID == factorySectionID && reservedOp.QuantityChange > 0 {
			// Operation 1: Debit from factory inventory
			factoryDebitOp := &models.Operation{
				UserID:          reservedOp.UserID,
				SectionID:       factorySectionID,
				ItemID:          reservedOp.ItemID,
				CollectionID:    reservedOp.CollectionID,
				QualityLevelID:  reservedOp.QualityLevelID,
				QuantityChange:  -reservedOp.QuantityChange, // Negative to debit
				OperationTypeID: returnTypeID,
				OperationID:     &req.OperationID,
				Comment:         stringPtr("Factory return - debit from factory"),
			}

			// Operation 2: Credit to main inventory
			mainCreditOp := &models.Operation{
				UserID:          reservedOp.UserID,
				SectionID:       mainSectionID,
				ItemID:          reservedOp.ItemID,
				CollectionID:    reservedOp.CollectionID,
				QualityLevelID:  reservedOp.QualityLevelID,
				QuantityChange:  reservedOp.QuantityChange, // Positive to credit
				OperationTypeID: returnTypeID,
				OperationID:     &req.OperationID,
				Comment:         stringPtr("Factory return - credit to main"),
			}

			returnOperations = append(returnOperations, factoryDebitOp, mainCreditOp)
		}
	}

	_, err = oc.CreateOperationsInTransaction(ctx, returnOperations)
	return err
}

// CreateConsumptionOperations creates operations for consuming reserved items
func (oc *operationCreator) CreateConsumptionOperations(ctx context.Context, req *models.ConsumeReserveRequest) error {
	if req == nil {
		return errors.New("consume reserve request cannot be nil")
	}

	// Find the original reservation operations
	reservedOps, err := oc.deps.Repositories.Inventory.GetOperationsByExternalID(ctx, req.OperationID)
	if err != nil {
		return errors.Wrap(err, "failed to get reserved operations")
	}

	if len(reservedOps) == 0 {
		return errors.New("no reservation found for the given operation ID")
	}

	// Get operation type ID for consumption
	operationMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return errors.Wrap(err, "failed to get operation type mapping")
	}

	consumeTypeID, found := operationMapping[models.OperationTypeFactoryConsumption]
	if !found {
		return errors.New("consumption operation type not found")
	}

	returnTypeID, returnFound := operationMapping[models.OperationTypeFactoryReturn]

	// Check if this reservation has already been returned or consumed
	for _, op := range reservedOps {
		if op.OperationTypeID == consumeTypeID ||
			(returnFound && op.OperationTypeID == returnTypeID) {
			return errors.Errorf("operation_id %s has already been returned or consumed", req.OperationID.String())
		}
	}

	// Get factory section ID
	sectionMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierInventorySection)
	if err != nil {
		return errors.Wrap(err, "failed to get section mapping")
	}

	factorySectionID, found := sectionMapping["factory"]
	if !found {
		return errors.New("factory inventory section not found")
	}

	// Create consumption operations (debit from factory inventory, destroying the items)
	// According to spec: only creates operations for factory section debits
	var consumeOperations []*models.Operation
	for _, reservedOp := range reservedOps {
		// Only process factory credit operations (skip main debit operations)
		if reservedOp.SectionID == factorySectionID && reservedOp.QuantityChange > 0 {
			consumeOp := &models.Operation{
				UserID:          reservedOp.UserID,
				SectionID:       factorySectionID,
				ItemID:          reservedOp.ItemID,
				CollectionID:    reservedOp.CollectionID,
				QualityLevelID:  reservedOp.QualityLevelID,
				QuantityChange:  -reservedOp.QuantityChange, // Negative to debit from factory
				OperationTypeID: consumeTypeID,
				OperationID:     &req.OperationID,
				Comment:         stringPtr("Factory consumption - items destroyed"),
			}

			consumeOperations = append(consumeOperations, consumeOp)
		}
	}

	_, err = oc.CreateOperationsInTransaction(ctx, consumeOperations)
	return err
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// getDefaultCollectionID returns the default collection UUID by looking up "base" code
func (oc *operationCreator) getDefaultCollectionID(ctx context.Context) (uuid.UUID, error) {
	collectionMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get collection mapping")
	}

	baseCollectionID, found := collectionMapping["base"]
	if !found {
		return uuid.Nil, errors.New("base collection not found in classifier mapping")
	}

	return baseCollectionID, nil
}

// getDefaultQualityLevelID returns the default quality level UUID by looking up "base" code
func (oc *operationCreator) getDefaultQualityLevelID(ctx context.Context) (uuid.UUID, error) {
	qualityMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to get quality level mapping")
	}

	baseQualityID, found := qualityMapping["base"]
	if !found {
		return uuid.Nil, errors.New("base quality level not found in classifier mapping")
	}

	return baseQualityID, nil
}

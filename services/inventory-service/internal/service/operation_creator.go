package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	
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
		// In a real application, you might want to use structured logging here
		_ = err
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
func (oc *operationCreator) CreateReservationOperations(ctx context.Context, req *models.ReserveItemsRequest) ([]uuid.UUID, error) {
	if req == nil {
		return nil, errors.New("reserve items request cannot be nil")
	}

	if err := models.ValidateReserveItemsRequest(req); err != nil {
		return nil, errors.Wrap(err, "reserve items request validation failed")
	}

	// TODO: Convert section name to UUID when implementing reservation operations
	// For now, assume section is provided as code and needs conversion
	// In practice, this would come from the request context or be validated earlier

	// Get operation type ID for reservation
	operationMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operation type mapping")
	}

	reservationTypeID, found := operationMapping[models.OperationTypeFactoryReservation]
	if !found {
		return nil, errors.New("reservation operation type not found")
	}

	// Create operations for each item (negative quantities for reservation)
	var operations []*models.Operation
	for _, item := range req.Items {
		// Convert item codes to UUIDs if necessary
		// This is simplified - in practice you'd handle code conversion properly

		op := &models.Operation{
			UserID:          req.UserID,
			SectionID:       uuid.New(), // Would be converted from section code
			ItemID:          item.ItemID,
			CollectionID:    uuid.New(), // Would be converted from collection code
			QualityLevelID:  uuid.New(), // Would be converted from quality level code
			QuantityChange:  -item.Quantity, // Negative for reservation
			OperationTypeID: reservationTypeID,
			OperationID:     &req.OperationID,
		}

		operations = append(operations, op)
	}

	return oc.CreateOperationsInTransaction(ctx, operations)
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

	// Get operation type ID for return
	operationMapping, err := oc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierOperationType)
	if err != nil {
		return errors.Wrap(err, "failed to get operation type mapping")
	}

	returnTypeID, found := operationMapping[models.OperationTypeFactoryReturn]
	if !found {
		return errors.New("return operation type not found")
	}

	// Create return operations (opposite of reservation)
	var returnOperations []*models.Operation
	for _, reservedOp := range reservedOps {
		returnOp := &models.Operation{
			UserID:          reservedOp.UserID,
			SectionID:       reservedOp.SectionID,
			ItemID:          reservedOp.ItemID,
			CollectionID:    reservedOp.CollectionID,
			QualityLevelID:  reservedOp.QualityLevelID,
			QuantityChange:  -reservedOp.QuantityChange, // Opposite of reservation
			OperationTypeID: returnTypeID,
			OperationID:     &req.OperationID,
		}

		returnOperations = append(returnOperations, returnOp)
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

	// Create consumption operations (no quantity change since items were already reserved)
	// This is more for audit trail
	var consumeOperations []*models.Operation
	for _, reservedOp := range reservedOps {
		consumeOp := &models.Operation{
			UserID:          reservedOp.UserID,
			SectionID:       reservedOp.SectionID,
			ItemID:          reservedOp.ItemID,
			CollectionID:    reservedOp.CollectionID,
			QualityLevelID:  reservedOp.QualityLevelID,
			QuantityChange:  0, // No additional change, items already reserved
			OperationTypeID: consumeTypeID,
			OperationID:     &req.OperationID,
			Comment:         stringPtr("Item consumption confirmed"),
		}

		consumeOperations = append(consumeOperations, consumeOp)
	}

	_, err = oc.CreateOperationsInTransaction(ctx, consumeOperations)
	return err
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/shard-legends/inventory-service/internal/models"
)

// inventoryRepo implements InventoryRepository
type inventoryRepo struct {
	db *sqlx.DB
}

// NewInventoryRepository creates a new InventoryRepository
func NewInventoryRepository(db *sqlx.DB) InventoryRepository {
	return &inventoryRepo{
		db: db,
	}
}

// GetDailyBalance retrieves a daily balance for a specific item
func (r *inventoryRepo) GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error) {
	// Normalize date to end of day in UTC
	endOfDay := date.UTC().Truncate(24*time.Hour).Add(24*time.Hour - time.Second)
	
	query := `
		SELECT user_id, section_id, item_id, collection_id, quality_level_id,
		       balance_date, quantity, created_at
		FROM daily_balance
		WHERE user_id = $1 AND section_id = $2 AND item_id = $3
		  AND collection_id = $4 AND quality_level_id = $5
		  AND balance_date = $6
	`
	
	var balance models.DailyBalance
	err := r.db.GetContext(ctx, &balance, query, userID, sectionID, itemID, collectionID, qualityLevelID, endOfDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get daily balance")
	}
	
	return &balance, nil
}

// GetLatestDailyBalance retrieves the most recent daily balance before or on the given date
func (r *inventoryRepo) GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error) {
	// Normalize date to end of day in UTC
	endOfDay := beforeDate.UTC().Truncate(24*time.Hour).Add(24*time.Hour - time.Second)
	
	query := `
		SELECT user_id, section_id, item_id, collection_id, quality_level_id,
		       balance_date, quantity, created_at
		FROM daily_balance
		WHERE user_id = $1 AND section_id = $2 AND item_id = $3
		  AND collection_id = $4 AND quality_level_id = $5
		  AND balance_date <= $6
		ORDER BY balance_date DESC
		LIMIT 1
	`
	
	var balance models.DailyBalance
	err := r.db.GetContext(ctx, &balance, query, userID, sectionID, itemID, collectionID, qualityLevelID, endOfDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get latest daily balance")
	}
	
	return &balance, nil
}

// CreateDailyBalance creates a new daily balance record
func (r *inventoryRepo) CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error {
	// Normalize balance date to end of day in UTC
	balance.BalanceDate = balance.BalanceDate.UTC().Truncate(24*time.Hour).Add(24*time.Hour - time.Second)
	balance.CreatedAt = time.Now().UTC()
	
	query := `
		INSERT INTO daily_balance (user_id, section_id, item_id, collection_id, quality_level_id,
		                          balance_date, quantity, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, section_id, item_id, collection_id, quality_level_id, balance_date)
		DO UPDATE SET quantity = $7, created_at = $8
	`
	
	_, err := r.db.ExecContext(ctx, query,
		balance.UserID, balance.SectionID, balance.ItemID, balance.CollectionID, balance.QualityLevelID,
		balance.BalanceDate, balance.Quantity, balance.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create daily balance")
	}
	
	return nil
}

// GetOperations retrieves operations for a specific item from a given date
func (r *inventoryRepo) GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error) {
	query := `
		SELECT id, user_id, section_id, item_id, collection_id, quality_level_id,
		       quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at
		FROM operation
		WHERE user_id = $1 AND section_id = $2 AND item_id = $3
		  AND collection_id = $4 AND quality_level_id = $5
		  AND created_at >= $6
		ORDER BY created_at
	`
	
	var operations []*models.Operation
	err := r.db.SelectContext(ctx, &operations, query,
		userID, sectionID, itemID, collectionID, qualityLevelID, fromDate,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operations")
	}
	
	return operations, nil
}

// GetOperationsByExternalID retrieves operations by their external operation ID
func (r *inventoryRepo) GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error) {
	query := `
		SELECT id, user_id, section_id, item_id, collection_id, quality_level_id,
		       quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at
		FROM operation
		WHERE operation_id = $1
		ORDER BY created_at
	`
	
	var operations []*models.Operation
	err := r.db.SelectContext(ctx, &operations, query, operationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operations by external ID")
	}
	
	return operations, nil
}

// CreateOperation creates a new operation
func (r *inventoryRepo) CreateOperation(ctx context.Context, operation *models.Operation) error {
	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}
	if operation.CreatedAt.IsZero() {
		operation.CreatedAt = time.Now().UTC()
	}
	
	query := `
		INSERT INTO operation (id, user_id, section_id, item_id, collection_id, quality_level_id,
		                      quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		operation.ID, operation.UserID, operation.SectionID, operation.ItemID,
		operation.CollectionID, operation.QualityLevelID, operation.QuantityChange,
		operation.OperationTypeID, operation.OperationID, operation.RecipeID,
		operation.Comment, operation.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create operation")
	}
	
	return nil
}

// CreateOperations creates multiple operations in a batch
func (r *inventoryRepo) CreateOperations(ctx context.Context, operations []*models.Operation) error {
	if len(operations) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()
	
	err = r.CreateOperationsInTransaction(ctx, tx, operations)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}

// CreateOperationsInTransaction creates operations within a transaction
func (r *inventoryRepo) CreateOperationsInTransaction(ctx context.Context, tx *sqlx.Tx, operations []*models.Operation) error {
	if len(operations) == 0 {
		return nil
	}
	
	// Prepare batch insert
	valueStrings := make([]string, 0, len(operations))
	valueArgs := make([]interface{}, 0, len(operations)*12)
	
	for i, op := range operations {
		if op.ID == uuid.Nil {
			op.ID = uuid.New()
		}
		if op.CreatedAt.IsZero() {
			op.CreatedAt = time.Now().UTC()
		}
		
		// Build placeholders for this operation
		placeholders := make([]string, 12)
		for j := 0; j < 12; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*12+j+1)
		}
		valueStrings = append(valueStrings, "("+strings.Join(placeholders, ", ")+")")
		
		// Add values
		valueArgs = append(valueArgs,
			op.ID, op.UserID, op.SectionID, op.ItemID,
			op.CollectionID, op.QualityLevelID, op.QuantityChange,
			op.OperationTypeID, op.OperationID, op.RecipeID,
			op.Comment, op.CreatedAt,
		)
	}
	
	query := `
		INSERT INTO operation (id, user_id, section_id, item_id, collection_id, quality_level_id,
		                      quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at)
		VALUES ` + strings.Join(valueStrings, ", ")
	
	_, err := tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return errors.Wrap(err, "failed to create operations in batch")
	}
	
	return nil
}

// GetUserInventoryItems retrieves all unique item combinations for a user in a section
func (r *inventoryRepo) GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error) {
	query := `
		SELECT DISTINCT user_id, section_id, item_id, collection_id, quality_level_id
		FROM (
			-- Items from daily balances
			SELECT user_id, section_id, item_id, collection_id, quality_level_id
			FROM daily_balance
			WHERE user_id = $1 AND section_id = $2 AND quantity > 0
			
			UNION
			
			-- Items from recent operations (last 30 days)
			SELECT user_id, section_id, item_id, collection_id, quality_level_id
			FROM operation
			WHERE user_id = $3 AND section_id = $4
			  AND created_at >= NOW() - INTERVAL '30 days'
		) AS items
		ORDER BY item_id, collection_id, quality_level_id
	`
	
	rows, err := r.db.QueryxContext(ctx, query, userID, sectionID, userID, sectionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user inventory items")
	}
	defer rows.Close()
	
	var items []*models.ItemKey
	for rows.Next() {
		var key models.ItemKey
		err := rows.Scan(&key.UserID, &key.SectionID, &key.ItemID, &key.CollectionID, &key.QualityLevelID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item key")
		}
		items = append(items, &key)
	}
	
	return items, nil
}

// BeginTransaction starts a new database transaction
func (r *inventoryRepo) BeginTransaction(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	return tx, nil
}

// CommitTransaction commits a transaction
func (r *inventoryRepo) CommitTransaction(tx *sqlx.Tx) error {
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

// RollbackTransaction rolls back a transaction
func (r *inventoryRepo) RollbackTransaction(tx *sqlx.Tx) error {
	if err := tx.Rollback(); err != nil {
		return errors.Wrap(err, "failed to rollback transaction")
	}
	return nil
}
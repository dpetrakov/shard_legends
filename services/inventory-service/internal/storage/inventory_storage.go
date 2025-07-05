package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/inventory-service/internal/database"
	internalerrors "github.com/shard-legends/inventory-service/internal/errors"
	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/internal/service"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// InventoryStorage implements inventory data access using PostgreSQL and Redis
type InventoryStorage struct {
	pool    *pgxpool.Pool
	redis   *database.RedisDB
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	// Inventory operations
	GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error)
	GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error)
	GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error)
	CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error

	// Operations
	GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error)
	GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error)
	CreateOperations(ctx context.Context, operations []*models.Operation) error
	CreateOperationsInTransaction(ctx context.Context, tx interface{}, operations []*models.Operation) error

	// Transactions (matching service interface)
	BeginTransaction(ctx context.Context) (interface{}, error)
	CommitTransaction(tx interface{}) error
	RollbackTransaction(tx interface{}) error

	// Atomic balance checking with row-level locking
	CheckAndLockBalances(ctx context.Context, tx interface{}, items []service.BalanceLockRequest) ([]service.BalanceLockResult, error)

	// D-15: Optimized methods to eliminate N+1 queries
	GetUserInventoryOptimized(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error)
}

// ClassifierRepository defines the interface for classifier data access
type ClassifierRepository interface {
	GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error)
	GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
	GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
	InvalidateCache(ctx context.Context, classifierCode string) error
}

// ItemRepository defines the interface for item data access
type ItemRepository interface {
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
	GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error)
	GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error)

	// I18n and batch operations
	GetItemsBatch(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID]*models.ItemWithDetails, error)
	GetTranslationsBatch(ctx context.Context, entityType string, entityIDs []uuid.UUID, languageCode string) (map[uuid.UUID]map[string]string, error)
	GetDefaultLanguage(ctx context.Context) (*models.Language, error)
	GetItemImagesBatch(ctx context.Context, requests []models.ItemDetailRequestItem) (map[string]string, error)
}

// NewInventoryStorage creates a new inventory storage instance
func NewInventoryStorage(pool *pgxpool.Pool, redis *database.RedisDB, logger *slog.Logger, metrics *metrics.Metrics) InventoryRepository {
	return &InventoryStorage{
		pool:    pool,
		redis:   redis,
		logger:  logger,
		metrics: metrics,
	}
}

// NewClassifierStorage creates a new classifier storage instance
func NewClassifierStorage(pool *pgxpool.Pool, redis *database.RedisDB, logger *slog.Logger, metrics *metrics.Metrics) ClassifierRepository {
	return &classifierStorage{
		pool:    pool,
		redis:   redis,
		logger:  logger,
		metrics: metrics,
	}
}

// NewItemStorage creates a new item storage instance
func NewItemStorage(pool *pgxpool.Pool, logger *slog.Logger, metrics service.MetricsInterface, cache service.CacheInterface) ItemRepository {
	return &itemStorage{
		pool:    pool,
		logger:  logger,
		metrics: metrics,
		cache:   cache,
	}
}

// Inventory storage implementation

func (s *InventoryStorage) GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error) {
	s.logger.Info("GetUserInventoryItems called", "user_id", userID, "section_id", sectionID)

	query := `
		SELECT DISTINCT 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id
		FROM inventory.daily_balances 
		WHERE user_id = $1 AND section_id = $2
		UNION
		SELECT DISTINCT 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id
		FROM inventory.operations 
		WHERE user_id = $1 AND section_id = $2 AND created_at >= CURRENT_DATE
	`

	rows, err := s.pool.Query(ctx, query, userID, sectionID)
	if err != nil {
		s.logger.Error("Failed to get user inventory items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []*models.ItemKey
	for rows.Next() {
		var item models.ItemKey
		if err := rows.Scan(
			&item.UserID,
			&item.SectionID,
			&item.ItemID,
			&item.CollectionID,
			&item.QualityLevelID,
		); err != nil {
			s.logger.Error("Failed to scan item key", "error", err)
			return nil, err
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}

	s.logger.Info("Found inventory items", "count", len(items))
	return items, nil
}

func (s *InventoryStorage) GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error) {
	s.logger.Info("GetDailyBalance called", "user_id", userID, "date", date.Format("2006-01-02"))

	query := `
		SELECT 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			balance_date, 
			quantity, 
			created_at
		FROM inventory.daily_balances 
		WHERE user_id = $1 
			AND section_id = $2 
			AND item_id = $3 
			AND collection_id = $4 
			AND quality_level_id = $5 
			AND balance_date = $6
	`

	var balance models.DailyBalance
	err := s.pool.QueryRow(ctx, query, userID, sectionID, itemID, collectionID, qualityLevelID, date.Truncate(24*time.Hour)).Scan(
		&balance.UserID,
		&balance.SectionID,
		&balance.ItemID,
		&balance.CollectionID,
		&balance.QualityLevelID,
		&balance.BalanceDate,
		&balance.Quantity,
		&balance.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("No daily balance found", "date", date.Format("2006-01-02"))
			return nil, nil
		}
		s.logger.Error("Failed to get daily balance", "error", err)
		return nil, err
	}

	return &balance, nil
}

func (s *InventoryStorage) GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error) {
	s.logger.Info("GetLatestDailyBalance called", "user_id", userID, "before_date", beforeDate.Format("2006-01-02"))

	query := `
		SELECT 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			balance_date, 
			quantity, 
			created_at
		FROM inventory.daily_balances 
		WHERE user_id = $1 
			AND section_id = $2 
			AND item_id = $3 
			AND collection_id = $4 
			AND quality_level_id = $5 
			AND balance_date < $6
		ORDER BY balance_date DESC
		LIMIT 1
	`

	var balance models.DailyBalance
	err := s.pool.QueryRow(ctx, query, userID, sectionID, itemID, collectionID, qualityLevelID, beforeDate.Truncate(24*time.Hour)).Scan(
		&balance.UserID,
		&balance.SectionID,
		&balance.ItemID,
		&balance.CollectionID,
		&balance.QualityLevelID,
		&balance.BalanceDate,
		&balance.Quantity,
		&balance.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("No latest daily balance found", "before_date", beforeDate.Format("2006-01-02"))
			return nil, nil
		}
		s.logger.Error("Failed to get latest daily balance", "error", err)
		return nil, err
	}

	return &balance, nil
}

func (s *InventoryStorage) CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error {
	s.logger.Info("CreateDailyBalance called", "user_id", balance.UserID, "quantity", balance.Quantity)

	query := `
		INSERT INTO inventory.daily_balances (
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			balance_date, 
			quantity
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, section_id, item_id, collection_id, quality_level_id, balance_date)
		DO UPDATE SET quantity = EXCLUDED.quantity
	`

	_, err := s.pool.Exec(ctx, query,
		balance.UserID,
		balance.SectionID,
		balance.ItemID,
		balance.CollectionID,
		balance.QualityLevelID,
		balance.BalanceDate.Truncate(24*time.Hour),
		balance.Quantity,
	)

	if err != nil {
		s.logger.Error("Failed to create daily balance", "error", err)
		return err
	}

	s.logger.Debug("Daily balance created successfully")
	return nil
}

func (s *InventoryStorage) GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error) {
	s.logger.Info("GetOperations called", "user_id", userID, "from_date", fromDate.Format("2006-01-02 15:04:05"))

	query := `
		SELECT 
			id, 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			quantity_change, 
			operation_type_id, 
			operation_id, 
			recipe_id, 
			comment, 
			created_at
		FROM inventory.operations 
		WHERE user_id = $1 
			AND section_id = $2 
			AND item_id = $3 
			AND collection_id = $4 
			AND quality_level_id = $5 
			AND created_at >= $6
		ORDER BY created_at ASC
	`

	rows, err := s.pool.Query(ctx, query, userID, sectionID, itemID, collectionID, qualityLevelID, fromDate)
	if err != nil {
		s.logger.Error("Failed to get operations", "error", err)
		return nil, err
	}
	defer rows.Close()

	var operations []*models.Operation
	for rows.Next() {
		var op models.Operation
		if err := rows.Scan(
			&op.ID,
			&op.UserID,
			&op.SectionID,
			&op.ItemID,
			&op.CollectionID,
			&op.QualityLevelID,
			&op.QuantityChange,
			&op.OperationTypeID,
			&op.OperationID,
			&op.RecipeID,
			&op.Comment,
			&op.CreatedAt,
		); err != nil {
			s.logger.Error("Failed to scan operation", "error", err)
			return nil, err
		}
		operations = append(operations, &op)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}

	s.logger.Info("Found operations", "count", len(operations))
	return operations, nil
}

func (s *InventoryStorage) GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error) {
	s.logger.Info("GetOperationsByExternalID called", "operation_id", operationID)

	query := `
		SELECT 
			id, 
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			quantity_change, 
			operation_type_id, 
			operation_id, 
			recipe_id, 
			comment, 
			created_at
		FROM inventory.operations 
		WHERE operation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := s.pool.Query(ctx, query, operationID)
	if err != nil {
		s.logger.Error("Failed to get operations by external ID", "error", err)
		return nil, err
	}
	defer rows.Close()

	var operations []*models.Operation
	for rows.Next() {
		var op models.Operation
		if err := rows.Scan(
			&op.ID,
			&op.UserID,
			&op.SectionID,
			&op.ItemID,
			&op.CollectionID,
			&op.QualityLevelID,
			&op.QuantityChange,
			&op.OperationTypeID,
			&op.OperationID,
			&op.RecipeID,
			&op.Comment,
			&op.CreatedAt,
		); err != nil {
			s.logger.Error("Failed to scan operation", "error", err)
			return nil, err
		}
		operations = append(operations, &op)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}

	s.logger.Info("Found operations by external ID", "count", len(operations))
	return operations, nil
}

func (s *InventoryStorage) CreateOperations(ctx context.Context, operations []*models.Operation) error {
	s.logger.Info("CreateOperations called", "count", len(operations))

	if len(operations) == 0 {
		return nil
	}

	// Use batch insert for better performance
	query := `
		INSERT INTO inventory.operations (
			id,
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			quantity_change, 
			operation_type_id, 
			operation_id, 
			recipe_id, 
			comment
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	batch := &pgx.Batch{}
	for _, op := range operations {
		// Generate ID if not provided
		if op.ID == uuid.Nil {
			op.ID = uuid.New()
		}

		batch.Queue(query,
			op.ID,
			op.UserID,
			op.SectionID,
			op.ItemID,
			op.CollectionID,
			op.QualityLevelID,
			op.QuantityChange,
			op.OperationTypeID,
			op.OperationID,
			op.RecipeID,
			op.Comment,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(operations); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("Failed to create operation", "index", i, "error", err)
			// Handle database errors appropriately
			if dbErr := internalerrors.HandleDatabaseError(err, "create_operation"); dbErr != nil {
				return dbErr
			}
			return err
		}
	}

	s.logger.Info("Operations created successfully", "count", len(operations))
	return nil
}

func (s *InventoryStorage) BeginTransaction(ctx context.Context) (interface{}, error) {
	return s.pool.Begin(ctx)
}

func (s *InventoryStorage) CommitTransaction(tx interface{}) error {
	if pgxTx, ok := tx.(pgx.Tx); ok {
		return pgxTx.Commit(context.Background())
	}
	return nil
}

func (s *InventoryStorage) RollbackTransaction(tx interface{}) error {
	if pgxTx, ok := tx.(pgx.Tx); ok {
		return pgxTx.Rollback(context.Background())
	}
	return nil
}

// CheckAndLockBalances atomically checks and locks balances for multiple items
// Uses SELECT ... FOR UPDATE to prevent race conditions
func (s *InventoryStorage) CheckAndLockBalances(ctx context.Context, tx interface{}, items []service.BalanceLockRequest) ([]service.BalanceLockResult, error) {
	s.logger.Info("CheckAndLockBalances called", "items_count", len(items))

	if len(items) == 0 {
		return nil, nil
	}

	// Cast to pgx.Tx
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type")
	}

	// Group items by unique combination to avoid locking the same row multiple times
	type itemKey struct {
		UserID         uuid.UUID
		SectionID      uuid.UUID
		ItemID         uuid.UUID
		CollectionID   uuid.UUID
		QualityLevelID uuid.UUID
	}

	groupedItems := make(map[itemKey][]int) // map[itemKey][]originalIndex
	for i, item := range items {
		key := itemKey{
			UserID:         item.UserID,
			SectionID:      item.SectionID,
			ItemID:         item.ItemID,
			CollectionID:   item.CollectionID,
			QualityLevelID: item.QualityLevelID,
		}
		groupedItems[key] = append(groupedItems[key], i)
	}

	s.logger.Info("Grouped items for balance checking",
		"original_count", len(items),
		"unique_groups", len(groupedItems))

	results := make([]service.BalanceLockResult, len(items))

	// Check and lock balance for each unique item group
	for key, indices := range groupedItems {
		// Calculate total required quantity for this group
		totalRequired := int64(0)
		for _, idx := range indices {
			totalRequired += items[idx].RequiredQty
		}

		// Lock the daily balance row for this item (if exists)
		// This prevents concurrent modifications to the same item
		lockQuery := `
			SELECT quantity 
			FROM inventory.daily_balances 
			WHERE user_id = $1 
			  AND section_id = $2 
			  AND item_id = $3 
			  AND collection_id = $4 
			  AND quality_level_id = $5 
			  AND balance_date <= CURRENT_DATE
			ORDER BY balance_date DESC 
			LIMIT 1
			FOR UPDATE NOWAIT
		`

		var dailyBalance int64 = 0
		err := pgxTx.QueryRow(ctx, lockQuery,
			key.UserID, key.SectionID, key.ItemID,
			key.CollectionID, key.QualityLevelID).Scan(&dailyBalance)

		if err != nil && err != pgx.ErrNoRows {
			// Handle database errors appropriately
			for _, idx := range indices {
				result := service.BalanceLockResult{
					BalanceLockRequest: items[idx],
				}
				if dbErr := internalerrors.HandleDatabaseError(err, "lock_balance_row"); dbErr != nil {
					result.Error = dbErr
				} else {
					result.Error = err
				}
				results[idx] = result
			}
			continue
		}

		// Calculate operations sum since the last daily_balance date
		// This matches the logic used in GetUserInventoryOptimized
		operationsQuery := `
			SELECT COALESCE(SUM(quantity_change), 0)
			FROM inventory.operations 
			WHERE user_id = $1 
			  AND section_id = $2 
			  AND item_id = $3 
			  AND collection_id = $4 
			  AND quality_level_id = $5 
			  AND created_at > (
				SELECT COALESCE(MAX(db.balance_date) + INTERVAL '1 day', '1970-01-01'::timestamp)
				FROM inventory.daily_balances db
				WHERE db.user_id = $1 
				  AND db.section_id = $2
				  AND db.item_id = $3
				  AND db.collection_id = $4
				  AND db.quality_level_id = $5
			  )
		`

		var operationsSum int64 = 0
		s.logger.Info("Executing operations query with parameters",
			"user_id", key.UserID,
			"section_id", key.SectionID,
			"item_id", key.ItemID,
			"collection_id", key.CollectionID,
			"quality_level_id", key.QualityLevelID,
			"total_required", totalRequired)

		err = pgxTx.QueryRow(ctx, operationsQuery,
			key.UserID, key.SectionID, key.ItemID,
			key.CollectionID, key.QualityLevelID).Scan(&operationsSum)

		if err != nil {
			for _, idx := range indices {
				result := service.BalanceLockResult{
					BalanceLockRequest: items[idx],
					Error:              err,
				}
				results[idx] = result
			}
			continue
		}

		// Calculate available balance
		availableBalance := dailyBalance + operationsSum
		groupSufficient := availableBalance >= totalRequired

		s.logger.Info("Balance locked and checked for group",
			"user_id", key.UserID,
			"item_id", key.ItemID,
			"daily_balance", dailyBalance,
			"operations_sum", operationsSum,
			"available", availableBalance,
			"total_required", totalRequired,
			"group_sufficient", groupSufficient,
			"group_size", len(indices))

		// Set results for all items in this group
		for _, idx := range indices {
			result := service.BalanceLockResult{
				BalanceLockRequest: items[idx],
				AvailableQty:       availableBalance,
				Sufficient:         groupSufficient && availableBalance >= items[idx].RequiredQty,
			}
			results[idx] = result
		}
	}

	return results, nil
}

func (s *InventoryStorage) CreateOperationsInTransaction(ctx context.Context, tx interface{}, operations []*models.Operation) error {
	s.logger.Info("CreateOperationsInTransaction called", "count", len(operations))

	if len(operations) == 0 {
		return nil
	}

	// If no transaction provided, create operations normally
	if tx == nil {
		return s.CreateOperations(ctx, operations)
	}

	// Cast to pgx.Tx
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		s.logger.Error("Invalid transaction type")
		return fmt.Errorf("invalid transaction type")
	}

	// Use batch insert for better performance within transaction
	query := `
		INSERT INTO inventory.operations (
			id,
			user_id, 
			section_id, 
			item_id, 
			collection_id, 
			quality_level_id, 
			quantity_change, 
			operation_type_id, 
			operation_id, 
			recipe_id, 
			comment
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	batch := &pgx.Batch{}
	for _, op := range operations {
		// Generate ID if not provided
		if op.ID == uuid.Nil {
			op.ID = uuid.New()
		}

		batch.Queue(query,
			op.ID,
			op.UserID,
			op.SectionID,
			op.ItemID,
			op.CollectionID,
			op.QualityLevelID,
			op.QuantityChange,
			op.OperationTypeID,
			op.OperationID,
			op.RecipeID,
			op.Comment,
		)
	}

	br := pgxTx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(operations); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("Failed to create operation in transaction", "error", err, "operation_index", i)
			// Handle database errors appropriately
			if dbErr := internalerrors.HandleDatabaseError(err, "create_operation_in_transaction"); dbErr != nil {
				return dbErr
			}
			return err
		}
	}

	s.logger.Info("Operations created successfully in transaction", "count", len(operations))
	return nil
}

// GetUserInventoryOptimized - D-15: оптимизированный метод для устранения N+1 запросов
// Использует единый JOIN запрос вместо множественных отдельных запросов
func (s *InventoryStorage) GetUserInventoryOptimized(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error) {
	s.logger.Info("GetUserInventoryOptimized called", "user_id", userID, "section_id", sectionID)

	// Единый запрос с JOIN для получения всех данных за раз
	// Устраняет N+1 проблему путем объединения:
	// 1. daily_balances и operations для расчета текущего баланса
	// 2. items, classifier_items для получения деталей предмета и кодов
	query := `
		WITH daily_balances_latest AS (
			-- Получаем актуальные остатки на конец дня
			SELECT DISTINCT ON (db.user_id, db.section_id, db.item_id, db.collection_id, db.quality_level_id)
				db.user_id,
				db.section_id, 
				db.item_id,
				db.collection_id,
				db.quality_level_id,
				db.quantity as daily_quantity,
				db.balance_date
			FROM inventory.daily_balances db
			WHERE db.user_id = $1 AND db.section_id = $2
			ORDER BY db.user_id, db.section_id, db.item_id, db.collection_id, db.quality_level_id, db.balance_date DESC
		),
		operations_only AS (
			-- Добавляем предметы, которые есть только в операциях (без daily_balance)
			SELECT DISTINCT 
				op.user_id,
				op.section_id,
				op.item_id,
				op.collection_id,
				op.quality_level_id,
				0 as daily_quantity,
				CURRENT_DATE as balance_date
			FROM inventory.operations op
			WHERE op.user_id = $1 AND op.section_id = $2 
				AND NOT EXISTS (
					SELECT 1 FROM inventory.daily_balances db2 
					WHERE db2.user_id = op.user_id 
						AND db2.section_id = op.section_id
						AND db2.item_id = op.item_id
						AND db2.collection_id = op.collection_id
						AND db2.quality_level_id = op.quality_level_id
				)
		),
		current_balances AS (
			SELECT * FROM daily_balances_latest
			UNION ALL
			SELECT * FROM operations_only
		),
		operations_sum AS (
			-- Сумма операций с даты последнего daily_balance
			SELECT 
				op.user_id,
				op.section_id,
				op.item_id,
				op.collection_id,
				op.quality_level_id,
				COALESCE(SUM(op.quantity_change), 0) as today_operations
			FROM inventory.operations op
			WHERE op.user_id = $1 AND op.section_id = $2 
				AND op.created_at > (
					SELECT COALESCE(MAX(db.balance_date) + INTERVAL '1 day', '1970-01-01'::timestamp)
					FROM inventory.daily_balances db
					WHERE db.user_id = op.user_id 
						AND db.section_id = op.section_id
						AND db.item_id = op.item_id
						AND db.collection_id = op.collection_id
						AND db.quality_level_id = op.quality_level_id
				)
			GROUP BY op.user_id, op.section_id, op.item_id, op.collection_id, op.quality_level_id
		),
		final_balances AS (
			-- Итоговые остатки = дневной остаток + операции за сегодня
			SELECT 
				cb.user_id,
				cb.section_id,
				cb.item_id,
				cb.collection_id,
				cb.quality_level_id,
				cb.daily_quantity + COALESCE(os.today_operations, 0) as final_quantity
			FROM current_balances cb
			LEFT JOIN operations_sum os ON (
				cb.user_id = os.user_id 
				AND cb.section_id = os.section_id
				AND cb.item_id = os.item_id
				AND cb.collection_id = os.collection_id
				AND cb.quality_level_id = os.quality_level_id
			)
			WHERE cb.daily_quantity + COALESCE(os.today_operations, 0) > 0
		)
		-- Основной запрос с JOIN для получения всех данных
		SELECT 
			fb.item_id,
			fb.final_quantity,
			-- Детали предмета
			i.item_class_id,
			i.item_type_id,
			-- Коды классификаторов
			ic.code as item_class_code,
			it.code as item_type_code,
			coll.code as collection_code,
			qual.code as quality_code
		FROM final_balances fb
		-- JOIN с таблицей предметов
		INNER JOIN inventory.items i ON fb.item_id = i.id
		-- JOIN для получения кода класса предмета
		INNER JOIN inventory.classifier_items ic ON i.item_class_id = ic.id
		-- JOIN для получения кода типа предмета  
		INNER JOIN inventory.classifier_items it ON i.item_type_id = it.id
		-- JOIN для получения кода коллекции
		INNER JOIN inventory.classifier_items coll ON fb.collection_id = coll.id
		-- JOIN для получения кода качества
		INNER JOIN inventory.classifier_items qual ON fb.quality_level_id = qual.id
		ORDER BY fb.item_id
	`

	rows, err := s.pool.Query(ctx, query, userID, sectionID)
	if err != nil {
		s.logger.Error("Failed to execute optimized inventory query", "error", err)
		if s.metrics != nil {
			s.metrics.RecordInventoryOperation("get_inventory_optimized", sectionID.String(), "error")
		}
		return nil, err
	}
	defer rows.Close()

	var result []*models.InventoryItemResponse
	for rows.Next() {
		var item models.InventoryItemResponse
		var collectionCode, qualityCode string

		if err := rows.Scan(
			&item.ItemID,
			&item.Quantity,
			// Пропускаем item_class_id и item_type_id так как нам нужны коды
			new(interface{}), // item_class_id
			new(interface{}), // item_type_id
			&item.ItemClass,
			&item.ItemType,
			&collectionCode,
			&qualityCode,
		); err != nil {
			s.logger.Error("Failed to scan optimized inventory item", "error", err)
			return nil, err
		}

		// Устанавливаем коды коллекции и качества
		item.Collection = &collectionCode
		item.QualityLevel = &qualityCode

		result = append(result, &item)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error in optimized inventory", "error", err)
		return nil, err
	}

	// Записываем метрики успеха
	if s.metrics != nil {
		s.metrics.RecordInventoryOperation("get_inventory_optimized", sectionID.String(), "success")
	}

	s.logger.Info("Optimized inventory query completed",
		"user_id", userID,
		"section_id", sectionID,
		"items_count", len(result))

	return result, nil
}

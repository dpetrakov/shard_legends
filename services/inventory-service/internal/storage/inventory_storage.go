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
	"github.com/shard-legends/inventory-service/internal/models"
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
func NewItemStorage(pool *pgxpool.Pool, logger *slog.Logger, metrics *metrics.Metrics) ItemRepository {
	return &itemStorage{
		pool:    pool,
		logger:  logger,
		metrics: metrics,
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
		WHERE user_id = $1 AND section_id = $2 AND quantity > 0
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
			return err
		}
	}
	
	s.logger.Info("Operations created successfully in transaction", "count", len(operations))
	return nil
}


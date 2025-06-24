package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	
	"github.com/shard-legends/inventory-service/internal/models"
)

// MockClassifierRepository is a mock implementation of ClassifierRepository
type MockClassifierRepository struct {
	GetClassifierByCodeFunc      func(ctx context.Context, code string) (*models.Classifier, error)
	GetClassifierItemsFunc       func(ctx context.Context, classifierID uuid.UUID) ([]*models.ClassifierItem, error)
	GetClassifierItemByCodeFunc  func(ctx context.Context, classifierID uuid.UUID, code string) (*models.ClassifierItem, error)
	GetAllClassifiersWithItemsFunc func(ctx context.Context) (map[string][]*models.ClassifierItem, error)
	GetCodeToUUIDMappingFunc     func(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
	GetUUIDToCodeMappingFunc     func(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
	InvalidateCacheFunc          func(ctx context.Context, classifierCode string) error
}

func (m *MockClassifierRepository) GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error) {
	if m.GetClassifierByCodeFunc != nil {
		return m.GetClassifierByCodeFunc(ctx, code)
	}
	return nil, ErrNotFound
}

func (m *MockClassifierRepository) GetClassifierItems(ctx context.Context, classifierID uuid.UUID) ([]*models.ClassifierItem, error) {
	if m.GetClassifierItemsFunc != nil {
		return m.GetClassifierItemsFunc(ctx, classifierID)
	}
	return nil, nil
}

func (m *MockClassifierRepository) GetClassifierItemByCode(ctx context.Context, classifierID uuid.UUID, code string) (*models.ClassifierItem, error) {
	if m.GetClassifierItemByCodeFunc != nil {
		return m.GetClassifierItemByCodeFunc(ctx, classifierID, code)
	}
	return nil, ErrNotFound
}

func (m *MockClassifierRepository) GetAllClassifiersWithItems(ctx context.Context) (map[string][]*models.ClassifierItem, error) {
	if m.GetAllClassifiersWithItemsFunc != nil {
		return m.GetAllClassifiersWithItemsFunc(ctx)
	}
	return make(map[string][]*models.ClassifierItem), nil
}

func (m *MockClassifierRepository) GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	if m.GetCodeToUUIDMappingFunc != nil {
		return m.GetCodeToUUIDMappingFunc(ctx, classifierCode)
	}
	return make(map[string]uuid.UUID), nil
}

func (m *MockClassifierRepository) GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	if m.GetUUIDToCodeMappingFunc != nil {
		return m.GetUUIDToCodeMappingFunc(ctx, classifierCode)
	}
	return make(map[uuid.UUID]string), nil
}

func (m *MockClassifierRepository) InvalidateCache(ctx context.Context, classifierCode string) error {
	if m.InvalidateCacheFunc != nil {
		return m.InvalidateCacheFunc(ctx, classifierCode)
	}
	return nil
}

// MockItemRepository is a mock implementation of ItemRepository
type MockItemRepository struct {
	GetItemByIDFunc            func(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
	GetItemsByClassFunc        func(ctx context.Context, classCode string) ([]*models.Item, error)
	GetItemsByClassAndTypeFunc func(ctx context.Context, classCode, typeCode string) ([]*models.Item, error)
	GetItemImageFunc           func(ctx context.Context, itemID, collectionID, qualityLevelID uuid.UUID) (*models.ItemImage, error)
	GetItemImagesFunc          func(ctx context.Context, itemID uuid.UUID) ([]*models.ItemImage, error)
	GetItemWithDetailsFunc     func(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error)
}

func (m *MockItemRepository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error) {
	if m.GetItemByIDFunc != nil {
		return m.GetItemByIDFunc(ctx, itemID)
	}
	return nil, ErrNotFound
}

func (m *MockItemRepository) GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error) {
	if m.GetItemsByClassFunc != nil {
		return m.GetItemsByClassFunc(ctx, classCode)
	}
	return nil, nil
}

func (m *MockItemRepository) GetItemsByClassAndType(ctx context.Context, classCode, typeCode string) ([]*models.Item, error) {
	if m.GetItemsByClassAndTypeFunc != nil {
		return m.GetItemsByClassAndTypeFunc(ctx, classCode, typeCode)
	}
	return nil, nil
}

func (m *MockItemRepository) GetItemImage(ctx context.Context, itemID, collectionID, qualityLevelID uuid.UUID) (*models.ItemImage, error) {
	if m.GetItemImageFunc != nil {
		return m.GetItemImageFunc(ctx, itemID, collectionID, qualityLevelID)
	}
	return nil, ErrNotFound
}

func (m *MockItemRepository) GetItemImages(ctx context.Context, itemID uuid.UUID) ([]*models.ItemImage, error) {
	if m.GetItemImagesFunc != nil {
		return m.GetItemImagesFunc(ctx, itemID)
	}
	return nil, nil
}

func (m *MockItemRepository) GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error) {
	if m.GetItemWithDetailsFunc != nil {
		return m.GetItemWithDetailsFunc(ctx, itemID)
	}
	return nil, ErrNotFound
}

// MockInventoryRepository is a mock implementation of InventoryRepository
type MockInventoryRepository struct {
	GetDailyBalanceFunc                func(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error)
	GetLatestDailyBalanceFunc          func(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error)
	CreateDailyBalanceFunc             func(ctx context.Context, balance *models.DailyBalance) error
	GetOperationsFunc                  func(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error)
	GetOperationsByExternalIDFunc      func(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error)
	CreateOperationFunc                func(ctx context.Context, operation *models.Operation) error
	CreateOperationsFunc               func(ctx context.Context, operations []*models.Operation) error
	CreateOperationsInTransactionFunc  func(ctx context.Context, tx *sqlx.Tx, operations []*models.Operation) error
	GetUserInventoryItemsFunc          func(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error)
	BeginTransactionFunc               func(ctx context.Context) (*sqlx.Tx, error)
	CommitTransactionFunc              func(tx *sqlx.Tx) error
	RollbackTransactionFunc            func(tx *sqlx.Tx) error
}

func (m *MockInventoryRepository) GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error) {
	if m.GetDailyBalanceFunc != nil {
		return m.GetDailyBalanceFunc(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, date)
	}
	return nil, ErrNotFound
}

func (m *MockInventoryRepository) GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error) {
	if m.GetLatestDailyBalanceFunc != nil {
		return m.GetLatestDailyBalanceFunc(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, beforeDate)
	}
	return nil, ErrNotFound
}

func (m *MockInventoryRepository) CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error {
	if m.CreateDailyBalanceFunc != nil {
		return m.CreateDailyBalanceFunc(ctx, balance)
	}
	return nil
}

func (m *MockInventoryRepository) GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error) {
	if m.GetOperationsFunc != nil {
		return m.GetOperationsFunc(ctx, userID, sectionID, itemID, collectionID, qualityLevelID, fromDate)
	}
	return nil, nil
}

func (m *MockInventoryRepository) GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error) {
	if m.GetOperationsByExternalIDFunc != nil {
		return m.GetOperationsByExternalIDFunc(ctx, operationID)
	}
	return nil, nil
}

func (m *MockInventoryRepository) CreateOperation(ctx context.Context, operation *models.Operation) error {
	if m.CreateOperationFunc != nil {
		return m.CreateOperationFunc(ctx, operation)
	}
	return nil
}

func (m *MockInventoryRepository) CreateOperations(ctx context.Context, operations []*models.Operation) error {
	if m.CreateOperationsFunc != nil {
		return m.CreateOperationsFunc(ctx, operations)
	}
	return nil
}

func (m *MockInventoryRepository) CreateOperationsInTransaction(ctx context.Context, tx *sqlx.Tx, operations []*models.Operation) error {
	if m.CreateOperationsInTransactionFunc != nil {
		return m.CreateOperationsInTransactionFunc(ctx, tx, operations)
	}
	return nil
}

func (m *MockInventoryRepository) GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error) {
	if m.GetUserInventoryItemsFunc != nil {
		return m.GetUserInventoryItemsFunc(ctx, userID, sectionID)
	}
	return nil, nil
}

func (m *MockInventoryRepository) BeginTransaction(ctx context.Context) (*sqlx.Tx, error) {
	if m.BeginTransactionFunc != nil {
		return m.BeginTransactionFunc(ctx)
	}
	return nil, nil
}

func (m *MockInventoryRepository) CommitTransaction(tx *sqlx.Tx) error {
	if m.CommitTransactionFunc != nil {
		return m.CommitTransactionFunc(tx)
	}
	return nil
}

func (m *MockInventoryRepository) RollbackTransaction(tx *sqlx.Tx) error {
	if m.RollbackTransactionFunc != nil {
		return m.RollbackTransactionFunc(tx)
	}
	return nil
}

// MockCache is a mock implementation of Cache
type MockCache struct {
	GetFunc           func(ctx context.Context, key string, value interface{}) error
	SetFunc           func(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteFunc        func(ctx context.Context, key string) error
	DeletePatternFunc func(ctx context.Context, pattern string) error
}

func (m *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key, value)
	}
	return nil
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	if m.DeletePatternFunc != nil {
		return m.DeletePatternFunc(ctx, pattern)
	}
	return nil
}
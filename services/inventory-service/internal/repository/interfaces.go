package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	
	"github.com/shard-legends/inventory-service/internal/models"
)

// ClassifierRepository defines methods for working with classifiers
type ClassifierRepository interface {
	// GetClassifierByCode retrieves a classifier by its code
	GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error)
	
	// GetClassifierItems retrieves all items for a classifier
	GetClassifierItems(ctx context.Context, classifierID uuid.UUID) ([]*models.ClassifierItem, error)
	
	// GetClassifierItemByCode retrieves a specific classifier item by code
	GetClassifierItemByCode(ctx context.Context, classifierID uuid.UUID, code string) (*models.ClassifierItem, error)
	
	// GetAllClassifiersWithItems retrieves all classifiers with their items
	GetAllClassifiersWithItems(ctx context.Context) (map[string][]*models.ClassifierItem, error)
	
	// GetCodeToUUIDMapping returns a mapping of codes to UUIDs for a classifier
	GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
	
	// GetUUIDToCodeMapping returns a mapping of UUIDs to codes for a classifier
	GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
	
	// InvalidateCache invalidates the cache for a specific classifier
	InvalidateCache(ctx context.Context, classifierCode string) error
}

// ItemRepository defines methods for working with items
type ItemRepository interface {
	// GetItemByID retrieves an item by its ID
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
	
	// GetItemsByClass retrieves all items for a specific class
	GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error)
	
	// GetItemsByClassAndType retrieves items by class and type codes
	GetItemsByClassAndType(ctx context.Context, classCode, typeCode string) ([]*models.Item, error)
	
	// GetItemImage retrieves an image for a specific item variant
	GetItemImage(ctx context.Context, itemID, collectionID, qualityLevelID uuid.UUID) (*models.ItemImage, error)
	
	// GetItemImages retrieves all active images for an item
	GetItemImages(ctx context.Context, itemID uuid.UUID) ([]*models.ItemImage, error)
	
	// GetItemWithDetails retrieves an item with classifier details loaded
	GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error)
}

// InventoryRepository defines methods for working with inventory operations and balances
type InventoryRepository interface {
	// GetDailyBalance retrieves a daily balance for a specific item
	GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error)
	
	// GetLatestDailyBalance retrieves the most recent daily balance before or on the given date
	GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error)
	
	// CreateDailyBalance creates a new daily balance record
	CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error
	
	// GetOperations retrieves operations for a specific item from a given date
	GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error)
	
	// GetOperationsByExternalID retrieves operations by their external operation ID
	GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error)
	
	// CreateOperation creates a new operation
	CreateOperation(ctx context.Context, operation *models.Operation) error
	
	// CreateOperations creates multiple operations in a batch
	CreateOperations(ctx context.Context, operations []*models.Operation) error
	
	// CreateOperationsInTransaction creates operations within a transaction
	CreateOperationsInTransaction(ctx context.Context, tx *sqlx.Tx, operations []*models.Operation) error
	
	// GetUserInventoryItems retrieves all unique item combinations for a user in a section
	GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error)
	
	// BeginTransaction starts a new database transaction
	BeginTransaction(ctx context.Context) (*sqlx.Tx, error)
	
	// CommitTransaction commits a transaction
	CommitTransaction(tx *sqlx.Tx) error
	
	// RollbackTransaction rolls back a transaction
	RollbackTransaction(tx *sqlx.Tx) error
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	Classifier ClassifierRepository
	Item       ItemRepository
	Inventory  InventoryRepository
}

// NewRepositories creates a new Repositories instance
func NewRepositories(db *sqlx.DB, cache Cache) *Repositories {
	return &Repositories{
		Classifier: NewClassifierRepository(db, cache),
		Item:       NewItemRepository(db),
		Inventory:  NewInventoryRepository(db),
	}
}

// Cache defines caching interface
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string, value interface{}) error
	
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error
	
	// DeletePattern removes all keys matching a pattern
	DeletePattern(ctx context.Context, pattern string) error
}

// Common errors
var (
	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = sql.ErrNoRows
	
	// ErrDuplicateKey indicates a unique constraint violation
	ErrDuplicateKey = &DuplicateKeyError{}
)

// DuplicateKeyError represents a duplicate key error
type DuplicateKeyError struct {
	Key string
}

func (e *DuplicateKeyError) Error() string {
	if e.Key != "" {
		return "duplicate key: " + e.Key
	}
	return "duplicate key violation"
}

// IsDuplicateKeyError checks if an error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	_, ok := err.(*DuplicateKeyError)
	return ok
}
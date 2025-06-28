package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/shard-legends/inventory-service/internal/models"
)

// InventoryService defines the main business logic interface for inventory operations
type InventoryService interface {
	// Core operations
	CalculateCurrentBalance(ctx context.Context, req *BalanceRequest) (int64, error)
	CreateDailyBalance(ctx context.Context, req *DailyBalanceRequest) (*models.DailyBalance, error)
	CheckSufficientBalance(ctx context.Context, req *SufficientBalanceRequest) error
	CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error)

	// Utility operations
	ConvertClassifierCodes(ctx context.Context, req *CodeConversionRequest) (*CodeConversionResponse, error)
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error

	// High-level business operations
	GetUserInventory(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error)
	// D-15: Legacy method for performance comparison (not in public interface)
	GetUserInventoryLegacy(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error)
	AddItems(ctx context.Context, req *models.AddItemsRequest) ([]uuid.UUID, error)
	AdjustInventory(ctx context.Context, req *models.AdjustInventoryRequest) (*models.AdjustInventoryResponse, error)

	// Reservation operations
	ReserveItems(ctx context.Context, req *models.ReserveItemsRequest) ([]uuid.UUID, error)
	ReturnReservedItems(ctx context.Context, req *models.ReturnReserveRequest) error
	ConsumeReservedItems(ctx context.Context, req *models.ConsumeReserveRequest) error
	GetReservationStatus(ctx context.Context, operationID uuid.UUID) (*models.ReservationStatusResponse, error)

	// Item details with i18n
	GetItemsDetails(ctx context.Context, req *models.ItemDetailsRequest, languageCode string) (*models.ItemDetailsResponse, error)

	// I18n utilities
	GetDefaultLanguage(ctx context.Context) (*models.Language, error)
}

// ClassifierService defines methods for working with classifiers and their mappings
type ClassifierService interface {
	GetClassifierMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
	GetReverseClassifierMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
	RefreshClassifierCache(ctx context.Context, classifierCode string) error
}

// BalanceCalculator defines the interface for balance calculation algorithms
type BalanceCalculator interface {
	CalculateCurrentBalance(ctx context.Context, req *BalanceRequest) (int64, error)
}

// DailyBalanceCreator defines the interface for daily balance creation
type DailyBalanceCreator interface {
	CreateDailyBalance(ctx context.Context, req *DailyBalanceRequest) (*models.DailyBalance, error)
}

// CodeConverter defines the interface for classifier code conversion
type CodeConverter interface {
	ConvertClassifierCodes(ctx context.Context, req *CodeConversionRequest) (*CodeConversionResponse, error)
}

// BalanceChecker defines the interface for balance sufficiency checks
type BalanceChecker interface {
	CheckSufficientBalance(ctx context.Context, req *SufficientBalanceRequest) error
}

// OperationCreator defines the interface for creating operations in transactions
type OperationCreator interface {
	CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error)
	CreateReservationOperations(ctx context.Context, req *models.ReserveItemsRequest) ([]uuid.UUID, error)
	CreateReturnOperations(ctx context.Context, req *models.ReturnReserveRequest) error
	CreateConsumptionOperations(ctx context.Context, req *models.ConsumeReserveRequest) error
}

// CacheManager defines the interface for cache management operations
type CacheManager interface {
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
	InvalidateClassifierCache(ctx context.Context, classifierCode string) error
}

// Request/Response types for service operations

// BalanceRequest represents a request to calculate current balance
type BalanceRequest struct {
	UserID         uuid.UUID
	SectionID      uuid.UUID
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
}

// DailyBalanceRequest represents a request to create a daily balance
type DailyBalanceRequest struct {
	UserID         uuid.UUID
	SectionID      uuid.UUID
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
	TargetDate     time.Time // usually yesterday's date
}

// SufficientBalanceRequest represents a request to check if user has sufficient items
type SufficientBalanceRequest struct {
	UserID    uuid.UUID
	SectionID uuid.UUID
	Items     []ItemQuantityCheck
}

// ItemQuantityCheck represents an item with required quantity for balance checking
type ItemQuantityCheck struct {
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
	RequiredQty    int64
}

// CodeConversionRequest represents a request to convert classifier codes
type CodeConversionRequest struct {
	Direction string                 // "toUUID" or "fromUUID"
	Data      map[string]interface{} // object with codes/UUIDs
}

// CodeConversionResponse represents the response from code conversion
type CodeConversionResponse struct {
	Data map[string]interface{} // converted object
}

// OperationBatch represents a batch of operations to be created together
type OperationBatch struct {
	Operations []*models.Operation
	UserID     uuid.UUID
	ExternalID uuid.UUID // External operation ID for linking
}

// InsufficientBalanceDetails represents details about insufficient items
type InsufficientBalanceDetails struct {
	MissingItems []MissingItemDetails `json:"missing_items"`
}

// MissingItemDetails represents details about a missing item
type MissingItemDetails struct {
	ItemID         uuid.UUID `json:"item_id"`
	CollectionID   uuid.UUID `json:"collection_id"`
	QualityLevelID uuid.UUID `json:"quality_level_id"`
	Required       int64     `json:"required"`
	Available      int64     `json:"available"`
	Missing        int64     `json:"missing"`
}

// RepositoryInterfaces defines all repository interfaces needed by services
type RepositoryInterfaces struct {
	Classifier ClassifierRepositoryInterface
	Item       ItemRepositoryInterface
	Inventory  InventoryRepositoryInterface
}

// ClassifierRepositoryInterface defines the classifier repository interface
type ClassifierRepositoryInterface interface {
	GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error)
	GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
	GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
	InvalidateCache(ctx context.Context, classifierCode string) error
}

// ItemRepositoryInterface defines the item repository interface
type ItemRepositoryInterface interface {
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
	GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error)

	// I18n and batch operations
	GetItemsBatch(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID]*models.ItemWithDetails, error)
	GetTranslationsBatch(ctx context.Context, entityType string, entityIDs []uuid.UUID, languageCode string) (map[uuid.UUID]map[string]string, error)
	GetDefaultLanguage(ctx context.Context) (*models.Language, error)
	GetItemImagesBatch(ctx context.Context, requests []models.ItemDetailRequestItem) (map[string]string, error)
}

// InventoryRepositoryInterface defines the inventory repository interface
type InventoryRepositoryInterface interface {
	GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error)
	GetLatestDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, beforeDate time.Time) (*models.DailyBalance, error)
	CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error
	GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error)
	GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error)
	CreateOperationsInTransaction(ctx context.Context, tx interface{}, operations []*models.Operation) error
	GetUserInventoryItems(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.ItemKey, error)
	BeginTransaction(ctx context.Context) (interface{}, error)
	CommitTransaction(tx interface{}) error
	RollbackTransaction(tx interface{}) error

	// Atomic balance checking with row-level locking
	CheckAndLockBalances(ctx context.Context, tx interface{}, items []BalanceLockRequest) ([]BalanceLockResult, error)

	// D-15: Optimized methods to eliminate N+1 queries
	GetUserInventoryOptimized(ctx context.Context, userID, sectionID uuid.UUID) ([]*models.InventoryItemResponse, error)
}

// CacheInterface defines the caching interface
type CacheInterface interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// ServiceDependencies aggregates all dependencies needed by services
type ServiceDependencies struct {
	Repositories   *RepositoryInterfaces
	Cache          CacheInterface
	Metrics        MetricsInterface
	CodeConverter  CodeConverter
	BalanceChecker BalanceChecker
}

// BalanceLockRequest represents a request to lock and check balance for a specific item
type BalanceLockRequest struct {
	UserID         uuid.UUID
	SectionID      uuid.UUID
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
	RequiredQty    int64
}

// BalanceLockResult represents the result of an atomic balance check and lock
type BalanceLockResult struct {
	BalanceLockRequest
	AvailableQty int64
	Sufficient   bool
	Error        error
}

// MetricsInterface defines the metrics collection interface
type MetricsInterface interface {
	RecordInventoryOperation(operationType, section, status string)
	RecordBalanceCalculation(section string, duration time.Duration, status string)
	RecordCacheHit(cacheType string)
	RecordCacheMiss(cacheType string)
	RecordTransactionMetrics(operationType string, operationCount int, duration time.Duration)
	RecordItemsPerInventory(section string, count int)
}

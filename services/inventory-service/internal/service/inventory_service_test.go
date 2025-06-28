package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shard-legends/inventory-service/internal/models"
)

// Extended mocks for full service testing
type MockClassifierRepo struct {
	mock.Mock
}

func (m *MockClassifierRepo) GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Classifier), args.Error(1)
}

func (m *MockClassifierRepo) GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	args := m.Called(ctx, classifierCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]uuid.UUID), args.Error(1)
}

func (m *MockClassifierRepo) GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	args := m.Called(ctx, classifierCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]string), args.Error(1)
}

func (m *MockClassifierRepo) InvalidateCache(ctx context.Context, classifierCode string) error {
	args := m.Called(ctx, classifierCode)
	return args.Error(0)
}

type MockItemRepo struct {
	mock.Mock
}

func (m *MockItemRepo) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemRepo) GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ItemWithDetails), args.Error(1)
}

func (m *MockItemRepo) GetItemsBatch(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID]*models.ItemWithDetails, error) {
	args := m.Called(ctx, itemIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*models.ItemWithDetails), args.Error(1)
}

func (m *MockItemRepo) GetTranslationsBatch(ctx context.Context, entityType string, entityIDs []uuid.UUID, languageCode string) (map[uuid.UUID]map[string]string, error) {
	args := m.Called(ctx, entityType, entityIDs, languageCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]map[string]string), args.Error(1)
}

func (m *MockItemRepo) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Language), args.Error(1)
}

func (m *MockItemRepo) GetItemImagesBatch(ctx context.Context, requests []models.ItemDetailRequestItem) (map[string]string, error) {
	args := m.Called(ctx, requests)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func createFullTestDeps() (*ServiceDependencies, *MockCache, *MockClassifierRepo, *MockItemRepo, *MockInventoryRepo) {
	cache := new(MockCache)
	classifierRepo := new(MockClassifierRepo)
	itemRepo := new(MockItemRepo)
	inventoryRepo := new(MockInventoryRepo)

	deps := &ServiceDependencies{
		Cache: cache,
		Repositories: &RepositoryInterfaces{
			Classifier: classifierRepo,
			Item:       itemRepo,
			Inventory:  inventoryRepo,
		},
	}

	// Initialize BalanceChecker to avoid nil pointer issues
	deps.BalanceChecker = NewBalanceChecker(deps)

	return deps, cache, classifierRepo, itemRepo, inventoryRepo
}

func TestInventoryService_CalculateCurrentBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, cache, _, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	req := &BalanceRequest{
		UserID:         uuid.New(),
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
	}

	expectedBalance := int64(100)
	cacheKey := "inventory:" + req.UserID.String() + ":" + req.SectionID.String() + ":" +
		req.ItemID.String() + ":" + req.CollectionID.String() + ":" + req.QualityLevelID.String()

	// Mock cache hit
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(nil).Run(func(args mock.Arguments) {
		balance := args.Get(2).(*int64)
		*balance = expectedBalance
	})

	// Act
	result, err := service.CalculateCurrentBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result)
	cache.AssertExpectations(t)
	inventoryRepo.AssertNotCalled(t, "GetLatestDailyBalance")
}

func TestInventoryService_GetUserInventory(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()

	// Mock the optimized method to return complete inventory data
	expectedItems := []*models.InventoryItemResponse{
		{
			ItemID:       itemID,
			ItemClass:    "resources",
			ItemType:     "stone",
			Collection:   stringPtr("common"),
			QualityLevel: stringPtr("basic"),
			Quantity:     50,
		},
	}

	inventoryRepo.On("GetUserInventoryOptimized", ctx, userID, sectionID).Return(expectedItems, nil)

	// Act
	result, err := service.GetUserInventory(ctx, userID, sectionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	item := result[0]
	assert.Equal(t, itemID, item.ItemID)
	assert.Equal(t, "resources", item.ItemClass)
	assert.Equal(t, "stone", item.ItemType)
	assert.Equal(t, "common", *item.Collection)
	assert.Equal(t, "basic", *item.QualityLevel)
	assert.Equal(t, int64(50), item.Quantity)

	// Verify all mocks
	inventoryRepo.AssertExpectations(t)
}

func TestInventoryService_AddItems(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, cache, classifierRepo, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	operationID := uuid.New()
	itemID := uuid.New()

	collection := "common"
	qualityLevel := "basic"

	req := &models.AddItemsRequest{
		UserID:        userID,
		Section:       models.SectionMain,
		OperationType: models.OperationTypeChestReward,
		OperationID:   operationID,
		Items: []models.ItemQuantityRequest{
			{
				ItemID:       itemID,
				Collection:   &collection,
				QualityLevel: &qualityLevel,
				Quantity:     50,
			},
		},
	}

	// Mock section mapping
	sectionMapping := map[string]uuid.UUID{
		models.SectionMain: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock operation type mapping
	operationMapping := map[string]uuid.UUID{
		models.OperationTypeChestReward: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierOperationType).Return(operationMapping, nil)

	// Mock collection mapping
	collectionMapping := map[string]uuid.UUID{
		collection: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)

	// Mock quality level mapping
	qualityMapping := map[string]uuid.UUID{
		qualityLevel: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)

	// Mock transaction
	tx := &struct{}{}
	inventoryRepo.On("BeginTransaction", ctx).Return(tx, nil)
	inventoryRepo.On("CreateOperationsInTransaction", ctx, tx, mock.MatchedBy(func(ops []*models.Operation) bool {
		return len(ops) == 1 && ops[0].QuantityChange == 50 && ops[0].UserID == userID
	})).Return(nil)
	inventoryRepo.On("CommitTransaction", tx).Return(nil)

	// Mock cache invalidation
	cache.On("DeletePattern", ctx, "inventory:"+userID.String()+":*").Return(nil)

	// Act
	result, err := service.AddItems(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	inventoryRepo.AssertExpectations(t)
	classifierRepo.AssertExpectations(t)
}

func TestInventoryService_ReserveItems_InsufficientBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, cache, classifierRepo, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	operationID := uuid.New()
	itemID := uuid.New()
	mainSectionID := uuid.New()
	factorySectionID := uuid.New()
	defaultCollectionID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	defaultQualityID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	req := &models.ReserveItemsRequest{
		UserID:      userID,
		OperationID: operationID,
		Items: []models.ItemQuantityRequest{
			{
				ItemID:   itemID,
				Quantity: 100, // Requesting more than available
			},
		},
	}

	// Mock section mapping
	sectionMapping := map[string]uuid.UUID{
		models.SectionMain:    mainSectionID,
		models.SectionFactory: factorySectionID,
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock operation type mapping - may not be called if balance check fails early
	operationMapping := map[string]uuid.UUID{
		models.OperationTypeFactoryReservation: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierOperationType).Return(operationMapping, nil).Maybe()

	// Mock collection and quality level mappings (using defaults) - these may not be called if balance check fails
	collectionMapping := map[string]uuid.UUID{}
	qualityMapping := map[string]uuid.UUID{}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil).Maybe()
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil).Maybe()

	// Mock balance calculation - insufficient balance (available = 50, required = 100)
	cacheKey := "inventory:" + userID.String() + ":" + mainSectionID.String() + ":" +
		itemID.String() + ":" + defaultCollectionID.String() + ":" + defaultQualityID.String()

	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError).Maybe()
	inventoryRepo.On("GetLatestDailyBalance", ctx, userID, mainSectionID, itemID, defaultCollectionID, defaultQualityID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

	// Mock GetDailyBalance call from CreateDailyBalance
	inventoryRepo.On("GetDailyBalance", ctx, userID, mainSectionID, itemID, defaultCollectionID, defaultQualityID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()

	// Mock operations that give us a balance of 25 (less than required 100)
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       mainSectionID,
			ItemID:          itemID,
			CollectionID:    defaultCollectionID,
			QualityLevelID:  defaultQualityID,
			QuantityChange:  25, // Available balance = 25, less than required 100
			OperationTypeID: uuid.New(),
			CreatedAt:       time.Now().UTC().AddDate(0, 0, -1),
		},
	}

	// Mock operations for CreateDailyBalance (called first)
	inventoryRepo.On("GetOperations", ctx, userID, mainSectionID, itemID, defaultCollectionID, defaultQualityID, time.Time{}).Return(operations, nil).Maybe()

	// Mock CreateDailyBalance call
	inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()

	// Mock operations for balance calculation (called second)
	inventoryRepo.On("GetOperations", ctx, userID, mainSectionID, itemID, defaultCollectionID, defaultQualityID, mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()

	// Mock GetOperationsByExternalID that's called in CreateReservationOperations
	inventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return([]*models.Operation{}, nil).Maybe()

	// Mock transaction methods for CreateReservationOperations
	tx := &struct{}{} // Simple mock transaction
	inventoryRepo.On("BeginTransaction", ctx).Return(tx, nil).Maybe()
	inventoryRepo.On("RollbackTransaction", tx).Return(nil).Maybe()
	inventoryRepo.On("CheckAndLockBalances", ctx, tx, mock.AnythingOfType("[]service.BalanceLockRequest")).Return(
		[]BalanceLockResult{
			{
				BalanceLockRequest: BalanceLockRequest{
					UserID:         userID,
					SectionID:      mainSectionID,
					ItemID:         itemID,
					CollectionID:   defaultCollectionID,
					QualityLevelID: defaultQualityID,
					RequiredQty:    100,
				},
				AvailableQty: 50,
				Sufficient:   false,
				Error:        nil,
			},
		}, nil).Maybe()

	cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()

	// Act
	result, err := service.ReserveItems(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	t.Logf("Error received: %v", err)
	t.Logf("Error type: %T", err)

	// The error should contain information about insufficient balance
	assert.Contains(t, err.Error(), "insufficient balance for reservation")

	inventoryRepo.AssertExpectations(t)
	classifierRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestInventoryService_InvalidateUserCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, cache, _, _, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	pattern := "inventory:" + userID.String() + ":*"

	cache.On("DeletePattern", ctx, pattern).Return(nil)

	// Act
	err := service.InvalidateUserCache(ctx, userID)

	// Assert
	assert.NoError(t, err)
	cache.AssertExpectations(t)
}

func TestInventoryService_ConvertClassifierCodes(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, classifierRepo, _, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	req := &CodeConversionRequest{
		Direction: "toUUID",
		Data: map[string]interface{}{
			"section": models.SectionMain,
		},
	}

	// Mock section mapping
	sectionMapping := map[string]uuid.UUID{
		models.SectionMain: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Act
	result, err := service.ConvertClassifierCodes(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Data, "section_id")
	assert.Equal(t, sectionMapping[models.SectionMain], result.Data["section_id"])
	assert.NotContains(t, result.Data, "section")

	classifierRepo.AssertExpectations(t)
}

func TestInventoryService_CreateDailyBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	req := &DailyBalanceRequest{
		UserID:         uuid.New(),
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
		TargetDate:     time.Now().UTC().AddDate(0, 0, -1),
	}

	expectedBalance := &models.DailyBalance{
		UserID:         req.UserID,
		SectionID:      req.SectionID,
		ItemID:         req.ItemID,
		CollectionID:   req.CollectionID,
		QualityLevelID: req.QualityLevelID,
		BalanceDate:    req.TargetDate.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second),
		Quantity:       75,
		CreatedAt:      time.Now().UTC(),
	}

	// Mock daily balance doesn't exist
	inventoryRepo.On("GetDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock no previous balance
	inventoryRepo.On("GetLatestDailyBalance", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError)

	// Mock operations
	inventoryRepo.On("GetOperations", ctx, req.UserID, req.SectionID, req.ItemID, req.CollectionID, req.QualityLevelID, time.Time{}).Return([]*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          req.UserID,
			SectionID:       req.SectionID,
			ItemID:          req.ItemID,
			CollectionID:    req.CollectionID,
			QualityLevelID:  req.QualityLevelID,
			QuantityChange:  75,
			OperationTypeID: uuid.New(),
			CreatedAt:       req.TargetDate.Add(-1 * time.Hour),
		},
	}, nil)

	// Mock create balance
	inventoryRepo.On("CreateDailyBalance", ctx, mock.MatchedBy(func(balance *models.DailyBalance) bool {
		return balance.Quantity == 75
	})).Return(nil)

	// Act
	result, err := service.CreateDailyBalance(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance.Quantity, result.Quantity)
	assert.Equal(t, req.UserID, result.UserID)

	inventoryRepo.AssertExpectations(t)
}

func TestInventoryService_GetItemsDetails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, itemRepo, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	itemID1 := uuid.New()
	itemID2 := uuid.New()

	req := &models.ItemDetailsRequest{
		Items: []models.ItemDetailRequestItem{
			{
				ItemID:       itemID1,
				Collection:   stringPtr("winter_2025"),
				QualityLevel: stringPtr("stone"),
			},
			{
				ItemID:       itemID2,
				Collection:   nil,
				QualityLevel: nil,
			},
		},
	}

	// Mock items batch
	itemsMap := map[uuid.UUID]*models.ItemWithDetails{
		itemID1: {
			Item: models.Item{
				ID: itemID1,
			},
			ItemClass: "resources",
			ItemType:  "stone",
		},
		itemID2: {
			Item: models.Item{
				ID: itemID2,
			},
			ItemClass: "reagents",
			ItemType:  "disc",
		},
	}
	itemRepo.On("GetItemsBatch", ctx, []uuid.UUID{itemID1, itemID2}).Return(itemsMap, nil)

	// Mock translations
	translationsMap := map[uuid.UUID]map[string]string{
		itemID1: {
			"name":        "Камень",
			"description": "Базовый строительный материал",
		},
		itemID2: {
			"name":        "Диск",
			"description": "Реагент для обработки",
		},
	}
	itemRepo.On("GetTranslationsBatch", ctx, "item", []uuid.UUID{itemID1, itemID2}, "ru").Return(translationsMap, nil)

	// Mock default language
	defaultLang := &models.Language{
		Code:      "en",
		Name:      "English",
		IsDefault: true,
		IsActive:  true,
	}
	itemRepo.On("GetDefaultLanguage", ctx).Return(defaultLang, nil)

	// Mock fallback translations (not called since language is different)
	fallbackTranslations := map[uuid.UUID]map[string]string{
		itemID1: {
			"name":        "Stone",
			"description": "Basic building material",
		},
		itemID2: {
			"name":        "Disc",
			"description": "Processing reagent",
		},
	}
	itemRepo.On("GetTranslationsBatch", ctx, "item", []uuid.UUID{itemID1, itemID2}, "en").Return(fallbackTranslations, nil)

	// Mock images
	imagesMap := map[string]string{
		itemID1.String() + "_winter_2025_stone": "https://cdn.example.com/items/stone_winter_2025_stone.png",
		itemID2.String():                        "https://cdn.example.com/items/disc_default.png",
	}
	itemRepo.On("GetItemImagesBatch", ctx, req.Items).Return(imagesMap, nil)

	// Act
	result, err := service.GetItemsDetails(ctx, req, "ru")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)

	// Check first item
	item1 := result.Items[0]
	assert.Equal(t, itemID1, item1.ItemID)
	assert.Equal(t, "stone", item1.Code)
	assert.Equal(t, "Камень", item1.Name)
	assert.Equal(t, "Базовый строительный материал", item1.Description)
	assert.Equal(t, "https://cdn.example.com/items/stone_winter_2025_stone.png", item1.ImageURL)
	assert.Equal(t, "winter_2025", *item1.Collection)
	assert.Equal(t, "stone", *item1.QualityLevel)

	// Check second item
	item2 := result.Items[1]
	assert.Equal(t, itemID2, item2.ItemID)
	assert.Equal(t, "disc", item2.Code)
	assert.Equal(t, "Диск", item2.Name)
	assert.Equal(t, "Реагент для обработки", item2.Description)
	assert.Equal(t, "https://cdn.example.com/items/disc_default.png", item2.ImageURL)
	assert.Nil(t, item2.Collection)
	assert.Nil(t, item2.QualityLevel)

	itemRepo.AssertExpectations(t)
}

func TestInventoryService_GetItemsDetails_WithFallback(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, itemRepo, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	itemID := uuid.New()

	req := &models.ItemDetailsRequest{
		Items: []models.ItemDetailRequestItem{
			{
				ItemID:       itemID,
				Collection:   stringPtr("winter_2025"),
				QualityLevel: stringPtr("stone"),
			},
		},
	}

	// Mock items batch
	itemsMap := map[uuid.UUID]*models.ItemWithDetails{
		itemID: {
			Item: models.Item{
				ID: itemID,
			},
			ItemClass: "resources",
			ItemType:  "stone",
		},
	}
	itemRepo.On("GetItemsBatch", ctx, []uuid.UUID{itemID}).Return(itemsMap, nil)

	// Mock translations - missing for requested language
	translationsMap := map[uuid.UUID]map[string]string{
		itemID: {
			"description": "Базовый строительный материал", // Only description in Russian
		},
	}
	itemRepo.On("GetTranslationsBatch", ctx, "item", []uuid.UUID{itemID}, "ru").Return(translationsMap, nil)

	// Mock default language
	defaultLang := &models.Language{
		Code:      "en",
		Name:      "English",
		IsDefault: true,
		IsActive:  true,
	}
	itemRepo.On("GetDefaultLanguage", ctx).Return(defaultLang, nil)

	// Mock fallback translations
	fallbackTranslations := map[uuid.UUID]map[string]string{
		itemID: {
			"name":        "Stone", // Name available in fallback language
			"description": "Basic building material",
		},
	}
	itemRepo.On("GetTranslationsBatch", ctx, "item", []uuid.UUID{itemID}, "en").Return(fallbackTranslations, nil)

	// Mock images
	imagesMap := map[string]string{
		itemID.String() + "_winter_2025_stone": "https://cdn.example.com/items/stone_winter_2025_stone.png",
	}
	itemRepo.On("GetItemImagesBatch", ctx, req.Items).Return(imagesMap, nil)

	// Act
	result, err := service.GetItemsDetails(ctx, req, "ru")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)

	item := result.Items[0]
	assert.Equal(t, itemID, item.ItemID)
	assert.Equal(t, "Stone", item.Name)                                // Fallback name from English
	assert.Equal(t, "Базовый строительный материал", item.Description) // Russian description

	itemRepo.AssertExpectations(t)
}

func TestInventoryService_GetItemsDetails_EmptyRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, _, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	req := &models.ItemDetailsRequest{
		Items: []models.ItemDetailRequestItem{},
	}

	// Act
	result, err := service.GetItemsDetails(ctx, req, "ru")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 0)
}

func TestInventoryService_GetItemsDetails_NilRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, _, _, _, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	// Act
	result, err := service.GetItemsDetails(ctx, nil, "ru")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "item details request cannot be nil")
}

// Tests for GetReservationStatus

func TestInventoryService_GetReservationStatus_Success_ActiveReservation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, mockClassifierRepo, mockItemRepo, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	sectionID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()
	operationTypeID := uuid.New()

	// Mock reservation operations
	reservationOps := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  5,
			OperationTypeID: operationTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
	}

	// Setup mocks
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(reservationOps, nil)
	
	// Mock operation type mapping
	operationTypeMapping := map[uuid.UUID]string{
		operationTypeID: models.OperationTypeFactoryReservation,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierOperationType).Return(operationTypeMapping, nil)

	// Mock section mapping
	sectionMapping := map[uuid.UUID]string{
		sectionID: models.SectionFactory,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID: itemID,
		},
		ItemClass: "resources",
		ItemType:  "wood_plank",
	}
	mockItemRepo.On("GetItemWithDetails", ctx, itemID).Return(itemDetails, nil)

	// Mock collection and quality level mappings
	collectionMapping := map[uuid.UUID]string{
		collectionID: "basic",
	}
	qualityMapping := map[uuid.UUID]string{
		qualityLevelID: "common",
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)

	// Mock metrics
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "success").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ReservationExists)
	assert.Equal(t, operationID, *result.OperationID)
	assert.Equal(t, userID, *result.UserID)
	assert.Equal(t, "active", *result.Status)
	assert.Len(t, result.ReservedItems, 1)
	assert.Equal(t, "wood_plank", result.ReservedItems[0].ItemCode)
	assert.Equal(t, "basic", result.ReservedItems[0].CollectionCode)
	assert.Equal(t, "common", result.ReservedItems[0].QualityLevelCode)
	assert.Equal(t, int64(5), result.ReservedItems[0].Quantity)

	mockInventoryRepo.AssertExpectations(t)
	mockClassifierRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_Success_ConsumedReservation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, mockClassifierRepo, mockItemRepo, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	sectionID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()
	reservationTypeID := uuid.New()
	consumptionTypeID := uuid.New()

	// Mock reservation and consumption operations
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  10,
			OperationTypeID: reservationTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  -10,
			OperationTypeID: consumptionTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now().Add(time.Hour),
		},
	}

	// Setup mocks
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(operations, nil)
	
	// Mock operation type mapping
	operationTypeMapping := map[uuid.UUID]string{
		reservationTypeID: models.OperationTypeFactoryReservation,
		consumptionTypeID: models.OperationTypeFactoryConsumption,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierOperationType).Return(operationTypeMapping, nil)

	// Mock section mapping
	sectionMapping := map[uuid.UUID]string{
		sectionID: models.SectionFactory,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID: itemID,
		},
		ItemClass: "resources",
		ItemType:  "stone",
	}
	mockItemRepo.On("GetItemWithDetails", ctx, itemID).Return(itemDetails, nil)

	// Mock collection and quality level mappings
	collectionMapping := map[uuid.UUID]string{
		collectionID: "winter_2025",
	}
	qualityMapping := map[uuid.UUID]string{
		qualityLevelID: "stone",
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)

	// Mock metrics
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "success").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ReservationExists)
	assert.Equal(t, "consumed", *result.Status)

	mockInventoryRepo.AssertExpectations(t)
	mockClassifierRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_Success_ReturnedReservation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, mockClassifierRepo, mockItemRepo, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	sectionID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()
	reservationTypeID := uuid.New()
	returnTypeID := uuid.New()

	// Mock reservation and return operations
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  3,
			OperationTypeID: reservationTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       uuid.New(), // Different section for return
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  3,
			OperationTypeID: returnTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now().Add(time.Hour),
		},
	}

	// Setup mocks
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(operations, nil)
	
	// Mock operation type mapping
	operationTypeMapping := map[uuid.UUID]string{
		reservationTypeID: models.OperationTypeFactoryReservation,
		returnTypeID:      models.OperationTypeFactoryReturn,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierOperationType).Return(operationTypeMapping, nil)

	// Mock section mapping
	sectionMapping := map[uuid.UUID]string{
		sectionID: models.SectionFactory,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID: itemID,
		},
		ItemClass: "resources",
		ItemType:  "ore",
	}
	mockItemRepo.On("GetItemWithDetails", ctx, itemID).Return(itemDetails, nil)

	// Mock collection and quality level mappings
	collectionMapping := map[uuid.UUID]string{
		collectionID: "basic",
	}
	qualityMapping := map[uuid.UUID]string{
		qualityLevelID: "metal",
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)

	// Mock metrics
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "success").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ReservationExists)
	assert.Equal(t, "returned", *result.Status)

	mockInventoryRepo.AssertExpectations(t)
	mockClassifierRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, _, _, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()

	// Setup mocks - no operations found
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return([]*models.Operation{}, nil)
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "not_found").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.ReservationExists)
	assert.Equal(t, "Reservation not found", *result.Error)

	mockInventoryRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, _, _, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()

	// Setup mocks - repository error
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(([]*models.Operation)(nil), assert.AnError)
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "error").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get operations by external ID")

	mockInventoryRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_NoReservationOperations(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, mockClassifierRepo, _, _ := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	sectionID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()
	otherTypeID := uuid.New()

	// Mock operations that are NOT reservation operations
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  5,
			OperationTypeID: otherTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
	}

	// Setup mocks
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(operations, nil)
	
	// Mock operation type mapping - not a reservation operation
	operationTypeMapping := map[uuid.UUID]string{
		otherTypeID: models.OperationTypeChestReward,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierOperationType).Return(operationTypeMapping, nil)

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.ReservationExists)
	assert.Equal(t, "Reservation not found", *result.Error)

	mockInventoryRepo.AssertExpectations(t)
	mockClassifierRepo.AssertExpectations(t)
}

func TestInventoryService_GetReservationStatus_MultipleItems(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, mockInventoryRepo, mockClassifierRepo, mockItemRepo, mockMetrics := createFullTestDeps()
	service := NewInventoryService(deps)

	operationID := uuid.New()
	userID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()
	sectionID := uuid.New()
	collection1ID := uuid.New()
	collection2ID := uuid.New()
	quality1ID := uuid.New()
	quality2ID := uuid.New()
	operationTypeID := uuid.New()

	// Mock multiple reservation operations
	reservationOps := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          item1ID,
			CollectionID:    collection1ID,
			QualityLevelID:  quality1ID,
			QuantityChange:  5,
			OperationTypeID: operationTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          item2ID,
			CollectionID:    collection2ID,
			QualityLevelID:  quality2ID,
			QuantityChange:  3,
			OperationTypeID: operationTypeID,
			OperationID:     &operationID,
			CreatedAt:       time.Now(),
		},
	}

	// Setup mocks
	mockInventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return(reservationOps, nil)
	
	// Mock operation type mapping
	operationTypeMapping := map[uuid.UUID]string{
		operationTypeID: models.OperationTypeFactoryReservation,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierOperationType).Return(operationTypeMapping, nil)

	// Mock section mapping
	sectionMapping := map[uuid.UUID]string{
		sectionID: models.SectionFactory,
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)

	// Mock item details for both items
	item1Details := &models.ItemWithDetails{
		Item: models.Item{
			ID: item1ID,
		},
		ItemClass: "resources",
		ItemType:  "wood_plank",
	}
	item2Details := &models.ItemWithDetails{
		Item: models.Item{
			ID: item2ID,
		},
		ItemClass: "resources",
		ItemType:  "stone",
	}
	mockItemRepo.On("GetItemWithDetails", ctx, item1ID).Return(item1Details, nil)
	mockItemRepo.On("GetItemWithDetails", ctx, item2ID).Return(item2Details, nil)

	// Mock collection and quality level mappings
	collectionMapping := map[uuid.UUID]string{
		collection1ID: "basic",
		collection2ID: "winter_2025",
	}
	qualityMapping := map[uuid.UUID]string{
		quality1ID: "common",
		quality2ID: "stone",
	}
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)
	mockClassifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)

	// Mock metrics
	mockMetrics.On("RecordInventoryOperation", "get_reservation_status", "factory", "success").Return()

	// Act
	result, err := service.GetReservationStatus(ctx, operationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ReservationExists)
	assert.Equal(t, operationID, *result.OperationID)
	assert.Equal(t, userID, *result.UserID)
	assert.Equal(t, "active", *result.Status)
	assert.Len(t, result.ReservedItems, 2)

	// Check that both items are present (order may vary)
	itemCodes := make(map[string]int64)
	for _, item := range result.ReservedItems {
		itemCodes[item.ItemCode] = item.Quantity
	}
	assert.Equal(t, int64(5), itemCodes["wood_plank"])
	assert.Equal(t, int64(3), itemCodes["stone"])

	mockInventoryRepo.AssertExpectations(t)
	mockClassifierRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

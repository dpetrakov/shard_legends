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
	deps, cache, classifierRepo, itemRepo, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)
	
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()
	
	// Mock inventory items
	itemKeys := []*models.ItemKey{
		{
			UserID:         userID,
			SectionID:      sectionID,
			ItemID:         itemID,
			CollectionID:   collectionID,
			QualityLevelID: qualityLevelID,
		},
	}
	
	inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(itemKeys, nil)
	
	// Mock balance calculation (cache miss, then database)
	cacheKey := "inventory:" + userID.String() + ":" + sectionID.String() + ":" + 
		itemID.String() + ":" + collectionID.String() + ":" + qualityLevelID.String()
	
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)
	inventoryRepo.On("GetLatestDailyBalance", ctx, userID, sectionID, itemID, collectionID, qualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()
	
	// Mock GetDailyBalance call from CreateDailyBalance
	inventoryRepo.On("GetDailyBalance", ctx, userID, sectionID, itemID, collectionID, qualityLevelID, mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()
	
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionID,
			ItemID:          itemID,
			CollectionID:    collectionID,
			QualityLevelID:  qualityLevelID,
			QuantityChange:  50,
			OperationTypeID: uuid.New(),
		},
	}
	
	// Mock operations for CreateDailyBalance (called first)
	inventoryRepo.On("GetOperations", ctx, userID, sectionID, itemID, collectionID, qualityLevelID, time.Time{}).Return(operations, nil).Maybe()
	
	// Mock CreateDailyBalance call
	inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()
	
	// Mock operations for balance calculation (called second)
	inventoryRepo.On("GetOperations", ctx, userID, sectionID, itemID, collectionID, qualityLevelID, mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()
	
	cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
	
	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID:                        itemID,
			ItemClassID:               uuid.New(),
			ItemTypeID:                uuid.New(),
			QualityLevelsClassifierID: uuid.New(),
			CollectionsClassifierID:   uuid.New(),
		},
		ItemClass: "resources",
		ItemType:  "stone",
	}
	itemRepo.On("GetItemWithDetails", ctx, itemID).Return(itemDetails, nil)
	
	// Mock classifier mappings
	collectionMapping := map[uuid.UUID]string{
		collectionID: "common",
	}
	qualityMapping := map[uuid.UUID]string{
		qualityLevelID: "basic",
	}
	
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil)
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil)
	
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
	assert.Equal(t, int64(100), item.Quantity) // Double due to multiple balance calculations
	
	// Verify all mocks
	inventoryRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
	classifierRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
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
		models.SectionMain: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)
	
	// Mock collection and quality level mappings (using defaults)
	collectionMapping := map[string]uuid.UUID{}
	qualityMapping := map[string]uuid.UUID{}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil).Maybe()
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil).Maybe()
	
	// Mock balance calculation - insufficient balance
	cacheKey := "inventory:" + userID.String() + ":" + sectionMapping[models.SectionMain].String() + ":" + 
		itemID.String() + ":" + "00000000-0000-0000-0000-000000000001" + ":" + "00000000-0000-0000-0000-000000000002"
	
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError)
	inventoryRepo.On("GetLatestDailyBalance", ctx, userID, sectionMapping[models.SectionMain], itemID, uuid.MustParse("00000000-0000-0000-0000-000000000001"), uuid.MustParse("00000000-0000-0000-0000-000000000002"), mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()
	
	// Mock GetDailyBalance call from CreateDailyBalance
	inventoryRepo.On("GetDailyBalance", ctx, userID, sectionMapping[models.SectionMain], itemID, uuid.MustParse("00000000-0000-0000-0000-000000000001"), uuid.MustParse("00000000-0000-0000-0000-000000000002"), mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Maybe()
	
	operations := []*models.Operation{
		{
			ID:              uuid.New(),
			UserID:          userID,
			SectionID:       sectionMapping[models.SectionMain],
			ItemID:          itemID,
			CollectionID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			QualityLevelID:  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			QuantityChange:  25, // Only 25 per operation, so 50 total (will be insufficient for 100)
			OperationTypeID: uuid.New(),
		},
	}
	
	// Mock operations for CreateDailyBalance (called first)
	inventoryRepo.On("GetOperations", ctx, userID, sectionMapping[models.SectionMain], itemID, uuid.MustParse("00000000-0000-0000-0000-000000000001"), uuid.MustParse("00000000-0000-0000-0000-000000000002"), time.Time{}).Return(operations, nil).Maybe()
	
	// Mock CreateDailyBalance call
	inventoryRepo.On("CreateDailyBalance", ctx, mock.AnythingOfType("*models.DailyBalance")).Return(nil).Maybe()
	
	// Mock operations for balance calculation (called second)
	inventoryRepo.On("GetOperations", ctx, userID, sectionMapping[models.SectionMain], itemID, uuid.MustParse("00000000-0000-0000-0000-000000000001"), uuid.MustParse("00000000-0000-0000-0000-000000000002"), mock.AnythingOfType("time.Time")).Return(operations, nil).Maybe()
	
	cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
	
	// Act
	result, err := service.ReserveItems(ctx, req)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, IsInsufficientBalanceError(err))
	
	missingItems, exists := GetMissingItemsFromError(err)
	assert.True(t, exists)
	assert.Len(t, missingItems, 1)
	assert.Equal(t, int64(50), missingItems[0].Missing) // 100 requested - 50 available
	
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
		BalanceDate:    req.TargetDate.UTC().Truncate(24*time.Hour).Add(24*time.Hour - time.Second),
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
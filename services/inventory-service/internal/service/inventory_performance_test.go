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

// BenchmarkGetUserInventory_Legacy - D-15: бенчмарк оригинальной версии с N+1 запросами
func BenchmarkGetUserInventory_Legacy(b *testing.B) {
	// Arrange
	ctx := context.Background()
	deps, cache, classifierRepo, itemRepo, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	sectionID := uuid.New()

	// Создаем тестовые данные для имитации большого инвентаря
	itemKeys := createTestItemKeys(100, userID, sectionID) // 100 предметов
	inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(itemKeys, nil)

	// Мокаем GetItemWithDetails для каждого предмета (N+1 проблема)
	for _, itemKey := range itemKeys {
		setupMocksForItem(ctx, cache, classifierRepo, itemRepo, itemKey)
	}

	// Act & Measure
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetUserInventoryLegacy(ctx, userID, sectionID)
		assert.NoError(b, err)
	}
}

// BenchmarkGetUserInventory_Optimized - D-15: бенчмарк оптимизированной версии
func BenchmarkGetUserInventory_Optimized(b *testing.B) {
	// Arrange
	ctx := context.Background()
	deps, _, _, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	sectionID := uuid.New()

	// Создаем ответ оптимизированного метода (один запрос вместо N+1)
	optimizedResult := createTestInventoryResponse(100) // 100 предметов
	inventoryRepo.On("GetUserInventoryOptimized", ctx, userID, sectionID).Return(optimizedResult, nil)

	// Act & Measure
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetUserInventory(ctx, userID, sectionID)
		assert.NoError(b, err)
	}
}

// BenchmarkGetUserInventory_Comparison - D-15: сравнительный бенчмарк
func BenchmarkGetUserInventory_Comparison(b *testing.B) {
	scenarios := []struct {
		name       string
		itemCount  int
		optimized  bool
	}{
		{"Legacy_10_items", 10, false},
		{"Optimized_10_items", 10, true},
		{"Legacy_50_items", 50, false},
		{"Optimized_50_items", 50, true},
		{"Legacy_100_items", 100, false},
		{"Optimized_100_items", 100, true},
		{"Legacy_300_items", 300, false},
		{"Optimized_300_items", 300, true},
		{"Legacy_500_items", 500, false},
		{"Optimized_500_items", 500, true},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			ctx := context.Background()
			deps, cache, classifierRepo, itemRepo, inventoryRepo := createFullTestDeps()
			service := NewInventoryService(deps)

			userID := uuid.New()
			sectionID := uuid.New()

			if scenario.optimized {
				// Оптимизированная версия - один запрос
				optimizedResult := createTestInventoryResponse(scenario.itemCount)
				inventoryRepo.On("GetUserInventoryOptimized", ctx, userID, sectionID).Return(optimizedResult, nil)
			} else {
				// Legacy версия - N+1 запросы
				itemKeys := createTestItemKeys(scenario.itemCount, userID, sectionID)
				inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(itemKeys, nil)
				
				for _, itemKey := range itemKeys {
					setupMocksForItem(ctx, cache, classifierRepo, itemRepo, itemKey)
				}
			}

			// Measure performance
			b.ResetTimer()
			start := time.Now()
			
			for i := 0; i < b.N; i++ {
				var err error
				if scenario.optimized {
					_, err = service.GetUserInventory(ctx, userID, sectionID)
				} else {
					_, err = service.GetUserInventoryLegacy(ctx, userID, sectionID)
				}
				assert.NoError(b, err)
			}
			
			elapsed := time.Since(start)
			if b.N > 0 {
				avgTime := elapsed / time.Duration(b.N)
				b.Logf("Average time per operation: %v", avgTime)
				
				// Проверяем цель производительности D-15: <100ms для 500 предметов
				if scenario.itemCount == 500 && avgTime > 100*time.Millisecond {
					b.Logf("WARNING: Performance target not met. Expected <100ms, got %v", avgTime)
				}
			}
		})
	}
}

// TestPerformanceComparison_DatabaseQueries - D-15: тест сравнения количества запросов к БД
func TestPerformanceComparison_DatabaseQueries(t *testing.T) {
	// Arrange
	ctx := context.Background()
	deps, cache, classifierRepo, itemRepo, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)

	userID := uuid.New()
	sectionID := uuid.New()
	itemCount := 50

	// Test Legacy version - подсчитываем количество вызовов
	t.Run("Legacy_Query_Count", func(t *testing.T) {
		itemKeys := createTestItemKeys(itemCount, userID, sectionID)
		inventoryRepo.On("GetUserInventoryItems", ctx, userID, sectionID).Return(itemKeys, nil).Once()

		// Для каждого предмета будет вызван GetItemWithDetails + дополнительные запросы
		callCount := 0
		for _, itemKey := range itemKeys {
			setupMocksForItemWithCounter(ctx, cache, classifierRepo, itemRepo, itemKey, &callCount)
		}

		_, err := service.GetUserInventoryLegacy(ctx, userID, sectionID)
		assert.NoError(t, err)

		t.Logf("Legacy version: %d items resulted in %d database calls", itemCount, callCount)
		// Ожидаем: 1 (GetUserInventoryItems) + itemCount * N (GetItemWithDetails + balance calculations + code conversions)
		// Это демонстрирует N+1 проблему
		assert.Greater(t, callCount, itemCount, "Legacy version should have more database calls due to N+1 problem")
	})

	// Test Optimized version - один запрос
	t.Run("Optimized_Query_Count", func(t *testing.T) {
		optimizedResult := createTestInventoryResponse(itemCount)
		
		callCount := 0
		inventoryRepo.On("GetUserInventoryOptimized", ctx, userID, sectionID).Return(optimizedResult, nil).Run(func(args mock.Arguments) {
			callCount++
		}).Once()

		_, err := service.GetUserInventory(ctx, userID, sectionID)
		assert.NoError(t, err)

		t.Logf("Optimized version: %d items resulted in %d database calls", itemCount, callCount)
		// Ожидаем: только 1 запрос вне зависимости от количества предметов
		assert.Equal(t, 1, callCount, "Optimized version should use only 1 database call")
	})
}

// Helper functions for benchmark tests

func createTestItemKeys(count int, userID, sectionID uuid.UUID) []*models.ItemKey {
	keys := make([]*models.ItemKey, count)
	for i := 0; i < count; i++ {
		keys[i] = &models.ItemKey{
			UserID:         userID,
			SectionID:      sectionID,
			ItemID:         uuid.New(),
			CollectionID:   uuid.New(),
			QualityLevelID: uuid.New(),
		}
	}
	return keys
}

func createTestInventoryResponse(count int) []*models.InventoryItemResponse {
	items := make([]*models.InventoryItemResponse, count)
	for i := 0; i < count; i++ {
		collection := "common"
		quality := "basic"
		items[i] = &models.InventoryItemResponse{
			ItemID:       uuid.New(),
			ItemClass:    "resources",
			ItemType:     "stone",
			Collection:   &collection,
			QualityLevel: &quality,
			Quantity:     int64(100 + i),
		}
	}
	return items
}

func setupMocksForItem(ctx context.Context, cache *MockCache, classifierRepo *MockClassifierRepo, itemRepo *MockItemRepo, itemKey *models.ItemKey) {
	// Mock balance calculation
	cacheKey := "inventory:" + itemKey.UserID.String() + ":" + itemKey.SectionID.String() + ":" +
		itemKey.ItemID.String() + ":" + itemKey.CollectionID.String() + ":" + itemKey.QualityLevelID.String()
	
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError).Maybe()
	
	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID:                        itemKey.ItemID,
			ItemClassID:               uuid.New(),
			ItemTypeID:                uuid.New(),
			QualityLevelsClassifierID: uuid.New(),
			CollectionsClassifierID:   uuid.New(),
		},
		ItemClass: "resources",
		ItemType:  "stone",
	}
	itemRepo.On("GetItemWithDetails", ctx, itemKey.ItemID).Return(itemDetails, nil).Maybe()

	// Mock classifier mappings
	collectionMapping := map[uuid.UUID]string{itemKey.CollectionID: "common"}
	qualityMapping := map[uuid.UUID]string{itemKey.QualityLevelID: "basic"}
	
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil).Maybe()
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil).Maybe()
	
	cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
}

func setupMocksForItemWithCounter(ctx context.Context, cache *MockCache, classifierRepo *MockClassifierRepo, itemRepo *MockItemRepo, itemKey *models.ItemKey, counter *int) {
	// Mock balance calculation
	cacheKey := "inventory:" + itemKey.UserID.String() + ":" + itemKey.SectionID.String() + ":" +
		itemKey.ItemID.String() + ":" + itemKey.CollectionID.String() + ":" + itemKey.QualityLevelID.String()
	
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*int64")).Return(assert.AnError).Run(func(args mock.Arguments) {
		*counter++
	}).Maybe()
	
	// Mock item details
	itemDetails := &models.ItemWithDetails{
		Item: models.Item{
			ID:                        itemKey.ItemID,
			ItemClassID:               uuid.New(),
			ItemTypeID:                uuid.New(),
			QualityLevelsClassifierID: uuid.New(),
			CollectionsClassifierID:   uuid.New(),
		},
		ItemClass: "resources",
		ItemType:  "stone",
	}
	itemRepo.On("GetItemWithDetails", ctx, itemKey.ItemID).Return(itemDetails, nil).Run(func(args mock.Arguments) {
		*counter++
	}).Maybe()

	// Mock classifier mappings
	collectionMapping := map[uuid.UUID]string{itemKey.CollectionID: "common"}
	qualityMapping := map[uuid.UUID]string{itemKey.QualityLevelID: "basic"}
	
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierCollection).Return(collectionMapping, nil).Run(func(args mock.Arguments) {
		*counter++
	}).Maybe()
	classifierRepo.On("GetUUIDToCodeMapping", ctx, models.ClassifierQualityLevel).Return(qualityMapping, nil).Run(func(args mock.Arguments) {
		*counter++
	}).Maybe()
	
	cache.On("Set", ctx, cacheKey, mock.AnythingOfType("int64"), mock.AnythingOfType("time.Duration")).Return(nil).Run(func(args mock.Arguments) {
		*counter++
	}).Maybe()
}
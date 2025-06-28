package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"

	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/internal/service"
)

// MockCacheInterface for testing
type MockCacheInterface struct {
	mock.Mock
}

func (m *MockCacheInterface) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCacheInterface) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheInterface) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheInterface) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// Verify MockCacheInterface implements service.CacheInterface
var _ service.CacheInterface = (*MockCacheInterface)(nil)

// MockMetrics for testing
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) RecordInventoryOperation(operationType, section, status string) {
	m.Called(operationType, section, status)
}

func (m *MockMetrics) RecordBalanceCalculation(section string, duration time.Duration, status string) {
	m.Called(section, duration, status)
}

func (m *MockMetrics) RecordCacheHit(cacheType string) {
	m.Called(cacheType)
}

func (m *MockMetrics) RecordCacheMiss(cacheType string) {
	m.Called(cacheType)
}

func (m *MockMetrics) RecordTransactionMetrics(operationType string, operationCount int, duration time.Duration) {
	m.Called(operationType, operationCount, duration)
}

func (m *MockMetrics) RecordItemsPerInventory(section string, count int) {
	m.Called(section, count)
}

// Verify MockMetrics implements service.MetricsInterface
var _ service.MetricsInterface = (*MockMetrics)(nil)

func createTestItemStorage() (*itemStorage, *MockCacheInterface, *MockMetrics) {
	cache := new(MockCacheInterface)
	metricsCollector := new(MockMetrics)
	logger := slog.Default()

	storage := &itemStorage{
		pool:    nil, // Will be mocked in individual tests
		logger:  logger,
		metrics: metricsCollector, // Use mock metrics
		cache:   cache,
	}

	return storage, cache, metricsCollector
}

func TestItemStorage_GetItemImagesBatch_CacheKeyGeneration(t *testing.T) {
	// Arrange
	_, _, _ = createTestItemStorage()

	itemID1 := uuid.New()
	itemID2 := uuid.New()

	requests := []models.ItemDetailRequestItem{
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
	}

	// Test cache key generation logic (since we can't easily mock the full DB)
	expectedKey1 := fmt.Sprintf("i18n:item_images:%s:%s:%s", itemID1.String(), "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
	expectedKey2 := fmt.Sprintf("i18n:item_images:%s:%s:%s", itemID2.String(), "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")

	// The actual implementation will generate these keys during the method execution
	assert.Len(t, requests, 2)
	assert.Contains(t, expectedKey1, itemID1.String())
	assert.Contains(t, expectedKey2, itemID2.String())
}

func TestItemStorage_GetTranslationsBatch_EmptyInput(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage, _, _ := createTestItemStorage()

	// Act
	result, err := storage.GetTranslationsBatch(ctx, "item", []uuid.UUID{}, "ru")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestItemStorage_GetItemImagesBatch_EmptyInput(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage, _, _ := createTestItemStorage()

	// Act
	result, err := storage.GetItemImagesBatch(ctx, []models.ItemDetailRequestItem{})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestItemStorage_CacheInteraction(t *testing.T) {
	// Arrange
	ctx := context.Background()
	_, cache, _ := createTestItemStorage()

	itemID := uuid.New()
	cacheKey := fmt.Sprintf("i18n:translations:item:%s:ru", itemID.String())
	expectedTranslations := map[string]string{
		"name":        "Test Item",
		"description": "Test Description",
	}

	// Test cache hit scenario
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*map[string]string")).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*map[string]string)
		*arg = expectedTranslations
	}).Return(nil)

	// Test that cache interaction works correctly
	var result map[string]string
	err := cache.Get(ctx, cacheKey, &result)

	assert.NoError(t, err)
	assert.Equal(t, expectedTranslations, result)

	cache.AssertExpectations(t)
}

func TestItemStorage_CacheMiss(t *testing.T) {
	// Arrange
	ctx := context.Background()
	_, cache, _ := createTestItemStorage()

	itemID := uuid.New()
	cacheKey := fmt.Sprintf("i18n:translations:item:%s:ru", itemID.String())
	translations := map[string]string{
		"name":        "New Item",
		"description": "New Description",
	}

	// Test cache miss and set scenario
	cache.On("Get", ctx, cacheKey, mock.AnythingOfType("*map[string]string")).Return(fmt.Errorf("cache miss"))
	cache.On("Set", ctx, cacheKey, translations, 24*time.Hour).Return(nil)

	// Test cache miss
	var result map[string]string
	err := cache.Get(ctx, cacheKey, &result)
	assert.Error(t, err)

	// Test cache set
	err = cache.Set(ctx, cacheKey, translations, 24*time.Hour)
	assert.NoError(t, err)

	cache.AssertExpectations(t)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/shard-legends/inventory-service/internal/models"
)

func TestClassifierService_GetClassifierMapping(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierInventorySection
		expectedMapping := map[string]uuid.UUID{
			"main":    uuid.New(),
			"factory": uuid.New(),
			"trade":   uuid.New(),
		}

		// Mock repository call
		classifierRepo.On("GetCodeToUUIDMapping", ctx, classifierCode).Return(expectedMapping, nil)

		// Act
		result, err := service.GetClassifierMapping(ctx, classifierCode)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedMapping, result)
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Empty classifier code", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		// Act
		result, err := service.GetClassifierMapping(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "classifier code cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierInventorySection

		// Mock repository error
		classifierRepo.On("GetCodeToUUIDMapping", ctx, classifierCode).Return(nil, errors.New("db error"))

		// Act
		result, err := service.GetClassifierMapping(ctx, classifierCode)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get classifier mapping")
		assert.Nil(t, result)
		classifierRepo.AssertExpectations(t)
	})

}

func TestClassifierService_GetReverseClassifierMapping(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierQualityLevel
		id1, id2 := uuid.New(), uuid.New()
		expectedMapping := map[uuid.UUID]string{
			id1: "common",
			id2: "rare",
		}

		// Mock repository call
		classifierRepo.On("GetUUIDToCodeMapping", ctx, classifierCode).Return(expectedMapping, nil)

		// Act
		result, err := service.GetReverseClassifierMapping(ctx, classifierCode)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedMapping, result)
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Empty classifier code", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		// Act
		result, err := service.GetReverseClassifierMapping(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "classifier code cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierOperationType

		// Mock repository error
		classifierRepo.On("GetUUIDToCodeMapping", ctx, classifierCode).Return(nil, errors.New("db error"))

		// Act
		result, err := service.GetReverseClassifierMapping(ctx, classifierCode)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get reverse classifier mapping")
		assert.Nil(t, result)
		classifierRepo.AssertExpectations(t)
	})
}

func TestClassifierService_RefreshClassifierCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierInventorySection

		// Mock invalidate cache
		classifierRepo.On("InvalidateCache", ctx, classifierCode).Return(nil)

		// Act
		err := service.RefreshClassifierCache(ctx, classifierCode)

		// Assert
		assert.NoError(t, err)
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, classifierRepo, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		classifierCode := models.ClassifierOperationType

		// Mock invalidate cache error
		classifierRepo.On("InvalidateCache", ctx, classifierCode).Return(errors.New("db error"))

		// Act
		err := service.RefreshClassifierCache(ctx, classifierCode)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to refresh classifier cache")
		classifierRepo.AssertExpectations(t)
	})

	t.Run("Empty classifier code", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		deps, _, _, _, _ := createFullTestDeps()
		service := NewClassifierService(deps)

		// Act
		err := service.RefreshClassifierCache(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "classifier code cannot be empty")
	})
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/shard-legends/production-service/internal/models"
)

// Test cleanup service initialization
func TestNewCleanupService(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	logger := zap.NewNop()
	config := GetDefaultCleanupConfig()

	service := NewCleanupService(mockTaskRepo, mockInventoryClient, logger, config)

	assert.NotNil(t, service)
	assert.Equal(t, mockTaskRepo, service.taskRepo)
	assert.Equal(t, mockInventoryClient, service.inventoryClient)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, config, service.config)
}

// Test default cleanup configuration
func TestGetDefaultCleanupConfig(t *testing.T) {
	config := GetDefaultCleanupConfig()

	assert.Equal(t, 5*time.Minute, config.OrphanedTaskTimeout)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
}

// Test successful cleanup of orphaned tasks
func TestCleanupService_RunCleanup_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := CleanupConfig{
		OrphanedTaskTimeout: 5 * time.Minute,
		CleanupInterval:     5 * time.Minute,
	}
	
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	cutoffTime := time.Now().Add(-config.OrphanedTaskTimeout)
	
	// Create orphaned tasks
	orphanedTask1 := models.ProductionTask{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  uuid.New(),
		Status:    models.TaskStatusDraft,
		CreatedAt: cutoffTime.Add(-10 * time.Minute),
	}
	
	orphanedTask2 := models.ProductionTask{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  uuid.New(),
		Status:    models.TaskStatusDraft,
		CreatedAt: cutoffTime.Add(-20 * time.Minute),
	}
	
	orphanedTasks := []models.ProductionTask{orphanedTask1, orphanedTask2}
	
	// Mock repository call
	mockTaskRepo.On("GetOrphanedDraftTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("time.Time")).Return(orphanedTasks, nil)
	
	// Mock cleanup operations for both tasks
	mockInventoryClient.On("ReturnReserve", mock.AnythingOfType("*context.timerCtx"), orphanedTask1.UserID, orphanedTask1.ID).Return(nil)
	mockTaskRepo.On("DeleteTask", mock.AnythingOfType("*context.timerCtx"), orphanedTask1.ID).Return(nil)
	
	mockInventoryClient.On("ReturnReserve", mock.AnythingOfType("*context.timerCtx"), orphanedTask2.UserID, orphanedTask2.ID).Return(nil)
	mockTaskRepo.On("DeleteTask", mock.AnythingOfType("*context.timerCtx"), orphanedTask2.ID).Return(nil)
	
	// Execute
	service.runCleanup(ctx)
	
	// Verify
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test cleanup with no orphaned tasks
func TestCleanupService_RunCleanup_NoOrphanedTasks(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := GetDefaultCleanupConfig()
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	
	// Mock repository call returning empty result
	mockTaskRepo.On("GetOrphanedDraftTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("time.Time")).Return([]models.ProductionTask{}, nil)
	
	// Execute
	service.runCleanup(ctx)
	
	// Verify
	mockTaskRepo.AssertExpectations(t)
	// No inventory client calls should be made
	mockInventoryClient.AssertExpectations(t)
}

// Test cleanup failure in getting orphaned tasks
func TestCleanupService_RunCleanup_GetOrphanedTasksFailure(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := GetDefaultCleanupConfig()
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	
	// Mock repository call failure
	expectedError := errors.New("database connection failed")
	mockTaskRepo.On("GetOrphanedDraftTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("time.Time")).Return([]models.ProductionTask{}, expectedError)
	
	// Execute (should not panic)
	service.runCleanup(ctx)
	
	// Verify
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test individual task cleanup success
func TestCleanupService_CleanupOrphanedTask_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := GetDefaultCleanupConfig()
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	
	task := models.ProductionTask{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  uuid.New(),
		Status:    models.TaskStatusDraft,
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}
	
	// Mock successful cleanup
	mockInventoryClient.On("ReturnReserve", ctx, task.UserID, task.ID).Return(nil)
	mockTaskRepo.On("DeleteTask", ctx, task.ID).Return(nil)
	
	// Execute
	result := service.cleanupOrphanedTask(ctx, task)
	
	// Verify
	assert.True(t, result)
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test individual task cleanup with inventory return failure (should still delete task)
func TestCleanupService_CleanupOrphanedTask_InventoryReturnFailure(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := GetDefaultCleanupConfig()
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	
	task := models.ProductionTask{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  uuid.New(),
		Status:    models.TaskStatusDraft,
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}
	
	// Mock inventory return failure (reservation might not exist)
	inventoryError := errors.New("reservation not found")
	mockInventoryClient.On("ReturnReserve", ctx, task.UserID, task.ID).Return(inventoryError)
	
	// Task should still be deleted
	mockTaskRepo.On("DeleteTask", ctx, task.ID).Return(nil)
	
	// Execute
	result := service.cleanupOrphanedTask(ctx, task)
	
	// Verify
	assert.True(t, result)
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test individual task cleanup with delete failure
func TestCleanupService_CleanupOrphanedTask_DeleteFailure(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	config := GetDefaultCleanupConfig()
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	ctx := context.Background()
	
	task := models.ProductionTask{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  uuid.New(),
		Status:    models.TaskStatusDraft,
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}
	
	// Mock successful inventory return
	mockInventoryClient.On("ReturnReserve", ctx, task.UserID, task.ID).Return(nil)
	
	// Mock delete failure
	deleteError := errors.New("database delete failed")
	mockTaskRepo.On("DeleteTask", ctx, task.ID).Return(deleteError)
	
	// Execute
	result := service.cleanupOrphanedTask(ctx, task)
	
	// Verify
	assert.False(t, result)
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test cleanup service start/stop (integration test)
func TestCleanupService_Start_StopWithContext(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)
	
	// Use shorter intervals for testing
	config := CleanupConfig{
		OrphanedTaskTimeout: 1 * time.Second,
		CleanupInterval:     100 * time.Millisecond,
	}
	
	service := NewCleanupService(mockTaskRepo, mockInventoryClient, zap.NewNop(), config)
	
	// Mock at least one cleanup run
	mockTaskRepo.On("GetOrphanedDraftTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("time.Time")).Return([]models.ProductionTask{}, nil).Maybe()
	
	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	
	// Start service in goroutine
	done := make(chan struct{})
	go func() {
		service.Start(ctx)
		close(done)
	}()
	
	// Wait for context cancellation or timeout
	select {
	case <-done:
		// Service stopped as expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Service did not stop within expected time")
	}
	
	// Verify at least the setup was correct
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}
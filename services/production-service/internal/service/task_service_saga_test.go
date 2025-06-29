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

// Mock implementations for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, task *models.ProductionTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.ProductionTask, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*models.ProductionTask), args.Error(1)
}

func (m *MockTaskRepository) GetUserTasks(ctx context.Context, userID uuid.UUID, statuses []string) ([]models.ProductionTask, error) {
	args := m.Called(ctx, userID, statuses)
	return args.Get(0).([]models.ProductionTask), args.Error(1)
}

func (m *MockTaskRepository) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskRepository) GetOrphanedDraftTasks(ctx context.Context, olderThan time.Time) ([]models.ProductionTask, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).([]models.ProductionTask), args.Error(1)
}

func (m *MockTaskRepository) GetTasksStats(ctx context.Context, filters map[string]interface{}, page, limit int) (*models.TasksStatsResponse, error) {
	args := m.Called(ctx, filters, page, limit)
	return args.Get(0).(*models.TasksStatsResponse), args.Error(1)
}

func (m *MockTaskRepository) UpdateTaskSlotNumber(ctx context.Context, taskID uuid.UUID, slotNumber int) error {
	args := m.Called(ctx, taskID, slotNumber)
	return args.Error(0)
}

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) ReserveItems(ctx context.Context, userID uuid.UUID, operationID uuid.UUID, items []models.ReservationItem) error {
	args := m.Called(ctx, userID, operationID, items)
	return args.Error(0)
}

func (m *MockInventoryClient) ReturnReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error {
	args := m.Called(ctx, userID, operationID)
	return args.Error(0)
}

func (m *MockInventoryClient) ConsumeReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error {
	args := m.Called(ctx, userID, operationID)
	return args.Error(0)
}

func (m *MockInventoryClient) AddItems(ctx context.Context, userID uuid.UUID, section string, operationType string, operationID uuid.UUID, items []models.AddItem) error {
	args := m.Called(ctx, userID, section, operationType, operationID, items)
	return args.Error(0)
}

// Test createTaskWithReservation success scenario (Saga pattern)
func TestTaskService_CreateTaskWithReservation_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)

	// Create task service instance
	taskService := &TaskService{
		taskRepo:        mockTaskRepo,
		inventoryClient: mockInventoryClient,
		logger:          zap.NewNop(),
	}

	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &models.ProductionTask{
		ID:     taskID,
		UserID: userID,
	}

	itemsToReserve := []models.ReservationItem{
		{ItemID: uuid.New(), Quantity: 5},
	}

	// Phase 1: Create draft task
	mockTaskRepo.On("CreateTask", ctx, task).Return(nil).Run(func(args mock.Arguments) {
		// Verify task status is set to DRAFT
		createdTask := args.Get(1).(*models.ProductionTask)
		assert.Equal(t, models.TaskStatusDraft, createdTask.Status)
	})

	// Phase 2: Reserve inventory
	mockInventoryClient.On("ReserveItems", ctx, userID, taskID, itemsToReserve).Return(nil)

	// Phase 3: Confirm task (update to PENDING)
	mockTaskRepo.On("UpdateTaskStatus", ctx, taskID, models.TaskStatusPending).Return(nil)

	// Execute
	err := taskService.createTaskWithReservation(ctx, task, itemsToReserve)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, models.TaskStatusDraft, task.Status) // Status should be set during execution
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test createTaskWithReservation failure in Phase 1 (draft creation)
func TestTaskService_CreateTaskWithReservation_FailurePhase1(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)

	taskService := &TaskService{
		taskRepo:        mockTaskRepo,
		inventoryClient: mockInventoryClient,
		logger:          zap.NewNop(),
	}

	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &models.ProductionTask{
		ID:     taskID,
		UserID: userID,
	}

	itemsToReserve := []models.ReservationItem{
		{ItemID: uuid.New(), Quantity: 5},
	}

	// Phase 1: Create draft task fails
	expectedError := errors.New("database error")
	mockTaskRepo.On("CreateTask", ctx, task).Return(expectedError)

	// No other phases should be called

	// Execute
	err := taskService.createTaskWithReservation(ctx, task, itemsToReserve)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create draft task")
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test createTaskWithReservation failure in Phase 2 (inventory reservation)
func TestTaskService_CreateTaskWithReservation_FailurePhase2(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)

	taskService := &TaskService{
		taskRepo:        mockTaskRepo,
		inventoryClient: mockInventoryClient,
		logger:          zap.NewNop(),
	}

	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &models.ProductionTask{
		ID:     taskID,
		UserID: userID,
	}

	itemsToReserve := []models.ReservationItem{
		{ItemID: uuid.New(), Quantity: 5},
	}

	// Phase 1: Create draft task succeeds
	mockTaskRepo.On("CreateTask", ctx, task).Return(nil)

	// Phase 2: Reserve inventory fails
	reservationError := errors.New("insufficient items")
	mockInventoryClient.On("ReserveItems", ctx, userID, taskID, itemsToReserve).Return(reservationError)

	// Compensation: Delete draft task
	mockTaskRepo.On("DeleteTask", ctx, taskID).Return(nil)

	// Execute
	err := taskService.createTaskWithReservation(ctx, task, itemsToReserve)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reserve inventory")
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test createTaskWithReservation failure in Phase 3 (task confirmation)
func TestTaskService_CreateTaskWithReservation_FailurePhase3(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)

	taskService := &TaskService{
		taskRepo:        mockTaskRepo,
		inventoryClient: mockInventoryClient,
		logger:          zap.NewNop(),
	}

	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &models.ProductionTask{
		ID:     taskID,
		UserID: userID,
	}

	itemsToReserve := []models.ReservationItem{
		{ItemID: uuid.New(), Quantity: 5},
	}

	// Phase 1: Create draft task succeeds
	mockTaskRepo.On("CreateTask", ctx, task).Return(nil)

	// Phase 2: Reserve inventory succeeds
	mockInventoryClient.On("ReserveItems", ctx, userID, taskID, itemsToReserve).Return(nil)

	// Phase 3: Update status fails
	confirmationError := errors.New("database update failed")
	mockTaskRepo.On("UpdateTaskStatus", ctx, taskID, models.TaskStatusPending).Return(confirmationError)

	// Compensation: Return reservation and delete task
	mockInventoryClient.On("ReturnReserve", ctx, userID, taskID).Return(nil)
	mockTaskRepo.On("DeleteTask", ctx, taskID).Return(nil)

	// Execute
	err := taskService.createTaskWithReservation(ctx, task, itemsToReserve)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to confirm task")
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Test compensation failure scenarios (logging should occur but not block)
func TestTaskService_CreateTaskWithReservation_CompensationFailures(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockInventoryClient := new(MockInventoryClient)

	taskService := &TaskService{
		taskRepo:        mockTaskRepo,
		inventoryClient: mockInventoryClient,
		logger:          zap.NewNop(),
	}

	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &models.ProductionTask{
		ID:     taskID,
		UserID: userID,
	}

	itemsToReserve := []models.ReservationItem{
		{ItemID: uuid.New(), Quantity: 5},
	}

	// Phase 1: Create draft task succeeds
	mockTaskRepo.On("CreateTask", ctx, task).Return(nil)

	// Phase 2: Reserve inventory succeeds
	mockInventoryClient.On("ReserveItems", ctx, userID, taskID, itemsToReserve).Return(nil)

	// Phase 3: Update status fails
	confirmationError := errors.New("database update failed")
	mockTaskRepo.On("UpdateTaskStatus", ctx, taskID, models.TaskStatusPending).Return(confirmationError)

	// Compensation: Return reservation fails but delete succeeds
	returnError := errors.New("return reservation failed")
	mockInventoryClient.On("ReturnReserve", ctx, userID, taskID).Return(returnError)
	mockTaskRepo.On("DeleteTask", ctx, taskID).Return(nil)

	// Execute
	err := taskService.createTaskWithReservation(ctx, task, itemsToReserve)

	// Verify - main error should still be returned despite compensation failure
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to confirm task")
	mockTaskRepo.AssertExpectations(t)
	mockInventoryClient.AssertExpectations(t)
}

// Mock recipe repository for testing
type MockRecipeRepository struct {
	mock.Mock
}

func (m *MockRecipeRepository) GetActiveRecipes(ctx context.Context, filters *models.RecipeFilters) ([]models.ProductionRecipe, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.ProductionRecipe), args.Error(1)
}

func (m *MockRecipeRepository) GetRecipeByID(ctx context.Context, recipeID uuid.UUID) (*models.ProductionRecipe, error) {
	args := m.Called(ctx, recipeID)
	return args.Get(0).(*models.ProductionRecipe), args.Error(1)
}

func (m *MockRecipeRepository) GetRecipeLimits(ctx context.Context, recipeID uuid.UUID) ([]models.RecipeLimit, error) {
	args := m.Called(ctx, recipeID)
	return args.Get(0).([]models.RecipeLimit), args.Error(1)
}

func (m *MockRecipeRepository) GetRecipeUsageStats(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, limitType string, periodStart, periodEnd time.Time) (int, error) {
	args := m.Called(ctx, userID, recipeID, limitType, periodStart, periodEnd)
	return args.Int(0), args.Error(1)
}

func (m *MockRecipeRepository) CheckRecipeLimits(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, requestedExecutions int) ([]models.UserRecipeLimit, error) {
	args := m.Called(ctx, userID, recipeID, requestedExecutions)
	return args.Get(0).([]models.UserRecipeLimit), args.Error(1)
}

// Test TaskService structure validation
func TestTaskService_StructureValidation(t *testing.T) {
	// Test that TaskService can be created with minimal dependencies
	taskService := &TaskService{
		logger: zap.NewNop(),
	}

	assert.NotNil(t, taskService)
	assert.NotNil(t, taskService.logger)
}

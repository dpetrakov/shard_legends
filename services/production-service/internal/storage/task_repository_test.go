package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabaseInterface - мок для DatabaseInterface
type MockDatabaseInterface struct {
	mock.Mock
}

func (m *MockDatabaseInterface) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(Row)
}

func (m *MockDatabaseInterface) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(Rows), mockArgs.Error(1)
}

func (m *MockDatabaseInterface) Exec(ctx context.Context, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Error(0)
}

func (m *MockDatabaseInterface) BeginTx(ctx context.Context) (Tx, error) {
	mockArgs := m.Called(ctx)
	return mockArgs.Get(0).(Tx), mockArgs.Error(1)
}

func (m *MockDatabaseInterface) Health(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
}

// MockTx - мок для транзакции
type MockTx struct {
	mock.Mock
}

func (m *MockTx) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(Row)
}

func (m *MockTx) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(Rows), mockArgs.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Error(0)
}

func (m *MockTx) Commit() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

func (m *MockTx) Rollback() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

// MockCacheInterface - мок для CacheInterface
type MockCacheInterface struct {
	mock.Mock
}

func (m *MockCacheInterface) Get(ctx context.Context, key string) (string, error) {
	mockArgs := m.Called(ctx, key)
	return mockArgs.String(0), mockArgs.Error(1)
}

func (m *MockCacheInterface) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	mockArgs := m.Called(ctx, key, value, ttl)
	return mockArgs.Error(0)
}

func (m *MockCacheInterface) Del(ctx context.Context, key string) error {
	mockArgs := m.Called(ctx, key)
	return mockArgs.Error(0)
}

func (m *MockCacheInterface) Health(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
}

// MockMetricsInterface - мок для MetricsInterface
type MockMetricsInterface struct {
	mock.Mock
}

func (m *MockMetricsInterface) IncDBQuery(operation string) {
	m.Called(operation)
}

func (m *MockMetricsInterface) IncCacheHit(cacheType string) {
	m.Called(cacheType)
}

func (m *MockMetricsInterface) IncCacheMiss(cacheType string) {
	m.Called(cacheType)
}

func (m *MockMetricsInterface) ObserveDBQueryDuration(operation string, duration time.Duration) {
	m.Called(operation, duration)
}

// MockRows - мок для Rows
type MockRows struct {
	mock.Mock
	data [][]interface{}
	pos  int
}

func (m *MockRows) Next() bool {
	m.pos++
	return m.pos <= len(m.data)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.pos <= 0 || m.pos > len(m.data) {
		return nil
	}

	row := m.data[m.pos-1]
	for i, dest := range dest {
		if i < len(row) {
			switch d := dest.(type) {
			case *uuid.UUID:
				*d = row[i].(uuid.UUID)
			case *string:
				*d = row[i].(string)
			case *int:
				*d = row[i].(int)
			case *time.Time:
				*d = row[i].(time.Time)
			case **time.Time:
				if row[i] != nil {
					t := row[i].(time.Time)
					*d = &t
				}
			case *models.AppliedModifiers:
				*d = row[i].(models.AppliedModifiers)
			}
		}
	}
	return nil
}

func (m *MockRows) Err() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

func (m *MockRows) Close() {
	m.Called()
}

func TestTaskRepository_CreateTask(t *testing.T) {
	// Arrange
	mockDB := &MockDatabaseInterface{}
	mockCache := &MockCacheInterface{}
	mockMetrics := &MockMetricsInterface{}
	mockTx := &MockTx{}

	deps := &RepositoryDependencies{
		DB:               mockDB,
		Cache:            mockCache,
		MetricsCollector: mockMetrics,
	}

	repo := NewTaskRepository(deps)

	task := &models.ProductionTask{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		RecipeID:         uuid.New(),
		SlotNumber:       1,
		Status:           models.TaskStatusPending,
		CreatedAt:        time.Now(),
		ModifiersApplied: models.AppliedModifiers{"test": "value"},
		OutputItems: []models.TaskOutputItem{
			{
				ItemID:   uuid.New(),
				Quantity: 5,
			},
		},
	}

	// Setup mocks
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockTx.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)
	mockMetrics.On("IncDBQuery", "task_create")

	// Act
	err := repo.CreateTask(context.Background(), task)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestTaskRepository_GetUserTasks(t *testing.T) {
	// Arrange
	mockDB := &MockDatabaseInterface{}
	mockCache := &MockCacheInterface{}
	mockMetrics := &MockMetricsInterface{}

	deps := &RepositoryDependencies{
		DB:               mockDB,
		Cache:            mockCache,
		MetricsCollector: mockMetrics,
	}

	repo := NewTaskRepository(deps)

	userID := uuid.New()
	statuses := []string{models.TaskStatusPending, models.TaskStatusInProgress}

	taskID := uuid.New()
	recipeID := uuid.New()
	now := time.Now()

	// Setup mock data
	mockRows := &MockRows{
		data: [][]interface{}{
			{
				taskID,                                // id
				userID,                                // user_id  
				recipeID,                              // recipe_id
				1,                                     // slot_number
				models.TaskStatusPending,              // status
				nil,                                   // started_at
				nil,                                   // completion_time
				nil,                                   // claimed_at
				models.AppliedModifiers{},             // pre_calculated_results
				models.AppliedModifiers{},             // modifiers_applied
				nil,                                   // reservation_id
				now,                                   // created_at
				now,                                   // updated_at
			},
		},
	}

	mockOutputRows := &MockRows{
		data: [][]interface{}{},
	}

	// Setup mocks
	mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockRows, nil).Once()
	mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockOutputRows, nil).Maybe()
	mockRows.On("Err").Return(nil)
	mockRows.On("Close")
	mockOutputRows.On("Err").Return(nil)
	mockOutputRows.On("Close")

	// Act
	tasks, err := repo.GetUserTasks(context.Background(), userID, statuses)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, taskID, tasks[0].ID)
	assert.Equal(t, userID, tasks[0].UserID)
	assert.Equal(t, models.TaskStatusPending, tasks[0].Status)

	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
	mockOutputRows.AssertExpectations(t)
}

func TestTaskRepository_UpdateTaskStatus(t *testing.T) {
	// Arrange
	mockDB := &MockDatabaseInterface{}
	mockCache := &MockCacheInterface{}
	mockMetrics := &MockMetricsInterface{}

	deps := &RepositoryDependencies{
		DB:               mockDB,
		Cache:            mockCache,
		MetricsCollector: mockMetrics,
	}

	repo := NewTaskRepository(deps)

	taskID := uuid.New()
	newStatus := models.TaskStatusInProgress

	// Setup mocks
	mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	mockMetrics.On("IncDBQuery", "task_status_update")

	// Act
	err := repo.UpdateTaskStatus(context.Background(), taskID, newStatus)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

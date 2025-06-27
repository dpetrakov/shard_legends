package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/shard-legends/production-service/internal/models"
)

// Тест для NewTaskService
func TestNewTaskService_Creation(t *testing.T) {
	// Создаем nil зависимости для простого теста конструктора
	service := NewTaskService(
		nil, // taskRepo
		nil, // recipeRepo
		nil, // classifierRepo
		nil, // codeConverter
		nil, // inventoryClient
		nil, // userClient
		zap.NewNop(),
	)

	assert.NotNil(t, service)
	assert.NotNil(t, service.modifierService)
	assert.NotNil(t, service.calculator)
}

// Тест для CalculateSlotInfo (существующий метод)
func TestTaskService_CalculateSlotInfo(t *testing.T) {
	service := NewTaskService(
		nil, nil, nil, nil, nil, nil, zap.NewNop(),
	)

	tests := []struct {
		name  string
		tasks []models.ProductionTask
	}{
		{
			name:  "empty tasks",
			tasks: []models.ProductionTask{},
		},
		{
			name: "single task",
			tasks: []models.ProductionTask{
				{
					ID:                 uuid.New(),
					Status:             models.TaskStatusInProgress,
					OperationClassCode: "basic",
				},
			},
		},
		{
			name: "multiple tasks",
			tasks: []models.ProductionTask{
				{
					ID:                 uuid.New(),
					Status:             models.TaskStatusInProgress,
					OperationClassCode: "basic",
				},
				{
					ID:                 uuid.New(),
					Status:             models.TaskStatusPending,
					OperationClassCode: "advanced",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateSlotInfo(tt.tasks)
			assert.NotNil(t, result)
			// CalculateSlotInfo всегда возвращает результат
		})
	}
}

// Простой тест для существующего функционала
func TestTaskService_Constants(t *testing.T) {
	// Проверяем, что константы определены правильно
	assert.Equal(t, "pending", models.TaskStatusPending)
	assert.Equal(t, "in_progress", models.TaskStatusInProgress)
	assert.Equal(t, "completed", models.TaskStatusCompleted)
	assert.Equal(t, "claimed", models.TaskStatusClaimed)
	assert.Equal(t, "cancelled", models.TaskStatusCancelled)
	assert.Equal(t, "failed", models.TaskStatusFailed)
}
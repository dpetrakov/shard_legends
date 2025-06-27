package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/shard-legends/production-service/internal/models"
)

// Тест для slotSupportsOperation
func TestTaskService_SlotSupportsOperation_Extended(t *testing.T) {
	service := NewTaskService(nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name      string
		slot      models.ProductionSlot
		operation string
		expected  bool
	}{
		{
			name: "universal slot supports crafting",
			slot: models.ProductionSlot{
				SlotType: "universal",
			},
			operation: "crafting",
			expected:  true,
		},
		{
			name: "universal slot supports smelting",
			slot: models.ProductionSlot{
				SlotType: "universal",
			},
			operation: "smelting",
			expected:  true,
		},
		{
			name: "specialized slot supports listed operation",
			slot: models.ProductionSlot{
				SlotType:            "specialized",
				SupportedOperations: []string{"crafting", "enchanting"},
			},
			operation: "crafting",
			expected:  true,
		},
		{
			name: "specialized slot supports another listed operation",
			slot: models.ProductionSlot{
				SlotType:            "specialized",
				SupportedOperations: []string{"crafting", "enchanting"},
			},
			operation: "enchanting",
			expected:  true,
		},
		{
			name: "specialized slot does not support unlisted operation",
			slot: models.ProductionSlot{
				SlotType:            "specialized",
				SupportedOperations: []string{"crafting"},
			},
			operation: "smelting",
			expected:  false,
		},
		{
			name: "specialized slot with empty operations",
			slot: models.ProductionSlot{
				SlotType:            "specialized",
				SupportedOperations: []string{},
			},
			operation: "crafting",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.slotSupportsOperation(tt.slot, tt.operation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Тест для CalculateSlotInfo с различными комбинациями статусов
func TestTaskService_CalculateSlotInfo_Extended(t *testing.T) {
	service := NewTaskService(nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name         string
		tasks        []models.ProductionTask
		expectedUsed int
		expectedFree int
	}{
		{
			name:         "no tasks",
			tasks:        []models.ProductionTask{},
			expectedUsed: 0,
			expectedFree: 2,
		},
		{
			name: "all in_progress",
			tasks: []models.ProductionTask{
				{ID: uuid.New(), Status: models.TaskStatusInProgress},
				{ID: uuid.New(), Status: models.TaskStatusInProgress},
			},
			expectedUsed: 2,
			expectedFree: 0,
		},
		{
			name: "all pending", 
			tasks: []models.ProductionTask{
				{ID: uuid.New(), Status: models.TaskStatusPending},
				{ID: uuid.New(), Status: models.TaskStatusPending},
			},
			expectedUsed: 0,
			expectedFree: 2,
		},
		{
			name: "all completed",
			tasks: []models.ProductionTask{
				{ID: uuid.New(), Status: models.TaskStatusCompleted},
				{ID: uuid.New(), Status: models.TaskStatusCompleted},
			},
			expectedUsed: 0,
			expectedFree: 2,
		},
		{
			name: "mixed statuses",
			tasks: []models.ProductionTask{
				{ID: uuid.New(), Status: models.TaskStatusInProgress},
				{ID: uuid.New(), Status: models.TaskStatusPending},
				{ID: uuid.New(), Status: models.TaskStatusCompleted},
				{ID: uuid.New(), Status: models.TaskStatusInProgress},
			},
			expectedUsed: 2,
			expectedFree: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateSlotInfo(tt.tasks)
			assert.Equal(t, 2, result.Total) // Захардкоженное значение
			assert.Equal(t, tt.expectedUsed, result.Used)
			assert.Equal(t, tt.expectedFree, result.Free)
		})
	}
}

// Тест для prepareItemsForReservation с различными сценариями
func TestTaskService_PrepareItemsForReservation_Extended(t *testing.T) {
	service := NewTaskService(nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name           string
		inputItems     []models.RecipeInputItem
		request        models.StartProductionRequest
		expectedCount  int
		expectedTotal  int
	}{
		{
			name: "no boosters",
			inputItems: []models.RecipeInputItem{
				{ItemID: uuid.New(), Quantity: 5},
			},
			request: models.StartProductionRequest{
				ExecutionCount: 2,
				Boosters:       []models.BoosterItem{},
			},
			expectedCount: 1,
			expectedTotal: 10, // 5 * 2
		},
		{
			name: "with boosters",
			inputItems: []models.RecipeInputItem{
				{ItemID: uuid.New(), Quantity: 3},
			},
			request: models.StartProductionRequest{
				ExecutionCount: 3,
				Boosters: []models.BoosterItem{
					{ItemID: uuid.New(), Quantity: 1},
					{ItemID: uuid.New(), Quantity: 2},
				},
			},
			expectedCount: 3, // 1 input + 2 boosters
			expectedTotal: 12, // 9 + 1 + 2
		},
		{
			name: "multiple inputs",
			inputItems: []models.RecipeInputItem{
				{ItemID: uuid.New(), Quantity: 2},
				{ItemID: uuid.New(), Quantity: 3},
			},
			request: models.StartProductionRequest{
				ExecutionCount: 2,
				Boosters: []models.BoosterItem{
					{ItemID: uuid.New(), Quantity: 1},
				},
			},
			expectedCount: 3, // 2 inputs + 1 booster
			expectedTotal: 11, // (2*2) + (3*2) + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.prepareItemsForReservation(nil, tt.inputItems, tt.request)
			
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
			
			totalQuantity := 0
			for _, item := range result {
				totalQuantity += item.Quantity
			}
			assert.Equal(t, tt.expectedTotal, totalQuantity)
		})
	}
}
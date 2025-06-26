package service

import (
	"testing"

	"github.com/shard-legends/production-service/internal/models"
	"github.com/stretchr/testify/assert"
)

// Test CalculateSlotInfo function which is used in claim operations
func TestCalculateSlotInfo(t *testing.T) {
	// Create a minimal TaskService instance for testing public methods
	service := &TaskService{}

	// Test with mixed task statuses
	tasks := []models.ProductionTask{
		{Status: models.TaskStatusInProgress},
		{Status: models.TaskStatusInProgress},
		{Status: models.TaskStatusPending},
		{Status: models.TaskStatusCompleted},
		{Status: models.TaskStatusClaimed},
	}

	slotInfo := service.CalculateSlotInfo(tasks)

	// Verify slot calculation
	assert.Equal(t, 2, slotInfo.Total, "Total slots should be 2 (hardcoded)")
	assert.Equal(t, 2, slotInfo.Used, "Used slots should count only in_progress tasks")
	assert.Equal(t, 0, slotInfo.Free, "Free slots should be total - used")
}

func TestCalculateSlotInfo_NoActiveTasks(t *testing.T) {
	service := &TaskService{}

	tasks := []models.ProductionTask{
		{Status: models.TaskStatusPending},
		{Status: models.TaskStatusCompleted},
		{Status: models.TaskStatusClaimed},
		{Status: models.TaskStatusCancelled},
	}

	slotInfo := service.CalculateSlotInfo(tasks)

	assert.Equal(t, 2, slotInfo.Total)
	assert.Equal(t, 0, slotInfo.Used, "No tasks should be using slots")
	assert.Equal(t, 2, slotInfo.Free, "All slots should be free")
}

func TestCalculateSlotInfo_AllSlotsUsed(t *testing.T) {
	service := &TaskService{}

	tasks := []models.ProductionTask{
		{Status: models.TaskStatusInProgress},
		{Status: models.TaskStatusInProgress},
		{Status: models.TaskStatusInProgress}, // Third task won't fit
	}

	slotInfo := service.CalculateSlotInfo(tasks)

	assert.Equal(t, 2, slotInfo.Total)
	assert.Equal(t, 3, slotInfo.Used, "Should count all in_progress tasks even if exceeding total")
	assert.Equal(t, -1, slotInfo.Free, "Free can be negative when overbooked")
}

func TestCalculateSlotInfo_EmptyList(t *testing.T) {
	service := &TaskService{}

	var tasks []models.ProductionTask

	slotInfo := service.CalculateSlotInfo(tasks)

	assert.Equal(t, 2, slotInfo.Total)
	assert.Equal(t, 0, slotInfo.Used)
	assert.Equal(t, 2, slotInfo.Free)
}

// Test helper for status constants validation
func TestTaskStatusConstants(t *testing.T) {
	// Ensure all status constants are properly defined
	statuses := []string{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
		models.TaskStatusClaimed,
		models.TaskStatusCancelled,
		models.TaskStatusFailed,
	}

	expectedStatuses := []string{
		"pending",
		"in_progress",
		"completed",
		"claimed",
		"cancelled",
		"failed",
	}

	assert.Equal(t, expectedStatuses, statuses, "Task status constants should match expected values")
}

// Test for operation class constants used in slot calculation
func TestOperationClassConstants(t *testing.T) {
	classes := []string{
		models.OperationClassCrafting,
		models.OperationClassSmelting,
		models.OperationClassChestOpening,
		models.OperationClassResourceGathering,
		models.OperationClassSpecial,
	}

	expectedClasses := []string{
		"crafting",
		"smelting",
		"chest_opening",
		"resource_gathering",
		"special",
	}

	assert.Equal(t, expectedClasses, classes, "Operation class constants should match expected values")
}

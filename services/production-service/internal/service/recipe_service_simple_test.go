package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Тест для RecipeLimitExceededError
func TestRecipeLimitExceededError_Error(t *testing.T) {
	recipeID := uuid.New()
	err := &RecipeLimitExceededError{
		RecipeID:     recipeID,
		LimitType:    "daily",
		LimitObject:  "recipe",
		CurrentUsage: 15,
		MaxAllowed:   10,
	}

	errorMessage := err.Error()
	assert.Contains(t, errorMessage, "recipe limit exceeded")
	assert.Contains(t, errorMessage, recipeID.String())
	assert.Contains(t, errorMessage, "daily")
	assert.Contains(t, errorMessage, "recipe")
	assert.Contains(t, errorMessage, "15/10")
}

// Тест для IsRecipeLimitExceededError
func TestIsRecipeLimitExceededError(t *testing.T) {
	recipeID := uuid.New()
	limitErr := &RecipeLimitExceededError{
		RecipeID:     recipeID,
		LimitType:    "daily", 
		LimitObject:  "recipe",
		CurrentUsage: 15,
		MaxAllowed:   10,
	}

	// Тест с правильной ошибкой
	result, ok := IsRecipeLimitExceededError(limitErr)
	assert.True(t, ok)
	assert.Equal(t, limitErr, result)

	// Тест с другой ошибкой
	otherErr := assert.AnError
	result, ok = IsRecipeLimitExceededError(otherErr)
	assert.False(t, ok)
	assert.Nil(t, result)
}

// Тест создания RecipeService с зависимостями
func TestRecipeService_Creation(t *testing.T) {
	deps := &ServiceDependencies{
		Repository: nil,
		Cache:      nil,
		Metrics:    nil,
	}

	service := NewRecipeService(deps)
	assert.NotNil(t, service)
}
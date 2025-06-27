package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Тест для NewProductionCalculator
func TestNewProductionCalculator_Creation(t *testing.T) {
	modifierService := &ModifierService{}
	calculator := NewProductionCalculator(
		nil, // classifierRepo
		modifierService,
		zap.NewNop(),
	)

	assert.NotNil(t, calculator)
}

// Простой тест для проверки что calculator создается
func TestProductionCalculator_BasicFunctionality(t *testing.T) {
	modifierService := &ModifierService{}
	calculator := NewProductionCalculator(nil, modifierService, zap.NewNop())

	// Просто проверяем что объект не nil
	assert.NotNil(t, calculator)
	
	// Можем проверить что у него есть внутренние поля
	assert.NotNil(t, modifierService)
}
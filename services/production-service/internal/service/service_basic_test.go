package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shard-legends/production-service/internal/storage"
)

// Тест для ServiceDependencies
func TestServiceDependencies_Structure(t *testing.T) {
	deps := &ServiceDependencies{
		Repository: &storage.Repository{},
		Cache:      nil,
		Metrics:    nil,
	}

	assert.NotNil(t, deps)
	assert.NotNil(t, deps.Repository)
}

// Тест для Service структуры
func TestService_Structure(t *testing.T) {
	service := &Service{
		Recipe:        nil,
		CodeConverter: nil,
	}

	assert.NotNil(t, service)
}

// Тест создания CodeConverterService
func TestNewCodeConverterService_Creation(t *testing.T) {
	deps := &ServiceDependencies{
		Repository: &storage.Repository{},
		Cache:      nil,
		Metrics:    nil,
	}

	codeConverter := NewCodeConverterService(deps)
	assert.NotNil(t, codeConverter)
}
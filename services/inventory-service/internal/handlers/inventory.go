package handlers

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/shard-legends/inventory-service/internal/service"
)

// InventoryHandler handles inventory-related HTTP requests
type InventoryHandler struct {
	inventoryService  service.InventoryService
	classifierService service.ClassifierService
	logger            *slog.Logger
	validator         *validator.Validate
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(
	inventoryService service.InventoryService,
	classifierService service.ClassifierService,
	logger *slog.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		inventoryService:  inventoryService,
		classifierService: classifierService,
		logger:            logger,
		validator:         validator.New(),
	}
}

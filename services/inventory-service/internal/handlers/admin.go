package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/internal/models"
)

// AdjustInventory handles POST /admin/inventory/adjust
func (h *InventoryHandler) AdjustInventory(c *gin.Context) {
	h.logger.Info("AdjustInventory called")

	var req models.AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	response, err := h.inventoryService.AdjustInventory(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to adjust inventory", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to adjust inventory",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

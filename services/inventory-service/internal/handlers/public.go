package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/inventory-service/internal/auth"
	"github.com/shard-legends/inventory-service/internal/models"
)

// GetUserInventory handles GET /inventory (public, requires JWT)
func (h *InventoryHandler) GetUserInventory(c *gin.Context) {
	h.logger.Info("GetUserInventory with JWT called")

	// Extract user from JWT token (middleware should set it)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not found in token",
		})
		return
	}

	user, ok := userInterface.(*auth.UserContext)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
		return
	}

	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get section from query params (default to "main")
	sectionParam := c.DefaultQuery("section", "main")

	// Get section UUID from classifier
	sectionMapping, err := h.classifierService.GetClassifierMapping(c.Request.Context(), "inventory_section")
	if err != nil {
		h.logger.Error("Failed to get section mapping", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve section",
		})
		return
	}

	sectionID, exists := sectionMapping[sectionParam]
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_section",
			Message: "Invalid section specified",
		})
		return
	}

	// Get user inventory
	inventory, err := h.inventoryService.GetUserInventory(c.Request.Context(), userID, sectionID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve inventory",
		})
		return
	}

	c.JSON(http.StatusOK, models.InventoryResponse{
		Items: func() []models.InventoryItemResponse {
			result := make([]models.InventoryItemResponse, len(inventory))
			for i, item := range inventory {
				result[i] = *item
			}
			return result
		}(),
	})
}

// GetUserInventoryNoAuth handles GET /inventory (development only, no auth)
func (h *InventoryHandler) GetUserInventoryNoAuth(c *gin.Context) {
	h.logger.Info("GetUserInventory without JWT called")

	// For development only - user_id should be provided as query param
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_user_id",
			Message: "user_id query parameter is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get section from query params (default to "main")
	sectionParam := c.DefaultQuery("section", "main")

	// Get section UUID from classifier
	sectionMapping, err := h.classifierService.GetClassifierMapping(c.Request.Context(), "inventory_section")
	if err != nil {
		h.logger.Error("Failed to get section mapping", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve section",
		})
		return
	}

	sectionID, exists := sectionMapping[sectionParam]
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_section",
			Message: "Invalid section specified",
		})
		return
	}

	// Get user inventory
	inventory, err := h.inventoryService.GetUserInventory(c.Request.Context(), userID, sectionID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve inventory",
		})
		return
	}

	c.JSON(http.StatusOK, models.InventoryResponse{
		Items: func() []models.InventoryItemResponse {
			result := make([]models.InventoryItemResponse, len(inventory))
			for i, item := range inventory {
				result[i] = *item
			}
			return result
		}(),
	})
}
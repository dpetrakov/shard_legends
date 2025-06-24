package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/inventory-service/internal/auth"
	"github.com/shard-legends/inventory-service/internal/models"
)

// GetUserInventory handles GET /inventory
func (h *InventoryHandler) GetUserInventory(c *gin.Context) {
	// 1. Извлечь user_id из JWT токена
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// 2. Валидировать query параметры (section)
	section := c.DefaultQuery("section", "main")

	var sectionID uuid.UUID
	switch section {
	case "main":
		sectionID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") // main section UUID
	case "factory":
		sectionID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001") // factory section UUID
	case "trade":
		sectionID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440002") // trade section UUID
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_section",
			Message: "Section must be one of: main, factory, trade",
		})
		return
	}

	// 3. Получить инвентарь пользователя через сервис
	items, err := h.inventoryService.GetUserInventory(c.Request.Context(), userID, sectionID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", "user_id", userID, "section", section, "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve inventory",
		})
		return
	}

	// 4. Вернуть InventoryResponse
	responseItems := make([]models.InventoryItemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, *item)
	}

	response := models.InventoryResponse{
		Items: responseItems,
	}

	h.logger.Info("Successfully retrieved user inventory",
		"user_id", userID,
		"section", section,
		"items_count", len(items))

	c.JSON(http.StatusOK, response)
}

// GetHealthInfo provides a simple health check for public endpoints
func (h *InventoryHandler) GetHealthInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "inventory-service",
		"version": "1.0.0",
	})
}

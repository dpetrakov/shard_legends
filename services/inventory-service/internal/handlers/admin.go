package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/inventory-service/internal/auth"
	"github.com/shard-legends/inventory-service/internal/models"
)

// AdjustInventory handles POST /admin/inventory/adjust
func (h *InventoryHandler) AdjustInventory(c *gin.Context) {
	var req models.AdjustInventoryRequest

	// 1. Проверка административных прав доступа (должно быть проверено в middleware)
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// 2. Парсинг AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse adjust inventory request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 3. Валидация входных данных
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 4. Валидация административных прав (placeholder - должно быть в middleware)
	adminUserID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid admin user ID", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid admin user ID format",
		})
		return
	}

	// TODO: Реализовать административную корректировку через сервис
	// Пока что возвращаем placeholder response
	h.logger.Info("Admin inventory adjustment requested",
		"admin_user_id", adminUserID,
		"target_user_id", req.UserID,
		"section", req.Section,
		"items_count", len(req.Items),
		"reason", req.Reason)

	// Placeholder response - будет заменено реальной реализацией
	response := models.AdjustInventoryResponse{
		Success:       true,
		OperationIDs:  []uuid.UUID{uuid.New()},          // Placeholder
		FinalBalances: []models.InventoryItemResponse{}, // Will be calculated by service
	}

	c.JSON(http.StatusOK, response)
}

// CheckAdminPermissions is a placeholder for admin permission checking
// This should be implemented as middleware in production
func (h *InventoryHandler) CheckAdminPermissions(c *gin.Context) {
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		c.Abort()
		return
	}

	// TODO: Implement real admin permission checking
	// For now, allow all authenticated users (should be restricted in production)
	h.logger.Info("Admin permission check", "user_id", user.UserID)

	c.Next()
}

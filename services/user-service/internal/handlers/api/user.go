package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"user-service/internal/services"
	"user-service/pkg/logger"
)

// UserHandler handles internal user-related requests
type UserHandler struct {
	mockDataService *services.MockDataService
	logger          *logger.Logger
}

// NewUserHandler creates a new internal user handler
func NewUserHandler(mockDataService *services.MockDataService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		mockDataService: mockDataService,
		logger:          logger,
	}
}

// GetProductionSlots returns user production slots for internal services
// GET /internal/users/{user_id}/production-slots
func (h *UserHandler) GetProductionSlots(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.logger.Error("Missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_user_id",
			"message": "User ID parameter is required",
		})
		return
	}

	h.logger.Info("Getting production slots for internal service", map[string]interface{}{
		"user_id": userID,
	})

	// Получение информации о слотах (временная версия всегда возвращает одинаковые данные)
	slotsData := h.mockDataService.GetProductionSlots(userID)

	h.logger.Info("Internal production slots data generated", map[string]interface{}{
		"user_id":     userID,
		"total_slots": slotsData.TotalSlots,
	})

	c.JSON(http.StatusOK, slotsData)
}

// GetProductionModifiers returns user production modifiers for internal services
// GET /internal/users/{user_id}/production-modifiers
func (h *UserHandler) GetProductionModifiers(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.logger.Error("Missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_user_id",
			"message": "User ID parameter is required",
		})
		return
	}

	h.logger.Info("Getting production modifiers for internal service", map[string]interface{}{
		"user_id": userID,
	})

	// Получение модификаторов (временная версия всегда возвращает нулевые значения)
	modifiersData := h.mockDataService.GetProductionModifiers(userID)

	h.logger.Info("Internal production modifiers data generated", map[string]interface{}{
		"user_id": userID,
		"vip_level": modifiersData.Modifiers.VIPStatus.Level,
		"character_level": modifiersData.Modifiers.CharacterLevel.Level,
	})

	c.JSON(http.StatusOK, modifiersData)
}
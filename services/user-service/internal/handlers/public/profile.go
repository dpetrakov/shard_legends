package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"user-service/internal/middleware"
	"user-service/internal/services"
	"user-service/pkg/logger"
)

// ProfileHandler handles profile-related requests
type ProfileHandler struct {
	mockDataService *services.MockDataService
	logger          *logger.Logger
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(mockDataService *services.MockDataService, logger *logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		mockDataService: mockDataService,
		logger:          logger,
	}
}

// GetProfile returns user profile data
// GET /profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Извлечение пользователя из контекста (добавлено middleware)
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "user_not_found",
			"message": "User information not found in request context",
		})
		return
	}

	h.logger.Info("Getting profile for user", map[string]interface{}{
		"user_id":     user.UserID,
		"telegram_id": user.TelegramID,
	})

	// Получение моковых данных профиля
	profileData := h.mockDataService.GetProfileData(user.UserID, user.TelegramID)

	h.logger.Info("Profile data generated", map[string]interface{}{
		"user_id":    user.UserID,
		"vip_active": profileData.VIPStatus.IsActive,
		"level":      profileData.Level,
	})

	c.JSON(http.StatusOK, profileData)
}

// GetProductionSlots returns user production slots information
// GET /production-slots
func (h *ProfileHandler) GetProductionSlots(c *gin.Context) {
	// Извлечение пользователя из контекста
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "user_not_found",
			"message": "User information not found in request context",
		})
		return
	}

	h.logger.Info("Getting production slots for user", map[string]interface{}{
		"user_id": user.UserID,
	})

	// Получение информации о слотах
	slotsData := h.mockDataService.GetProductionSlots(user.UserID)

	h.logger.Info("Production slots data generated", map[string]interface{}{
		"user_id":     user.UserID,
		"total_slots": slotsData.TotalSlots,
	})

	c.JSON(http.StatusOK, slotsData)
}
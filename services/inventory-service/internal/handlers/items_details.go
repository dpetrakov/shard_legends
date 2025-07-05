package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/internal/models"
)

// GetItemsDetails handles POST /items/details (public, requires JWT)
func (h *InventoryHandler) GetItemsDetails(c *gin.Context) {
	h.logger.Info("GetItemsDetails called")

	// Extract user from JWT token (middleware should set it)
	_, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not found in token",
		})
		return
	}

	// Determine language: use query param if provided, otherwise fetch default from DB
	var languageCode string
	if lang, exists := c.GetQuery("lang"); exists && lang != "" {
		languageCode = lang
	} else {
		defaultLang, err := h.inventoryService.GetDefaultLanguage(c.Request.Context())
		if err != nil {
			h.logger.Warn("Failed to get default language, fallback to 'ru'", "error", err)
		}
		if defaultLang != nil {
			languageCode = defaultLang.Code
		} else {
			languageCode = "ru" // application-level fallback
		}
	}

	// Parse request body
	var req models.ItemDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	// Log request details for debugging
	for i, item := range req.Items {
		h.logger.Info("DEBUG: Received item details request",
			"index", i,
			"item_id", item.ItemID,
			"collection", item.Collection,
			"quality_level", item.QualityLevel)
	}

	// Validate request
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "empty_request",
			Message: "Items list cannot be empty",
		})
		return
	}

	if len(req.Items) > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "too_many_items",
			Message: "Maximum 100 items allowed per request",
		})
		return
	}

	// Get item details with translations
	response, err := h.inventoryService.GetItemsDetails(c.Request.Context(), &req, languageCode)
	if err != nil {
		h.logger.Error("Failed to get items details", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve item details",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

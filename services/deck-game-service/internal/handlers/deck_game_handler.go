package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/shard-legends/deck-game-service/internal/auth"
	"github.com/shard-legends/deck-game-service/internal/models"
	"github.com/shard-legends/deck-game-service/internal/service"
)

// DeckGameHandler handles deck game related HTTP requests
type DeckGameHandler struct {
	deckGameService service.DeckGameService
	logger          *slog.Logger
}

// NewDeckGameHandler creates a new deck game handler
func NewDeckGameHandler(deckGameService service.DeckGameService, logger *slog.Logger) *DeckGameHandler {
	return &DeckGameHandler{
		deckGameService: deckGameService,
		logger:          logger,
	}
}

// GetDailyChestStatus handles GET /deck/daily-chest/status
func (h *DeckGameHandler) GetDailyChestStatus(c *gin.Context) {
	h.logger.Info("GetDailyChestStatus called")

	// Extract user from JWT context
	userValue, exists := c.Get("user")
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_user_context",
			Message: "User context not found",
		})
		return
	}

	user, ok := userValue.(*auth.UserContext)
	if !ok {
		h.logger.Error("Invalid user context type")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_user_context",
			Message: "Invalid user context type",
		})
		return
	}

	// Parse user ID from UUID string
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID format", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get daily chest status
	status, err := h.deckGameService.GetDailyChestStatus(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get daily chest status", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get daily chest status",
		})
		return
	}

	h.logger.Info("Daily chest status retrieved successfully", "user_id", userID, "status", status)
	c.JSON(http.StatusOK, status)
}

// ClaimDailyChest handles POST /deck/daily-chest/claim
func (h *DeckGameHandler) ClaimDailyChest(c *gin.Context) {
	h.logger.Info("ClaimDailyChest called")

	// Extract user from JWT context
	userValue, exists := c.Get("user")
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_user_context",
			Message: "User context not found",
		})
		return
	}

	user, ok := userValue.(*auth.UserContext)
	if !ok {
		h.logger.Error("Invalid user context type")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_user_context",
			Message: "Invalid user context type",
		})
		return
	}

	// Parse user ID from UUID string
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID format", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Parse request body
	var request models.ClaimRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Extract JWT token from context
	jwtToken, exists := c.Get("jwt_token")
	if !exists {
		h.logger.Error("JWT token not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_jwt_token",
			Message: "JWT token not found in context",
		})
		return
	}

	jwtTokenStr, ok := jwtToken.(string)
	if !ok {
		h.logger.Error("Invalid JWT token type in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_jwt_token",
			Message: "Invalid JWT token type in context",
		})
		return
	}

	h.logger.Info("Processing claim request",
		"user_id", userID,
		"combo", request.Combo,
		"chest_indices", request.ChestIndices)

	// Process claim
	response, err := h.deckGameService.ClaimDailyChest(c.Request.Context(), jwtTokenStr, userID, &request)
	if err != nil {
		h.logger.Error("Failed to claim daily chest", "user_id", userID, "error", err)

		// Handle business errors
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "invalid_combo"):
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_combo",
				Message: "Provided combo is less than expected",
			})
		case strings.Contains(errorMessage, "daily_finished"):
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "daily_finished",
				Message: "Daily limit reached",
			})
		case strings.Contains(errorMessage, "recipe_not_found") || strings.Contains(errorMessage, "invalid recipe ID"):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "recipe_not_found",
				Message: "Daily chest recipe not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to claim daily chest",
			})
		}
		return
	}

	h.logger.Info("Daily chest claimed successfully", "user_id", userID, "items_count", len(response.Items))
	c.JSON(http.StatusOK, response)
}

// OpenChest handles POST /deck/chest/open
func (h *DeckGameHandler) OpenChest(c *gin.Context) {
	h.logger.Info("OpenChest called")

	// Extract user from JWT context
	userValue, exists := c.Get("user")
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_user_context",
			Message: "User context not found",
		})
		return
	}

	user, ok := userValue.(*auth.UserContext)
	if !ok {
		h.logger.Error("Invalid user context type")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_user_context",
			Message: "Invalid user context type",
		})
		return
	}

	// Parse user ID from UUID string
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID format", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Parse request body
	var request models.OpenChestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Extract JWT token from context
	jwtToken, exists := c.Get("jwt_token")
	if !exists {
		h.logger.Error("JWT token not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_jwt_token",
			Message: "JWT token not found in context",
		})
		return
	}

	jwtTokenStr, ok := jwtToken.(string)
	if !ok {
		h.logger.Error("Invalid JWT token type in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_jwt_token",
			Message: "Invalid JWT token type in context",
		})
		return
	}

	h.logger.Info("Processing chest opening request",
		"user_id", userID,
		"chest_type", request.ChestType,
		"quality_level", request.QualityLevel,
		"quantity", request.Quantity,
		"open_all", request.OpenAll)

	// Process chest opening
	response, err := h.deckGameService.OpenChest(c.Request.Context(), jwtTokenStr, userID, &request)
	if err != nil {
		h.logger.Error("Failed to open chest", "user_id", userID, "error", err)

		// Handle business errors
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "invalid_input"):
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_input",
				Message: "Invalid request parameters",
			})
		case strings.Contains(errorMessage, "recipe_not_found"):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "recipe_not_found",
				Message: "Recipe for this chest type not found",
			})
		case strings.Contains(errorMessage, "insufficient_chests"), strings.Contains(errorMessage, "insufficient_items"):
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "insufficient_chests",
				Message: "Not enough chests in inventory",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to open chest",
			})
		}
		return
	}

	h.logger.Info("Chest opened successfully", "user_id", userID, "items_count", len(response.Items), "quantity_opened", response.QuantityOpened)
	c.JSON(http.StatusOK, response)
}

// BuyItem handles POST /api/deck-game/buy-item
func (h *DeckGameHandler) BuyItem(c *gin.Context) {
	h.logger.Info("BuyItem called")

	// Extract user from JWT context
	userValue, exists := c.Get("user")
	if !exists {
		h.logger.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_user_context",
			Message: "User context not found",
		})
		return
	}

	user, ok := userValue.(*auth.UserContext)
	if !ok {
		h.logger.Error("Invalid user context type")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_user_context",
			Message: "Invalid user context type",
		})
		return
	}

	// Parse user ID from UUID string
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID format", "user_id", user.UserID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Parse request body
	var request models.BuyItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		h.logger.Error("Invalid buy item request", "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// Extract JWT token from context
	jwtToken, exists := c.Get("jwt_token")
	if !exists {
		h.logger.Error("JWT token not found in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "missing_jwt_token",
			Message: "JWT token not found in context",
		})
		return
	}

	jwtTokenStr, ok := jwtToken.(string)
	if !ok {
		h.logger.Error("Invalid JWT token type in context")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_jwt_token",
			Message: "Invalid JWT token type in context",
		})
		return
	}

	h.logger.Info("Processing buy item request",
		"user_id", userID,
		"item_code", request.ItemCode,
		"quantity", request.Quantity)

	// Process buy item
	response, err := h.deckGameService.BuyItem(c.Request.Context(), jwtTokenStr, userID, &request)
	if err != nil {
		h.logger.Error("Failed to buy item", "user_id", userID, "error", err)

		// Handle errors
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "recipe_not_found"):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "recipe_not_found",
				Message: "Recipe not found",
			})
		case strings.Contains(errorMessage, "recipe_ambiguous"):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "recipe_ambiguous",
				Message: "Multiple suitable recipes found",
			})
		case strings.Contains(errorMessage, "production_error"):
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "production_error",
				Message: "Error during production",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to buy item",
			})
		}
		return
	}

	h.logger.Info("Item bought successfully", "user_id", userID, "items_count", len(response.Items))
	c.JSON(http.StatusOK, response)
}

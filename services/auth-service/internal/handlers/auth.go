package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	logger            *slog.Logger
	telegramValidator *services.TelegramValidator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger *slog.Logger, telegramValidator *services.TelegramValidator) *AuthHandler {
	return &AuthHandler{
		logger:            logger,
		telegramValidator: telegramValidator,
	}
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success   bool          `json:"success"`
	Token     string        `json:"token,omitempty"`
	ExpiresAt string        `json:"expires_at,omitempty"`
	User      *UserResponse `json:"user,omitempty"`
	Error     string        `json:"error,omitempty"`
	Message   string        `json:"message,omitempty"`
}

// UserResponse represents user data in auth response
type UserResponse struct {
	ID         string `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name,omitempty"`
	IsNewUser  bool   `json:"is_new_user"`
}

// Auth handles POST /auth requests
func (h *AuthHandler) Auth(c *gin.Context) {
	// Get X-Telegram-Init-Data header
	initData := c.GetHeader("X-Telegram-Init-Data")
	if initData == "" {
		h.logger.Error("Missing X-Telegram-Init-Data header")
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "missing_telegram_data",
			Message: "X-Telegram-Init-Data header is required",
		})
		return
	}

	h.logger.Info("Processing auth request",
		"method", c.Request.Method,
		"headers", len(c.Request.Header),
		"data_length", len(initData))

	// Validate Telegram data
	telegramData, err := h.telegramValidator.ValidateTelegramData(initData)
	if err != nil {
		h.logger.Error("Telegram data validation failed", "error", err)
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Error:   "invalid_telegram_data",
			Message: "Telegram authentication failed",
		})
		return
	}

	h.logger.Info("Telegram data validation successful",
		"user_id", telegramData.User.ID,
		"username", telegramData.User.Username,
		"first_name", telegramData.User.FirstName)

	// For now, return a mock successful response
	// In future implementations, this will:
	// 1. Check if user exists in database
	// 2. Create user if new
	// 3. Generate JWT token
	// 4. Store token in Redis
	userResponse := &UserResponse{
		ID:         "mock-uuid-user-id",
		TelegramID: telegramData.User.ID,
		Username:   telegramData.User.Username,
		FirstName:  telegramData.User.FirstName,
		LastName:   telegramData.User.LastName,
		IsNewUser:  true, // Mock value
	}

	response := AuthResponse{
		Success:   true,
		Token:     "mock-jwt-token-will-be-implemented-in-d3",
		ExpiresAt: "2024-12-22T10:30:00Z",
		User:      userResponse,
	}

	h.logger.Info("Authentication successful",
		"telegram_id", telegramData.User.ID,
		"is_new_user", userResponse.IsNewUser)

	c.JSON(http.StatusOK, response)
}

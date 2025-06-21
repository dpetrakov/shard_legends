package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	logger            *slog.Logger
	telegramValidator *services.TelegramValidator
	jwtService        *services.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger *slog.Logger, telegramValidator *services.TelegramValidator, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		logger:            logger,
		telegramValidator: telegramValidator,
		jwtService:        jwtService,
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
	ID           string `json:"id"`
	TelegramID   int64  `json:"telegram_id"`
	Username     string `json:"username,omitempty"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	IsPremium    bool   `json:"is_premium"`
	PhotoURL     string `json:"photo_url,omitempty"`
	IsNewUser    bool   `json:"is_new_user"`
}

// Auth handles POST /auth requests
func (h *AuthHandler) Auth(c *gin.Context) {
	// Get X-Telegram-Init-Data header
	initData := c.GetHeader("X-Telegram-Init-Data")
	if initData == "" {
		h.logger.Error("Missing X-Telegram-Init-Data header")
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "missing_init_data",
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
			Error:   "invalid_telegram_signature",
			Message: "Telegram authentication failed",
		})
		return
	}

	h.logger.Info("Telegram data validation successful",
		"user_id", telegramData.User.ID,
		"username", telegramData.User.Username,
		"first_name", telegramData.User.FirstName)

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(telegramData.User.ID)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", "error", err, "telegram_id", telegramData.User.ID)
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Error:   "internal_server_error",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// Calculate expiration time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// TODO: In future implementations, this will:
	// 1. Check if user exists in database
	// 2. Create user if new
	// 3. Store token in Redis for session management
	userResponse := &UserResponse{
		ID:           "mock-uuid-user-id", // TODO: Generate real UUID when DB is connected
		TelegramID:   telegramData.User.ID,
		Username:     telegramData.User.Username,
		FirstName:    telegramData.User.FirstName,
		LastName:     telegramData.User.LastName,
		LanguageCode: telegramData.User.LanguageCode,
		IsPremium:    telegramData.User.IsPremium,
		PhotoURL:     telegramData.User.PhotoURL,
		IsNewUser:    true, // TODO: Check database for existing user
	}

	response := AuthResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      userResponse,
	}

	h.logger.Info("Authentication successful",
		"telegram_id", telegramData.User.ID,
		"is_new_user", userResponse.IsNewUser)

	c.JSON(http.StatusOK, response)
}

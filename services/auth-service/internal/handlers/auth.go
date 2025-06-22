package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/models"
	"github.com/shard-legends/auth-service/internal/services"
	"github.com/shard-legends/auth-service/internal/storage"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	logger            *slog.Logger
	telegramValidator *services.TelegramValidator
	jwtService        *services.JWTService
	userRepo          storage.UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger *slog.Logger, telegramValidator *services.TelegramValidator, jwtService *services.JWTService, userRepo storage.UserRepository) *AuthHandler {
	return &AuthHandler{
		logger:            logger,
		telegramValidator: telegramValidator,
		jwtService:        jwtService,
		userRepo:          userRepo,
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

	// Get or create user in database
	ctx := context.Background()
	user, err := h.getOrCreateUser(ctx, *telegramData.User)
	if err != nil {
		h.logger.Error("Failed to get or create user", "error", err, "telegram_id", telegramData.User.ID)
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Error:   "internal_server_error",
			Message: "Failed to process user data",
		})
		return
	}

	// Update last login timestamp
	if err := h.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		h.logger.Warn("Failed to update last login", "error", err, "user_id", user.ID.String())
		// Don't fail the request for this non-critical error
	}
	
	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID, telegramData.User.ID)
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

	// Create user response
	userResponse := &UserResponse{
		ID:           user.ID.String(),
		TelegramID:   user.TelegramID,
		Username:     stringPtrToString(user.Username),
		FirstName:    user.FirstName,
		LastName:     stringPtrToString(user.LastName),
		LanguageCode: stringPtrToString(user.LanguageCode),
		IsPremium:    user.IsPremium,
		PhotoURL:     stringPtrToString(user.PhotoURL),
		IsNewUser:    user.CreatedAt.After(time.Now().Add(-time.Minute)), // Consider new if created within last minute
	}

	response := AuthResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      userResponse,
	}

	h.logger.Info("Authentication successful",
		"user_id", user.ID.String(),
		"telegram_id", telegramData.User.ID,
		"is_new_user", userResponse.IsNewUser)

	c.JSON(http.StatusOK, response)
}

// getOrCreateUser gets an existing user by telegram_id or creates a new one
func (h *AuthHandler) getOrCreateUser(ctx context.Context, telegramUser services.TelegramUser) (*models.User, error) {
	// Try to get existing user by telegram_id
	existingUser, err := h.userRepo.GetUserByTelegramID(ctx, telegramUser.ID)
	if err == nil {
		// User exists, update their information if needed
		h.logger.Info("Existing user found", 
			"user_id", existingUser.ID.String(),
			"telegram_id", telegramUser.ID)
		
		// Update user information with latest Telegram data
		updateReq := &models.UpdateUserRequest{
			Username:     stringToStringPtr(telegramUser.Username),
			FirstName:    &telegramUser.FirstName,
			LastName:     stringToStringPtr(telegramUser.LastName),
			LanguageCode: stringToStringPtr(telegramUser.LanguageCode),
			IsPremium:    &telegramUser.IsPremium,
			PhotoURL:     stringToStringPtr(telegramUser.PhotoURL),
		}
		
		updatedUser, err := h.userRepo.UpdateUser(ctx, existingUser.ID, updateReq)
		if err != nil {
			h.logger.Warn("Failed to update existing user", "error", err, "user_id", existingUser.ID.String())
			// Return the existing user even if update failed
			return existingUser, nil
		}
		
		return updatedUser, nil
	}
	
	// Check if error is "user not found", otherwise it's a real error
	if err != storage.ErrUserNotFound {
		return nil, err
	}
	
	// User doesn't exist, create new one
	h.logger.Info("Creating new user", "telegram_id", telegramUser.ID)
	
	createReq := &models.CreateUserRequest{
		TelegramID:   telegramUser.ID,
		Username:     stringToStringPtr(telegramUser.Username),
		FirstName:    telegramUser.FirstName,
		LastName:     stringToStringPtr(telegramUser.LastName),
		LanguageCode: stringToStringPtr(telegramUser.LanguageCode),
		IsPremium:    telegramUser.IsPremium,
		PhotoURL:     stringToStringPtr(telegramUser.PhotoURL),
	}
	
	newUser, err := h.userRepo.CreateUser(ctx, createReq)
	if err != nil {
		return nil, err
	}
	
	h.logger.Info("New user created successfully",
		"user_id", newUser.ID.String(),
		"telegram_id", newUser.TelegramID)
	
	return newUser, nil
}

// Helper functions for string pointer conversions
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringToStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

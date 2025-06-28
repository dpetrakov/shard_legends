package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/auth-service/internal/storage"
)

// AdminHandler handles admin endpoints for token management
type AdminHandler struct {
	logger       *slog.Logger
	tokenStorage storage.TokenStorage
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(logger *slog.Logger, tokenStorage storage.TokenStorage) *AdminHandler {
	return &AdminHandler{
		logger:       logger,
		tokenStorage: tokenStorage,
	}
}

// TokenStatsResponse represents token statistics
type TokenStatsResponse struct {
	ActiveTokenCount int64  `json:"active_token_count"`
	Timestamp        string `json:"timestamp"`
}

// UserTokensResponse represents user's active tokens
type UserTokensResponse struct {
	UserID       string   `json:"user_id"`
	ActiveTokens []string `json:"active_tokens"`
	Count        int      `json:"count"`
}

// CleanupResponse represents cleanup operation result
type CleanupResponse struct {
	CleanedCount int64  `json:"cleaned_count"`
	Timestamp    string `json:"timestamp"`
}

// GetTokenStats returns statistics about active tokens
func (h *AdminHandler) GetTokenStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := h.tokenStorage.GetActiveTokenCount(ctx)
	if err != nil {
		h.logger.Error("Failed to get token count", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get token statistics",
		})
		return
	}

	response := TokenStatsResponse{
		ActiveTokenCount: count,
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserTokens returns active tokens for a specific user
func (h *AdminHandler) GetUserTokens(c *gin.Context) {
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokens, err := h.tokenStorage.GetUserActiveTokens(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user tokens", "error", err, "user_id", userID.String())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user tokens",
		})
		return
	}

	response := UserTokensResponse{
		UserID:       userID.String(),
		ActiveTokens: tokens,
		Count:        len(tokens),
	}

	c.JSON(http.StatusOK, response)
}

// RevokeUserTokens revokes all tokens for a specific user
func (h *AdminHandler) RevokeUserTokens(c *gin.Context) {
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = h.tokenStorage.RevokeUserTokens(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to revoke user tokens", "error", err, "user_id", userID.String())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke user tokens",
		})
		return
	}

	h.logger.Info("User tokens revoked", "user_id", userID.String())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All user tokens have been revoked",
		"user_id": userID.String(),
	})
}

// RevokeToken revokes a specific token
func (h *AdminHandler) RevokeToken(c *gin.Context) {
	jti := c.Param("jti")
	if jti == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token ID (JTI) is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.tokenStorage.RevokeToken(ctx, jti)
	if err != nil {
		h.logger.Error("Failed to revoke token", "error", err, "jti", jti)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke token",
		})
		return
	}

	h.logger.Info("Token revoked", "jti", jti)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token has been revoked",
		"jti":     jti,
	})
}

// CleanupExpiredTokens manually triggers cleanup of expired tokens
func (h *AdminHandler) CleanupExpiredTokens(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cleanedCount, err := h.tokenStorage.CleanupExpiredTokens(ctx)
	if err != nil {
		h.logger.Error("Failed to cleanup expired tokens", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup expired tokens",
		})
		return
	}

	h.logger.Info("Manual token cleanup completed", "cleaned_count", cleanedCount)

	response := CleanupResponse{
		CleanedCount: cleanedCount,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

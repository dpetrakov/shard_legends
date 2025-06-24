package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/inventory-service/internal/auth"
)

// AdminMiddleware provides admin authorization
type AdminMiddleware struct {
	logger *slog.Logger
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(logger *slog.Logger) *AdminMiddleware {
	return &AdminMiddleware{
		logger: logger,
	}
}

// RequireAdmin checks if user has admin permissions
func (m *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// User должен быть аутентифицирован через JWT middleware
		user, exists := auth.GetUserFromContext(c)
		if !exists {
			m.logger.Error("User not found in context for admin check")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// TODO: Implement real admin permission checking
		// For now, this is a placeholder that allows all authenticated users
		// In production, this should check against admin roles/permissions

		isAdmin := m.checkAdminPermissions(user.UserID)
		if !isAdmin {
			m.logger.Warn("User attempted admin operation without permissions",
				"user_id", user.UserID,
				"telegram_id", user.TelegramID)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions for admin operation",
			})
			c.Abort()
			return
		}

		m.logger.Info("Admin access granted",
			"user_id", user.UserID,
			"endpoint", c.Request.URL.Path,
			"method", c.Request.Method)

		c.Next()
	}
}

// checkAdminPermissions checks if user has admin permissions
// This is a placeholder implementation
func (m *AdminMiddleware) checkAdminPermissions(userID string) bool {
	// TODO: Implement real admin permission checking
	// This could check against:
	// - Admin roles in database
	// - Admin permissions cache
	// - External admin service
	// - Hardcoded admin user IDs for development

	// For now, return true to allow development/testing
	// In production, this should be properly implemented
	m.logger.Warn("Using placeholder admin check - implement proper admin permissions", "user_id", userID)
	return true
}

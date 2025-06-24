package auth

import "github.com/gin-gonic/gin"

// UserContext represents authenticated user context
type UserContext struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
}

// GetUserFromContext extracts authenticated user from gin context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*UserContext)
	if !ok {
		return nil, false
	}

	return user, true
}

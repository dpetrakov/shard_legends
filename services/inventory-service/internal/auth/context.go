package auth

// UserContext представляет контекст пользователя из JWT токена
type UserContext struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
}

// GetUserFromContext извлекает пользователя из gin.Context (заглушка для совместимости)
func GetUserFromContext(c interface{}) (*UserContext, bool) {
	// Эта функция не используется, т.к. логика перенесена в JWT middleware
	return nil, false
}

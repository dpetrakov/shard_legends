package models

import (
	"time"
)

// User представляет базовую информацию о пользователе
type User struct {
	ID          string `json:"id"`
	TelegramID  int64  `json:"telegram_id"`
	DisplayName string `json:"display_name"`
}

// VIPStatus представляет VIP статус пользователя
type VIPStatus struct {
	IsActive  bool       `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Clan представляет клан пользователя
type Clan struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ProfileResponse - ответ эндпоинта GET /profile
type ProfileResponse struct {
	UserID            string    `json:"user_id"`
	TelegramID        int64     `json:"telegram_id"`
	DisplayName       string    `json:"display_name"`
	VIPStatus         VIPStatus `json:"vip_status"`
	ProfileImage      string    `json:"profile_image"`
	Level             int       `json:"level"`
	Experience        int       `json:"experience"`
	AchievementsCount int       `json:"achievements_count"`
	Clan              *Clan     `json:"clan,omitempty"`
}
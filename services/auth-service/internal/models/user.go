package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system from Telegram Web App
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TelegramID   int64      `json:"telegram_id" db:"telegram_id"`
	Username     *string    `json:"username,omitempty" db:"username"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     *string    `json:"last_name,omitempty" db:"last_name"`
	LanguageCode *string    `json:"language_code,omitempty" db:"language_code"`
	IsPremium    bool       `json:"is_premium" db:"is_premium"`
	PhotoURL     *string    `json:"photo_url,omitempty" db:"photo_url"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	TelegramID   int64   `json:"telegram_id" validate:"required,min=1"`
	Username     *string `json:"username,omitempty"`
	FirstName    string  `json:"first_name" validate:"required,min=1"`
	LastName     *string `json:"last_name,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
	IsPremium    bool    `json:"is_premium"`
	PhotoURL     *string `json:"photo_url,omitempty"`
}

// UpdateUserRequest represents a request to update an existing user
type UpdateUserRequest struct {
	Username     *string `json:"username,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
	IsPremium    *bool   `json:"is_premium,omitempty"`
	PhotoURL     *string `json:"photo_url,omitempty"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// ToUser converts CreateUserRequest to User model
func (req *CreateUserRequest) ToUser() *User {
	return &User{
		ID:           uuid.New(),
		TelegramID:   req.TelegramID,
		Username:     req.Username,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		LanguageCode: req.LanguageCode,
		IsPremium:    req.IsPremium,
		PhotoURL:     req.PhotoURL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
}
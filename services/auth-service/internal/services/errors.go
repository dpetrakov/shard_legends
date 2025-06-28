package services

import (
	"errors"
	"fmt"
)

// Common service errors
var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
)

// JWTError represents a JWT-related error
type JWTError struct {
	Message string
	Cause   error
}

func (e *JWTError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("JWT error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("JWT error: %s", e.Message)
}

func (e *JWTError) Unwrap() error {
	return e.Cause
}

// TelegramError represents a Telegram validation error
type TelegramError struct {
	Message string
	Cause   error
}

func (e *TelegramError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Telegram error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("Telegram error: %s", e.Message)
}

func (e *TelegramError) Unwrap() error {
	return e.Cause
}

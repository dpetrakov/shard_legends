package auth

import (
	"context"
	"fmt"
)

type contextKey string

const userContextKey contextKey = "user"

type UserContext struct {
	UserID     string
	TelegramID int64
	IsAdmin    bool
}

func WithUser(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

func GetUserID(ctx context.Context) (string, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return "", err
	}
	return user.UserID, nil
}

func GetTelegramID(ctx context.Context) (int64, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return 0, err
	}
	return user.TelegramID, nil
}

func IsAdmin(ctx context.Context) bool {
	user, err := GetUser(ctx)
	if err != nil {
		return false
	}
	return user.IsAdmin
}
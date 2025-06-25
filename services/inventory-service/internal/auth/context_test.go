package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserContext_Structure(t *testing.T) {
	// Arrange & Act
	user := &UserContext{
		UserID:     "test-user-id",
		TelegramID: 123456789,
	}

	// Assert
	assert.Equal(t, "test-user-id", user.UserID)
	assert.Equal(t, int64(123456789), user.TelegramID)
}

func TestGetUserFromContext(t *testing.T) {
	// Arrange
	mockContext := struct{}{}

	// Act
	user, exists := GetUserFromContext(mockContext)

	// Assert
	assert.Nil(t, user)
	assert.False(t, exists)
}

func TestGetUserFromContext_WithNilContext(t *testing.T) {
	// Act
	user, exists := GetUserFromContext(nil)

	// Assert
	assert.Nil(t, user)
	assert.False(t, exists)
}
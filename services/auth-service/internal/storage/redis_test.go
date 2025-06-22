package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis creates a test Redis client for testing
func setupTestRedis(t *testing.T) (*RedisTokenStorage, func()) {
	// Use Redis database 15 for testing (to avoid conflicts)
	redisURL := "redis://localhost:6379/15"
	
	// Check if Redis is available
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Skipf("Invalid Redis URL: %v", err)
	}
	
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skipf("Redis not available for testing: %v", err)
	}
	client.Close()

	// Create logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce log noise in tests
	}))

	// Create Redis token storage
	storage, err := NewRedisTokenStorage(redisURL, 5, logger, nil) // Pass nil for metrics in tests
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		ctx := context.Background()
		
		// Clear test database
		storage.client.FlushDB(ctx)
		storage.Close()
	}

	// Clear any existing data
	storage.client.FlushDB(context.Background())

	return storage, cleanup
}

func TestRedisTokenStorage_StoreActiveToken(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	jti := uuid.New().String()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Store token
	err := storage.StoreActiveToken(ctx, jti, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Verify token is stored
	tokenInfo, err := storage.GetTokenInfo(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to get token info: %v", err)
	}

	if tokenInfo.JTI != jti {
		t.Errorf("Expected JTI %s, got %s", jti, tokenInfo.JTI)
	}

	if tokenInfo.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, tokenInfo.UserID)
	}

	if tokenInfo.TelegramID != telegramID {
		t.Errorf("Expected TelegramID %d, got %d", telegramID, tokenInfo.TelegramID)
	}

	if tokenInfo.IsRevoked {
		t.Error("Token should not be revoked")
	}
}

func TestRedisTokenStorage_IsTokenActive(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	jti := uuid.New().String()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Token should not be active before storing
	isActive, err := storage.IsTokenActive(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to check token status: %v", err)
	}
	if isActive {
		t.Error("Token should not be active before storing")
	}

	// Store token
	err = storage.StoreActiveToken(ctx, jti, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Token should be active after storing
	isActive, err = storage.IsTokenActive(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to check token status: %v", err)
	}
	if !isActive {
		t.Error("Token should be active after storing")
	}
}

func TestRedisTokenStorage_RevokeToken(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	jti := uuid.New().String()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Store token
	err := storage.StoreActiveToken(ctx, jti, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Token should be active
	isActive, err := storage.IsTokenActive(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to check token status: %v", err)
	}
	if !isActive {
		t.Error("Token should be active before revoking")
	}

	// Revoke token
	err = storage.RevokeToken(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Token should not be active after revoking
	isActive, err = storage.IsTokenActive(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to check token status: %v", err)
	}
	if isActive {
		t.Error("Token should not be active after revoking")
	}

	// Check revocation status
	isRevoked, err := storage.IsTokenRevoked(ctx, jti)
	if err != nil {
		t.Fatalf("Failed to check revocation status: %v", err)
	}
	if !isRevoked {
		t.Error("Token should be marked as revoked")
	}
}

func TestRedisTokenStorage_GetUserActiveTokens(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Store multiple tokens for the same user
	jti1 := uuid.New().String()
	jti2 := uuid.New().String()
	jti3 := uuid.New().String()

	err := storage.StoreActiveToken(ctx, jti1, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 1: %v", err)
	}

	err = storage.StoreActiveToken(ctx, jti2, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 2: %v", err)
	}

	err = storage.StoreActiveToken(ctx, jti3, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 3: %v", err)
	}

	// Get user tokens
	tokens, err := storage.GetUserActiveTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user tokens: %v", err)
	}

	if len(tokens) != 3 {
		t.Errorf("Expected 3 active tokens, got %d", len(tokens))
	}

	// Revoke one token
	err = storage.RevokeToken(ctx, jti2)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Should have 2 active tokens now
	tokens, err = storage.GetUserActiveTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user tokens: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 active tokens after revocation, got %d", len(tokens))
	}
}

func TestRedisTokenStorage_RevokeUserTokens(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Store multiple tokens for the same user
	jti1 := uuid.New().String()
	jti2 := uuid.New().String()

	err := storage.StoreActiveToken(ctx, jti1, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 1: %v", err)
	}

	err = storage.StoreActiveToken(ctx, jti2, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 2: %v", err)
	}

	// Verify tokens are active
	tokens, err := storage.GetUserActiveTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user tokens: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 active tokens, got %d", len(tokens))
	}

	// Revoke all user tokens
	err = storage.RevokeUserTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to revoke user tokens: %v", err)
	}

	// Should have no active tokens
	tokens, err = storage.GetUserActiveTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user tokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("Expected 0 active tokens after revoking all, got %d", len(tokens))
	}
}

func TestRedisTokenStorage_CleanupExpiredTokens(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()
	telegramID := int64(123456789)

	// Store token that expires soon
	jti1 := uuid.New().String()
	shortExpiry := time.Now().Add(100 * time.Millisecond)

	err := storage.StoreActiveToken(ctx, jti1, userID, telegramID, shortExpiry)
	if err != nil {
		t.Fatalf("Failed to store short-lived token: %v", err)
	}

	// Store token with normal expiry
	jti2 := uuid.New().String()
	normalExpiry := time.Now().Add(time.Hour)

	err = storage.StoreActiveToken(ctx, jti2, userID, telegramID, normalExpiry)
	if err != nil {
		t.Fatalf("Failed to store normal token: %v", err)
	}

	// Wait for first token to expire
	time.Sleep(200 * time.Millisecond)

	// Run cleanup
	cleanedCount, err := storage.CleanupExpiredTokens(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup expired tokens: %v", err)
	}

	// Should have cleaned at least the expired token
	if cleanedCount == 0 {
		t.Error("Expected at least 1 token to be cleaned up")
	}

	// Normal token should still be active
	isActive, err := storage.IsTokenActive(ctx, jti2)
	if err != nil {
		t.Fatalf("Failed to check token status: %v", err)
	}
	if !isActive {
		t.Error("Normal token should still be active")
	}
}

func TestRedisTokenStorage_Health(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	// Health check should pass
	err := storage.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestRedisTokenStorage_GetActiveTokenCount(t *testing.T) {
	storage, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()
	telegramID := int64(123456789)
	expiresAt := time.Now().Add(time.Hour)

	// Initially should have 0 tokens
	count, err := storage.GetActiveTokenCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get token count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tokens initially, got %d", count)
	}

	// Store some tokens
	jti1 := uuid.New().String()
	jti2 := uuid.New().String()

	err = storage.StoreActiveToken(ctx, jti1, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 1: %v", err)
	}

	err = storage.StoreActiveToken(ctx, jti2, userID, telegramID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store token 2: %v", err)
	}

	// Should have 2 active tokens
	count, err = storage.GetActiveTokenCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get token count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 active tokens, got %d", count)
	}

	// Revoke one token
	err = storage.RevokeToken(ctx, jti1)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Should have 1 active token
	count, err = storage.GetActiveTokenCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get token count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 active token after revocation, got %d", count)
	}
}
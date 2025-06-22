package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TokenStorage defines the interface for token management
type TokenStorage interface {
	// StoreActiveToken stores an active JWT token with TTL
	StoreActiveToken(ctx context.Context, jti string, userID uuid.UUID, telegramID int64, expiresAt time.Time) error

	// RevokeToken marks a token as revoked
	RevokeToken(ctx context.Context, jti string) error

	// IsTokenRevoked checks if a token has been revoked
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)

	// IsTokenActive checks if a token is active (not revoked and not expired)
	IsTokenActive(ctx context.Context, jti string) (bool, error)

	// GetTokenInfo retrieves token information
	GetTokenInfo(ctx context.Context, jti string) (*TokenInfo, error)

	// CleanupExpiredTokens removes expired tokens from storage
	CleanupExpiredTokens(ctx context.Context) (int64, error)

	// GetActiveTokenCount returns the number of active tokens
	GetActiveTokenCount(ctx context.Context) (int64, error)

	// GetUserActiveTokens returns all active tokens for a specific user
	GetUserActiveTokens(ctx context.Context, userID uuid.UUID) ([]string, error)

	// RevokeUserTokens revokes all tokens for a specific user
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error

	// Health checks the Redis connection health
	Health(ctx context.Context) error

	// Close closes the Redis connection
	Close() error
}

// TokenInfo represents stored token information
type TokenInfo struct {
	JTI        string    `json:"jti"`
	UserID     uuid.UUID `json:"user_id"`
	TelegramID int64     `json:"telegram_id"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsRevoked  bool      `json:"is_revoked"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// RedisTokenStorage implements TokenStorage interface using Redis
type RedisTokenStorage struct {
	client *redis.Client
	logger *slog.Logger
}

// Redis key prefixes
const (
	// Active tokens: active_token:{jti} -> TokenInfo JSON
	activeTokenPrefix = "active_token:"
	// Revoked tokens: revoked_token:{jti} -> revocation timestamp
	revokedTokenPrefix = "revoked_token:"
	// User tokens index: user_tokens:{user_id} -> Set of JTIs
	userTokensPrefix = "user_tokens:"
	// Token expiry index: token_expiry:{unix_timestamp} -> Set of JTIs
	tokenExpiryPrefix = "token_expiry:"
)

// NewRedisTokenStorage creates a new Redis token storage instance
func NewRedisTokenStorage(redisURL string, maxConns int, logger *slog.Logger) (*RedisTokenStorage, error) {
	// Parse Redis URL
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure connection pool
	opts.PoolSize = maxConns
	opts.MinIdleConns = 1
	opts.ConnMaxLifetime = 30 * time.Minute
	opts.ConnMaxIdleTime = 5 * time.Minute

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established",
		slog.String("max_conns", fmt.Sprintf("%d", maxConns)),
	)

	return &RedisTokenStorage{
		client: client,
		logger: logger,
	}, nil
}

// StoreActiveToken stores an active JWT token with TTL
func (r *RedisTokenStorage) StoreActiveToken(ctx context.Context, jti string, userID uuid.UUID, telegramID int64, expiresAt time.Time) error {
	tokenInfo := &TokenInfo{
		JTI:        jti,
		UserID:     userID,
		TelegramID: telegramID,
		IssuedAt:   time.Now(),
		ExpiresAt:  expiresAt,
		IsRevoked:  false,
	}

	// Serialize token info
	tokenData, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	// Calculate TTL (add buffer for cleanup)
	ttl := time.Until(expiresAt) + time.Hour

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store active token
	pipe.Set(ctx, activeTokenPrefix+jti, tokenData, ttl)

	// Add to user tokens index
	pipe.SAdd(ctx, userTokensPrefix+userID.String(), jti)
	pipe.Expire(ctx, userTokensPrefix+userID.String(), ttl)

	// Add to expiry index for cleanup
	expiryKey := tokenExpiryPrefix + strconv.FormatInt(expiresAt.Unix(), 10)
	pipe.SAdd(ctx, expiryKey, jti)
	pipe.ExpireAt(ctx, expiryKey, expiresAt.Add(time.Hour))

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to store active token",
			slog.String("error", err.Error()),
			slog.String("jti", jti),
		)
		return fmt.Errorf("failed to store active token: %w", err)
	}

	r.logger.Info("Active token stored successfully",
		slog.String("jti", jti),
		slog.String("user_id", userID.String()),
		slog.Time("expires_at", expiresAt),
	)

	return nil
}

// RevokeToken marks a token as revoked
func (r *RedisTokenStorage) RevokeToken(ctx context.Context, jti string) error {
	now := time.Now()

	// Get token info to determine TTL
	tokenInfo, err := r.GetTokenInfo(ctx, jti)
	if err != nil {
		return fmt.Errorf("failed to get token info for revocation: %w", err)
	}

	// Calculate TTL for revoked token (until original expiry)
	ttl := time.Until(tokenInfo.ExpiresAt)
	if ttl <= 0 {
		// Token already expired, no need to revoke
		return nil
	}

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store revoked token
	pipe.Set(ctx, revokedTokenPrefix+jti, now.Unix(), ttl)

	// Update token info to mark as revoked
	tokenInfo.IsRevoked = true
	tokenInfo.RevokedAt = &now

	tokenData, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated token info: %w", err)
	}

	pipe.Set(ctx, activeTokenPrefix+jti, tokenData, ttl)

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to revoke token",
			slog.String("error", err.Error()),
			slog.String("jti", jti),
		)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	r.logger.Info("Token revoked successfully",
		slog.String("jti", jti),
		slog.Time("revoked_at", now),
	)

	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (r *RedisTokenStorage) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	exists, err := r.client.Exists(ctx, revokedTokenPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation status: %w", err)
	}

	return exists > 0, nil
}

// IsTokenActive checks if a token is active (not revoked and not expired)
func (r *RedisTokenStorage) IsTokenActive(ctx context.Context, jti string) (bool, error) {
	// Check if token exists in active tokens
	exists, err := r.client.Exists(ctx, activeTokenPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	if exists == 0 {
		return false, nil // Token not found or expired
	}

	// Check if token is revoked
	isRevoked, err := r.IsTokenRevoked(ctx, jti)
	if err != nil {
		return false, err
	}

	return !isRevoked, nil
}

// GetTokenInfo retrieves token information
func (r *RedisTokenStorage) GetTokenInfo(ctx context.Context, jti string) (*TokenInfo, error) {
	tokenData, err := r.client.Get(ctx, activeTokenPrefix+jti).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found: %s", jti)
		}
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	var tokenInfo TokenInfo
	if err := json.Unmarshal([]byte(tokenData), &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token info: %w", err)
	}

	return &tokenInfo, nil
}

// CleanupExpiredTokens removes expired tokens from storage
func (r *RedisTokenStorage) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	now := time.Now()
	cleanedCount := int64(0)

	// Find expired token keys using scan
	var cursor uint64
	var expiredKeys []string

	for {
		// Scan for active token keys
		keys, nextCursor, err := r.client.Scan(ctx, cursor, activeTokenPrefix+"*", 100).Result()
		if err != nil {
			return cleanedCount, fmt.Errorf("failed to scan for expired tokens: %w", err)
		}

		// Check each token's expiration time
		for _, key := range keys {
			// Get token info to check expires_at
			tokenData, err := r.client.Get(ctx, key).Result()
			if err != nil {
				if err == redis.Nil {
					// Key doesn't exist, add to cleanup list
					expiredKeys = append(expiredKeys, key)
				}
				continue // Skip on other errors
			}

			var tokenInfo TokenInfo
			if err := json.Unmarshal([]byte(tokenData), &tokenInfo); err != nil {
				// Invalid token data, add to cleanup list
				expiredKeys = append(expiredKeys, key)
				continue
			}

			// Check if token is expired based on expires_at field
			if tokenInfo.ExpiresAt.Before(now) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// Remove expired keys in batches
	if len(expiredKeys) > 0 {
		pipe := r.client.Pipeline()

		for _, key := range expiredKeys {
			pipe.Del(ctx, key)
			
			// Extract JTI from key
			jti := key[len(activeTokenPrefix):]
			
			// Also clean up related keys
			pipe.Del(ctx, revokedTokenPrefix+jti)
		}

		results, err := pipe.Exec(ctx)
		if err != nil {
			return cleanedCount, fmt.Errorf("failed to cleanup expired tokens: %w", err)
		}

		// Count successful deletions
		for _, result := range results {
			if cmd, ok := result.(*redis.IntCmd); ok {
				cleanedCount += cmd.Val()
			}
		}
	}

	// Clean up expired revoked tokens
	cursor = 0
	var expiredRevokedKeys []string

	for {
		// Scan for revoked token keys
		keys, nextCursor, err := r.client.Scan(ctx, cursor, revokedTokenPrefix+"*", 100).Result()
		if err != nil {
			break // Continue with what we have
		}

		// Check each revoked token's expiration time by getting corresponding active token info
		for _, key := range keys {
			// Extract JTI from revoked token key
			jti := key[len(revokedTokenPrefix):]
			
			// Get token info from active token to check expires_at
			activeKey := activeTokenPrefix + jti
			tokenData, err := r.client.Get(ctx, activeKey).Result()
			if err != nil {
				// If active token doesn't exist, revoked token can be cleaned
				expiredRevokedKeys = append(expiredRevokedKeys, key)
				continue
			}

			var tokenInfo TokenInfo
			if err := json.Unmarshal([]byte(tokenData), &tokenInfo); err != nil {
				// Invalid token data, clean up revoked token
				expiredRevokedKeys = append(expiredRevokedKeys, key)
				continue
			}

			// If token is expired, clean up the revoked token entry
			if tokenInfo.ExpiresAt.Before(now) {
				expiredRevokedKeys = append(expiredRevokedKeys, key)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// Remove expired revoked tokens
	if len(expiredRevokedKeys) > 0 {
		pipe := r.client.Pipeline()
		for _, key := range expiredRevokedKeys {
			pipe.Del(ctx, key)
		}

		results, err := pipe.Exec(ctx)
		if err == nil {
			// Count successful deletions of revoked tokens
			for _, result := range results {
				if cmd, ok := result.(*redis.IntCmd); ok {
					cleanedCount += cmd.Val()
				}
			}
		}
	}

	// Clean up expired expiry index keys
	expiryPattern := tokenExpiryPrefix + "*"
	cursor = 0

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, expiryPattern, 100).Result()
		if err != nil {
			break // Continue with what we have
		}

		pipe := r.client.Pipeline()
		for _, key := range keys {
			// Extract timestamp from key
			timestampStr := key[len(tokenExpiryPrefix):]
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				continue
			}

			// If timestamp is in the past, remove the key
			if time.Unix(timestamp, 0).Before(now) {
				pipe.Del(ctx, key)
			}
		}

		pipe.Exec(ctx)

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	r.logger.Info("Token cleanup completed",
		slog.Int64("cleaned_count", cleanedCount),
	)

	return cleanedCount, nil
}

// GetActiveTokenCount returns the number of active tokens
func (r *RedisTokenStorage) GetActiveTokenCount(ctx context.Context) (int64, error) {
	var cursor uint64
	var count int64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, activeTokenPrefix+"*", 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan active tokens: %w", err)
		}

		// Check which keys are actually not revoked
		for _, key := range keys {
			jti := key[len(activeTokenPrefix):]
			isActive, err := r.IsTokenActive(ctx, jti)
			if err == nil && isActive {
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// GetUserActiveTokens returns all active tokens for a specific user
func (r *RedisTokenStorage) GetUserActiveTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	jtis, err := r.client.SMembers(ctx, userTokensPrefix+userID.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	var activeJTIs []string
	for _, jti := range jtis {
		isActive, err := r.IsTokenActive(ctx, jti)
		if err == nil && isActive {
			activeJTIs = append(activeJTIs, jti)
		}
	}

	return activeJTIs, nil
}

// RevokeUserTokens revokes all tokens for a specific user
func (r *RedisTokenStorage) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	jtis, err := r.GetUserActiveTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user tokens for revocation: %w", err)
	}

	var revokeErrors []error
	revokedCount := 0

	for _, jti := range jtis {
		if err := r.RevokeToken(ctx, jti); err != nil {
			revokeErrors = append(revokeErrors, err)
		} else {
			revokedCount++
		}
	}

	r.logger.Info("User tokens revocation completed",
		slog.String("user_id", userID.String()),
		slog.Int("revoked_count", revokedCount),
		slog.Int("errors_count", len(revokeErrors)),
	)

	if len(revokeErrors) > 0 {
		return fmt.Errorf("failed to revoke %d tokens", len(revokeErrors))
	}

	return nil
}

// Health checks the Redis connection health
func (r *RedisTokenStorage) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisTokenStorage) Close() error {
	if err := r.client.Close(); err != nil {
		r.logger.Error("Failed to close Redis connection", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("Redis connection closed")
	return nil
}
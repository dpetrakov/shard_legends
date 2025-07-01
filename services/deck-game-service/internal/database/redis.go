package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shard-legends/deck-game-service/pkg/metrics"
)

// RedisDB wraps redis.Client for JWT auth validation only
type RedisDB struct {
	client  *redis.Client
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// NewRedisDB creates a new Redis connection for JWT auth validation
func NewRedisDB(redisURL string, maxConns int, logger *slog.Logger, metrics *metrics.Metrics) (*RedisDB, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure connection pool
	opt.PoolSize = maxConns
	opt.MinIdleConns = 1
	opt.ConnMaxLifetime = time.Hour
	opt.ConnMaxIdleTime = time.Minute * 30
	opt.PoolTimeout = time.Second * 10

	// Create Redis client
	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established",
		"pool_size", maxConns,
		"database", opt.DB)

	return &RedisDB{
		client:  client,
		logger:  logger,
		metrics: metrics,
	}, nil
}

// Client returns the underlying Redis client
func (r *RedisDB) Client() *redis.Client {
	return r.client
}

// Health checks the Redis connection health
func (r *RedisDB) Health(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		r.logger.Error("Redis health check failed", "error", err)
		return err
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisDB) Close() {
	if r.client != nil {
		r.logger.Info("Closing Redis connection")
		r.client.Close()
	}
}

// UpdateMetrics updates Redis connection metrics
func (r *RedisDB) UpdateMetrics() {
	if r.metrics != nil {
		stats := r.client.PoolStats()
		r.metrics.RedisConnections.Set(float64(stats.TotalConns))
	}
}

// IsTokenRevoked checks if a JWT token is revoked
func (r *RedisDB) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	result, err := r.client.Exists(ctx, fmt.Sprintf("revoked:%s", tokenID)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return result > 0, nil
}

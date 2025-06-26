package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shard-legends/production-service/internal/config"
	"github.com/shard-legends/production-service/pkg/logger"
	"go.uber.org/zap"
)

type RedisClient struct {
	client     *redis.Client
	authClient *redis.Client // Клиент для JWT revocation (база 0)
}

func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	opt.MaxRetries = 3
	opt.PoolSize = cfg.MaxConnections
	opt.ReadTimeout = cfg.ReadTimeout
	opt.WriteTimeout = cfg.WriteTimeout

	client := redis.NewClient(opt)

	// Создание клиента для Auth базы (JWT revocation)
	authOpt, err := redis.ParseURL(cfg.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis auth URL: %w", err)
	}

	authOpt.MaxRetries = 3
	authOpt.PoolSize = cfg.MaxConnections
	authOpt.ReadTimeout = cfg.ReadTimeout
	authOpt.WriteTimeout = cfg.WriteTimeout

	authClient := redis.NewClient(authOpt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверка основного клиента
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	// Проверка auth клиента
	if err := authClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis auth: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.Int("max_connections", cfg.MaxConnections),
		zap.Duration("read_timeout", cfg.ReadTimeout),
		zap.Duration("write_timeout", cfg.WriteTimeout),
		zap.String("cache_url", cfg.URL),
		zap.String("auth_url", cfg.AuthURL),
	)

	return &RedisClient{
		client:     client,
		authClient: authClient,
	}, nil
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	var errs []error

	if r.client != nil {
		if err := r.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis cache connection: %w", err))
		}
	}

	if r.authClient != nil {
		if err := r.authClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis auth connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("redis close errors: %v", errs)
	}

	logger.Info("Redis connections closed")
	return nil
}

func (r *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Проверка основного клиента
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis cache health check failed: %w", err)
	}

	// Проверка auth клиента
	if err := r.authClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis auth health check failed: %w", err)
	}

	return nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence of keys: %w", err)
	}
	return count, nil
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx key %s: %w", key, err)
	}
	return ok, nil
}

func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ttl for key %s: %w", key, err)
	}
	return ttl, nil
}

// IsJWTRevoked проверяет отозван ли JWT токен в auth базе Redis
func (r *RedisClient) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	count, err := r.authClient.Exists(ctx, fmt.Sprintf("revoked:%s", jti)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check jwt revocation for jti %s: %w", jti, err)
	}
	return count > 0, nil
}

package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// RedisDB wraps redis.Client with additional functionality
type RedisDB struct {
	client     *redis.Client // Основной клиент для кеша
	authClient *redis.Client // Клиент для JWT revocation (база 0)
	logger     *slog.Logger
	metrics    *metrics.Metrics
}

// NewRedisDB creates a new Redis client with dual connections
func NewRedisDB(redisURL, redisAuthURL string, maxConns int, logger *slog.Logger, metricsCollector *metrics.Metrics) (*RedisDB, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure connection pool
	opt.PoolSize = maxConns
	opt.MinIdleConns = 1
	opt.MaxIdleConns = maxConns / 2
	opt.ConnMaxLifetime = time.Hour
	opt.ConnMaxIdleTime = time.Minute * 30
	opt.PoolTimeout = time.Second * 30
	opt.ReadTimeout = time.Second * 10
	opt.WriteTimeout = time.Second * 10

	// Create main Redis client
	client := redis.NewClient(opt)

	// Parse Redis Auth URL
	authOpt, err := redis.ParseURL(redisAuthURL)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to parse Redis Auth URL: %w", err)
	}

	// Configure auth client connection pool
	authOpt.PoolSize = maxConns
	authOpt.MinIdleConns = 1
	authOpt.MaxIdleConns = maxConns / 2
	authOpt.ConnMaxLifetime = time.Hour
	authOpt.ConnMaxIdleTime = time.Minute * 30
	authOpt.PoolTimeout = time.Second * 30
	authOpt.ReadTimeout = time.Second * 10
	authOpt.WriteTimeout = time.Second * 10

	// Create auth Redis client
	authClient := redis.NewClient(authOpt)

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		authClient.Close()
		return nil, fmt.Errorf("failed to ping Redis cache: %w", err)
	}

	if err := authClient.Ping(ctx).Err(); err != nil {
		client.Close()
		authClient.Close()
		return nil, fmt.Errorf("failed to ping Redis auth: %w", err)
	}

	rdb := &RedisDB{
		client:     client,
		authClient: authClient,
		logger:     logger,
		metrics:    metricsCollector,
	}

	// Update metrics
	if metricsCollector != nil {
		metricsCollector.RedisConnections.Set(float64(maxConns))
		metricsCollector.UpdateDependencyHealth("redis", true)
	}

	logger.Info("Redis connections established",
		"max_conns", maxConns,
		"cache_addr", opt.Addr,
		"cache_db", opt.DB,
		"auth_addr", authOpt.Addr,
		"auth_db", authOpt.DB,
	)

	return rdb, nil
}

// Client returns the underlying redis.Client
func (r *RedisDB) Client() *redis.Client {
	return r.client
}

// Health checks the health of both Redis connections
func (r *RedisDB) Health(ctx context.Context) error {
	// Check main cache connection
	if err := r.client.Ping(ctx).Err(); err != nil {
		if r.metrics != nil {
			r.metrics.UpdateDependencyHealth("redis", false)
		}
		return fmt.Errorf("redis cache health check failed: %w", err)
	}

	// Check auth connection
	if err := r.authClient.Ping(ctx).Err(); err != nil {
		if r.metrics != nil {
			r.metrics.UpdateDependencyHealth("redis", false)
		}
		return fmt.Errorf("redis auth health check failed: %w", err)
	}

	if r.metrics != nil {
		r.metrics.UpdateDependencyHealth("redis", true)

		// Update connection stats
		stats := r.client.PoolStats()
		r.metrics.RedisConnections.Set(float64(stats.TotalConns))
	}

	return nil
}

// Close closes both Redis clients
func (r *RedisDB) Close() error {
	var errs []error

	if r.client != nil {
		if err := r.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis cache: %w", err))
		}
	}

	if r.authClient != nil {
		if err := r.authClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis auth: %w", err))
		}
	}

	if r.logger != nil {
		r.logger.Info("Redis connections closed")
	}

	if r.metrics != nil {
		r.metrics.RedisConnections.Set(0)
		r.metrics.UpdateDependencyHealth("redis", false)
	}

	if len(errs) > 0 {
		return fmt.Errorf("redis close errors: %v", errs)
	}

	return nil
}

// Stats returns Redis connection pool statistics
func (r *RedisDB) Stats() map[string]interface{} {
	if r.client == nil {
		return map[string]interface{}{
			"status": "disconnected",
		}
	}

	stats := r.client.PoolStats()
	return map[string]interface{}{
		"status":      "connected",
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
	}
}

// Set sets a key-value pair with optional expiration
func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RedisCommandDuration.WithLabelValues("set").Observe(time.Since(start).Seconds())
		}
	}()

	err := r.client.Set(ctx, key, value, expiration).Err()

	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.RedisCommandsTotal.WithLabelValues("set", status).Inc()
	}

	return err
}

// Get gets a value by key
func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RedisCommandDuration.WithLabelValues("get").Observe(time.Since(start).Seconds())
		}
	}()

	result, err := r.client.Get(ctx, key).Result()

	if r.metrics != nil {
		status := "success"
		if err != nil {
			if err == redis.Nil {
				status = "not_found"
			} else {
				status = "error"
			}
		}
		r.metrics.RedisCommandsTotal.WithLabelValues("get", status).Inc()
	}

	return result, err
}

// Del deletes keys
func (r *RedisDB) Del(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RedisCommandDuration.WithLabelValues("del").Observe(time.Since(start).Seconds())
		}
	}()

	result, err := r.client.Del(ctx, keys...).Result()

	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.RedisCommandsTotal.WithLabelValues("del", status).Inc()
	}

	return result, err
}

// Keys finds keys matching a pattern
func (r *RedisDB) Keys(ctx context.Context, pattern string) ([]string, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RedisCommandDuration.WithLabelValues("keys").Observe(time.Since(start).Seconds())
		}
	}()

	result, err := r.client.Keys(ctx, pattern).Result()

	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.RedisCommandsTotal.WithLabelValues("keys", status).Inc()
	}

	return result, err
}

// IsJWTRevoked проверяет отозван ли JWT токен в auth базе Redis
func (r *RedisDB) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RedisCommandDuration.WithLabelValues("exists").Observe(time.Since(start).Seconds())
		}
	}()

	count, err := r.authClient.Exists(ctx, fmt.Sprintf("revoked:%s", jti)).Result()

	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.RedisCommandsTotal.WithLabelValues("exists", status).Inc()
	}

	if err != nil {
		return false, fmt.Errorf("failed to check jwt revocation for jti %s: %w", jti, err)
	}

	return count > 0, nil
}

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
	client  *redis.Client
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// NewRedisDB creates a new Redis client
func NewRedisDB(redisURL string, maxConns int, logger *slog.Logger, metricsCollector *metrics.Metrics) (*RedisDB, error) {
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

	// Create Redis client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	rdb := &RedisDB{
		client:  client,
		logger:  logger,
		metrics: metricsCollector,
	}

	// Update metrics
	if metricsCollector != nil {
		metricsCollector.RedisConnections.Set(float64(maxConns))
		metricsCollector.UpdateDependencyHealth("redis", true)
	}

	logger.Info("Redis connection established",
		"max_conns", maxConns,
		"addr", opt.Addr,
		"db", opt.DB,
	)

	return rdb, nil
}

// Client returns the underlying redis.Client
func (r *RedisDB) Client() *redis.Client {
	return r.client
}

// Health checks the health of the Redis connection
func (r *RedisDB) Health(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		if r.metrics != nil {
			r.metrics.UpdateDependencyHealth("redis", false)
		}
		return err
	}

	if r.metrics != nil {
		r.metrics.UpdateDependencyHealth("redis", true)
		
		// Update connection stats
		stats := r.client.PoolStats()
		r.metrics.RedisConnections.Set(float64(stats.TotalConns))
	}

	return nil
}

// Close closes the Redis client
func (r *RedisDB) Close() error {
	if r.client != nil {
		err := r.client.Close()
		r.logger.Info("Redis connection closed")
		
		if r.metrics != nil {
			r.metrics.RedisConnections.Set(0)
			r.metrics.UpdateDependencyHealth("redis", false)
		}
		
		return err
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
		"status":       "connected",
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
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
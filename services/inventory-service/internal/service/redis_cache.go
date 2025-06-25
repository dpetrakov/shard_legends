package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shard-legends/inventory-service/internal/database"
)

// redisCache implements CacheInterface
type redisCache struct {
	redis *database.RedisDB
}

// NewRedisCache creates a new Redis cache implementation
func NewRedisCache(redis *database.RedisDB) CacheInterface {
	return &redisCache{
		redis: redis,
	}
}

// Get retrieves a value from cache
func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	result, err := c.redis.Get(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(result), value)
}

// Set stores a value in cache
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return c.redis.Set(ctx, key, data, ttl)
}

// Delete removes a value from cache
func (c *redisCache) Delete(ctx context.Context, key string) error {
	_, err := c.redis.Del(ctx, key)
	return err
}

// DeletePattern removes all keys matching a pattern
func (c *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.redis.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		_, err := c.redis.Del(ctx, keys...)
		return err
	}
	
	return nil
}
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCache implements CacheInterface
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache implementation
func NewRedisCache(client *redis.Client) CacheInterface {
	return &redisCache{
		client: client,
	}
}

// Get retrieves a value from cache
func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	result, err := c.client.Get(ctx, key).Result()
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
	
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from cache
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	
	return nil
}
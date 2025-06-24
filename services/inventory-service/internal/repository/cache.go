package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// redisCache implements Cache interface using Redis
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis-based cache
func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{
		client: client,
	}
}

// Get retrieves a value from cache
func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.New("key not found")
		}
		return errors.Wrap(err, "failed to get from cache")
	}
	
	// Try to unmarshal as JSON first
	if err := json.Unmarshal(data, value); err != nil {
		// If it's a []byte, copy directly
		if v, ok := value.(*[]byte); ok {
			*v = data
			return nil
		}
		return errors.Wrap(err, "failed to unmarshal cache value")
	}
	
	return nil
}

// Set stores a value in cache with TTL
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var data []byte
	var err error
	
	// If it's already []byte, use it directly
	if v, ok := value.([]byte); ok {
		data = v
	} else {
		// Otherwise, marshal to JSON
		data, err = json.Marshal(value)
		if err != nil {
			return errors.Wrap(err, "failed to marshal cache value")
		}
	}
	
	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return errors.Wrap(err, "failed to set cache value")
	}
	
	return nil
}

// Delete removes a value from cache
func (c *redisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return errors.Wrap(err, "failed to delete from cache")
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	// Use SCAN to find all matching keys
	var cursor uint64
	var keys []string
	
	for {
		var err error
		var batch []string
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return errors.Wrap(err, "failed to scan keys")
		}
		
		keys = append(keys, batch...)
		
		if cursor == 0 {
			break
		}
	}
	
	// Delete keys in batches
	if len(keys) > 0 {
		// Split into batches of 1000 keys
		batchSize := 1000
		for i := 0; i < len(keys); i += batchSize {
			end := i + batchSize
			if end > len(keys) {
				end = len(keys)
			}
			
			err := c.client.Del(ctx, keys[i:end]...).Err()
			if err != nil {
				return errors.Wrap(err, "failed to delete keys")
			}
		}
	}
	
	return nil
}

// memoryCache implements Cache interface using in-memory storage (for testing)
type memoryCache struct {
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache (mainly for testing)
func NewMemoryCache() Cache {
	return &memoryCache{
		data: make(map[string]cacheEntry),
	}
}

// Get retrieves a value from cache
func (c *memoryCache) Get(ctx context.Context, key string, value interface{}) error {
	entry, exists := c.data[key]
	if !exists {
		return errors.New("key not found")
	}
	
	// Check if expired
	if time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return errors.New("key expired")
	}
	
	// Unmarshal value
	if v, ok := value.(*[]byte); ok {
		*v = entry.value
		return nil
	}
	
	return json.Unmarshal(entry.value, value)
}

// Set stores a value in cache with TTL
func (c *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var data []byte
	var err error
	
	if v, ok := value.([]byte); ok {
		data = v
	} else {
		data, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}
	
	c.data[key] = cacheEntry{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	
	return nil
}

// Delete removes a value from cache
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *memoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// Simple implementation - just delete keys that start with pattern
	prefix := pattern
	if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
		prefix = prefix[:len(prefix)-1]
	}
	
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
	
	return nil
}

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build creates a cache key from components
func (b *CacheKeyBuilder) Build(components ...string) string {
	if b.prefix != "" {
		components = append([]string{b.prefix}, components...)
	}
	
	key := ""
	for i, comp := range components {
		if i > 0 {
			key += ":"
		}
		key += comp
	}
	
	return key
}

// UserInventoryKey builds a key for user inventory cache
func (b *CacheKeyBuilder) UserInventoryKey(userID, sectionID, itemID, collectionID, qualityLevelID string) string {
	return fmt.Sprintf("inventory:%s:%s:%s:%s:%s", userID, sectionID, itemID, collectionID, qualityLevelID)
}
package adapters

import (
	"context"
	"time"

	"github.com/shard-legends/production-service/internal/database"
	"github.com/shard-legends/production-service/internal/storage"
)

// CacheAdapter адаптирует database.RedisClient для storage.CacheInterface
type CacheAdapter struct {
	redis *database.RedisClient
}

// NewCacheAdapter создает новый адаптер для Redis
func NewCacheAdapter(redis *database.RedisClient) storage.CacheInterface {
	return &CacheAdapter{redis: redis}
}

// Get получает значение по ключу
func (a *CacheAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.redis.Get(ctx, key)
}

// Set устанавливает значение с TTL
func (a *CacheAdapter) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return a.redis.Set(ctx, key, value, ttl)
}

// Del удаляет ключ
func (a *CacheAdapter) Del(ctx context.Context, key string) error {
	return a.redis.Delete(ctx, key)
}

// Health проверяет состояние Redis
func (a *CacheAdapter) Health(ctx context.Context) error {
	return a.redis.Health(ctx)
}
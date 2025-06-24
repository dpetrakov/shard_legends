package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_EdgeCases(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	t.Run("Set nil value", func(t *testing.T) {
		err := cache.Set(ctx, "nil_key", nil, time.Hour)
		assert.NoError(t, err)

		var result interface{}
		err = cache.Get(ctx, "nil_key", &result)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Set with zero duration", func(t *testing.T) {
		err := cache.Set(ctx, "zero_duration", "value", 0)
		assert.NoError(t, err)

		// Should expire immediately
		time.Sleep(time.Millisecond)
		
		var result string
		err = cache.Get(ctx, "zero_duration", &result)
		assert.Error(t, err)
	})

	t.Run("Get with wrong type", func(t *testing.T) {
		err := cache.Set(ctx, "string_key", "string_value", time.Hour)
		assert.NoError(t, err)

		var result int
		err = cache.Get(ctx, "string_key", &result)
		assert.Error(t, err)
	})
}

func TestCacheKeyBuilder_EdgeCases(t *testing.T) {
	builder := NewCacheKeyBuilder("inventory")

	t.Run("Build with empty parts", func(t *testing.T) {
		key := builder.Build("prefix", "", "suffix")
		assert.Equal(t, "inventory:prefix::suffix", key)
	})

	t.Run("Build with single part", func(t *testing.T) {
		key := builder.Build("single")
		assert.Equal(t, "inventory:single", key)
	})

	t.Run("UserInventoryKey with all parameters", func(t *testing.T) {
		key := builder.UserInventoryKey("user123", "section456", "item789", "collection1", "quality1")
		assert.Equal(t, "inventory:user123:section456:item789:collection1:quality1", key)
	})
}
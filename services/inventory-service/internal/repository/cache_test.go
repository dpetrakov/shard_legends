package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()
	
	t.Run("Set and Get string", func(t *testing.T) {
		key := "test_key"
		value := "test_value"
		
		err := cache.Set(ctx, key, value, 1*time.Hour)
		assert.NoError(t, err)
		
		var result string
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})
	
	t.Run("Set and Get struct", func(t *testing.T) {
		type testStruct struct {
			ID   string
			Name string
		}
		
		key := "test_struct"
		value := testStruct{ID: "123", Name: "Test"}
		
		err := cache.Set(ctx, key, value, 1*time.Hour)
		assert.NoError(t, err)
		
		var result testStruct
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})
	
	t.Run("Set and Get bytes", func(t *testing.T) {
		key := "test_bytes"
		value := []byte("test data")
		
		err := cache.Set(ctx, key, value, 1*time.Hour)
		assert.NoError(t, err)
		
		var result []byte
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})
	
	t.Run("Get non-existent key", func(t *testing.T) {
		var result string
		err := cache.Get(ctx, "non_existent", &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key not found")
	})
	
	t.Run("Delete key", func(t *testing.T) {
		key := "test_delete"
		value := "test"
		
		// Set value
		err := cache.Set(ctx, key, value, 1*time.Hour)
		assert.NoError(t, err)
		
		// Verify it exists
		var result string
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		
		// Delete it
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)
		
		// Verify it's gone
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err)
	})
	
	t.Run("Delete pattern", func(t *testing.T) {
		// Set multiple keys
		keys := []string{"prefix:key1", "prefix:key2", "other:key3"}
		for _, key := range keys {
			err := cache.Set(ctx, key, "value", 1*time.Hour)
			assert.NoError(t, err)
		}
		
		// Delete by pattern
		err := cache.DeletePattern(ctx, "prefix:*")
		assert.NoError(t, err)
		
		// Check what remains
		var result string
		err = cache.Get(ctx, "prefix:key1", &result)
		assert.Error(t, err)
		
		err = cache.Get(ctx, "prefix:key2", &result)
		assert.Error(t, err)
		
		err = cache.Get(ctx, "other:key3", &result)
		assert.NoError(t, err)
	})
	
	t.Run("Expiration", func(t *testing.T) {
		key := "test_expire"
		value := "test"
		
		// Set with very short TTL
		err := cache.Set(ctx, key, value, 1*time.Millisecond)
		assert.NoError(t, err)
		
		// Wait for expiration
		time.Sleep(2 * time.Millisecond)
		
		// Try to get
		var result string
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key expired")
	})
}

func TestCacheKeyBuilder(t *testing.T) {
	t.Run("Build without prefix", func(t *testing.T) {
		builder := NewCacheKeyBuilder("")
		key := builder.Build("comp1", "comp2", "comp3")
		assert.Equal(t, "comp1:comp2:comp3", key)
	})
	
	t.Run("Build with prefix", func(t *testing.T) {
		builder := NewCacheKeyBuilder("prefix")
		key := builder.Build("comp1", "comp2")
		assert.Equal(t, "prefix:comp1:comp2", key)
	})
	
	t.Run("UserInventoryKey", func(t *testing.T) {
		builder := NewCacheKeyBuilder("")
		key := builder.UserInventoryKey("user1", "section1", "item1", "coll1", "qual1")
		assert.Equal(t, "inventory:user1:section1:item1:coll1:qual1", key)
	})
}

func TestJSONMarshaling(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()
	
	type complexStruct struct {
		ID        string
		Numbers   []int
		Nested    map[string]interface{}
		Timestamp time.Time
	}
	
	original := complexStruct{
		ID:      "test",
		Numbers: []int{1, 2, 3},
		Nested: map[string]interface{}{
			"key": "value",
			"num": 42,
		},
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}
	
	// Test marshaling through cache
	err := cache.Set(ctx, "complex", original, 1*time.Hour)
	assert.NoError(t, err)
	
	var result complexStruct
	err = cache.Get(ctx, "complex", &result)
	assert.NoError(t, err)
	
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Numbers, result.Numbers)
	assert.Equal(t, original.Nested["key"], result.Nested["key"])
	assert.Equal(t, float64(original.Nested["num"].(int)), result.Nested["num"]) // JSON numbers are float64
	assert.Equal(t, original.Timestamp.Unix(), result.Timestamp.Unix())
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache()
	ctx := context.Background()
	
	data := map[string]interface{}{
		"id":   "test",
		"name": "benchmark",
		"data": []byte("some data"),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i%1000))
		_ = cache.Set(ctx, key, data, 1*time.Hour)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache()
	ctx := context.Background()
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := "key_" + string(rune(i))
		data := map[string]interface{}{"id": i}
		_ = cache.Set(ctx, key, data, 1*time.Hour)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i%1000))
		var result map[string]interface{}
		_ = cache.Get(ctx, key, &result)
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	data := map[string]interface{}{
		"id":     "test",
		"name":   "benchmark",
		"number": 42,
		"nested": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}
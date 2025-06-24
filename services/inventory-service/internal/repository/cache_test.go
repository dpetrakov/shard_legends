package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockRedisClient для расширенного тестирования кэширования
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redis.NewStatusCmd(ctx)
	if err := args.Error(0); err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal("OK")
	}
	return cmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewStringCmd(ctx)
	if err := args.Error(1); err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(args.String(0))
	}
	return cmd
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	cmd := redis.NewIntCmd(ctx)
	if err := args.Error(1); err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(args.Get(0).(int64))
	}
	return cmd
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	cmd := redis.NewStringSliceCmd(ctx)
	if err := args.Error(1); err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(args.Get(0).([]string))
	}
	return cmd
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	cmd := redis.NewStatusCmd(ctx)
	if err := args.Error(0); err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal("PONG")
	}
	return cmd
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Расширенные тесты для edge cases и error handling
func TestMemoryCache_EdgeCases(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()
	
	t.Run("Set nil value", func(t *testing.T) {
		err := cache.Set(ctx, "nil_key", nil, time.Hour)
		assert.NoError(t, err)
		
		var result interface{}
		err = cache.Get(ctx, "nil_key", &result)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
	
	t.Run("Set with zero duration", func(t *testing.T) {
		err := cache.Set(ctx, "zero_ttl", "value", 0)
		assert.NoError(t, err)
		
		// Should be expired immediately
		var result string
		err = cache.Get(ctx, "zero_ttl", &result)
		assert.Error(t, err)
	})
	
	t.Run("Get with wrong type", func(t *testing.T) {
		// Set as string
		err := cache.Set(ctx, "type_key", "string_value", time.Hour)
		assert.NoError(t, err)
		
		// Try to get as int
		var result int
		err = cache.Get(ctx, "type_key", &result)
		assert.Error(t, err)
	})
}

func TestCacheKeyBuilder_EdgeCases(t *testing.T) {
	t.Run("Build with empty parts", func(t *testing.T) {
		builder := NewCacheKeyBuilder("prefix")
		key := builder.Build("", "part2", "")
		assert.Equal(t, "prefix::part2:", key)
	})
	
	t.Run("Build with single part", func(t *testing.T) {
		builder := NewCacheKeyBuilder("")
		key := builder.Build("single")
		assert.Equal(t, "single", key)
	})
	
	t.Run("UserInventoryKey with all parameters", func(t *testing.T) {
		builder := NewCacheKeyBuilder("test")
		key := builder.UserInventoryKey("user", "section", "item", "collection", "quality")
		assert.Equal(t, "inventory:user:section:item:collection:quality", key)
	})
}

// Тесты для Redis mock client
func TestRedisClient_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	
	t.Run("redis connection failure during set", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		expectedErr := errors.New("connection refused")
		
		mockClient.On("Set", ctx, "test_key", "test_value", time.Hour).Return(expectedErr)
		
		err := mockClient.Set(ctx, "test_key", "test_value", time.Hour).Err()
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		
		mockClient.AssertExpectations(t)
	})
	
	t.Run("redis key not found", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		
		mockClient.On("Get", ctx, "nonexistent_key").Return("", redis.Nil)
		
		result, err := mockClient.Get(ctx, "nonexistent_key").Result()
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
		assert.Empty(t, result)
		
		mockClient.AssertExpectations(t)
	})
}

func TestCache_SerializationErrors(t *testing.T) {
	t.Run("json marshaling error", func(t *testing.T) {
		// Test with data that can't be marshaled to JSON
		invalidData := make(chan int) // channels can't be marshaled to JSON
		
		_, err := json.Marshal(invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json: unsupported type")
	})
	
	t.Run("json unmarshaling error", func(t *testing.T) {
		invalidJSON := "{invalid json"
		
		var data map[string]interface{}
		err := json.Unmarshal([]byte(invalidJSON), &data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})
}

func TestCache_ConcurrentOperations(t *testing.T) {
	// Заметка: MemoryCache не является thread-safe по дизайну
	// Этот тест демонстрирует ожидаемое поведение для single-threaded использования
	ctx := context.Background()
	
	t.Run("sequential operations simulate concurrent access patterns", func(t *testing.T) {
		cache := NewMemoryCache()
		
		// Simulate what would be concurrent operations in sequential manner
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("test_key_%d", i)
			value := fmt.Sprintf("test_value_%d", i)
			err := cache.Set(ctx, key, value, time.Hour)
			assert.NoError(t, err)
		}
		
		// Verify all values were set correctly
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("test_key_%d", i)
			expectedValue := fmt.Sprintf("test_value_%d", i)
			
			var result string
			err := cache.Get(ctx, key, &result)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, result)
		}
	})
}

func TestCache_ContextHandling(t *testing.T) {
	cache := NewMemoryCache()
	
	t.Run("context cancellation during operation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		// Operations should still work with memory cache as it doesn't use context
		err := cache.Set(ctx, "test_key", "test_value", time.Hour)
		assert.NoError(t, err)
		
		var result string
		err = cache.Get(ctx, "test_key", &result)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", result)
	})
}

func TestCache_DataTypes(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()
	
	t.Run("different data types serialization", func(t *testing.T) {
		testData := []struct {
			name  string
			value interface{}
		}{
			{"string", "test_string"},
			{"integer", 42},
			{"float", 3.14159},
			{"boolean", true},
			{"array", []string{"a", "b", "c"}},
			{"map", map[string]int{"one": 1, "two": 2}},
			{"struct", struct{ Name string; Age int }{"John", 30}},
		}
		
		for _, td := range testData {
			t.Run("serialize "+td.name, func(t *testing.T) {
				key := "test_" + td.name
				
				// Set value
				err := cache.Set(ctx, key, td.value, time.Hour)
				assert.NoError(t, err)
				
				// Get value back
				var result interface{}
				err = cache.Get(ctx, key, &result)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			})
		}
	})
}
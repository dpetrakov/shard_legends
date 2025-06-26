package database

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisDB_InvalidURL(t *testing.T) {
	logger := slog.Default()

	t.Run("invalid redis URL", func(t *testing.T) {
		db, err := NewRedisDB("invalid-url", "redis://localhost:6379/0", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to parse Redis URL")
	})

	t.Run("empty redis URL", func(t *testing.T) {
		db, err := NewRedisDB("", "redis://localhost:6379/0", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to parse Redis URL")
	})

	t.Run("wrong scheme", func(t *testing.T) {
		db, err := NewRedisDB("postgresql://localhost:5432", "redis://localhost:6379/0", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to parse Redis URL")
	})
}

func TestNewRedisDB_ConnectionFail(t *testing.T) {
	logger := slog.Default()

	t.Run("connection fails", func(t *testing.T) {
		// Use localhost with wrong port for immediate connection refused
		invalidURL := "redis://localhost:1"

		db, err := NewRedisDB(invalidURL, "redis://localhost:6379/0", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to ping Redis")
	})
}

func TestRedisDB_Methods(t *testing.T) {
	// Test with nil client to test method safety
	db := &RedisDB{
		client:  nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	t.Run("Client returns nil safely", func(t *testing.T) {
		client := db.Client()
		assert.Nil(t, client)
	})

	t.Run("Close with nil client", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			err := db.Close()
			_ = err
		})
	})

	t.Run("Stats with nil client", func(t *testing.T) {
		stats := db.Stats()
		expected := map[string]interface{}{
			"status": "disconnected",
		}
		assert.Equal(t, expected, stats)
	})

	t.Run("Health with nil client panics", func(t *testing.T) {
		ctx := context.Background()
		// Health method will panic with nil client
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})

	t.Run("Set with nil client panics", func(t *testing.T) {
		ctx := context.Background()
		assert.Panics(t, func() {
			db.Set(ctx, "key", "value", time.Hour)
		})
	})

	t.Run("Get with nil client panics", func(t *testing.T) {
		ctx := context.Background()
		assert.Panics(t, func() {
			_, _ = db.Get(ctx, "key")
		})
	})

	t.Run("Del with nil client panics", func(t *testing.T) {
		ctx := context.Background()
		assert.Panics(t, func() {
			_, _ = db.Del(ctx, "key")
		})
	})

	t.Run("Keys with nil client panics", func(t *testing.T) {
		ctx := context.Background()
		assert.Panics(t, func() {
			_, _ = db.Keys(ctx, "*")
		})
	})
}

func TestRedisDB_ConfigValidation(t *testing.T) {
	logger := slog.Default()

	t.Run("max connections validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			maxConns int
			wantErr  bool
			panics   bool
		}{
			{"valid max conns", 10, true, false},
			{"zero max conns", 0, true, false},
			{"negative max conns", -1, false, true}, // Redis client panics on negative
			{"large max conns", 1000, true, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Use localhost with wrong port for immediate connection refused
				url := "redis://localhost:1"

				if tc.panics {
					assert.Panics(t, func() {
						NewRedisDB(url, "redis://localhost:6379/0", tc.maxConns, logger, nil)
					})
				} else {
					_, err := NewRedisDB(url, "redis://localhost:6379/0", tc.maxConns, logger, nil)
					if tc.wantErr {
						assert.Error(t, err)
						// Should not be a parse error
						assert.NotContains(t, err.Error(), "failed to parse")
					} else {
						assert.NoError(t, err)
					}
				}
			})
		}
	})
}

func TestRedisDB_ContextHandling(t *testing.T) {
	db := &RedisDB{
		client:  nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	t.Run("Health with cancelled context panics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Health method will panic with nil client regardless of context
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})

	t.Run("Set with timeout context panics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(1 * time.Millisecond)

		// Set method will panic with nil client regardless of context
		assert.Panics(t, func() {
			db.Set(ctx, "key", "value", time.Hour)
		})
	})
}

func TestRedisDB_URLParsing(t *testing.T) {
	logger := slog.Default()

	testCases := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid redis URL",
			url:     "redis://localhost:6379",
			wantErr: true, // Connection will fail but parsing should succeed
			errMsg:  "failed to ping Redis",
		},
		{
			name:    "redis URL with auth",
			url:     "redis://user:pass@localhost:6379",
			wantErr: true,
			errMsg:  "failed to ping Redis",
		},
		{
			name:    "redis URL with database",
			url:     "redis://localhost:6379/0",
			wantErr: true,
			errMsg:  "failed to ping Redis",
		},
		{
			name:    "redis URL with all options",
			url:     "redis://user:pass@localhost:6379/1",
			wantErr: true,
			errMsg:  "failed to ping Redis",
		},
		{
			name:    "invalid scheme",
			url:     "postgresql://localhost:5432/db",
			wantErr: true,
			errMsg:  "failed to parse Redis URL",
		},
		{
			name:    "malformed URL",
			url:     "not-a-url",
			wantErr: true,
			errMsg:  "failed to parse Redis URL",
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "failed to parse Redis URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewRedisDB(tc.url, "redis://localhost:6379/0", 10, logger, nil)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				db.Close() // Clean up
			}
		})
	}
}

func TestRedisDB_EdgeCases(t *testing.T) {
	t.Run("nil logger", func(t *testing.T) {
		// Should not panic with nil logger
		assert.NotPanics(t, func() {
			_, _ = NewRedisDB("redis://localhost:1", "redis://localhost:6379/0", 10, nil, nil)
		})
	})

	t.Run("concurrent Close calls", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  slog.Default(),
			metrics: nil,
		}

		// Multiple concurrent Close calls should be safe
		for i := 0; i < 10; i++ {
			go func() {
				db.Close()
			}()
		}

		// Give goroutines time to complete
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Stats field completeness", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  slog.Default(),
			metrics: nil,
		}

		stats := db.Stats()

		// Should have status field
		_, hasStatus := stats["status"]
		assert.True(t, hasStatus)
		assert.Equal(t, "disconnected", stats["status"])

		// Should not have connection stats when disconnected
		_, hasTotal := stats["total_conns"]
		assert.False(t, hasTotal)
	})
}

func TestRedisDB_MethodsWithContext(t *testing.T) {
	db := &RedisDB{
		client:  nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	ctx := context.Background()

	t.Run("Set with various values panics", func(t *testing.T) {
		values := []interface{}{
			"string",
			123,
			true,
			[]byte("bytes"),
			map[string]string{"key": "value"},
		}

		for _, value := range values {
			assert.Panics(t, func() {
				db.Set(ctx, "key", value, time.Hour)
			})
		}
	})

	t.Run("Del with multiple keys panics", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = db.Del(ctx, "key1", "key2", "key3")
		})
	})

	t.Run("Keys with different patterns panics", func(t *testing.T) {
		patterns := []string{"*", "user:*", "session:*", "cache:*"}

		for _, pattern := range patterns {
			assert.Panics(t, func() {
				_, _ = db.Keys(ctx, pattern)
			})
		}
	})
}

func TestRedisDB_MetricsIntegration(t *testing.T) {
	logger := slog.Default()

	t.Run("metrics called on connection failure", func(t *testing.T) {
		// We can't easily test successful connection without a real Redis
		// But we can test that metrics are handled safely with nil metrics

		invalidURL := "redis://localhost:1"
		db, err := NewRedisDB(invalidURL, "redis://localhost:6379/0", 10, logger, nil)

		assert.Error(t, err)
		assert.Nil(t, db)

		// Should not panic with nil metrics
	})

	t.Run("redis URL variations", func(t *testing.T) {
		testURLs := []string{
			"redis://localhost",                  // Default port
			"redis://localhost:6379",             // With port
			"redis://localhost:6379/0",           // With database
			"redis://localhost:6379/1",           // Different database
			"redis://user@localhost:6379",        // With user
			"redis://user:pass@localhost:6379",   // With auth
			"redis://user:pass@localhost:6379/1", // Full URL
		}

		for _, url := range testURLs {
			t.Run(url, func(t *testing.T) {
				// All should parse successfully but fail to connect
				db, err := NewRedisDB(url, "redis://localhost:6379/0", 10, logger, nil)
				assert.Error(t, err)
				assert.Nil(t, db)

				// Error should be connection-related, not parsing
				assert.NotContains(t, err.Error(), "failed to parse Redis URL")
				assert.Contains(t, err.Error(), "failed to ping Redis")
			})
		}
	})
}

// Расширенные тесты для продвинутых сценариев

func TestRedisDB_AdvancedConcurrentOperations(t *testing.T) {
	logger := slog.Default()

	t.Run("concurrent close operations", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  logger,
			metrics: nil,
		}

		// Launch multiple concurrent Close operations
		for i := 0; i < 50; i++ {
			go func() {
				db.Close()
			}()
		}

		// Wait for all goroutines to complete
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("concurrent stats operations", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  logger,
			metrics: nil,
		}

		// Launch multiple concurrent Stats operations
		results := make(chan map[string]interface{}, 50)

		for i := 0; i < 50; i++ {
			go func() {
				stats := db.Stats()
				results <- stats
			}()
		}

		// Collect all results
		for i := 0; i < 50; i++ {
			stats := <-results
			assert.Equal(t, "disconnected", stats["status"])
		}
	})
}

func TestRedisDB_AdvancedOperationErrorHandling(t *testing.T) {
	logger := slog.Default()

	t.Run("redis operations with various contexts", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  logger,
			metrics: nil,
		}

		testOperations := []struct {
			name string
			op   func(context.Context)
		}{
			{
				name: "Set operation",
				op: func(ctx context.Context) {
					db.Set(ctx, "key", "value", time.Hour)
				},
			},
			{
				name: "Get operation",
				op: func(ctx context.Context) {
					db.Get(ctx, "key")
				},
			},
			{
				name: "Del operation",
				op: func(ctx context.Context) {
					db.Del(ctx, "key1", "key2")
				},
			},
			{
				name: "Keys operation",
				op: func(ctx context.Context) {
					db.Keys(ctx, "*")
				},
			},
		}

		for _, testOp := range testOperations {
			t.Run(testOp.name+" with cancelled context", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				// Should panic with nil client
				assert.Panics(t, func() {
					testOp.op(ctx)
				})
			})

			t.Run(testOp.name+" with timeout context", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()

				time.Sleep(1 * time.Millisecond)

				// Should panic with nil client
				assert.Panics(t, func() {
					testOp.op(ctx)
				})
			})
		}
	})
}

func TestRedisDB_AdvancedDataTypeOperations(t *testing.T) {
	db := &RedisDB{
		client:  nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	ctx := context.Background()

	t.Run("Set with various data types", func(t *testing.T) {
		testValues := []struct {
			name  string
			value interface{}
		}{
			{"string", "test_string"},
			{"integer", 42},
			{"float", 3.14},
			{"boolean", true},
			{"byte slice", []byte("test_bytes")},
			{"map", map[string]string{"key": "value"}},
			{"struct", struct{ Name string }{"test"}},
		}

		for _, tv := range testValues {
			t.Run(tv.name, func(t *testing.T) {
				// Should panic with nil client
				assert.Panics(t, func() {
					db.Set(ctx, "test_key", tv.value, time.Hour)
				})
			})
		}
	})

	t.Run("Del with multiple keys", func(t *testing.T) {
		keySets := [][]string{
			{"key1"},
			{"key1", "key2"},
			{"key1", "key2", "key3"},
			{"key1", "key2", "key3", "key4", "key5"},
		}

		for _, keys := range keySets {
			t.Run(fmt.Sprintf("delete %d keys", len(keys)), func(t *testing.T) {
				// Should panic with nil client
				assert.Panics(t, func() {
					db.Del(ctx, keys...)
				})
			})
		}
	})
}

func TestRedisDB_AdvancedLoggerHandling(t *testing.T) {
	t.Run("operations with nil logger", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  nil, // Nil logger
			metrics: nil,
		}

		// These operations should not panic even with nil logger
		assert.NotPanics(t, func() {
			stats := db.Stats()
			assert.Equal(t, "disconnected", stats["status"])
		})

		assert.NotPanics(t, func() {
			db.Close()
		})

		assert.NotPanics(t, func() {
			client := db.Client()
			assert.Nil(t, client)
		})
	})
}

func TestRedisDB_AdvancedStatsCompleteness(t *testing.T) {
	t.Run("disconnected stats structure", func(t *testing.T) {
		db := &RedisDB{
			client:  nil,
			logger:  slog.Default(),
			metrics: nil,
		}

		stats := db.Stats()

		// Check all expected fields for disconnected state
		expectedFields := []string{"status"}

		for _, field := range expectedFields {
			_, exists := stats[field]
			assert.True(t, exists, fmt.Sprintf("Field %s should exist", field))
		}

		// Check that connection-specific fields don't exist
		connectionFields := []string{
			"total_conns", "idle_conns", "stale_conns",
			"hits", "misses", "timeouts",
		}

		for _, field := range connectionFields {
			_, exists := stats[field]
			assert.False(t, exists, fmt.Sprintf("Field %s should not exist when disconnected", field))
		}
	})
}

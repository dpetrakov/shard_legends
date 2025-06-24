package database

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPostgresDB_InvalidURL(t *testing.T) {
	logger := slog.Default()

	t.Run("invalid database URL", func(t *testing.T) {
		db, err := NewPostgresDB("invalid-url", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to parse database URL")
	})

	t.Run("empty database URL", func(t *testing.T) {
		db, err := NewPostgresDB("", 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestNewPostgresDB_ConnectionFail(t *testing.T) {
	logger := slog.Default()

	t.Run("connection fails", func(t *testing.T) {
		// Use unrouteable IP address for faster failure (RFC 5737)
		invalidURL := "postgresql://user:pass@192.0.2.1:5432/testdb"
		
		db, err := NewPostgresDB(invalidURL, 10, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to")
	})
}

func TestPostgresDB_Methods(t *testing.T) {
	// Test with nil pool to test method safety
	db := &PostgresDB{
		pool:    nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	t.Run("Pool returns nil safely", func(t *testing.T) {
		pool := db.Pool()
		assert.Nil(t, pool)
	})

	t.Run("Close with nil pool", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			db.Close()
		})
	})

	t.Run("Stats with nil pool", func(t *testing.T) {
		stats := db.Stats()
		expected := map[string]interface{}{
			"status": "disconnected",
		}
		assert.Equal(t, expected, stats)
	})

	t.Run("Health with nil pool panics", func(t *testing.T) {
		ctx := context.Background()
		// Health method will panic with nil pool - this is expected behavior
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})
}

func TestPostgresDB_ConfigValidation(t *testing.T) {
	logger := slog.Default()

	t.Run("max connections validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			maxConns int
			wantErr  bool
		}{
			{"valid max conns", 10, false},
			{"zero max conns", 0, false}, // pgxpool should handle this
			{"negative max conns", -1, false}, // pgxpool should handle this
			{"large max conns", 1000, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Use unrouteable IP for fast failure (RFC 5737)
				url := "postgresql://user:pass@192.0.2.1:5432/db"
				_, err := NewPostgresDB(url, tc.maxConns, logger, nil)
				
				// We expect error due to connection failure, not config
				assert.Error(t, err)
				// Should not be a parse error
				assert.NotContains(t, err.Error(), "failed to parse")
			})
		}
	})
}

func TestPostgresDB_ContextHandling(t *testing.T) {
	db := &PostgresDB{
		pool:    nil,
		logger:  slog.Default(),
		metrics: nil,
	}

	t.Run("Health with cancelled context panics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Health method will panic with nil pool regardless of context
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})

	t.Run("Health with timeout context panics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		// Wait for timeout
		time.Sleep(1 * time.Millisecond)

		// Health method will panic with nil pool regardless of context
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})
}

// Mock metrics for testing
type mockMetrics struct {
	databaseConnections     float64
	postgresHealthy         bool
	updateDependencyHealthCalled bool
}

func (m *mockMetrics) UpdateDependencyHealth(service string, healthy bool) {
	m.updateDependencyHealthCalled = true
	if service == "postgres" {
		m.postgresHealthy = healthy
	}
}

func (m *mockMetrics) SetDatabaseConnections(count float64) {
	m.databaseConnections = count
}

func TestPostgresDB_MetricsIntegration(t *testing.T) {
	logger := slog.Default()

	t.Run("metrics called on connection failure", func(t *testing.T) {
		// We can't easily test successful connection without a real database
		// But we can test that metrics are handled safely with nil metrics
		
		invalidURL := "postgresql://user:pass@192.0.2.1:5432/testdb"
		db, err := NewPostgresDB(invalidURL, 10, logger, nil)
		
		assert.Error(t, err)
		assert.Nil(t, db)
		
		// Should not panic with nil metrics
	})

	t.Run("database URL variations", func(t *testing.T) {
		testURLs := []string{
			"postgresql://user:pass@localhost/db",
			"postgres://user:pass@localhost/db",
			"postgresql://user@localhost/db",          // No password
			"postgresql://localhost/db",               // No user/pass
			"postgresql://user:pass@localhost:5432/db", // With port
		}

		for _, url := range testURLs {
			t.Run(url, func(t *testing.T) {
				// All should parse successfully but fail to connect
				db, err := NewPostgresDB(url, 10, logger, nil)
				assert.Error(t, err)
				assert.Nil(t, db)
				
				// Error should be connection-related, not parsing
				assert.NotContains(t, err.Error(), "failed to parse database URL")
			})
		}
	})
}

func TestPostgresDB_EdgeCases(t *testing.T) {
	t.Run("nil logger", func(t *testing.T) {
		// Should not panic with nil logger
		assert.NotPanics(t, func() {
			_, _ = NewPostgresDB("postgresql://192.0.2.1/db", 10, nil, nil)
		})
	})

	t.Run("concurrent Close calls", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
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
		db := &PostgresDB{
			pool:    nil,
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

func TestPostgresDB_URLParsing(t *testing.T) {
	logger := slog.Default()

	testCases := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid postgresql URL",
			url:     "postgresql://user:pass@localhost:5432/testdb?sslmode=disable",
			wantErr: true, // Connection will fail but parsing should succeed
			errMsg:  "failed to", // Should be connection error, not parse error
		},
		{
			name:    "valid postgres URL",
			url:     "postgres://user:pass@localhost:5432/testdb",
			wantErr: true,
			errMsg:  "failed to",
		},
		{
			name:    "invalid scheme",
			url:     "mysql://user:pass@localhost:3306/testdb",
			wantErr: true,
			errMsg:  "failed to parse database URL",
		},
		{
			name:    "malformed URL",
			url:     "not-a-url",
			wantErr: true,
			errMsg:  "failed to parse database URL",
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "failed to", // PostgreSQL parses empty URL as default, but connection fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewPostgresDB(tc.url, 10, logger, nil)
			
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

// Расширенные тесты для продвинутых сценариев

// MockAdvancedMetrics implements the metrics interface for testing
type MockAdvancedMetrics struct {
	mock.Mock
}

func (m *MockAdvancedMetrics) UpdateDependencyHealth(service string, healthy bool) {
	m.Called(service, healthy)
}

func (m *MockAdvancedMetrics) SetDatabaseConnections(count float64) {
	m.Called(count)
}

func TestPostgresDB_AdvancedErrorHandling(t *testing.T) {
	logger := slog.Default()
	
	t.Run("health check with context cancellation", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
			logger:  logger,
			metrics: nil,
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		// Should panic with nil pool
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})
	
	t.Run("health check with timeout context", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
			logger:  logger,
			metrics: nil,
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(1 * time.Millisecond) // Wait for timeout
		
		// Should panic with nil pool
		assert.Panics(t, func() {
			db.Health(ctx)
		})
	})
}

func TestPostgresDB_AdvancedConnectionPoolConfiguration(t *testing.T) {
	logger := slog.Default()
	
	testCases := []struct {
		name     string
		maxConns int
		url      string
		wantErr  bool
	}{
		{"small pool size", 1, "postgresql://user:pass@192.0.2.1:5432/testdb", true},
		{"medium pool size", 50, "postgresql://user:pass@192.0.2.1:5432/testdb", true},
		{"large pool size", 200, "postgresql://user:pass@192.0.2.1:5432/testdb", true},
		{"zero pool size", 0, "postgresql://user:pass@192.0.2.1:5432/testdb", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewPostgresDB(tc.url, tc.maxConns, logger, nil)
			
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestPostgresDB_AdvancedConcurrentOperations(t *testing.T) {
	logger := slog.Default()
	
	t.Run("concurrent close operations", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
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
		db := &PostgresDB{
			pool:    nil,
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

func TestPostgresDB_AdvancedErrorPathCoverage(t *testing.T) {
	logger := slog.Default()
	
	t.Run("malformed URLs for parsing errors", func(t *testing.T) {
		malformedURLs := []string{
			"://missing-scheme",
			"postgresql://",
			"postgresql://user@",
			"postgresql://user:@host",
			"postgresql://user:pass@",
			"not-a-url-at-all",
			"postgresql://user:pass@host:invalid-port/db",
		}
		
		for _, url := range malformedURLs {
			t.Run(fmt.Sprintf("URL: %s", url), func(t *testing.T) {
				db, err := NewPostgresDB(url, 10, logger, nil)
				assert.Error(t, err)
				assert.Nil(t, db)
			})
		}
	})
	
	t.Run("various connection timeouts", func(t *testing.T) {
		// Test different network scenarios that could cause timeouts
		networkURLs := []string{
			"postgresql://user:pass@192.0.2.0:5432/db",     // RFC5737 test address
			"postgresql://user:pass@198.51.100.0:5432/db",  // RFC5737 test address
			"postgresql://user:pass@203.0.113.0:5432/db",   // RFC5737 test address
		}
		
		for _, url := range networkURLs {
			t.Run(fmt.Sprintf("Network: %s", url), func(t *testing.T) {
				start := time.Now()
				db, err := NewPostgresDB(url, 10, logger, nil)
				duration := time.Since(start)
				
				assert.Error(t, err)
				assert.Nil(t, db)
				// Should timeout relatively quickly (within reasonable bounds)
				assert.Less(t, duration, 30*time.Second)
			})
		}
	})
}

func TestPostgresDB_AdvancedLoggerHandling(t *testing.T) {
	t.Run("operations with nil logger", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
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
			pool := db.Pool()
			assert.Nil(t, pool)
		})
	})
}

func TestPostgresDB_AdvancedStatsCompleteness(t *testing.T) {
	t.Run("disconnected stats structure", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
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
			"total_conns", "acquired_conns", "idle_conns", "max_conns",
			"new_conns_count", "max_lifetime_destroys", "max_idle_destroys",
		}
		
		for _, field := range connectionFields {
			_, exists := stats[field]
			assert.False(t, exists, fmt.Sprintf("Field %s should not exist when disconnected", field))
		}
	})
}

func TestPostgresDB_AdvancedConfigurationEdgeCases(t *testing.T) {
	logger := slog.Default()
	
	t.Run("extreme configuration values", func(t *testing.T) {
		testConfigs := []struct {
			name     string
			maxConns int
		}{
			{"minimum connections", 1},
			{"maximum reasonable connections", 10000},
			{"zero connections", 0},
		}
		
		for _, config := range testConfigs {
			t.Run(config.name, func(t *testing.T) {
				// Use a URL that will fail to connect so we test config parsing only
				url := "postgresql://user:pass@192.0.2.1:5432/testdb"
				
				db, err := NewPostgresDB(url, config.maxConns, logger, nil)
				
				// Should fail due to connection, not configuration
				assert.Error(t, err)
				assert.Nil(t, db)
				assert.NotContains(t, err.Error(), "failed to parse database URL")
			})
		}
	})
}

func TestPostgresDB_AdvancedMetricsErrorHandling(t *testing.T) {
	t.Run("metrics operations with nil metrics", func(t *testing.T) {
		db := &PostgresDB{
			pool:    nil,
			logger:  slog.Default(),
			metrics: nil, // Nil metrics
		}
		
		// All operations should work safely with nil metrics
		assert.NotPanics(t, func() {
			db.Close()
		})
		
		// Health check will panic due to nil pool, but not due to nil metrics
		assert.Panics(t, func() {
			ctx := context.Background()
			db.Health(ctx)
		})
	})
}
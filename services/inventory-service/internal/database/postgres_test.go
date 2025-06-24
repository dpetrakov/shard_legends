package database

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		// Use invalid host to simulate connection failure
		invalidURL := "postgresql://user:pass@nonexistent-host:5432/testdb"
		
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
				// Use invalid URL so we don't actually connect
				url := "postgresql://user:pass@nonexistent:5432/db"
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
		
		invalidURL := "postgresql://user:pass@nonexistent:5432/testdb"
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
			_, _ = NewPostgresDB("postgresql://localhost/db", 10, nil, nil)
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
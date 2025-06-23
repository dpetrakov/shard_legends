package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// PostgresDB wraps pgxpool.Pool with additional functionality
type PostgresDB struct {
	pool    *pgxpool.Pool
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(databaseURL string, maxConns int, logger *slog.Logger, metricsCollector *metrics.Metrics) (*PostgresDB, error) {
	// Configure connection pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set pool configuration
	config.MaxConns = int32(maxConns)
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &PostgresDB{
		pool:    pool,
		logger:  logger,
		metrics: metricsCollector,
	}

	// Update metrics
	if metricsCollector != nil {
		metricsCollector.DatabaseConnections.Set(float64(maxConns))
		metricsCollector.UpdateDependencyHealth("postgres", true)
	}

	logger.Info("PostgreSQL connection established",
		"max_conns", maxConns,
		"database", config.ConnConfig.Database,
		"host", config.ConnConfig.Host,
		"port", config.ConnConfig.Port,
	)

	return db, nil
}

// Pool returns the underlying pgxpool.Pool
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Health checks the health of the database connection
func (db *PostgresDB) Health(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		if db.metrics != nil {
			db.metrics.UpdateDependencyHealth("postgres", false)
		}
		return err
	}

	if db.metrics != nil {
		db.metrics.UpdateDependencyHealth("postgres", true)
		
		// Update connection stats
		stats := db.pool.Stat()
		db.metrics.DatabaseConnections.Set(float64(stats.TotalConns()))
	}

	return nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
		db.logger.Info("PostgreSQL connection pool closed")
		
		if db.metrics != nil {
			db.metrics.DatabaseConnections.Set(0)
			db.metrics.UpdateDependencyHealth("postgres", false)
		}
	}
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() map[string]interface{} {
	if db.pool == nil {
		return map[string]interface{}{
			"status": "disconnected",
		}
	}

	stats := db.pool.Stat()
	return map[string]interface{}{
		"status":                 "connected",
		"total_conns":           stats.TotalConns(),
		"acquired_conns":        stats.AcquiredConns(),
		"idle_conns":            stats.IdleConns(),
		"max_conns":             stats.MaxConns(),
		"new_conns_count":       stats.NewConnsCount(),
		"max_lifetime_destroys": stats.MaxLifetimeDestroyCount(),
		"max_idle_destroys":     stats.MaxIdleDestroyCount(),
	}
}
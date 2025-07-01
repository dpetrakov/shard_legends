package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shard-legends/deck-game-service/pkg/metrics"
)

// PostgresDB wraps pgxpool.Pool with additional functionality
type PostgresDB struct {
	pool    *pgxpool.Pool
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(databaseURL string, maxConns int, logger *slog.Logger, metrics *metrics.Metrics) (*PostgresDB, error) {
	// Parse and configure connection
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set connection pool settings
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

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection established",
		"max_conns", maxConns,
		"database", maskDatabaseURL(databaseURL))

	return &PostgresDB{
		pool:    pool,
		logger:  logger,
		metrics: metrics,
	}, nil
}

// Pool returns the underlying connection pool
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Health checks the database connection health
func (db *PostgresDB) Health(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		db.logger.Error("Database health check failed", "error", err)
		return err
	}
	return nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.logger.Info("Closing PostgreSQL connection pool")
		db.pool.Close()
	}
}

// UpdateMetrics updates connection pool metrics
func (db *PostgresDB) UpdateMetrics() {
	if db.metrics != nil {
		stats := db.pool.Stat()
		db.metrics.DatabaseConnections.Set(float64(stats.TotalConns()))
	}
}

// maskDatabaseURL masks sensitive information in database URL
func maskDatabaseURL(url string) string {
	// Simple masking for logging
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-10:]
	}
	return "***"
}

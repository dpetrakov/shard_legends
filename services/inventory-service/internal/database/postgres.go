package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/inventory-service/pkg/metrics"
)

// PostgresDB wraps pgxpool.Pool with additional functionality
type PostgresDB struct {
	pool    pgxPool
	logger  *slog.Logger
	metrics *metrics.Metrics
}

// pgxPool defines the subset of pgxpool.Pool methods that PostgresDB relies on.
// It is also satisfied by pgxmock.PgxPoolIface, which simplifies testing.
type pgxPool interface {
	Ping(context.Context) error
	Close()
	Stat() *pgxpool.Stat
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Begin(context.Context) (pgx.Tx, error)
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(databaseURL string, maxConns int, logger *slog.Logger, metricsCollector *metrics.Metrics) (*PostgresDB, error) {
	connString := databaseURL
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = int32(maxConns)
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

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

	if logger != nil {
		logger.Info("PostgreSQL connection pool created")
	}
	if metricsCollector != nil {
		metricsCollector.UpdateDependencyHealth("postgres", true)
	}

	return db, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
		if db.logger != nil {
			db.logger.Info("PostgreSQL connection pool closed")
		}
		if db.metrics != nil {
			db.metrics.DatabaseConnections.Set(0)
			db.metrics.UpdateDependencyHealth("postgres", false)
		}
	}
}

// Health checks the health of the database connection
func (db *PostgresDB) Health(ctx context.Context) error {
	start := time.Now()
	err := db.pool.Ping(ctx)
	status := "success"
	if err != nil {
		status = "error"
		if db.metrics != nil {
			db.metrics.UpdateDependencyHealth("postgres", false)
		}
	} else {
		if db.metrics != nil {
			db.metrics.UpdateDependencyHealth("postgres", true)
			stats := db.pool.Stat()
			db.metrics.DatabaseConnections.Set(float64(stats.TotalConns()))
		}
	}
	if db.metrics != nil {
		db.metrics.DatabaseQueriesTotal.WithLabelValues("ping", status).Inc()
		db.metrics.DatabaseQueryDuration.WithLabelValues("ping").Observe(time.Since(start).Seconds())
	}
	return err
}

// Query executes a query that returns rows.
func (p *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := p.pool.Query(ctx, sql, args...)
	status := "success"
	if err != nil {
		status = "error"
	}
	p.metrics.DatabaseQueriesTotal.WithLabelValues("query", status).Inc()
	p.metrics.DatabaseQueryDuration.WithLabelValues("query").Observe(time.Since(start).Seconds())
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row.
func (p *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := p.pool.QueryRow(ctx, sql, args...)
	p.metrics.DatabaseQueriesTotal.WithLabelValues("query_row", "success").Inc()
	p.metrics.DatabaseQueryDuration.WithLabelValues("query_row").Observe(time.Since(start).Seconds())
	return row
}

// Exec executes a command on the database (e.g., INSERT, UPDATE, DELETE).
func (p *PostgresDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	cmdTag, err := p.pool.Exec(ctx, sql, args...)
	status := "success"
	if err != nil {
		status = "error"
	}
	p.metrics.DatabaseQueriesTotal.WithLabelValues("exec", status).Inc()
	p.metrics.DatabaseQueryDuration.WithLabelValues("exec").Observe(time.Since(start).Seconds())
	return cmdTag, err
}

// Begin starts a transaction.
func (p *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	start := time.Now()
	tx, err := p.pool.Begin(ctx)
	status := "success"
	if err != nil {
		status = "error"
	}
	p.metrics.DatabaseQueriesTotal.WithLabelValues("begin_transaction", status).Inc()
	p.metrics.DatabaseQueryDuration.WithLabelValues("begin_transaction").Observe(time.Since(start).Seconds())
	return tx, err
}

// Stats returns basic statistics about the underlying connection pool. If the pool is
// not initialised (nil) it returns a map with a single key "status" set to
// "disconnected". The concrete statistics returned mimic those exposed by
// pgxpool.Stat for ease of monitoring while keeping the method safe for tests.
func (db *PostgresDB) Stats() map[string]interface{} {
	if db == nil || db.pool == nil {
		return map[string]interface{}{
			"status": "disconnected",
		}
	}

	stats := db.pool.Stat()
	return map[string]interface{}{
		"status":                "connected",
		"total_conns":           stats.TotalConns(),
		"acquired_conns":        stats.AcquiredConns(),
		"idle_conns":            stats.IdleConns(),
		"max_conns":             stats.MaxConns(),
		"new_conns_count":       stats.NewConnsCount(),
		"max_lifetime_destroys": stats.MaxLifetimeDestroyCount(),
		"max_idle_destroys":     stats.MaxIdleDestroyCount(),
	}
}

// Pool returns the underlying *pgxpool.Pool when available. It returns nil when
// the pool is backed by a mock (used in tests) or when the pool is not
// initialised.
func (db *PostgresDB) Pool() *pgxpool.Pool {
	if db == nil || db.pool == nil {
		return nil
	}
	if p, ok := db.pool.(*pgxpool.Pool); ok {
		return p
	}
	return nil
}

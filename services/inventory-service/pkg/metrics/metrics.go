package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all Prometheus metrics for the inventory service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Database metrics
	DatabaseConnections  prometheus.Gauge
	DatabaseQueriesTotal *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec

	// Redis metrics
	RedisConnections     prometheus.Gauge
	RedisCommandsTotal   *prometheus.CounterVec
	RedisCommandDuration *prometheus.HistogramVec

	// Business metrics
	InventoryOperationsTotal *prometheus.CounterVec
	ActiveUsersTotal        prometheus.Gauge
	BalanceCalculations     *prometheus.CounterVec
	BalanceCalculationDuration *prometheus.HistogramVec
	ItemsPerInventory       *prometheus.HistogramVec
	CacheHits               *prometheus.CounterVec
	CacheMisses             *prometheus.CounterVec
	TransactionOperations   *prometheus.HistogramVec
	TransactionDuration     *prometheus.HistogramVec

	// Health metrics
	DependencyHealth *prometheus.GaugeVec
}

// New creates a new Metrics instance with all Prometheus metrics
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "inventory_service_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Database metrics
		DatabaseConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "inventory_service_database_connections",
				Help: "Current number of database connections",
			},
		),
		DatabaseQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "status"},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),

		// Redis metrics
		RedisConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "inventory_service_redis_connections",
				Help: "Current number of Redis connections",
			},
		),
		RedisCommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_redis_commands_total",
				Help: "Total number of Redis commands",
			},
			[]string{"command", "status"},
		),
		RedisCommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_redis_command_duration_seconds",
				Help:    "Duration of Redis commands in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"command"},
		),

		// Business metrics
		InventoryOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_operations_total",
				Help: "Total number of inventory operations",
			},
			[]string{"operation_type", "section", "status"},
		),
		ActiveUsersTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "inventory_service_active_users_total",
				Help: "Total number of active users",
			},
		),
		BalanceCalculations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_balance_calculations_total",
				Help: "Total number of balance calculations",
			},
			[]string{"section", "status"},
		),
		BalanceCalculationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_balance_calculation_duration_seconds",
				Help:    "Duration of balance calculations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"section"},
		),
		ItemsPerInventory: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_items_per_user",
				Help:    "Number of items per user inventory",
				Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"section"},
		),
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_service_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
		TransactionOperations: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_transaction_operations_count",
				Help:    "Number of operations per transaction",
				Buckets: []float64{1, 2, 5, 10, 25, 50, 100},
			},
			[]string{"operation_type"},
		),
		TransactionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_service_transaction_duration_seconds",
				Help:    "Duration of inventory transactions in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation_type"},
		),

		// Health metrics
		DependencyHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "inventory_service_dependency_health",
				Help: "Health status of dependencies (1 = healthy, 0 = unhealthy)",
			},
			[]string{"dependency"},
		),
	}
}

// Initialize sets up initial metric values
func (m *Metrics) Initialize() {
	// Set initial values for health metrics
	m.DependencyHealth.WithLabelValues("postgres").Set(0)
	m.DependencyHealth.WithLabelValues("redis").Set(0)
}

// UpdateDependencyHealth updates the health status of a dependency
func (m *Metrics) UpdateDependencyHealth(dependency string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.DependencyHealth.WithLabelValues(dependency).Set(value)
}

// Shutdown performs cleanup of metrics resources
func (m *Metrics) Shutdown() {
	// Currently no cleanup needed for Prometheus metrics
}

// RecordInventoryOperation records an inventory operation metric
func (m *Metrics) RecordInventoryOperation(operationType, section, status string) {
	if m.InventoryOperationsTotal != nil {
		m.InventoryOperationsTotal.WithLabelValues(operationType, section, status).Inc()
	}
}
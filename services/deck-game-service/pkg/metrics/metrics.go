package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all Prometheus metrics for the deck game service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Database metrics
	DatabaseConnections   prometheus.Gauge
	DatabaseQueriesTotal  *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec

	// Redis metrics
	RedisConnections     prometheus.Gauge
	RedisCommandsTotal   *prometheus.CounterVec
	RedisCommandDuration *prometheus.HistogramVec

	// Business metrics (from specification)
	DailyChestCraftsTotal   *prometheus.CounterVec
	DailyChestStatusChecks  *prometheus.CounterVec
	DailyChestClaimAttempts *prometheus.CounterVec
	DailyChestClaimDuration *prometheus.HistogramVec
	DailyChestComboValues   *prometheus.HistogramVec
	DailyChestCooldownHits  *prometheus.CounterVec

	// External service integration metrics
	ProductionServiceCalls  *prometheus.CounterVec
	InventoryServiceCalls   *prometheus.CounterVec
	ExternalServiceDuration *prometheus.HistogramVec

	// Health metrics
	DependencyHealth *prometheus.GaugeVec
}

// New creates a new Metrics instance with all Prometheus metrics
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "dgs_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Database metrics
		DatabaseConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "dgs_database_connections",
				Help: "Current number of database connections",
			},
		),
		DatabaseQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "status"},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),

		// Redis metrics
		RedisConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "dgs_redis_connections",
				Help: "Current number of Redis connections",
			},
		),
		RedisCommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_redis_commands_total",
				Help: "Total number of Redis commands",
			},
			[]string{"command", "status"},
		),
		RedisCommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_redis_command_duration_seconds",
				Help:    "Duration of Redis commands in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"command"},
		),

		// Business metrics
		DailyChestCraftsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_daily_chest_crafts_total",
				Help: "Total number of daily chest crafts",
			},
			[]string{"status"},
		),
		DailyChestStatusChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_daily_chest_status_checks_total",
				Help: "Total number of daily chest status checks",
			},
			[]string{"finished"},
		),
		DailyChestClaimAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_daily_chest_claim_attempts_total",
				Help: "Total number of daily chest claim attempts",
			},
			[]string{"status"},
		),
		DailyChestClaimDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_daily_chest_claim_duration_seconds",
				Help:    "Duration of daily chest claim operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
		DailyChestComboValues: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_daily_chest_combo_values",
				Help:    "Distribution of combo values in claim requests",
				Buckets: []float64{1, 3, 5, 7, 10, 15, 20, 25},
			},
			[]string{"expected_combo"},
		),
		DailyChestCooldownHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_daily_chest_cooldown_hits_total",
				Help: "Total number of cooldown hits during claim attempts",
			},
			[]string{"user_id"},
		),

		// External service integration metrics
		ProductionServiceCalls: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_production_service_calls_total",
				Help: "Total number of calls to Production Service",
			},
			[]string{"operation", "status"},
		),
		InventoryServiceCalls: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dgs_inventory_service_calls_total",
				Help: "Total number of calls to Inventory Service",
			},
			[]string{"operation", "status"},
		),
		ExternalServiceDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dgs_external_service_duration_seconds",
				Help:    "Duration of external service calls in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "operation"},
		),

		// Health metrics
		DependencyHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dgs_dependency_health",
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
	m.DependencyHealth.WithLabelValues("production-service").Set(0)
	m.DependencyHealth.WithLabelValues("inventory-service").Set(0)
}

// UpdateDependencyHealth updates the health status of a dependency
func (m *Metrics) UpdateDependencyHealth(dependency string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.DependencyHealth.WithLabelValues(dependency).Set(value)
}

// Shutdown performs cleanup operations
func (m *Metrics) Shutdown() {
	// Currently no cleanup needed for Prometheus metrics
}

// RecordDailyChestOperation records a daily chest operation
func (m *Metrics) RecordDailyChestOperation(operationType, status string) {
	switch operationType {
	case "status_check":
		m.DailyChestStatusChecks.WithLabelValues(status).Inc()
	case "claim_attempt":
		m.DailyChestClaimAttempts.WithLabelValues(status).Inc()
	case "craft":
		m.DailyChestCraftsTotal.WithLabelValues(status).Inc()
	}
}

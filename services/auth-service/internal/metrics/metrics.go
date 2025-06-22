package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "auth_service"
	subsystem = ""
)

// Metrics holds all Prometheus metrics for the auth service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal       *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
	HTTPRequestsInFlight    prometheus.Gauge

	// Auth metrics
	AuthRequestsTotal               *prometheus.CounterVec
	AuthRequestDuration             *prometheus.HistogramVec
	AuthTelegramValidationDuration  prometheus.Histogram
	AuthNewUsersTotal               prometheus.Counter
	AuthRateLimitHitsTotal          *prometheus.CounterVec

	// JWT metrics
	JWTTokensGeneratedTotal    prometheus.Counter
	JWTTokensValidatedTotal    *prometheus.CounterVec
	JWTKeyGenerationDuration   prometheus.Histogram
	JWTActiveTokensCount       prometheus.Gauge
	JWTTokensPerUserHistogram  prometheus.Histogram

	// Redis metrics
	RedisOperationsTotal         *prometheus.CounterVec
	RedisOperationDuration       *prometheus.HistogramVec
	RedisConnectionPoolActive    prometheus.Gauge
	RedisConnectionPoolIdle      prometheus.Gauge
	RedisTokenCleanupDuration    prometheus.Histogram
	RedisExpiredTokensCleaned    prometheus.Counter
	RedisCleanupProcessedUsers   prometheus.Counter

	// PostgreSQL metrics
	PostgresOperationsTotal         *prometheus.CounterVec
	PostgresOperationDuration       *prometheus.HistogramVec
	PostgresConnectionPoolActive    prometheus.Gauge
	PostgresConnectionPoolIdle      prometheus.Gauge
	PostgresConnectionPoolMax       prometheus.Gauge

	// System health metrics
	ServiceUp                prometheus.Gauge
	ServiceStartTime         prometheus.Gauge
	DependenciesHealthy      *prometheus.GaugeVec
	MemoryUsageBytes         prometheus.Gauge
	GoroutinesCount          prometheus.Gauge

	// Admin metrics
	AdminOperationsTotal      *prometheus.CounterVec
	AdminTokenRevocationsTotal *prometheus.CounterVec
	AdminCleanupOperationsTotal prometheus.Counter
}

// New creates and initializes all metrics
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// Auth metrics
		AuthRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_requests_total",
				Help:      "Total number of authentication requests",
			},
			[]string{"status", "reason"},
		),
		AuthRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_request_duration_seconds",
				Help:      "Authentication request processing time in seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"status"},
		),
		AuthTelegramValidationDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_telegram_validation_duration_seconds",
				Help:      "Time spent validating Telegram signatures in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
			},
		),
		AuthNewUsersTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_new_users_total",
				Help:      "Total number of new user registrations",
			},
		),
		AuthRateLimitHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_rate_limit_hits_total",
				Help:      "Total number of requests blocked by rate limiting",
			},
			[]string{"ip"},
		),

		// JWT metrics
		JWTTokensGeneratedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_generated_total",
				Help:      "Total number of JWT tokens generated",
			},
		),
		JWTTokensValidatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_validated_total",
				Help:      "Total number of JWT tokens validated",
			},
			[]string{"status"},
		),
		JWTKeyGenerationDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "jwt_key_generation_duration_seconds",
				Help:      "Time spent generating RSA keys in seconds",
				Buckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
		),
		JWTActiveTokensCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "jwt_active_tokens_count",
				Help:      "Current number of active JWT tokens",
			},
		),
		JWTTokensPerUserHistogram: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_per_user",
				Help:      "Distribution of number of tokens per user",
				Buckets:   []float64{1, 2, 3, 5, 10, 20, 50},
			},
		),

		// Redis metrics
		RedisOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_operations_total",
				Help:      "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),
		RedisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_operation_duration_seconds",
				Help:      "Redis operation duration in seconds",
				Buckets:   []float64{0.0001, 0.001, 0.01, 0.1, 0.5, 1.0},
			},
			[]string{"operation"},
		),
		RedisConnectionPoolActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "redis_connection_pool_active",
				Help:      "Number of active Redis connections",
			},
		),
		RedisConnectionPoolIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "redis_connection_pool_idle",
				Help:      "Number of idle Redis connections",
			},
		),
		RedisTokenCleanupDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_token_cleanup_duration_seconds",
				Help:      "Time spent cleaning up expired tokens in seconds",
				Buckets:   []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
			},
		),
		RedisExpiredTokensCleaned: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_expired_tokens_cleaned_total",
				Help:      "Total number of expired tokens cleaned up",
			},
		),
		RedisCleanupProcessedUsers: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_cleanup_processed_users_total",
				Help:      "Total number of users processed during cleanup",
			},
		),

		// PostgreSQL metrics
		PostgresOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "postgres_operations_total",
				Help:      "Total number of PostgreSQL operations",
			},
			[]string{"operation", "table", "status"},
		),
		PostgresOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "postgres_operation_duration_seconds",
				Help:      "PostgreSQL operation duration in seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"operation", "table"},
		),
		PostgresConnectionPoolActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_active",
				Help:      "Number of active PostgreSQL connections",
			},
		),
		PostgresConnectionPoolIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_idle",
				Help:      "Number of idle PostgreSQL connections",
			},
		),
		PostgresConnectionPoolMax: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_max",
				Help:      "Maximum number of PostgreSQL connections",
			},
		),

		// System health metrics
		ServiceUp: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_up",
				Help:      "Service availability (1 = up, 0 = down)",
			},
		),
		ServiceStartTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_start_time_seconds",
				Help:      "Service start time as unix timestamp",
			},
		),
		DependenciesHealthy: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dependencies_healthy",
				Help:      "Health status of service dependencies (1 = healthy, 0 = unhealthy)",
			},
			[]string{"dependency"},
		),
		MemoryUsageBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines_count",
				Help:      "Number of active goroutines",
			},
		),

		// Admin metrics
		AdminOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_operations_total",
				Help:      "Total number of admin operations",
			},
			[]string{"operation", "status"},
		),
		AdminTokenRevocationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_token_revocations_total",
				Help:      "Total number of token revocations",
			},
			[]string{"method"},
		),
		AdminCleanupOperationsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_cleanup_operations_total",
				Help:      "Total number of manual cleanup operations",
			},
		),
	}
}

// Initialize sets up initial metric values
func (m *Metrics) Initialize() {
	// Set service up metric
	m.ServiceUp.Set(1)
	
	// Set service start time
	m.ServiceStartTime.Set(float64(time.Now().Unix()))
	
	// Start background metrics collection
	go m.collectSystemMetrics()
}

// collectSystemMetrics collects system-level metrics in the background
func (m *Metrics) collectSystemMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Collect memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		m.MemoryUsageBytes.Set(float64(memStats.Alloc))
		
		// Collect goroutine count
		m.GoroutinesCount.Set(float64(runtime.NumGoroutine()))
	}
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordAuthRequest records authentication request metrics
func (m *Metrics) RecordAuthRequest(status, reason string, duration time.Duration) {
	m.AuthRequestsTotal.WithLabelValues(status, reason).Inc()
	m.AuthRequestDuration.WithLabelValues(status).Observe(duration.Seconds())
}

// RecordTelegramValidation records Telegram validation time
func (m *Metrics) RecordTelegramValidation(duration time.Duration) {
	m.AuthTelegramValidationDuration.Observe(duration.Seconds())
}

// RecordNewUser increments new user counter
func (m *Metrics) RecordNewUser() {
	m.AuthNewUsersTotal.Inc()
}

// RecordRateLimitHit records rate limit hit
func (m *Metrics) RecordRateLimitHit(ip string) {
	m.AuthRateLimitHitsTotal.WithLabelValues(ip).Inc()
}

// RecordJWTGenerated increments JWT generation counter
func (m *Metrics) RecordJWTGenerated() {
	m.JWTTokensGeneratedTotal.Inc()
}

// RecordJWTValidation records JWT validation result
func (m *Metrics) RecordJWTValidation(status string) {
	m.JWTTokensValidatedTotal.WithLabelValues(status).Inc()
}

// RecordKeyGeneration records RSA key generation time
func (m *Metrics) RecordKeyGeneration(duration time.Duration) {
	m.JWTKeyGenerationDuration.Observe(duration.Seconds())
}

// UpdateActiveTokensCount updates the active tokens gauge
func (m *Metrics) UpdateActiveTokensCount(count float64) {
	m.JWTActiveTokensCount.Set(count)
}

// RecordTokensPerUser records tokens per user distribution
func (m *Metrics) RecordTokensPerUser(count float64) {
	m.JWTTokensPerUserHistogram.Observe(count)
}

// RecordRedisOperation records Redis operation metrics
func (m *Metrics) RecordRedisOperation(operation, status string, duration time.Duration) {
	m.RedisOperationsTotal.WithLabelValues(operation, status).Inc()
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateRedisPoolStats updates Redis connection pool metrics
func (m *Metrics) UpdateRedisPoolStats(active, idle float64) {
	m.RedisConnectionPoolActive.Set(active)
	m.RedisConnectionPoolIdle.Set(idle)
}

// RecordTokenCleanup records token cleanup metrics
func (m *Metrics) RecordTokenCleanup(duration time.Duration, expiredTokens, processedUsers float64) {
	m.RedisTokenCleanupDuration.Observe(duration.Seconds())
	m.RedisExpiredTokensCleaned.Add(expiredTokens)
	m.RedisCleanupProcessedUsers.Add(processedUsers)
}

// RecordPostgresOperation records PostgreSQL operation metrics
func (m *Metrics) RecordPostgresOperation(operation, table, status string, duration time.Duration) {
	m.PostgresOperationsTotal.WithLabelValues(operation, table, status).Inc()
	m.PostgresOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdatePostgresPoolStats updates PostgreSQL connection pool metrics
func (m *Metrics) UpdatePostgresPoolStats(active, idle, max float64) {
	m.PostgresConnectionPoolActive.Set(active)
	m.PostgresConnectionPoolIdle.Set(idle)
	m.PostgresConnectionPoolMax.Set(max)
}

// UpdateDependencyHealth updates dependency health status
func (m *Metrics) UpdateDependencyHealth(dependency string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.DependenciesHealthy.WithLabelValues(dependency).Set(value)
}

// RecordAdminOperation records admin operation metrics
func (m *Metrics) RecordAdminOperation(operation, status string) {
	m.AdminOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordTokenRevocation records token revocation metrics
func (m *Metrics) RecordTokenRevocation(method string) {
	m.AdminTokenRevocationsTotal.WithLabelValues(method).Inc()
}

// RecordManualCleanup increments manual cleanup counter
func (m *Metrics) RecordManualCleanup() {
	m.AdminCleanupOperationsTotal.Inc()
}

// Shutdown updates service status
func (m *Metrics) Shutdown() {
	m.ServiceUp.Set(0)
}
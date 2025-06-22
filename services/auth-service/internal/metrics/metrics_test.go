package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// createTestMetrics creates a new metrics instance with a custom registry
// to avoid conflicts between tests
func createTestMetrics() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// Auth metrics
		AuthRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_requests_total",
				Help:      "Total number of authentication requests",
			},
			[]string{"status", "reason"},
		),
		AuthRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_request_duration_seconds",
				Help:      "Authentication request processing time in seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"status"},
		),
		AuthTelegramValidationDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_telegram_validation_duration_seconds",
				Help:      "Time spent validating Telegram signatures in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
			},
		),
		AuthNewUsersTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_new_users_total",
				Help:      "Total number of new user registrations",
			},
		),
		AuthRateLimitHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_rate_limit_hits_total",
				Help:      "Total number of requests blocked by rate limiting",
			},
			[]string{"ip"},
		),

		// JWT metrics
		JWTTokensGeneratedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_generated_total",
				Help:      "Total number of JWT tokens generated",
			},
		),
		JWTTokensValidatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_validated_total",
				Help:      "Total number of JWT tokens validated",
			},
			[]string{"status"},
		),
		JWTKeyGenerationDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "jwt_key_generation_duration_seconds",
				Help:      "Time spent generating RSA keys in seconds",
				Buckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
		),
		JWTActiveTokensCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "jwt_active_tokens_count",
				Help:      "Current number of active JWT tokens",
			},
		),
		JWTTokensPerUserHistogram: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "jwt_tokens_per_user",
				Help:      "Distribution of number of tokens per user",
				Buckets:   []float64{1, 2, 3, 5, 10, 20, 50},
			},
		),

		// Redis metrics
		RedisOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_operations_total",
				Help:      "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),
		RedisOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_operation_duration_seconds",
				Help:      "Redis operation duration in seconds",
				Buckets:   []float64{0.0001, 0.001, 0.01, 0.1, 0.5, 1.0},
			},
			[]string{"operation"},
		),
		RedisConnectionPoolActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "redis_connection_pool_active",
				Help:      "Number of active Redis connections",
			},
		),
		RedisConnectionPoolIdle: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "redis_connection_pool_idle",
				Help:      "Number of idle Redis connections",
			},
		),
		RedisTokenCleanupDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_token_cleanup_duration_seconds",
				Help:      "Time spent cleaning up expired tokens in seconds",
				Buckets:   []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
			},
		),
		RedisExpiredTokensCleaned: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_expired_tokens_cleaned_total",
				Help:      "Total number of expired tokens cleaned up",
			},
		),
		RedisCleanupProcessedUsers: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_cleanup_processed_users_total",
				Help:      "Total number of users processed during cleanup",
			},
		),

		// PostgreSQL metrics
		PostgresOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "postgres_operations_total",
				Help:      "Total number of PostgreSQL operations",
			},
			[]string{"operation", "table", "status"},
		),
		PostgresOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "postgres_operation_duration_seconds",
				Help:      "PostgreSQL operation duration in seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"operation", "table"},
		),
		PostgresConnectionPoolActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_active",
				Help:      "Number of active PostgreSQL connections",
			},
		),
		PostgresConnectionPoolIdle: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_idle",
				Help:      "Number of idle PostgreSQL connections",
			},
		),
		PostgresConnectionPoolMax: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "postgres_connection_pool_max",
				Help:      "Maximum number of PostgreSQL connections",
			},
		),

		// System health metrics
		ServiceUp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_up",
				Help:      "Service availability (1 = up, 0 = down)",
			},
		),
		ServiceStartTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_start_time_seconds",
				Help:      "Service start time as unix timestamp",
			},
		),
		DependenciesHealthy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dependencies_healthy",
				Help:      "Health status of service dependencies (1 = healthy, 0 = unhealthy)",
			},
			[]string{"dependency"},
		),
		MemoryUsageBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),
		GoroutinesCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines_count",
				Help:      "Number of active goroutines",
			},
		),

		// Admin metrics
		AdminOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_operations_total",
				Help:      "Total number of admin operations",
			},
			[]string{"operation", "status"},
		),
		AdminTokenRevocationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_token_revocations_total",
				Help:      "Total number of token revocations",
			},
			[]string{"method"},
		),
		AdminCleanupOperationsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "admin_cleanup_operations_total",
				Help:      "Total number of manual cleanup operations",
			},
		),
	}
}

func TestMetricsInitialization(t *testing.T) {
	// Create new metrics instance
	m := createTestMetrics()
	if m == nil {
		t.Fatal("Failed to create metrics instance")
	}

	// Initialize metrics
	m.Initialize()
	
	// Verify service up metric is set to 1
	serviceUpValue := getGaugeValue(t, m.ServiceUp)
	if serviceUpValue != 1.0 {
		t.Errorf("Expected ServiceUp to be 1.0, got %f", serviceUpValue)
	}

	// Verify service start time is set
	startTimeValue := getGaugeValue(t, m.ServiceStartTime)
	if startTimeValue <= 0 {
		t.Errorf("Expected ServiceStartTime to be > 0, got %f", startTimeValue)
	}

	// Test shutdown
	m.Shutdown()
	
	// Verify service up metric is set to 0
	serviceUpValue = getGaugeValue(t, m.ServiceUp)
	if serviceUpValue != 0.0 {
		t.Errorf("Expected ServiceUp to be 0.0 after shutdown, got %f", serviceUpValue)
	}
}

func TestHTTPMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Record HTTP request
	duration := 100 * time.Millisecond
	m.RecordHTTPRequest("POST", "/auth", "200", duration)
	
	// Verify counter metric
	counterValue := getCounterVecValue(t, m.HTTPRequestsTotal, "POST", "/auth", "200")
	if counterValue != 1.0 {
		t.Errorf("Expected HTTPRequestsTotal to be 1.0, got %f", counterValue)
	}
	
	// Verify histogram metric has observations
	histogramCount := getHistogramVecCount(t, m.HTTPRequestDuration, "POST", "/auth")
	if histogramCount != 1 {
		t.Errorf("Expected HTTPRequestDuration to have 1 observation, got %d", histogramCount)
	}
}

func TestAuthMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test successful auth
	duration := 50 * time.Millisecond
	m.RecordAuthRequest("success", "valid", duration)
	
	counterValue := getCounterVecValue(t, m.AuthRequestsTotal, "success", "valid")
	if counterValue != 1.0 {
		t.Errorf("Expected AuthRequestsTotal to be 1.0, got %f", counterValue)
	}
	
	// Test failed auth
	m.RecordAuthRequest("failed", "invalid_signature", duration)
	
	counterValue = getCounterVecValue(t, m.AuthRequestsTotal, "failed", "invalid_signature")
	if counterValue != 1.0 {
		t.Errorf("Expected AuthRequestsTotal to be 1.0, got %f", counterValue)
	}
	
	// Test telegram validation
	validationDuration := 10 * time.Millisecond
	m.RecordTelegramValidation(validationDuration)
	
	histogramCount := getHistogramCount(t, m.AuthTelegramValidationDuration)
	if histogramCount != 1 {
		t.Errorf("Expected AuthTelegramValidationDuration to have 1 observation, got %d", histogramCount)
	}
	
	// Test new user registration
	m.RecordNewUser()
	
	counterValue = getCounterValue(t, m.AuthNewUsersTotal)
	if counterValue != 1.0 {
		t.Errorf("Expected AuthNewUsersTotal to be 1.0, got %f", counterValue)
	}
	
	// Test rate limit hit
	m.RecordRateLimitHit("192.168.1.1")
	
	counterValue = getCounterVecValue(t, m.AuthRateLimitHitsTotal, "192.168.1.1")
	if counterValue != 1.0 {
		t.Errorf("Expected AuthRateLimitHitsTotal to be 1.0, got %f", counterValue)
	}
}

func TestJWTMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test JWT generation
	m.RecordJWTGenerated()
	
	counterValue := getCounterValue(t, m.JWTTokensGeneratedTotal)
	if counterValue != 1.0 {
		t.Errorf("Expected JWTTokensGeneratedTotal to be 1.0, got %f", counterValue)
	}
	
	// Test JWT validation
	m.RecordJWTValidation("valid")
	m.RecordJWTValidation("invalid")
	
	validCount := getCounterVecValue(t, m.JWTTokensValidatedTotal, "valid")
	if validCount != 1.0 {
		t.Errorf("Expected valid JWT tokens to be 1.0, got %f", validCount)
	}
	
	invalidCount := getCounterVecValue(t, m.JWTTokensValidatedTotal, "invalid")
	if invalidCount != 1.0 {
		t.Errorf("Expected invalid JWT tokens to be 1.0, got %f", invalidCount)
	}
	
	// Test key generation
	keyGenDuration := 500 * time.Millisecond
	m.RecordKeyGeneration(keyGenDuration)
	
	histogramCount := getHistogramCount(t, m.JWTKeyGenerationDuration)
	if histogramCount != 1 {
		t.Errorf("Expected JWTKeyGenerationDuration to have 1 observation, got %d", histogramCount)
	}
	
	// Test active tokens count
	m.UpdateActiveTokensCount(42.0)
	
	gaugeValue := getGaugeValue(t, m.JWTActiveTokensCount)
	if gaugeValue != 42.0 {
		t.Errorf("Expected JWTActiveTokensCount to be 42.0, got %f", gaugeValue)
	}
	
	// Test tokens per user
	m.RecordTokensPerUser(3.0)
	
	histogramCount = getHistogramCount(t, m.JWTTokensPerUserHistogram)
	if histogramCount != 1 {
		t.Errorf("Expected JWTTokensPerUserHistogram to have 1 observation, got %d", histogramCount)
	}
}

func TestRedisMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test Redis operation
	duration := 5 * time.Millisecond
	m.RecordRedisOperation("set", "success", duration)
	
	counterValue := getCounterVecValue(t, m.RedisOperationsTotal, "set", "success")
	if counterValue != 1.0 {
		t.Errorf("Expected RedisOperationsTotal to be 1.0, got %f", counterValue)
	}
	
	histogramCount := getHistogramVecCount(t, m.RedisOperationDuration, "set")
	if histogramCount != 1 {
		t.Errorf("Expected RedisOperationDuration to have 1 observation, got %d", histogramCount)
	}
	
	// Test pool stats
	m.UpdateRedisPoolStats(5.0, 3.0)
	
	activeValue := getGaugeValue(t, m.RedisConnectionPoolActive)
	if activeValue != 5.0 {
		t.Errorf("Expected RedisConnectionPoolActive to be 5.0, got %f", activeValue)
	}
	
	idleValue := getGaugeValue(t, m.RedisConnectionPoolIdle)
	if idleValue != 3.0 {
		t.Errorf("Expected RedisConnectionPoolIdle to be 3.0, got %f", idleValue)
	}
	
	// Test token cleanup
	cleanupDuration := 2 * time.Second
	m.RecordTokenCleanup(cleanupDuration, 10.0, 5.0)
	
	histogramCount = getHistogramCount(t, m.RedisTokenCleanupDuration)
	if histogramCount != 1 {
		t.Errorf("Expected RedisTokenCleanupDuration to have 1 observation, got %d", histogramCount)
	}
	
	expiredCount := getCounterValue(t, m.RedisExpiredTokensCleaned)
	if expiredCount != 10.0 {
		t.Errorf("Expected RedisExpiredTokensCleaned to be 10.0, got %f", expiredCount)
	}
	
	processedCount := getCounterValue(t, m.RedisCleanupProcessedUsers)
	if processedCount != 5.0 {
		t.Errorf("Expected RedisCleanupProcessedUsers to be 5.0, got %f", processedCount)
	}
}

func TestPostgresMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test Postgres operation
	duration := 15 * time.Millisecond
	m.RecordPostgresOperation("select", "users", "success", duration)
	
	counterValue := getCounterVecValue(t, m.PostgresOperationsTotal, "select", "users", "success")
	if counterValue != 1.0 {
		t.Errorf("Expected PostgresOperationsTotal to be 1.0, got %f", counterValue)
	}
	
	histogramCount := getHistogramVecCount(t, m.PostgresOperationDuration, "select", "users")
	if histogramCount != 1 {
		t.Errorf("Expected PostgresOperationDuration to have 1 observation, got %d", histogramCount)
	}
	
	// Test pool stats
	m.UpdatePostgresPoolStats(8.0, 2.0, 10.0)
	
	activeValue := getGaugeValue(t, m.PostgresConnectionPoolActive)
	if activeValue != 8.0 {
		t.Errorf("Expected PostgresConnectionPoolActive to be 8.0, got %f", activeValue)
	}
	
	idleValue := getGaugeValue(t, m.PostgresConnectionPoolIdle)
	if idleValue != 2.0 {
		t.Errorf("Expected PostgresConnectionPoolIdle to be 2.0, got %f", idleValue)
	}
	
	maxValue := getGaugeValue(t, m.PostgresConnectionPoolMax)
	if maxValue != 10.0 {
		t.Errorf("Expected PostgresConnectionPoolMax to be 10.0, got %f", maxValue)
	}
}

func TestDependencyHealthMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test healthy dependency
	m.UpdateDependencyHealth("postgres", true)
	
	healthyValue := getGaugeVecValue(t, m.DependenciesHealthy, "postgres")
	if healthyValue != 1.0 {
		t.Errorf("Expected healthy postgres to be 1.0, got %f", healthyValue)
	}
	
	// Test unhealthy dependency
	m.UpdateDependencyHealth("redis", false)
	
	unhealthyValue := getGaugeVecValue(t, m.DependenciesHealthy, "redis")
	if unhealthyValue != 0.0 {
		t.Errorf("Expected unhealthy redis to be 0.0, got %f", unhealthyValue)
	}
}

func TestAdminMetrics(t *testing.T) {
	m := createTestMetrics()
	
	// Test admin operation
	m.RecordAdminOperation("get_stats", "success")
	
	counterValue := getCounterVecValue(t, m.AdminOperationsTotal, "get_stats", "success")
	if counterValue != 1.0 {
		t.Errorf("Expected AdminOperationsTotal to be 1.0, got %f", counterValue)
	}
	
	// Test token revocation
	m.RecordTokenRevocation("single")
	m.RecordTokenRevocation("user_all")
	
	singleValue := getCounterVecValue(t, m.AdminTokenRevocationsTotal, "single")
	if singleValue != 1.0 {
		t.Errorf("Expected single token revocations to be 1.0, got %f", singleValue)
	}
	
	allValue := getCounterVecValue(t, m.AdminTokenRevocationsTotal, "user_all")
	if allValue != 1.0 {
		t.Errorf("Expected user_all token revocations to be 1.0, got %f", allValue)
	}
	
	// Test manual cleanup
	m.RecordManualCleanup()
	
	cleanupValue := getCounterValue(t, m.AdminCleanupOperationsTotal)
	if cleanupValue != 1.0 {
		t.Errorf("Expected AdminCleanupOperationsTotal to be 1.0, got %f", cleanupValue)
	}
}

// Helper functions for extracting metric values
func getCounterValue(t *testing.T, counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		t.Fatalf("Failed to write counter metric: %v", err)
	}
	return metric.GetCounter().GetValue()
}

func getCounterVecValue(t *testing.T, counterVec *prometheus.CounterVec, labelValues ...string) float64 {
	counter, err := counterVec.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		t.Fatalf("Failed to get counter with labels %v: %v", labelValues, err)
	}
	return getCounterValue(t, counter)
}

func getGaugeValue(t *testing.T, gauge prometheus.Gauge) float64 {
	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("Failed to write gauge metric: %v", err)
	}
	return metric.GetGauge().GetValue()
}

func getGaugeVecValue(t *testing.T, gaugeVec *prometheus.GaugeVec, labelValues ...string) float64 {
	gauge, err := gaugeVec.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		t.Fatalf("Failed to get gauge with labels %v: %v", labelValues, err)
	}
	return getGaugeValue(t, gauge)
}

func getHistogramCount(t *testing.T, histogram prometheus.Histogram) uint64 {
	metric := &dto.Metric{}
	if err := histogram.Write(metric); err != nil {
		t.Fatalf("Failed to write histogram metric: %v", err)
	}
	return metric.GetHistogram().GetSampleCount()
}

func getHistogramVecCount(t *testing.T, histogramVec *prometheus.HistogramVec, labelValues ...string) uint64 {
	observer, err := histogramVec.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		t.Fatalf("Failed to get histogram with labels %v: %v", labelValues, err)
	}
	
	// Convert Observer to Histogram to access Write method
	histogram, ok := observer.(prometheus.Histogram)
	if !ok {
		t.Fatalf("Failed to convert Observer to Histogram")
	}
	
	return getHistogramCount(t, histogram)
}
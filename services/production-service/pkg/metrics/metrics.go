package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "production_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"query_type", "table"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "production_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type", "table"},
	)

	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "production_redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	ProductionTasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_tasks_total",
			Help: "Total number of production tasks created",
		},
		[]string{"recipe_id", "status"},
	)

	ProductionTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "production_task_duration_seconds",
			Help:    "Production task duration in seconds",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"recipe_id"},
	)

	RecipeUsageTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_recipe_usage_total",
			Help: "Total number of times each recipe has been used",
		},
		[]string{"recipe_id", "recipe_name"},
	)

	InventoryAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "production_inventory_api_calls_total",
			Help: "Total number of calls to Inventory Service API",
		},
		[]string{"endpoint", "status"},
	)

	InventoryAPICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "production_inventory_api_call_duration_seconds",
			Help:    "Inventory Service API call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	ServiceUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "production_service_uptime_seconds",
			Help: "Time since Production Service started in seconds",
		},
	)

	ServiceInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "production_service_info",
			Help: "Production Service information",
		},
		[]string{"version", "build_time"},
	)
)

func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordDBQuery(queryType, table string, duration float64) {
	DBQueriesTotal.WithLabelValues(queryType, table).Inc()
	DBQueryDuration.WithLabelValues(queryType, table).Observe(duration)
}

func RecordRedisOperation(operation, status string, duration float64) {
	RedisOperationsTotal.WithLabelValues(operation, status).Inc()
	RedisOperationDuration.WithLabelValues(operation).Observe(duration)
}

func RecordProductionTask(recipeID, status string) {
	ProductionTasksTotal.WithLabelValues(recipeID, status).Inc()
}

func RecordProductionTaskDuration(recipeID string, duration float64) {
	ProductionTaskDuration.WithLabelValues(recipeID).Observe(duration)
}

func RecordRecipeUsage(recipeID, recipeName string) {
	RecipeUsageTotal.WithLabelValues(recipeID, recipeName).Inc()
}

func RecordInventoryAPICall(endpoint, status string, duration float64) {
	InventoryAPICallsTotal.WithLabelValues(endpoint, status).Inc()
	InventoryAPICallDuration.WithLabelValues(endpoint).Observe(duration)
}
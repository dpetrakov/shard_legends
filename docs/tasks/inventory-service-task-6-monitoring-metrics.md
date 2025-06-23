# Задача 6: Мониторинг и метрики

## Описание

Реализация комплексной системы мониторинга для inventory-service на основе Prometheus метрик и Grafana дашбордов, аналогично auth-service. Включает бизнес-метрики, технические метрики и алерты.

## Цели

1. Изучить существующую реализацию метрик в auth-service
2. Реализовать аналогичные метрики для inventory-service
3. Создать специфичные для инвентаря метрики
4. Настроить Grafana дашборды
5. Добавить алерты для критичных ситуаций

## Подзадачи

### 6.1. Анализ auth-service метрик
**Задача**: Изучить реализацию метрик в auth-service

**Файлы для изучения**:
- `services/auth-service/pkg/metrics/`
- `services/auth-service/internal/middleware/metrics.go`
- `monitoring/grafana/dashboards/auth-service.json`
- `monitoring/prometheus/rules/auth-service.yml`

**Что изучить**:
- Структура метрик
- Naming conventions
- Labels стратегия
- Dashboard layout
- Alert rules

### 6.2. Базовые HTTP метрики
**Файл**: `pkg/metrics/http.go`

**Метрики**:
```go
// HTTP request duration histogram
var HTTPRequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_http_request_duration_seconds",
        Help: "Duration of HTTP requests in seconds",
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "endpoint", "status_code"},
)

// HTTP request count counter
var HTTPRequestTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "endpoint", "status_code"},
)

// HTTP request size histogram
var HTTPRequestSize = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_http_request_size_bytes",
        Help: "Size of HTTP requests in bytes",
        Buckets: prometheus.ExponentialBuckets(100, 10, 6),
    },
    []string{"method", "endpoint"},
)

// HTTP response size histogram
var HTTPResponseSize = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_http_response_size_bytes",
        Help: "Size of HTTP responses in bytes",
        Buckets: prometheus.ExponentialBuckets(100, 10, 6),
    },
    []string{"method", "endpoint", "status_code"},
)
```

### 6.3. Бизнес метрики инвентаря
**Файл**: `pkg/metrics/business.go`

**Метрики**:
```go
// Inventory operations counter
var InventoryOperationsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_operations_total",
        Help: "Total number of inventory operations",
    },
    []string{"operation_type", "section", "status"},
)

// Items balance calculation duration
var BalanceCalculationDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_balance_calculation_duration_seconds",
        Help: "Duration of balance calculations in seconds",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
    },
    []string{"cache_hit"},
)

// Daily balance creation counter
var DailyBalanceCreated = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_daily_balances_created_total",
        Help: "Total number of daily balances created",
    },
    []string{"section"},
)

// Cache operations
var CacheOperations = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_cache_operations_total",
        Help: "Total number of cache operations",
    },
    []string{"operation", "cache_type", "status"},
)

// Cache hit ratio
var CacheHitRatio = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "inventory_cache_hit_ratio",
        Help: "Cache hit ratio",
    },
    []string{"cache_type"},
)

// Current inventory items count
var InventoryItemsCount = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "inventory_items_current_count",
        Help: "Current count of inventory items per user",
    },
    []string{"user_id", "section", "item_class"},
)

// Insufficient balance errors
var InsufficientBalanceErrors = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_insufficient_balance_errors_total",
        Help: "Total number of insufficient balance errors",
    },
    []string{"section", "item_class"},
)

// Transaction rollbacks
var TransactionRollbacks = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "inventory_transaction_rollbacks_total",
        Help: "Total number of transaction rollbacks",
    },
    []string{"operation_type", "reason"},
)
```

### 6.4. Технические метрики
**Файл**: `pkg/metrics/technical.go`

**Метрики**:
```go
// Database connection pool stats
var DBConnections = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "inventory_db_connections",
        Help: "Current database connections",
    },
    []string{"status"}, // active, idle, waiting
)

// Database query duration
var DBQueryDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_db_query_duration_seconds",
        Help: "Duration of database queries in seconds",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
    },
    []string{"query_type", "table"},
)

// Redis connection stats
var RedisConnections = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "inventory_redis_connections",
        Help: "Current Redis connections",
    },
    []string{"status"}, // active, idle
)

// Redis operation duration
var RedisOperationDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "inventory_redis_operation_duration_seconds",
        Help: "Duration of Redis operations in seconds",
        Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1},
    },
    []string{"operation", "status"},
)

// Go runtime metrics (built-in)
var GoMemoryUsage = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "inventory_go_memory_usage_bytes",
        Help: "Go memory usage in bytes",
    },
    []string{"type"}, // heap, stack, gc
)

// Goroutine count
var GoroutineCount = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "inventory_goroutines_total",
        Help: "Total number of goroutines",
    },
)
```

### 6.5. Metrics middleware
**Файл**: `internal/middleware/metrics.go`

```go
type MetricsMiddleware struct {
    metrics *Metrics
}

func NewMetricsMiddleware(m *Metrics) *MetricsMiddleware {
    return &MetricsMiddleware{metrics: m}
}

func (m *MetricsMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Record request size
        if c.Request.ContentLength > 0 {
            HTTPRequestSize.WithLabelValues(
                c.Request.Method,
                c.FullPath(),
            ).Observe(float64(c.Request.ContentLength))
        }
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        // Record request duration and count
        HTTPRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Observe(duration)
        
        HTTPRequestTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        // Record response size
        HTTPResponseSize.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Observe(float64(c.Writer.Size()))
    }
}
```

### 6.6. Business metrics instrumentation
**Файл**: `internal/service/instrumented_service.go`

```go
type InstrumentedInventoryService struct {
    service InventoryService
    metrics *Metrics
}

func (s *InstrumentedInventoryService) CalculateCurrentBalance(ctx context.Context, req BalanceRequest) (int64, error) {
    start := time.Now()
    
    balance, err := s.service.CalculateCurrentBalance(ctx, req)
    
    duration := time.Since(start).Seconds()
    cacheHit := "miss" // определить по логике кеширования
    
    BalanceCalculationDuration.WithLabelValues(cacheHit).Observe(duration)
    
    if err != nil {
        if isInsufficientBalanceError(err) {
            InsufficientBalanceErrors.WithLabelValues(
                req.SectionCode,
                req.ItemClassCode,
            ).Inc()
        }
    }
    
    return balance, err
}

func (s *InstrumentedInventoryService) CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error) {
    operationType := operations[0].OperationType // предполагаем однородные операции
    
    ids, err := s.service.CreateOperationsInTransaction(ctx, operations)
    
    status := "success"
    if err != nil {
        status = "error"
        TransactionRollbacks.WithLabelValues(operationType, getErrorReason(err)).Inc()
    }
    
    InventoryOperationsTotal.WithLabelValues(
        operationType,
        operations[0].Section,
        status,
    ).Add(float64(len(operations)))
    
    return ids, err
}
```

### 6.7. Grafana дашборд
**Файл**: `monitoring/grafana/dashboards/inventory-service.json`

**Панели дашборда**:

1. **Overview**
   - Request rate (req/s)
   - Error rate (%)
   - Response time percentiles (p50, p95, p99)
   - Active users count

2. **HTTP Metrics**
   - Request duration by endpoint
   - Request count by status code
   - Request/Response size histograms
   - Top slowest endpoints

3. **Business Metrics**
   - Inventory operations by type
   - Balance calculations per second
   - Cache hit ratio
   - Daily balances created
   - Insufficient balance errors

4. **Technical Metrics**
   - Database connection pool usage
   - Database query performance
   - Redis operation performance
   - Go memory usage
   - Goroutine count

5. **Alerts Status**
   - Current firing alerts
   - Alert history

### 6.8. Prometheus alert rules
**Файл**: `monitoring/prometheus/rules/inventory-service.yml`

```yaml
groups:
  - name: inventory-service
    rules:
      # High error rate
      - alert: InventoryHighErrorRate
        expr: rate(inventory_http_requests_total{status_code=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on inventory service"
          description: "Error rate is {{ $value }} errors per second"

      # High response time
      - alert: InventoryHighLatency
        expr: histogram_quantile(0.95, rate(inventory_http_request_duration_seconds_bucket[5m])) > 1.0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High latency on inventory service"
          description: "95th percentile latency is {{ $value }} seconds"

      # Database connection issues
      - alert: InventoryDBConnectionIssues
        expr: inventory_db_connections{status="active"} > 80
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted"
          description: "Active DB connections: {{ $value }}"

      # Cache hit ratio too low
      - alert: InventoryCacheHitRatioLow
        expr: inventory_cache_hit_ratio < 0.7
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Cache hit ratio is too low"
          description: "Cache hit ratio is {{ $value }}"

      # High insufficient balance errors
      - alert: InventoryInsufficientBalanceHigh
        expr: rate(inventory_insufficient_balance_errors_total[5m]) > 0.5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High rate of insufficient balance errors"
          description: "Rate: {{ $value }} errors per second"

      # Service down
      - alert: InventoryServiceDown
        expr: up{job="inventory-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Inventory service is down"
          description: "Inventory service has been down for more than 1 minute"
```

### 6.9. Metrics initialization
**Файл**: `pkg/metrics/init.go`

```go
type Metrics struct {
    Registry prometheus.Registerer
}

func NewMetrics() *Metrics {
    return &Metrics{
        Registry: prometheus.DefaultRegisterer,
    }
}

func (m *Metrics) MustRegister() {
    // HTTP metrics
    m.Registry.MustRegister(HTTPRequestDuration)
    m.Registry.MustRegister(HTTPRequestTotal)
    m.Registry.MustRegister(HTTPRequestSize)
    m.Registry.MustRegister(HTTPResponseSize)
    
    // Business metrics
    m.Registry.MustRegister(InventoryOperationsTotal)
    m.Registry.MustRegister(BalanceCalculationDuration)
    m.Registry.MustRegister(DailyBalanceCreated)
    m.Registry.MustRegister(CacheOperations)
    m.Registry.MustRegister(CacheHitRatio)
    m.Registry.MustRegister(InventoryItemsCount)
    m.Registry.MustRegister(InsufficientBalanceErrors)
    m.Registry.MustRegister(TransactionRollbacks)
    
    // Technical metrics
    m.Registry.MustRegister(DBConnections)
    m.Registry.MustRegister(DBQueryDuration)
    m.Registry.MustRegister(RedisConnections)
    m.Registry.MustRegister(RedisOperationDuration)
    m.Registry.MustRegister(GoMemoryUsage)
    m.Registry.MustRegister(GoroutineCount)
    
    // Go runtime metrics
    m.Registry.MustRegister(prometheus.NewGoCollector())
    m.Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}
```

### 6.10. Тесты метрик
**Файлы**: `pkg/metrics/*_test.go`

**Типы тестов**:
- Unit тесты для metrics middleware
- Integration тесты для business metrics
- Performance тесты для metrics overhead
- Тесты корректности labels

## Критерии готовности

### Функциональные
- [ ] Все метрики корректно собираются
- [ ] Grafana дашборд отображает данные
- [ ] Алерты срабатывают при тестовых условиях
- [ ] Метрики доступны на /metrics endpoint
- [ ] Labels корректно применяются

### Технические
- [ ] Overhead метрик < 1% CPU/memory
- [ ] Метрики не влияют на latency
- [ ] Дашборд загружается быстро
- [ ] Alert rules валидны

### Проверочные
- [ ] curl /metrics возвращает данные
- [ ] Grafana показывает live данные
- [ ] Prometheus targets healthy
- [ ] Alertmanager получает уведомления

## Методы тестирования

### 1. Functional тесты
```bash
# Проверка метрик endpoint
curl http://localhost:9090/metrics | grep inventory_

# Проверка Prometheus targets
curl http://prometheus:9090/api/v1/targets

# Проверка Grafana API
curl http://grafana:3000/api/dashboards/uid/inventory
```

### 2. Load тесты для метрик
```bash
# Нагрузочное тестирование с метриками
hey -n 10000 -c 100 http://localhost:8080/inventory

# Проверка overhead метрик
go test -bench=BenchmarkMetrics ./pkg/metrics/...
```

### 3. Alert тесты
```bash
# Симуляция high error rate
# Симуляция high latency
# Проверка срабатывания алертов
```

## Зависимости

### Входящие
- HTTP API (Задача 5)
- Service слой (Задача 4)
- auth-service metrics implementation
- Prometheus/Grafana infrastructure

### Исходящие
- Production-ready мониторинг
- Alerting для операционной команды
- Метрики для capacity planning

## Go зависимости

```go
// Metrics
github.com/prometheus/client_golang    // Prometheus client
github.com/prometheus/common           // Common Prometheus types

// Runtime metrics
github.com/prometheus/procfs           // Process metrics
```

## Заметки по реализации

### Best practices
- Consistent naming convention
- Meaningful labels
- Appropriate metric types (Counter, Gauge, Histogram)
- Reasonable cardinality

### Performance considerations
- Lazy metric initialization
- Efficient label handling
- Minimal allocation in hot paths

### Operational aspects
- Clear alert descriptions
- Actionable alerts only
- Proper severity levels
- Runbook references

## Риски и ограничения

- **Риск**: High cardinality metrics
  **Митигация**: Careful label design, limit user_id labels

- **Риск**: Metrics overhead
  **Митигация**: Performance testing, optimization

- **Ограничение**: Prometheus storage limitations
  **Решение**: Proper retention policies, downsampling
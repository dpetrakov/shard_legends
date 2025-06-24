# Задача 6: Мониторинг и метрики для Inventory Service

## Описание

Реализация комплексной системы мониторинга для inventory-service на основе Prometheus метрик и Grafana дашбордов, аналогично auth-service. Фокус на производительности балансовых расчетов, эффективности кеширования и мониторинге критичных бизнес-операций.

## Цели

1. Дополнить существующие базовые метрики специфичными для inventory-service
2. Создать Grafana дашборд по образцу auth-service-metrics.json
3. Настроить алерты для критичных ситуаций inventory-service
4. Интегрировать метрики в ключевые бизнес-алгоритмы
5. Обеспечить мониторинг производительности кеширования и балансовых расчетов

## Подзадачи

### 6.1. Дополнительные inventory-специфичные метрики
**Задача**: Дополнить существующие базовые метрики в `pkg/metrics/metrics.go`

**Текущие базовые метрики** (уже реализованы):
- HTTP metrics: requests_total, request_duration, requests_in_flight
- Database metrics: connections, queries_total, query_duration  
- Redis metrics: connections, commands_total, command_duration
- Basic business: inventory_operations_total, active_users_total
- Health: dependency_health

**Дополнительные метрики для реализации**:
```go
// Производительность расчета остатков
BalanceCalculationDuration *prometheus.HistogramVec // labels: cache_hit (true/false)

// Создание дневных остатков (ленивое создание)
DailyBalanceCreated *prometheus.CounterVec // labels: section

// Эффективность кеширования
CacheHitRatio *prometheus.GaugeVec // labels: cache_type (balances/classifiers)
CacheOperations *prometheus.CounterVec // labels: operation (get/set/delete), cache_type, status

// Преобразования классификаторов
ClassifierConversions *prometheus.CounterVec // labels: direction (to_uuid/from_uuid), classifier_type

// Бизнес-ошибки
InsufficientBalanceErrors *prometheus.CounterVec // labels: section, item_class
TransactionRollbacks *prometheus.CounterVec // labels: operation_type, reason

// Service lifecycle (для uptime панели)
ServiceUp prometheus.Gauge
ServiceStartTime prometheus.Gauge
```

### 6.2. Grafana дашборд
**Задача**: Создать дашборд по образцу `deploy/monitoring/grafana/dashboards/auth-service-metrics.json`

**Файл**: `deploy/monitoring/grafana/dashboards/inventory-service-metrics.json`

**Структура дашборда** (7 групп панелей):

1. **Service Overview**
   - Service Status (inventory_service_up: UP/DOWN)
   - Service Uptime (time() - inventory_service_start_time_seconds)
   - Memory Usage (process_resident_memory_bytes)
   - Goroutines (go_goroutines)

2. **HTTP Metrics**
   - HTTP Request Rate (inventory_http_requests_total)
   - HTTP Response Time (inventory_http_request_duration_seconds)
   - HTTP Status Codes (inventory_http_requests_total by status)
   - HTTP Requests In Flight (inventory_http_requests_in_flight)

3. **Inventory Business Metrics**
   - Inventory Operations Rate (inventory_operations_total by operation_type)
   - Balance Calculation Performance (inventory_balance_calculation_duration_seconds)
   - Daily Balance Creation Rate (inventory_daily_balance_created_total)
   - Insufficient Balance Errors (inventory_insufficient_balance_errors_total)

4. **Cache Performance**
   - Cache Hit Ratio (inventory_cache_hit_ratio by cache_type)
   - Cache Operations (inventory_cache_operations_total by operation)
   - Classifier Conversions (inventory_classifier_conversions_total)

5. **Database Metrics**
   - PostgreSQL Operations (inventory_database_queries_total, inventory_database_query_duration_seconds)
   - Redis Operations (inventory_redis_commands_total, inventory_redis_command_duration_seconds)
   - Connection Pools (inventory_database_connections, inventory_redis_connections)

6. **Dependencies Health**
   - Dependencies Status (inventory_dependency_health)
   - Transaction Rollbacks (inventory_transaction_rollbacks_total)

7. **Admin Operations**
   - Admin Operations (inventory_operations_total{operation_type="admin_adjustment"})
   - Active Users (inventory_active_users_total)

### 6.3. Prometheus алерты
**Задача**: Создать правила алертов для критичных ситуаций

**Файл**: `deploy/monitoring/prometheus/rules/inventory-service.yml`

**Алерты** (6 правил):

1. **InventoryServiceDown**
   ```yaml
   alert: InventoryServiceDown
   expr: inventory_service_up == 0
   for: 1m
   labels:
     severity: critical
   annotations:
     summary: "Inventory Service is down"
     description: "Inventory Service has been down for more than 1 minute"
   ```

2. **InventoryHighErrorRate**
   ```yaml
   alert: InventoryHighErrorRate
   expr: |
     (
       sum(rate(inventory_http_requests_total{status=~"5.."}[5m])) /
       sum(rate(inventory_http_requests_total[5m]))
     ) > 0.1
   for: 5m
   labels:
     severity: warning
   annotations:
     summary: "High error rate in Inventory Service"
     description: "Error rate is above 10% for 5 minutes"
   ```

3. **InventoryHighLatency**
   ```yaml
   alert: InventoryHighLatency
   expr: |
     histogram_quantile(0.95, 
       sum(rate(inventory_http_request_duration_seconds_bucket[5m])) by (le)
     ) > 2.0
   for: 5m
   labels:
     severity: warning
   annotations:
     summary: "High latency in Inventory Service"
     description: "95th percentile latency is above 2 seconds for 5 minutes"
   ```

4. **InventoryDatabaseIssues**
   ```yaml
   alert: InventoryDatabaseIssues
   expr: |
     (
       sum(rate(inventory_database_queries_total{status="error"}[5m])) /
       sum(rate(inventory_database_queries_total[5m]))
     ) > 0.05
     or
     histogram_quantile(0.95,
       sum(rate(inventory_database_query_duration_seconds_bucket[5m])) by (le)
     ) > 5.0
   for: 2m
   labels:
     severity: warning
   annotations:
     summary: "Database issues in Inventory Service"
     description: "Database error rate >5% or query latency >5s for 2 minutes"
   ```

5. **InventoryCacheProblems**
   ```yaml
   alert: InventoryCacheProblems
   expr: inventory_cache_hit_ratio < 0.7
   for: 10m
   labels:
     severity: warning
   annotations:
     summary: "Low cache hit ratio in Inventory Service"
     description: "Cache hit ratio is below 70% for 10 minutes"
   ```

6. **InventoryBalanceCalculationSlow**
   ```yaml
   alert: InventoryBalanceCalculationSlow
   expr: |
     histogram_quantile(0.95,
       sum(rate(inventory_balance_calculation_duration_seconds_bucket[5m])) by (le)
     ) > 1.0
   for: 5m
   labels:
     severity: warning
   annotations:
     summary: "Slow balance calculations in Inventory Service"
     description: "95th percentile balance calculation time >1s for 5 minutes"
   ```

### 6.4. Интеграция метрик в бизнес-алгоритмы
**Задача**: Добавить метрики в ключевые алгоритмы inventory-service

**Файлы для обновления**:
- `internal/service/balance_calculator.go` - метрики времени расчета и cache hit/miss
- `internal/service/daily_balance_creator.go` - счетчик создания дневных балансов
- `internal/service/cache_manager.go` - метрики cache operations и hit ratio
- `internal/service/code_converter.go` - счетчик преобразований классификаторов
- `internal/middleware/metrics.go` - интеграция новых метрик в HTTP middleware

**Примеры интеграции**:
```go
// В balance_calculator.go
func (bc *BalanceCalculator) CalculateCurrentBalance(...) (int64, error) {
    start := time.Now()
    cacheHit := false
    
    defer func() {
        bc.metrics.BalanceCalculationDuration.WithLabelValues(
            strconv.FormatBool(cacheHit),
        ).Observe(time.Since(start).Seconds())
    }()
    
    // Проверяем кеш
    if cachedBalance, found := bc.cache.Get(cacheKey); found {
        cacheHit = true
        bc.metrics.CacheOperations.WithLabelValues("get", "balances", "hit").Inc()
        return cachedBalance, nil
    }
    
    bc.metrics.CacheOperations.WithLabelValues("get", "balances", "miss").Inc()
    
    // Логика расчета баланса...
    balance := calculateBalance(...)
    
    // Сохранение в кеш
    bc.cache.Set(cacheKey, balance)
    bc.metrics.CacheOperations.WithLabelValues("set", "balances", "success").Inc()
    
    return balance, nil
}

// В daily_balance_creator.go
func (dbc *DailyBalanceCreator) CreateDailyBalance(...) error {
    err := dbc.repository.CreateDailyBalance(...)
    if err == nil {
        dbc.metrics.DailyBalanceCreated.WithLabelValues(sectionCode).Inc()
    }
    return err
}

// В code_converter.go
func (cc *CodeConverter) ConvertToUUID(codes map[string]string) (map[string]uuid.UUID, error) {
    for classifierType := range codes {
        cc.metrics.ClassifierConversions.WithLabelValues("to_uuid", classifierType).Inc()
    }
    return cc.doConversion(codes)
}
```

## Критерии готовности

### Метрики
- [ ] Дополнить `pkg/metrics/metrics.go` новыми inventory-специфичными метриками
- [ ] Обновить структуру Metrics для новых метрик
- [ ] Добавить инициализацию новых метрик в функции New() и Initialize()

### Дашборд
- [ ] Создать `deploy/monitoring/grafana/dashboards/inventory-service-metrics.json`
- [ ] Реализовать 7 групп панелей по образцу auth-service
- [ ] Настроить правильные Prometheus queries для всех панелей
- [ ] Проверить корректность отображения метрик

### Алерты
- [ ] Создать `deploy/monitoring/prometheus/rules/inventory-service.yml`
- [ ] Реализовать 6 алертов для критичных ситуаций
- [ ] Настроить корректные thresholds и временные интервалы
- [ ] Добавить meaningful описания для алертов

### Интеграция
- [ ] Интегрировать метрики в balance_calculator.go
- [ ] Интегрировать метрики в daily_balance_creator.go  
- [ ] Интегрировать метрики в cache_manager.go
- [ ] Интегрировать метрики в code_converter.go
- [ ] Обновить middleware/metrics.go для новых метрик

## Проверка результата

**Команды для тестирования**:
```bash
# Запуск inventory-service с метриками
go run ./cmd/server

# Проверка метрик endpoint
curl http://localhost:8080/metrics | grep inventory_

# Создание операций для генерации метрик
curl -X POST http://localhost:8080/inventory/reserve \
  -H "Authorization: Bearer <jwt>" \
  -d '{"items": [{"item_code": "wood", "quantity": 5}]}'

# Проверка конкретных метрик
curl http://localhost:8080/metrics | grep -E "(inventory_operations_total|inventory_balance_calculation)"
```

**Валидация дашборда**:
1. Открыть Grafana http://localhost:3000
2. Найти дашборд "Inventory Service Metrics"
3. Проверить отображение данных во всех панелях
4. Убедиться в корректности метрик и queries

**Валидация алертов**:
1. Открыть Prometheus http://localhost:9090/alerts
2. Найти правила inventory-service  
3. Создать нагрузку для тестирования алертов
4. Убедиться в корректном срабатывании

## Техническая реализация

### Ключевые решения
1. **Базовые метрики уже есть**: HTTP, Database, Redis метрики уже реализованы в `pkg/metrics/metrics.go`
2. **Фокус на inventory-специфичные метрики**: balance calculation, cache hit ratio, daily balance creation
3. **Структура дашборда по образцу auth-service**: 7 групп панелей с аналогичной визуализацией
4. **Targeted алерты**: 6 критичных алертов для производственных ситуаций
5. **Интеграция в бизнес-логику**: добавление метрик в ключевые алгоритмы без влияния на производительность

### Приоритеты реализации
1. Дополнить метрики в `pkg/metrics/metrics.go`
2. Создать Grafana дашборд по образцу auth-service
3. Настроить Prometheus алерты  
4. Интегрировать метрики в бизнес-алгоритмы
5. Протестировать полную систему мониторинга

### Успешность реализации
Задача считается успешно выполненной при:
- Корректной работе всех метрик на /metrics endpoint
- Отображении данных во всех панелях Grafana дашборда
- Срабатывании алертов при тестовых нагрузках
- Интеграции метрик в бизнес-алгоритмы без влияния на производительность
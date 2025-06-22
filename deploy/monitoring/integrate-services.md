# Интеграция сервисов с мониторингом

## Объяснение проблемы

После запуска стека мониторинга, ваши сервисы работают в разных Docker сетях:

```
┌─────────────────┐    ┌─────────────────┐
│  slcw-dev       │    │ slcw-monitoring │
│                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │auth-service │ │    │ │ prometheus  │ │
│ │ping-service │ │ ❌ │ │ loki        │ │
│ │telegram-bot │ │    │ │ grafana     │ │
│ │api-gateway  │ │    │ │ promtail    │ │
│ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘
     Не могут общаться между собой
```

## Решение: Подключение к обеим сетям

```
┌─────────────────────────────────────────┐
│  Контейнеры в двух сетях одновременно   │
│                                         │
│ ┌─────────────┐    ┌─────────────┐     │
│ │auth-service │◄──►│ prometheus  │     │
│ │ping-service │    │ loki        │     │
│ │telegram-bot │    │ grafana     │     │
│ │api-gateway  │    │ promtail    │     │
│ └─────────────┘    └─────────────┘     │
│      slcw-dev     slcw-monitoring      │
└─────────────────────────────────────────┘
```

## Интеграция с мониторингом

### Обновление docker-compose.yml основных сервисов

Добавьте в конец `deploy/dev/docker-compose.yml`:

```yaml
networks:
  slcw-dev:
    external: false
    name: slcw-dev
  slcw-monitoring:
    external: true
    name: slcw-monitoring
```

И для каждого сервиса обновите секцию networks:

```yaml
# Пример для auth-service
auth-service:
  build:
    context: ../../services/auth-service
    dockerfile: Dockerfile
  container_name: slcw-auth-service-dev
  environment:
    # ... существующие переменные ...
  volumes:
    - auth_jwt_keys:/etc/auth
  networks:
    - slcw-dev        # ← Для общения с другими сервисами
    - slcw-monitoring # ← Для мониторинга
  restart: unless-stopped
  labels:
    - "logging.service=auth-service"     # ← Для Promtail
    - "logging.environment=dev"          # ← Для Promtail
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
      labels: "logging.service,logging.environment"
```

## Добавление метрик в сервисы

### 1. Auth Service (Go)

Добавьте в `internal/handlers/metrics.go`:

```go
package handlers

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func MetricsHandler() http.Handler {
    return promhttp.Handler()
}
```

Добавьте в `cmd/main.go`:

```go
// Добавить эндпоинт метрик
http.Handle("/metrics", handlers.MetricsHandler())
```

### 2. Ping Service (Go)

В `main.go` добавьте:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func main() {
    http.HandleFunc("/ping", pingHandler)
    http.HandleFunc("/health", healthHandler)
    http.Handle("/metrics", promhttp.Handler()) // ← Добавить эту строку
    
    // ... остальной код
}
```

## Проверка работы мониторинга

### 1. Проверить что Prometheus видит сервисы

```bash
# Открыть Prometheus UI
open http://localhost:15090

# Или через API
curl http://localhost:15090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health, endpoint: .scrapeUrl}'
```

### 2. Проверить что логи поступают в Loki

```bash
# Открыть Grafana
open http://localhost:15000

# Или через API
curl -G http://localhost:15100/loki/api/v1/query \
  --data-urlencode 'query={service="auth-service"}' \
  --data-urlencode 'limit=10'
```

### 3. Создать тестовую активность

```bash
# Генерировать запросы к сервисам
for i in {1..10}; do
  curl -X POST http://localhost:9000/ping
  sleep 1
done

# Проверить метрики
curl http://localhost:15090/api/v1/query?query=http_requests_total
```

## Troubleshooting

### Сервисы не видны в Prometheus

```bash
# Проверить что сервисы в нужной сети
docker network inspect slcw-monitoring

# Проверить доступность метрик
docker exec slcw-prometheus wget -qO- http://auth-service:8080/metrics
```

### Логи не поступают в Loki

```bash
# Проверить конфигурацию Promtail
docker logs slcw-promtail

# Проверить что логи в JSON формате
docker logs slcw-auth-service-dev | head -1 | jq .
```

### Grafana не показывает данные

```bash
# Проверить datasources
curl http://localhost:15000/api/datasources

# Проверить что Loki отвечает
curl http://localhost:15100/ready
```
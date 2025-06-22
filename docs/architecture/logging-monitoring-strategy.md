# Стратегия логирования и мониторинга для Shard Legends: Clan Wars

## Обзор

Данный документ описывает стратегию логирования и мониторинга для всех нижних сред (dev, test, stage) проекта. Основные принципы: простота настройки, минимальная поддержка, максимальная польза.

**Ключевые принципы:**
- 🎯 **KISS** - Keep It Simple, Stupid
- 📦 **All-in-One** - Минимум компонентов для развертывания
- 🚀 **Quick Start** - Быстрый запуск из коробки
- 💰 **Cost Effective** - Минимальные требования к ресурсам

## Архитектура решения

### Выбранный стек

Для MVP мы используем **минималистичный, но эффективный стек**:

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Микросервисы   │────▶│   Promtail   │────▶│      Loki       │
│  (stdout logs)  │     │ (collector)  │     │  (log storage)  │
└─────────────────┘     └──────────────┘     └─────────────────┘
                                                      │
┌─────────────────┐     ┌──────────────┐            │
│  Микросервисы   │────▶│  Prometheus  │            │
│   (/metrics)    │     │  (metrics)   │            │
└─────────────────┘     └──────────────┘            │
                                │                    │
                                └────────────────────┘
                                          │
                                    ┌─────────────┐
                                    │   Grafana   │
                                    │(dashboards) │
                                    └─────────────┘
```

### Компоненты

1. **Loki** - Легковесная система хранения логов от Grafana Labs
   - Не требует индексации (как ElasticSearch)
   - Минимальные требования к ресурсам
   - Отлично работает с Docker labels
   - **Порт**: 15100

2. **Promtail** - Агент сбора логов для Loki
   - Автоматически собирает логи из Docker
   - Добавляет метаданные (labels)
   - Минимальная конфигурация
   - **Без внешнего порта** (внутренний сервис)

3. **Prometheus** - Система мониторинга метрик
   - Scraping метрик из /metrics эндпоинтов
   - Хранение временных рядов
   - Алертинг (опционально)
   - **Порт**: 15090

4. **Grafana** - Единая точка визуализации
   - Дашборды для логов и метрик
   - Готовые шаблоны для Docker/микросервисов
   - Алерты по email/Telegram
   - **Порт**: 15000

5. **cAdvisor** - Метрики контейнеров
   - Автоматический сбор метрик Docker
   - CPU, память, сеть, диск
   - **Порт**: 15081

### Порты мониторинга

| Сервис | Внешний порт | Внутренний порт | Описание |
|--------|-------------|----------------|----------|
| Grafana | 15000 | 3000 | Web UI дашбордов |
| Prometheus | 15090 | 9090 | Web UI метрик |
| Loki | 15100 | 3100 | API логов |
| cAdvisor | 15081 | 8080 | Метрики контейнеров |

**Примечание**: Все порты привязаны к 127.0.0.1 для безопасности.

## Стандарты логирования

### 1. Формат логов

**Все сервисы должны логировать в JSON формате:**

```json
{
  "timestamp": "2024-12-22T10:30:45Z",
  "level": "info",
  "service": "auth-service",
  "trace_id": "abc123",
  "message": "User authenticated",
  "user_id": "123",
  "duration_ms": 45
}
```

### 2. Уровни логирования

- **ERROR** - Критические ошибки, требующие внимания
- **WARN** - Предупреждения (rate limit, deprecated API)
- **INFO** - Важные бизнес-события (login, purchase)
- **DEBUG** - Отладочная информация (не в production)

### 3. Обязательные поля

```go
type LogEntry struct {
    Timestamp string `json:"timestamp"`      // ISO 8601
    Level     string `json:"level"`          // error|warn|info|debug
    Service   string `json:"service"`        // Имя сервиса
    TraceID   string `json:"trace_id"`       // Для distributed tracing
    Message   string `json:"message"`        // Человекочитаемое сообщение
}
```

### 4. Чувствительные данные

**НИКОГДА не логировать:**
- Пароли и токены
- Полные данные банковских карт
- Персональные данные (без обезличивания)

**Обезличивание:**
```json
{
  "user_email": "jo***@example.com",
  "phone": "+7******1234"
}
```

## Стандарты метрик

### 1. Обязательные метрики для каждого сервиса

```prometheus
# HTTP метрики
http_requests_total{service="auth-service",method="POST",endpoint="/auth",status="200"}
http_request_duration_seconds{service="auth-service",method="POST",endpoint="/auth"}

# Бизнес-метрики
business_events_total{service="auth-service",event="user_login",status="success"}
business_events_total{service="auth-service",event="user_registration",status="success"}

# Системные метрики (автоматически через cAdvisor)
container_cpu_usage_seconds_total
container_memory_usage_bytes
```

### 2. Эндпоинт метрик

Каждый сервис должен экспортировать метрики на `GET /metrics` в формате Prometheus:

```go
// Пример для Go с promhttp
http.Handle("/metrics", promhttp.Handler())
```

## Docker Compose конфигурация

### monitoring-stack.yml

```yaml
version: '3.8'

services:
  # Loki - хранилище логов
  loki:
    image: grafana/loki:2.9.0
    container_name: slcw-loki
    ports:
      - "3100:3100"
    volumes:
      - loki-data:/loki
      - ./monitoring/loki-config.yml:/etc/loki/local-config.yaml
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - monitoring

  # Promtail - сборщик логов
  promtail:
    image: grafana/promtail:2.9.0
    container_name: slcw-promtail
    volumes:
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - ./monitoring/promtail-config.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
    networks:
      - monitoring
    depends_on:
      - loki

  # Prometheus - метрики
  prometheus:
    image: prom/prometheus:v2.45.0
    container_name: slcw-prometheus
    ports:
      - "9090:9090"
    volumes:
      - prometheus-data:/prometheus
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=7d'
    networks:
      - monitoring

  # Grafana - визуализация
  grafana:
    image: grafana/grafana:10.0.0
    container_name: slcw-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin # Изменить в production!
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
    networks:
      - monitoring
    depends_on:
      - loki
      - prometheus

  # cAdvisor - метрики контейнеров (опционально)
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:v0.47.0
    container_name: slcw-cadvisor
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:ro
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
    networks:
      - monitoring

networks:
  monitoring:
    external: true
    name: slcw-monitoring

volumes:
  loki-data:
  prometheus-data:
  grafana-data:
```

## Конфигурация для микросервисов

### 1. Docker labels для логов

```yaml
# В docker-compose.yml для каждого сервиса
auth-service:
  labels:
    - "logging.service=auth-service"
    - "logging.environment=${ENVIRONMENT:-dev}"
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
      labels: "logging.service,logging.environment"
```

### 2. Структурированное логирование

**Go пример с slog:**
```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

logger = logger.With(
    "service", "auth-service",
    "version", "1.0.0",
)

// Использование
logger.Info("User authenticated",
    "user_id", userID,
    "duration_ms", duration.Milliseconds(),
    "trace_id", traceID,
)
```

## Дашборды Grafana

### 1. Системный дашборд

**Панели:**
- CPU использование по сервисам
- Память по сервисам
- Сетевой трафик
- Disk I/O

### 2. Бизнес-дашборд

**Панели:**
- Количество регистраций/логинов
- Активные пользователи
- Ошибки авторизации
- Среднее время ответа API

### 3. Дашборд ошибок

**Панели:**
- Error rate по сервисам
- Последние ошибки (логи)
- Top 10 ошибок
- Алерты

## Алертинг

### Базовые правила алертов

```yaml
# prometheus/alerts.yml
groups:
  - name: basic
    rules:
      - alert: ServiceDown
        expr: up == 0
        for: 5m
        annotations:
          summary: "Service {{ $labels.job }} is down"
          
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate in {{ $labels.service }}"
          
      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
        for: 5m
        annotations:
          summary: "High memory usage in {{ $labels.container_name }}"
```

## Развертывание

### 1. Подготовка

```bash
# Создание директории для конфигураций
mkdir -p deploy/monitoring/{loki,promtail,prometheus,grafana/provisioning}

# Создание сети
docker network create slcw-monitoring
```

### 2. Запуск мониторинга

```bash
# Запуск стека мониторинга
docker-compose -f deploy/monitoring/docker-compose.yml up -d

# Проверка статуса
docker-compose -f deploy/monitoring/docker-compose.yml ps
```

### 3. Доступ к интерфейсам

- **Grafana**: http://localhost:15000 (admin/admin)
- **Prometheus**: http://localhost:15090  
- **Loki**: http://localhost:15100
- **cAdvisor**: http://localhost:15081

**Примечание**: Используются 5-значные порты для избежания конфликтов с основными сервисами.

### 4. Интеграция с сервисами

Обновите `deploy/dev/docker-compose.yml` для подключения всех сервисов к мониторингу:

```yaml
# Пример для любого сервиса
your-service:
  # ... существующая конфигурация ...
  
  # Подключение к обеим сетям
  networks:
    - slcw-dev        # Для общения с другими сервисами
    - slcw-monitoring # Для мониторинга
  
  # Логирование (для Promtail)
  labels:
    - "logging.service=your-service"
    - "logging.environment=${ENVIRONMENT:-dev}"
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
      labels: "logging.service,logging.environment"

# Добавить сети в конец файла
networks:
  slcw-dev:
    external: false
    name: slcw-dev
  slcw-monitoring:
    external: true
    name: slcw-monitoring
```

После обновления:

```bash
# 1. Запустить мониторинг (создаст сеть автоматически)
cd deploy/monitoring
docker-compose up -d

# 2. Пересоздать основные сервисы
cd ../dev
docker-compose down
docker-compose up -d
```

## Хранение и ротация

### Политики хранения

- **Логи**: 7 дней (Loki)
- **Метрики**: 7 дней (Prometheus)
- **Дашборды**: Persistent (Grafana)

### Backup стратегия

```bash
# Backup скрипт (запускать по cron)
#!/bin/bash
DATE=$(date +%Y%m%d)
docker exec slcw-prometheus tar czf /tmp/prometheus-backup-$DATE.tar.gz /prometheus
docker exec slcw-grafana tar czf /tmp/grafana-backup-$DATE.tar.gz /var/lib/grafana
# Копирование на backup сервер
```

## Мониторинг для разработчиков

### Локальная разработка

```bash
# Просмотр логов сервиса
docker logs -f slcw-auth-service-dev

# Просмотр метрик
curl http://localhost:9000/metrics

# Grafana UI
open http://localhost:3000
```

### Отладка

1. **Trace ID** - Используйте для отслеживания запросов через сервисы
2. **Correlation** - Связывайте логи с метриками через labels
3. **Dashboards** - Создавайте временные дашборды для отладки

## Best Practices

### ✅ DO:
- Логировать все важные бизнес-события
- Использовать структурированные логи (JSON)
- Добавлять trace_id для distributed tracing
- Экспортировать бизнес-метрики
- Настраивать алерты на критические события

### ❌ DON'T:
- Не логировать чувствительные данные
- Не использовать print/console.log в production
- Не игнорировать ошибки логирования
- Не создавать слишком много метрик (cardinality)

## Roadmap

### Phase 1 (MVP) ✅
- [x] Базовый стек Loki + Prometheus + Grafana
- [x] Структурированное логирование
- [x] Основные метрики
- [x] Базовые дашборды

### Phase 2
- [ ] Distributed tracing (Jaeger/Tempo)
- [ ] Алертинг в Telegram
- [ ] Автоматические дашборды
- [ ] SLO/SLI метрики

### Phase 3
- [ ] APM (Application Performance Monitoring)
- [ ] Логи безопасности (WAF, IDS)
- [ ] ML-based anomaly detection
- [ ] Интеграция с внешними системами

## Заключение

Данная стратегия обеспечивает:
- 🚀 **Быстрый старт** - развертывание за 5 минут
- 💡 **Простоту** - минимум компонентов
- 📊 **Полноту** - логи + метрики + визуализация
- 💰 **Экономичность** - работает на 2GB RAM
- 🔧 **Расширяемость** - легко добавить новые возможности

Для production среды на Kubernetes будет отдельная стратегия с учетом cloud-native подходов.
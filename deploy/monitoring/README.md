# Мониторинг стек для Shard Legends

Полный стек мониторинга на основе Loki + Prometheus + Grafana для всех нижних сред.

## Быстрый старт

```bash
# Переход в директорию мониторинга
cd deploy/monitoring

# Запуск мониторинга
docker-compose up -d

# Проверка статуса
docker-compose ps
```

### 2. Доступ к интерфейсам

- **Grafana**: http://localhost:15000 (admin/admin)
- **Prometheus**: http://localhost:15090
- **Loki**: http://localhost:15100
- **cAdvisor**: http://localhost:15081

### 3. Интеграция основных сервисов

```bash
# Перезапуск основных сервисов для подключения к мониторингу
cd ../dev
docker-compose down
docker-compose up -d
```

## Настройка сервисов для мониторинга

### 1. Добавление в docker-compose.yml

```yaml
# Пример для любого сервиса
your-service:
  # ... существующая конфигурация ...
  
  # Логирование
  labels:
    - "logging.service=your-service"
    - "logging.environment=${ENVIRONMENT:-dev}"
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
      labels: "logging.service,logging.environment"
  
  # Подключение к сети мониторинга
  networks:
    - slcw-dev
    - slcw-monitoring

networks:
  slcw-monitoring:
    external: true
```

### 2. Экспорт метрик

Каждый сервис должен экспортировать метрики на `/metrics`:

**Go пример:**
```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Добавить в main.go
http.Handle("/metrics", promhttp.Handler())
```

**Node.js пример:**
```javascript
const client = require('prom-client');

// Создание registry
const register = new client.Registry();

// Базовые метрики
client.collectDefaultMetrics({ register });

// Эндпоинт метрик
app.get('/metrics', (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(register.metrics());
});
```

## Структурированное логирование

### Go с slog

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

logger = logger.With(
    "service", "your-service",
    "version", "1.0.0",
)

// Использование
logger.Info("Event occurred",
    "user_id", userID,
    "duration_ms", duration.Milliseconds(),
    "trace_id", traceID,
)
```

### Node.js с winston

```javascript
const winston = require('winston');

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  defaultMeta: { 
    service: 'your-service',
    version: '1.0.0'
  },
  transports: [
    new winston.transports.Console()
  ],
});

// Использование
logger.info('Event occurred', {
  user_id: userId,
  duration_ms: duration,
  trace_id: traceId
});
```

## Создание дашбордов

### 1. Импорт готовых дашбордов

В Grafana UI:
1. Перейти в "+" → "Import"
2. Использовать ID популярных дашбордов:
   - **Docker Container & Host Metrics**: 10619
   - **Prometheus Stats**: 2
   - **Loki Dashboard**: 13639

### 2. Создание кастомного дашборда

**Панель для логов ошибок:**
```
{level="error"} |= "your-service"
```

**Панель для метрик RPS:**
```
rate(http_requests_total[5m])
```

**Панель для memory usage:**
```
container_memory_usage_bytes{name=~".*your-service.*"}
```

## Алертинг

### 1. Настройка уведомлений в Telegram

1. Создать бота через @BotFather
2. Получить chat_id через @userinfobot
3. Добавить notification channel в Grafana:

```yaml
# В grafana/provisioning/notifiers/telegram.yml
notifiers:
  - name: telegram
    type: telegram
    settings:
      bottoken: "YOUR_BOT_TOKEN"
      chatid: YOUR_CHAT_ID
```

### 2. Создание алертов

В Grafana UI создавайте алерты для критических метрик:
- Service Down
- High Error Rate
- High Memory/CPU Usage
- Business Events (failures)

## Backup и восстановление

### Backup данных

```bash
#!/bin/bash
DATE=$(date +%Y%m%d)

# Backup Prometheus data
docker exec slcw-prometheus tar czf /tmp/prometheus-backup-$DATE.tar.gz /prometheus

# Backup Grafana data  
docker exec slcw-grafana tar czf /tmp/grafana-backup-$DATE.tar.gz /var/lib/grafana

# Copy backups to host
docker cp slcw-prometheus:/tmp/prometheus-backup-$DATE.tar.gz ./backups/
docker cp slcw-grafana:/tmp/grafana-backup-$DATE.tar.gz ./backups/
```

### Восстановление

```bash
# Остановить сервисы
docker-compose down

# Восстановить данные
docker run --rm -v prometheus-data:/data -v $(pwd)/backups:/backup alpine \
  tar xzf /backup/prometheus-backup-YYYYMMDD.tar.gz -C /data

# Запустить сервисы
docker-compose up -d
```

## Проверка портов

### Проверить что порты мониторинга свободны

```bash
# Быстрая проверка всех портов мониторинга
for port in 15000 15090 15100 15081; do 
  if lsof -i :$port > /dev/null 2>&1; then 
    echo "❌ Порт $port занят"; 
  else 
    echo "✅ Порт $port свободен"; 
  fi; 
done
```

### Проверить доступность после запуска

```bash
# Проверка health endpoints
curl -s -o /dev/null -w "Grafana (15000): %{http_code}\n" http://localhost:15000/api/health
curl -s -o /dev/null -w "Prometheus (15090): %{http_code}\n" http://localhost:15090/-/healthy  
curl -s -o /dev/null -w "Loki (15100): %{http_code}\n" http://localhost:15100/ready
curl -s -o /dev/null -w "cAdvisor (15081): %{http_code}\n" http://localhost:15081/healthz
```

## Troubleshooting

### Проблема: Promtail не собирает логи

**Решение:**
1. Проверить права доступа к Docker socket
2. Убедиться что логи в JSON формате
3. Проверить конфигурацию в promtail-config.yml

### Проблема: Метрики не поступают в Prometheus

**Решение:**
1. Проверить доступность эндпоинта `/metrics`
2. Убедиться что сервис в сети `slcw-monitoring`
3. Проверить targets в Prometheus UI

### Проблема: Grafana не может подключиться к Loki/Prometheus

**Решение:**
1. Проверить health check контейнеров
2. Убедиться что все сервисы в одной сети
3. Проверить конфигурацию datasources

## Мониторинг ресурсов

Весь стек мониторинга требует:
- **RAM**: ~1GB
- **CPU**: 0.5 core
- **Disk**: ~2GB для данных за неделю

Для production рекомендуется увеличить retention и добавить persistent volumes.
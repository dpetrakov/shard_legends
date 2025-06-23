# Задача 2: Создание базового Go приложения

## Описание

Создание основной структуры Go приложения для inventory-service с базовой конфигурацией, подключением к БД, логированием и health check эндпоинтами.

## Цели

1. Инициализировать Go модуль и базовую структуру проекта
2. Настроить подключение к PostgreSQL и Redis
3. Реализовать базовое логирование и конфигурацию
4. Создать health check и metrics эндпоинты
5. Настроить Docker для развертывания

## Подзадачи

### 2.1. Структура проекта и Go модуль
**Директория**: `services/inventory-service/`

**Структура**:
```
inventory-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── models/
│   ├── repository/
│   ├── service/
│   └── middleware/
├── pkg/
│   ├── logger/
│   └── metrics/
├── migrations/
├── docker/
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

### 2.2. Конфигурация приложения
**Файл**: `internal/config/config.go`

**Содержание**:
- Чтение переменных окружения
- Настройки БД (PostgreSQL, Redis)
- Настройки сервера (порт, таймауты)
- Настройки логирования
- Настройки метрик

**Поддерживаемые переменные**:
```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shard_legends
DB_USER=postgres
DB_PASSWORD=password
REDIS_HOST=localhost
REDIS_PORT=6379
SERVER_PORT=8080
LOG_LEVEL=info
METRICS_PORT=9090
```

### 2.3. Подключение к базе данных
**Файл**: `internal/database/postgres.go`

**Содержание**:
- Инициализация connection pool
- Ping проверка соединения
- Graceful shutdown
- Миграции (опционально)

**Файл**: `internal/database/redis.go`

**Содержание**:
- Подключение к Redis
- Connection pool настройки
- Health check методы

### 2.4. Логирование
**Файл**: `pkg/logger/logger.go`

**Содержание**:
- Structured logging (JSON format)
- Различные уровни логирования
- Request ID для трассировки
- Интеграция с middleware

### 2.5. Health Check эндпоинты
**Файл**: `internal/handlers/health.go`

**Эндпоинты**:
```
GET /health        - Простая проверка доступности
GET /health/ready  - Проверка готовности (БД, Redis)
GET /health/live   - Liveness probe для Kubernetes
```

**Ответы**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "database": "connected",
    "redis": "connected"
  }
}
```

### 2.6. Метрики (базовые)
**Файл**: `pkg/metrics/metrics.go`

**Метрики**:
- HTTP request duration
- HTTP request count
- Database connection pool stats
- Redis connection stats
- Go runtime metrics

### 2.7. HTTP сервер и роутинг
**Файл**: `cmd/server/main.go`

**Содержание**:
- Инициализация всех компонентов
- Graceful shutdown
- Error handling
- Signal handling

**Файл**: `internal/handlers/router.go`

**Содержание**:
- Настройка маршрутов
- Middleware подключение
- CORS настройки (если нужно)

### 2.8. Docker конфигурация
**Файл**: `Dockerfile`

**Содержание**:
- Multi-stage build
- Оптимизированный образ
- Non-root user
- Health check команда

**Файл**: `docker-compose.yml`

**Содержание**:
- inventory-service
- PostgreSQL
- Redis
- Сеть и volumes

## Критерии готовности

### Функциональные
- [ ] Приложение запускается без ошибок
- [ ] Health check эндпоинты возвращают корректные ответы
- [ ] Подключение к PostgreSQL работает
- [ ] Подключение к Redis работает
- [ ] Логи выводятся в JSON формате
- [ ] Метрики доступны на /metrics

### Технические
- [ ] Go модуль корректно настроен
- [ ] Все зависимости установлены
- [ ] Docker образ собирается успешно
- [ ] docker-compose поднимает стек

### Проверочные
- [ ] curl -X GET localhost:8080/health возвращает 200
- [ ] curl -X GET localhost:8080/health/ready возвращает статус БД
- [ ] curl -X GET localhost:9090/metrics возвращает метрики
- [ ] Логи содержат timestamp и level

## Методы тестирования

### 1. Локальная разработка
```bash
# Запуск локально
go run cmd/server/main.go

# Проверка health
curl http://localhost:8080/health

# Проверка метрик
curl http://localhost:9090/metrics
```

### 2. Docker тестирование
```bash
# Сборка образа
docker build -t inventory-service .

# Запуск стека
docker-compose up -d

# Проверка логов
docker-compose logs inventory-service
```

### 3. Интеграционные тесты
```bash
# Тесты подключения к БД
go test ./internal/database/...

# Тесты health endpoints
go test ./internal/handlers/...
```

## Зависимости

### Входящие
- Готовые миграции БД (Задача 1)
- PostgreSQL 17 и Redis 8.0.2
- Go 1.21+

### Исходящие
- Готовая платформа для API разработки
- Базовые метрики и мониторинг
- Docker образ для развертывания

## Go зависимости

```go
// Основные зависимости
github.com/gin-gonic/gin           // HTTP framework
github.com/lib/pq                  // PostgreSQL driver
github.com/go-redis/redis/v8       // Redis client
github.com/prometheus/client_golang // Metrics
github.com/sirupsen/logrus         // Logging
github.com/kelseyhightower/envconfig // Config
github.com/google/uuid             // UUID generation

// Dev зависимости
github.com/stretchr/testify        // Testing
github.com/golang/mock             // Mocking
```

## Заметки по реализации

### Архитектурные принципы
- Clean Architecture подход
- Dependency Injection
- Interface-based design
- Graceful shutdown паттерн

### Безопасность
- Не логировать чувствительные данные
- Валидация всех входных параметров
- Rate limiting (в будущих задачах)

### Производительность
- Connection pooling для БД
- Context-based cancellation
- Оптимизированные Docker образы

## Риски и ограничения

- **Риск**: Проблемы с connection pooling
  **Митигация**: Тщательное тестирование настроек

- **Риск**: Memory leaks в long-running процессах
  **Митигация**: Профилирование и мониторинг

- **Ограничение**: Требует Docker для полного тестирования
  **Решение**: Документировать локальную разработку
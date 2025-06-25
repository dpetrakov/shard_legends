# Inventory Service

Inventory Service - это микросервис для управления инвентарем пользователей в игре Shard Legends: Clan Wars.

## Функциональность

- Управление инвентарем пользователей
- Система классификаторов для справочных данных
- Health check эндпоинты
- Метрики Prometheus
- Structured JSON логирование

## Технологии

- **Язык**: Golang 1.23
- **База данных**: PostgreSQL 17
- **Кеширование**: Redis 8.0.2
- **HTTP Framework**: Gin
- **Метрики**: Prometheus
- **Контейнеризация**: Docker

## Быстрый старт

### Запуск в рамках dev стенда

```bash
# Перейти в директорию dev окружения
cd deploy/dev

# Запустить полный стек (включая inventory-service)
docker-compose up -d

# Проверить статус
docker-compose ps

# Проверить логи inventory-service
docker-compose logs inventory-service

# Проверить работу через API Gateway
curl http://localhost:9000/inventory/health
```

### Локальная разработка

```bash
# Установить зависимости
go mod download

# Запустить только базы данных из dev стенда
cd deploy/dev
docker-compose up -d postgres redis

# Настроить переменные окружения
export DATABASE_URL="postgres://slcw_user:dev_password_2024@localhost:5432/shard_legends_dev?sslmode=disable"
export REDIS_URL="redis://localhost:6379/1"

# Вернуться в директорию сервиса и запустить
cd ../../services/inventory-service
go run cmd/server/main.go
```

## API Endpoints

### Доступ через API Gateway
Все эндпоинты доступны через API Gateway с префиксом `/api/inventory`:
- External: `https://dev.slcw.dimlight.online/api/inventory/*`
- Internal: `http://localhost:9000/inventory/*`

### Основные эндпоинты
- `GET /api/inventory` - Получить инвентарь пользователя (требует JWT)
- `POST /api/inventory/adjust` - Административная корректировка инвентаря (требует JWT + admin права)

### Health Check
- `GET /health` - Проверка работоспособности (БД + Redis + общий статус)

### Метрики
- `GET /metrics` - Метрики Prometheus (только внутренний доступ)

## Конфигурация

Сервис настраивается через переменные окружения:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `INVENTORY_SERVICE_HOST` | Хост сервиса | `0.0.0.0` |
| `INVENTORY_SERVICE_PORT` | Порт сервиса | `8080` |
| `DATABASE_URL` | URL PostgreSQL | **обязательно** |
| `DATABASE_MAX_CONNECTIONS` | Макс. соединений с БД | `10` |
| `REDIS_URL` | URL Redis | **обязательно** |
| `REDIS_MAX_CONNECTIONS` | Макс. соединений с Redis | `10` |
| `LOG_LEVEL` | Уровень логирования | `info` |
| `METRICS_PORT` | Порт для метрик | `9090` |

## Архитектура

```
inventory-service/
├── cmd/server/          # Точка входа приложения
├── internal/            # Внутренняя логика
│   ├── config/         # Конфигурация
│   ├── database/       # Подключения к БД
│   ├── handlers/       # HTTP обработчики
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Модели данных
│   ├── repository/     # Слой данных
│   └── service/        # Бизнес-логика
├── pkg/                # Переиспользуемые пакеты
│   ├── logger/         # Логирование
│   └── metrics/        # Метрики
├── migrations/         # Миграции БД
└── docker/            # Docker конфигурации
```

## Мониторинг

### Метрики Prometheus

Сервис экспортирует следующие метрики:

- `inventory_http_requests_total` - Общее количество HTTP запросов
- `inventory_http_request_duration_seconds` - Время выполнения HTTP запросов
- `inventory_database_connections` - Количество соединений с БД
- `inventory_redis_connections` - Количество соединений с Redis
- `inventory_dependency_health` - Статус здоровья зависимостей

### Health Check

- **Basic**: `/health` - проверяет работоспособность сервиса и его зависимостей

## Разработка

### Требования

- Go 1.23+
- Docker и Docker Compose
- PostgreSQL 17
- Redis 8.0.2

### Запуск тестов

```bash
# Unit тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Integration тесты (требуют запущенные БД)
go test -tags=integration ./...
```

### Линтинг

```bash
# Запуск линтера
golangci-lint run

# Форматирование кода
go fmt ./...
```

## Docker

### Сборка образа

```bash
docker build -t inventory-service .
```

### Multi-stage build

Dockerfile использует multi-stage build для оптимизации размера образа:
- Build stage: компиляция на основе golang:1.23-alpine
- Runtime stage: минимальный alpine образ с бинарником

## Безопасность

- Приложение запускается от непривилегированного пользователя
- Нет sensitive данных в логах
- Graceful shutdown с таймаутом
- Health checks для мониторинга

## Логирование

Структурированное JSON логирование с полями:
- `timestamp` - время события
- `level` - уровень логирования
- `msg` - сообщение
- `method`, `path`, `status` - для HTTP запросов
- `error` - для ошибок

## Troubleshooting

### Частые проблемы

1. **Сервис не стартует**
   ```bash
   # Проверить логи
   docker-compose logs inventory-service
   
   # Проверить конфигурацию
   docker-compose config
   ```

2. **Проблемы с БД**
   ```bash
   # Проверить статус PostgreSQL
   docker-compose exec postgres pg_isready -U postgres
   
   # Проверить логи БД
   docker-compose logs postgres
   ```

3. **Проблемы с Redis**
   ```bash
   # Проверить статус Redis
   docker-compose exec redis redis-cli ping
   
   # Проверить логи Redis
   docker-compose logs redis
   ```

### Полезные команды

```bash
# Перезапустить только сервис
docker-compose restart inventory-service

# Пересобрать и запустить
docker-compose up --build

# Остановить и удалить все
docker-compose down -v

# Мониторинг логов в реальном времени
docker-compose logs -f inventory-service
```
# Deck Game Service

Микросервис для управления игровой механикой ежедневной выдачи сундуков в мини-игре «Дека» проекта Shard Legends: Clan Wars.

## Описание

Deck Game Service (DGS) обеспечивает:
- Расчёт ожидаемого комбо для следующей награды
- Валидацию и выдачу награды за выполненное комбо 
- Интеграцию с Production Service для создания сундуков
- Актуализацию инвентаря игрока через Inventory Service
- JWT-аутентификацию и метрики Prometheus

## Порты

- `8080` — публичный API (`/deck/*`)
- `8090` — внутренний API (`/health`, `/metrics`)

## Зависимости

- PostgreSQL 17 (schema production)
- Redis 8 база 0 (проверка отзыва JWT, разделяется с auth-service)
- Production Service (создание сундуков)
- Inventory Service (детализация предметов)  
- Auth Service (JWT ключи)

## API Endpoints

### Публичный API
- `GET /deck/daily-chest/status` - статус ежедневных сундуков
- `POST /deck/daily-chest/claim` - заявить награду

### Внутренний API  
- `GET /health` - проверка здоровья сервиса
- `GET /metrics` - метрики Prometheus

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT_PUBLIC` | `8080` | HTTP порт публичного API |
| `PORT_INTERNAL` | `8090` | HTTP порт внутреннего API |
| `DATABASE_URL` | — | PostgreSQL DSN (обязательный) |
| `REDIS_AUTH_URL` | `redis://redis:6379/0` | Redis URL для проверки JWT |
| `PRODUCTION_EXTERNAL_URL` | `http://production-service:8080` | Production Service API |
| `INVENTORY_INTERNAL_URL` | `http://inventory-service:8080` | Inventory Service API |
| `AUTH_PUBLIC_KEY_URL` | `http://auth-service:8090/public-key.pem` | Публичный RSA-ключ |
| `COOLDOWN_SEC` | `30` | Минимальный интервал между наградами |
| `DAILY_CHEST_RECIPE_ID` | `9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2` | ID рецепта сундука |
| `LOG_LEVEL` | `info` | Уровень логирования |

## Разработка

```bash
# Сборка
cd services/deck-game-service
go build -o deck-game-service ./cmd/server

# Запуск локально
export DATABASE_URL="postgresql://user:pass@localhost:5432/db"
./deck-game-service

# Запуск в Docker
cd deploy/dev
docker compose up deck-game-service
```

## Статус реализации

- ✅ D-Deck-001: Каркас сервиса (текущая задача)
- ⏳ D-Deck-002: Конфигурация и переменные окружения
- ⏳ D-Deck-003: JWT Middleware
- ⏳ D-Deck-004: Endpoint GET /deck/daily-chest/status
- ⏳ D-Deck-005: Endpoint POST /deck/daily-chest/claim

Полная спецификация: [docs/specs/deck-game-service.md](../../docs/specs/deck-game-service.md) 
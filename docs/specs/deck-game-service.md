# Спецификация сервиса Deck Game Service (Мини-игра «Дека»)

## Обзор

Deck Game Service (далее — **DGS**) отвечает за игровую механику ежедневной выдачи сундуков в мини-игре «Дека». Сервис предоставляет публичный REST API для фронтенда, внутренний REST API для мониторинга и использует внутренние API production- и inventory-сервисов для мгновенного создания и получения наград.

**Ключевые возможности:**
1. Расчёт ожидаемого комбо для следующей награды (5 → 15).
2. Валидация и выдача награды за выполненное комбо с соблюдением cooldown.
3. Интеграция с Production Service для запуска рецептов и мгновенного claim.
4. Актуализация инвентаря игрока через Inventory Service.
5. JWT-аутентификация, rate limiting, метрики Prometheus, полное покрытие тестами.

**Порты по умолчанию:**
- `8080` — публичный API (`/deck/*`)
- `8090` — внутренний API (`/metrics`, `/health`)

**Зависимости:** PostgreSQL 17 (schema production), Redis 8 (проверка отзыва JWT), Production Service, Inventory Service.


## Глоссарий
| Термин | Описание |
|-------|----------|
| Combo | Количество совпавших карт, которое игрок собрал в раунде мини-игры |
| Daily Chest | Сундук, выдаваемый за выполнение combo, создаётся рецептом `DAILY_CHEST_RECIPE_ID` |
| Cooldown | Минимальное время между двумя успешными наградами, по умолчанию 30 сек |
| Crafts Count | Количество успешных крафтов сундука конкретным игроком за календарный день |


## Константы бизнес-логики
```text
DAILY_CHEST_RECIPE_ID = "9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2"
COOLDOWN_DEFAULT_SEC  = 30
MAX_DAILY_CRAFTS      = 10
BASE_COMBO            = 5   // для 0 крафтов
MAX_COMBO             = 15  // для 10 крафтов (не выдаём сундуки)
```

Формула ожидаемого комбо:
```text
expected_combo = BASE_COMBO + crafts_count // crafts_count ∈ [0;9]
```
Если `crafts_count ≥ MAX_DAILY_CRAFTS` → награды на сегодня завершены.


## Потоки данных
```mermaid
sequenceDiagram
    participant FE as Frontend (Mini App)
    participant DGS as Deck Game Service
    participant PSDB as Production DB (shared)
    participant PSAPI as Production Service (internal)
    participant IS as Inventory Service (internal)
    FE->>DGS: GET /deck/daily-chest/status (JWT)
    DGS->>PSDB: SELECT COUNT(completed tasks) WHERE recipe_id = DAILY_CHEST_RECIPE_ID AND today
    DGS-->>FE: 200 {expected_combo, finished}
    
    FE->>DGS: POST /deck/daily-chest/claim {combo,chestIndex}
    DGS->>PSDB: SELECT COUNT(completed tasks) WHERE recipe_id = DAILY_CHEST_RECIPE_ID AND today
    DGS-->>FE: 400 daily_finished if combo < expected OR finished
    DGS->>PSAPI: POST /production/factory/start
    PSAPI-->>DGS: 200 {task_id}
    DGS->>PSAPI: POST /production/factory/claim {task_id}
    PSAPI-->>DGS: 200 {item_id}
    DGS->>IS: POST /items/details {item_id}
    IS-->>DGS: 200 {item_details}
    DGS-->>FE: 200 {item_details, next_expected_combo}
```


## Публичный API
### 1. Получить статус ежедневных сундуков
`GET /deck/daily-chest/status`

| Параметр | Тип | Описание |
|----------|-----|----------|
| — | — | Нет query-параметров |

**Заголовки:** `Authorization: Bearer <jwt>`

**Ответ 200**
```json
{
  "expected_combo": 7,
  "finished": false,
  "crafts_done": 2,
  "last_reward_at": "2024-07-01T10:45:23Z"
}
```
Если `finished=true`, поле `expected_combo` опускается.

### 2. Заявить награду
`POST /deck/daily-chest/claim`

**Payload**
```json
{
  "combo": 8,
  "chest_index": 3
}
```

**Ответ 200**
```json
{
  "items": [
    {
      "item_id": "bb3b0f2d-…",
      "code": "daily_chest",
      "name": "Daily Chest",
      "description": "Сундук с ежедневной наградой",
      "collection": null,
      "quality_level": "rare",
      "quantity": 1,
      "image_url": "https://…/daily_chest_rare.png"
    }
  ],
  "next_expected_combo": 9,
  "crafts_done": 3
}
```

**Ошибки**
| Код | Короткое сообщение | Причина |
|-----|--------------------|---------|
| 400 | `invalid_combo` | Переданное `combo` < ожидаемого |
| 400 | `daily_finished` | Лимит 10 наград достигнут **или запрос пришёл раньше минимального интервала между наградами** |
| 404 | `recipe_not_found` | Константа `DAILY_CHEST_RECIPE_ID` неконсистентна |


## Внутренний API
| Метод | Роут | Описание |
|-------|------|----------|
| GET | `/health` | Liveness/Readiness |
| GET | `/metrics` | Prometheus |


## Хранение данных и вычисление прогресса

Deck Game Service не хранит собственную таблицу прогресса. Информация о выданных сундуках определяется запросом к БД Production Service:

```sql
SELECT COUNT(*)
  FROM production.production_tasks
WHERE user_id = :user_id
  AND recipe_id = :DAILY_CHEST_RECIPE_ID
  AND status = 'completed'
  AND completed_at::date = current_date;
```

Этот запрос возвращает количество успешных крафтов сундука за текущие сутки. Для расчёта `expected_combo` и проверки лимита используется именно это значение.


## Конфигурация (ENV)
| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT_PUBLIC` | `8080` | HTTP порт публичного API |
| `PORT_INTERNAL` | `8090` | HTTP порт внутреннего API |
| `DATABASE_URL` | — | PostgreSQL DSN (production schema) |
| `PRODUCTION_INTERNAL_URL` | `http://production-service:8080` | Внутренние эндпоинты (start/claim) |
| `PRODUCTION_EXTERNAL_URL` | `https://api.shardlegends.com/production` | Внешние эндпоинты через API Gateway |
| `INVENTORY_INTERNAL_URL` | `http://inventory-service:8080` | Внутренние эндпоинты Inventory |
| `COOLDOWN_SEC` | `30` | Минимальный интервал между наградами |
| `DAILY_CHEST_RECIPE_ID` | см. выше | Идентификатор рецепта |
| `JWT_PUBLIC_KEYS_URL` | `http://auth-service:8090/keys` | Для валидации JWT |
| `AUTH_PUBLIC_KEY_PEM_URL` | `http://auth-service:8080/public-key.pem` | PEM-ключ (используется по умолчанию, как в Inventory) |
| `REDIS_URL` | `redis://redis:6379/0` | Проверка отзыва токенов |


## Интеграции с другими сервисами

### Production Service
| Действие | HTTP | Тело |
|----------|------|------|
| Запуск задания | `POST /production/factory/start` | `{ "recipe_id": DAILY_CHEST_RECIPE_ID, "execution_count": 1 }` |
| Claim награды | `POST /production/factory/claim` | `{ "task_id": "<uuid>" }` |

Ответ `POST /production/factory/claim` соответствует схеме `ClaimResponse` Production Service. DGS преобразует поле `items_received` в массив `items` (см. пример ответа выше) через Inventory Service.

### Inventory Service
| Действие | HTTP | Тело |
|----------|------|------|
| Получить локализованные данные предметов | `POST /items/details?lang=<ru|en>` | `{ "items": [ { "item_id": "…" } ] }` |


## Безопасность
- Все публичные эндпоинты защищены JWT (RS256, валидация по JWKS от Auth Service).
- Внутренний API доступен только в сети docker-compose, mTLS по необходимости.
- Rate limiter: не более 20 POST `/deck/daily-chest/claim` в минуту на IP.


## Метрики (Prometheus)
| Метрика | Тип | Лейблы | Описание |
|---------|-----|--------|----------|
| `dgs_http_requests_total` | counter | `endpoint`,`method`,`status` | Счётчик запросов |
| `dgs_daily_craft_total`   | counter | `user_id` | Кол-во выданных сундуков |
| `dgs_invalid_combo_total` | counter | `user_id` | Ошибки `invalid_combo` |
| `dgs_cooldown_violation_total` | counter | `user_id` | Ранний запрос (cooldown) |
| `dgs_processing_latency_ms` | histogram | `endpoint` | Время обработки |

Grafana dashboard ID — TBD.


## Тестирование
1. **Unit-тесты**: бизнес-логика расчёта combo, cooldown, интеграция с Production stub (coverage ≥ 90%).
2. **Integration-тесты**: happy-path и edge-cases с test-containers PostgreSQL.
3. **E2E**: сценарии через public API + mock Production/Inventory.


## SLO
- **Availability** ≥ 99.5 % за 30 дней.
- **P95 latency** ≤ 120 мс для GET status, ≤ 250 мс для POST claim.


## Roadmap
- v1.0: реализовать описанные эндпоинты.
- v1.1: вынести Production call в async очередь для лучшей отказоустойчивости.
- v1.2: добавить геймификацию (leaderboard по combo).

---

# JWT Аутентификация

Все публичные эндпоинты DGS требуют валидного пользовательского JWT токена с RS256-подписью от Auth Service.

### Алгоритм проверки токена (аналогично Inventory Service)

1. **Извлечение токена**
   • Заголовок `Authorization: Bearer <token>` обязателен.  
   • Ошибка `missing_token` при отсутствии.  
   • Проверяется префикс `Bearer ` → ошибка `invalid_token_format` при несоответствии.

2. **Валидация RS256 подписи**
   • Публичный ключ загружается с Auth Service (см. `JWT_PUBLIC_KEYS_URL`) в формате PEM.  
   • Загрузка выполняется при старте сервиса и кэшируется в памяти; обновление каждые 12 часов (cron-job), либо при ошибке валидации выполняется «lazy refresh».  
   • Используется библиотека `github.com/golang-jwt/jwt/v5`, метод `jwt.Parse` с проверкой алгоритма `RS256`.

3. **Проверка времени жизни**  
   • Библиотека автоматически валидирует `exp`.  
   • При истечении срока ответа — `token_expired`.

4. **Проверка отзыва токена в Redis**
   • Из claim `jti` извлекается идентификатор токена.  
   • Redis-ключ: `revoked:{jti}`.  
   • Если ключ существует → `token_revoked` (401).  
   • В случае недоступности Redis запрос не прерывается (только warning-лог), идентичное поведение Inventory Service.

5. **Извлечение пользовательских данных**  
   • `sub` → `user_id` (UUID).  
   • `telegram_id` (float64 → int64).  
   • При отсутствии — ошибки `missing_user_id` или `missing_telegram_id`.

6. **Запись в контекст запроса**  
   В Gin-контекст помещаются `user` (структура `{user_id, telegram_id}`) и `jti`.

### Заголовок запроса
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Возможные ошибки аутентификации (401)
| Код | Сообщение |
|-----|-----------|
| `missing_token` | Отсутствует Authorization header |
| `invalid_token_format` | Неправильный формат Bearer токена |
| `invalid_token_signature` | Неверная RS256-подпись |
| `invalid_token` | Структурно некорректный токен |
| `token_expired` | Токен просрочен (`exp`) |
| `token_revoked` | Токен отозван в Redis |
| `missing_token_id` | Отсутствует JTI в claims |
| `missing_user_id` | Отсутствует user_id (`sub`) |
| `missing_telegram_id` | Отсутствует telegram_id |

### ENV-переменные для JWT
| Переменная | По умолчанию | Назначение |
|------------|--------------|-------------|
| `JWT_PUBLIC_KEYS_URL` | `http://auth-service:8090/keys` | JWKS endpoint (fallback / rotation) |
| `AUTH_PUBLIC_KEY_PEM_URL` | `http://auth-service:8080/public-key.pem` | PEM-ключ (используется по умолчанию, как в Inventory) |
| `REDIS_URL` | `redis://redis:6379/0` | Проверка отзыва токенов |

---
# Аудит Auth Service

*Дата:* 27.06.2025  
*Версия исходного кода:* ветка `develop`, commit текущей рабочей копии

## 1. Методология аудита

1. Изучены документы:
   - `docs/architecture/architecture.md` — системная архитектура
   - `docs/specs/auth-service.md` — функциональная и техническая спецификация Auth Service
2. Проанализирована структура проекта `services/auth-service` (слои `cmd`, `internal`, `pkg`).
3. Сопоставлены реализованные функции с требованиями спецификации и внутренними стандартами (Clean Architecture, PocketFlow, KISS).
4. Запущены unit-тесты (go test ./...) и проанализирован отчёт покрытия.
5. Проверены Dockerfile/Compose, конфигурационные параметры, дэшборды Grafana.
6. Выявлены отклонения, риски и потенциальные зоны технического долга.

## 2. Высокоуровневый обзор реализации

### 2.1 Структура каталогов

- `cmd` — точка входа (`main.go`), инициализация зависимостей, запуск Gin.
- `internal/handlers` — HTTP-эндпоинты (`/auth`, `/admin/*`, `/jwks`).
- `internal/middleware` — логирование, метрики, rate-limit.
- `internal/services` — бизнес-логика (TelegramValidator, JWTService).
- `internal/storage` — Postgres/Redis репозитории.
- `internal/metrics` — сбор Prometheus-метрик.
- `pkg/utils` — логгер, вспомогательные функции.

Структура, в целом, соответствует Clean Architecture, однако в ряде мест наблюдается смешение ответственности (см. P-7).

### 2.2 Сопоставление с целевой архитектурой и спецификацией

| Компонент | Требование спецификации | Реализация | Оценка |
|-----------|------------------------|------------|--------|
| Путь через API Gateway | `/api/auth` | В `main.go` роутер слушает `/auth`, префикс `/api` добавляется nginx — соответствует | ✅ |
| HMAC-валидация Telegram | key = *bot_token*, data = "WebAppData" | В `telegram_validator.go` перепутаны параметры (key=data, data=bot_token) | ❌ |
| JWT подпись | RS256, RSA-2048, kid в header | RSA-2048 ✅, kid отсутствует ❌ |
| JWKS формат | RFC 7517, поля `kty`, `alg`, `n`, `e`, `kid` | Выдаётся custom JSON c полем `pem`, нет `n`, `e` | ❌ |
| Rate-Limit | Значения из ENV | Хардкод 10 req/min | ⚠️ |
| Токены в Redis | TTL или ZSET-очистка | `SCAN active_token:*` каждые N ч | ⚠️ |
| Метрики | HTTP + бизнес, экспорт `/metrics` | HTTP-метрики ✅, бизнес-метрик мало ⚠️ |

## 3. Положительные аспекты реализации

1. **Чёткое разделение слоёв** — handlers / services / storage повышают читаемость.
2. **Богатый набор unit-тестов** (coverage ~82 %).
3. **Structured logging** через `slog` + единый формат.
4. **Prometheus-метрики**: latency, error_rate, активные токены.
5. **Поддержка нескольких bot-token** в конфиге.
6. **Health-checks** `/health` и readiness-probe.

## 4. Предварительный список проблем

| № | Категория | Краткое описание |
|---|-----------|------------------|
| P-1 | Security | Неверная HMAC-валидация Telegram initData |
| P-2 | Security | `/admin/*` и `/metrics` доступны без аутентификации |
| P-3 | Key Management | RSA-ключи генерируются в контейнере, нет ротации, отсутствует `kid` |
| P-4 | Compatibility | JWKS отдается в нестандарте, сервисы-клиенты не смогут импортировать ключ |
| P-5 | Performance | Очистка токенов O(N) через `SCAN` → нагрузка при 10^6 токенов |
| P-6 | Configuration | Rate-limit и другие параметры захардкожены, ENV не используется полностью |
| P-7 | Observability | Нет trace-id, бизнес-метрики ограничены |
| P-8 | Architecture | Нарушение PocketFlow: handlers вызывают Redis/DB/logic без prep-exec-post |

## 5. Детальный разбор ключевых проблем

### 5.1 P-1 — Неверная HMAC-валидация Telegram initData

**Факты**
1. Спецификация (§1.1) предписывает `secret_key = HMAC_SHA256(bot_token, "WebAppData")`.
2. В `internal/services/telegram_validator.go:generateSecretKeyForToken` используется `HMAC(key="WebAppData", data=botToken)`.
3. Следствием сервис **принимает недействительные данные** и отклоняет валидные.

**Риски**
- Пользователи не смогут войти (401).  
- Возможность подделать initData при знании алгоритма.

**Рекомендации**
1. Исправить порядок аргументов при генерации secret_key.  
2. Добавить unit-тест, покрывающий happy-/negative-path на публичных примерах Telegram.  
3. Выпустить hot-fix до выхода в прод.

### 5.2 P-2 — Не защищены эндпоинты `/admin/*` и `/metrics`

**Факты**
- В `internal/handlers/admin.go` комментарий *"should be protected"*, но middleware не подключён.
- `/metrics` доступен публично через Gateway.

**Риски**
- Утечка метрик и чувствительных данных.  
- Неавторизованное управление токенами.

**Рекомендации**
1. Ввести BasicAuth или JWT c ролью `admin`.  
2. Ограничить доступ по IP-листу (CIDR DevOps).  
3. Вынести `/metrics` на отдельный порт или behind-the-gate.

### 5.3 P-3 — Управление RSA-ключами

**Факты**
- Ключи генерируются при старте и сохраняются в `/etc/auth/*.pem` **внутри контейнера**.
- При пересборке образа ключ будет новым → инвалидация токенов.
- В JWT `kid` не проставляется.

**Риски**
- Массовые 401 при деплое.  
- Отсутствие graceful-rotation делает невозможным blue-green rollout.

**Рекомендации**
1. Хранить приватный ключ в external secret (Vault, Kubernetes Secret).  
2. Генерировать `kid = sha256(pubKey)` и добавлять в header JWT + JWKS.  
3. Реализовать двухфазную ротацию: publish JWKS with two keys → switch signing → retire old.

### 5.4 P-4 — JWKS формат не соответствует RFC 7517

**Факты**
- Endpoint `/jwks` возвращает:
  ```json
  {"pem":"-----BEGIN PUBLIC KEY-----..."}
  ```
- Отсутствуют обязательные поля `kty`, `alg`, `n`, `e`, `kid`.

**Риски**
- Клиентские библиотеки (auth0, go-jwx, jose) не смогут импортировать ключ.

**Рекомендации**
1. Использовать `encoding/base64url` для `n`, `e`.  
2. Структура:
   ```json
   {"keys":[{"kty":"RSA","alg":"RS256","use":"sig","kid":"...","n":"...","e":"AQAB"}]}
   ```
3. Добавить integration-test `jwkslint`.

### 5.5 P-5 — Очистка токенов через `SCAN`

**Факты**
- Функция `CleanupExpiredTokens` каждые 6 ч делает `SCAN active_token:*`.

**Риски**
- O(N) операция блокирует Redis CPU на больших объёмах.

**Рекомендации**
1. Хранить активные токены с `EXPIRE` (Redis TTL) и избавиться от ручной очистки.  
2. Либо перейти на `ZSET` (score = expires_at) + `ZRANGEBYSCORE`.

### 5.6 P-6 — Жёстко заданный rate-limit

**Факты**
- В `internal/middleware/rate_limit.go` используется `NewRateLimiter(10, time.Minute)`.
- ENV-переменные `RATE_LIMIT_REQUESTS`, `RATE_LIMIT_WINDOW` объявлены, но не читаются.

**Риски**
- Невозможно динамически настроить лимит под нагрузку.

**Рекомендации**
1. Загружать значения из конфиг-структуры (`internal/config`).  
2. Fail-fast при отсутствии обязательных значений.

### 5.7 P-7 — Низкая Observability и Traceability

**Факты**
- В логах отсутствует `trace_id` / `request_id`, middleware не добавляет `X-Request-ID`.
- Бизнес-метрики (кол-во регистраций, неуспешных авторизаций) частично реализованы, но не экспортируются.

**Риски**
- Сложно отследить цепочку запросов через Gateway → Auth → другие сервисы.

**Рекомендации**
1. Добавить `request_id` middleware (UUID v4), писать в контекст и логи.  
2. Экспортировать бизнес-метрики: `auth_success_total`, `auth_failed_total`, `new_users_total`.

### 5.8 P-8 — Нарушение PocketFlow и смешение слоёв

**Факты**
- Handler `Auth` выполняет валидацию, запись БД, вызовы Redis, генерацию JWT и метрики «в одной куче».

**Риски**
- Сложность тестирования, повторного использования, ретраев.

**Рекомендации**
1. Разбить процесс на:
   - `prep`: валидация Telegram.
   - `exec`: транзакция DB + запись Redis.
   - `post`: метрики, логи.
2. Вынести Redis/DB операции в сервис-слой, handler → orchestrator.

## 6. Итоговые рекомендации

1. **P-1, P-2, P-3, P-4** — блокеры, исправить до выхода в production.  
2. **P-5, P-6** — выполнить в ближайшем спринте; уменьшат эксплуатационные риски.  
3. **P-7, P-8** — включить в refactor roadmap, совместить с внедрением tracing stack (Jaeger).

---

*Конец отчёта.*
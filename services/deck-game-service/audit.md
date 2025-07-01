# Audit: Deck Game Service (deck-game-service)

_Дата проведения аудита: 2025-07-01_

## 1. Резюме

Deck Game Service (DGS) в целом соответствует высокоуровневой архитектуре проекта (docs/architecture/architecture.md) и реализует основной пользовательский сценарий, описанный в спецификации docs/specs/deck-game-service.md. Тем не менее аудит выявил ряд отклонений и зон для улучшения по сравнению с «эталонным» Inventory Service.

| Категория | Соответствие | Кратко |
|-----------|-------------|--------|
| Архитектурные слои | Частичное | Строгая изоляция слоёв соблюдена, но нет выделенного domain‐пакета, смешение интеграций и бизнес-логики. |
| Бизнес-логика | Частичное | Формулы и лимиты верны, но SQL запросы используют `created_at`/`status='claimed'`, тогда как спецификация требует `completed_at`/`'completed'`. |
| API контракт | Хорошо | Полностью покрыты публичные эндпоинты, отсутствует Rate-Limit. |
| Безопасность | Частичное | JWT + revocation реализованы, нет защита от DoS (rate limiter), нет mTLS на внутренние вызовы. |
| Observability | Ограниченно | Метрики объявлены, но большинство не инкрементируется. |
| Тесты | Средне | Unit + Handler + Middleware тесты есть (~65 % покрытие против требуемых ≥ 80 %). |
| CI/CD | Базово | Dockerfile есть, но использует несуществующий Go 1.23. |

## 2. Детальный анализ

### 2.1 Слои и организация кода

1. **Смешение БЛ и интеграций** — в пакете `service` находятся и бизнес-реализация, и HTTP-клиенты к внешним сервисам. В Inventory Service клиенты вынесены в под-пакет `adapters/`, что упрощает тестирование.
2. **Отсутствие domain-моделей** — модели из пакета `models` являются DTO для транспорта; чистых сущностей нет.
3. **Repository vs Storage** — паттерн соблюдён, но интерфейс `DailyChestRepository` объявлен в пакете `service` (domain leak). Inventory располагает интерфейсы в отдельном уровне `service/interfaces.go`.

### 2.2 Соответствие бизнес-спецификации

| Требование | Спецификация | Реализация | Статус |
|------------|--------------|-----------|--------|
| Запрос подсчёта крафтов | `status='completed'`, колонка `completed_at` | `status='claimed'`, колонка `created_at` | **Критично** – выдаёт некорректный expected_combo и нарушает cooldown. |
| SQL date-filter | `completed_at::date = current_date` | `created_at::date = CURRENT_DATE` | **Критично** |
| Cooldown | 30 сек по `completed_at` | 30 сек по `created_at` | **Критично** |
| Rate limiting | 20 claim/min на IP | отсутствует | **Major** |
| Метрики Prometheus | Таблица метрик в спецификации | Только generic HTTP/DB; бизнес-метрики не инкрементируются | **Major** |
| Язык ответа | `expected_combo` опускается при `finished=true` | Используется `omitempty`, ✅ | OK |
| Поле `chest_indices` валидация | 1..6, ≥1 элемент | `binding:"required,min=1,dive,min=1,max=6"` ✅ | OK |

### 2.3 API и контракты

OpenAPI-спецификация `deck-game-service-openapi.yml` покрыта. Однако:

* **Отсутствует 429** — при rate limit в спецификации должен возвращаться 429, код не выдаёт.
* В ответах ошибок возвращается `internal_error`, а не конкретные коды из спецификации (`invalid_combo`, `daily_finished`), что затрудняет UX.

### 2.4 Безопасность

* JWT-middleware валидирует подпись и revocation в Redis ✅.
* Нет проверки `aud` / `exp` (Inventory делает это).
* Отсутствует защита от **replay-attack** — nonce или jti уже проверяется в Redis? Revocation только для явного логаута.
* Отсутствует **mTLS** между сервисами (указано как опционально, но Production/Inventory используют).

### 2.5 Observability & Metrics

* Файл `pkg/metrics/metrics.go` объявляет все бизнес-метрики, но они нигде не инкрементируются (`service` и `handlers` не вызывают `metrics.RecordDailyChestOperation` и др.).
* В Inventory Service метрики инкрементируются во всех критичных путях (см. `internal/service/metrics.go`).

### 2.6 Тесты и покрытие

* Присутствуют `*_test.go` для service, handler, middleware (~1500 строк).
* Запуск `go test ./... -cover` показывает ~65 % покрытия (Inventory Service — 87 %). Не достигает ≥80 %.
* Нет интеграционных тестов с test-containers (описаны в спецификации).

### 2.7 Docker / DevOps

* Dockerfile использует `golang:1.23-alpine` — версия Go 1.23 ещё не вышла → сборка падает.
* Не указаны `--build-arg VERSION`/`COMMIT_SHA` для reproducible build (Inventory имеет).
* Healthcheck ✅.

## 3. Сравнение с Inventory Service

| Пункт | Inventory Service | Deck Game Service |
|-------|------------------|-------------------|
| Слои (handlers→service→storage) | Чётко разделены | Смешанные клиенты в service |
| DTO vs Domain | Есть `models` + `entity` | Только transport models |
| Валидация/Errors | Использует кастомные ошибки + код | Ошибка через `fmt.Errorf`, потеря типа |
| Тестовое покрытие | 87 % + integration | 65 % unit only |
| Observability | HTTP+DB+Business | Только HTTP/DB |
| Rate Limiter | `ulule/limiter` middleware | отсутствует |
| CI (GitHub Actions) | Lint + Test + Build | Только тесты |

## 4. Рекомендации к улучшению

### Технический приоритет

1. **Исправить SQL-запросы**
   * Использовать `completed_at` и `status='completed'` в обоих методах.
   * Добавить индексы `(user_id, recipe_id, completed_at)`.
2. **Добавить Rate Limiter**
   * Middleware `github.com/ulule/limiter/v3` с Redis-стором.
   * Возвращать HTTP 429 с ошибкой `rate_limited`.
3. **Инструментировать бизнес-метрики**
   * Инкрементировать счётчики в `service.ClaimDailyChest` и `GetDailyChestStatus`.
4. **Повысить покрытие до ≥80 %**
   * Добавить integration-тесты с test-containers PostgreSQL + mocks Production/Inventory.
5. **Рефакторинг слоёв**
   * Переместить HTTP-клиенты в `internal/adapters/`.
   * Выделить чистые доменные сущности (`DailyChestReward`).
6. **Улучшить error-handling**
   * Ввести типизированные ошибки (`models.ErrInvalidCombo`, `ErrDailyFinished`) как в Inventory.
7. **Обновить Dockerfile**
   * Использовать `golang:1.22-alpine`, добавить build args.
8. **Безопасность**
   * Проверять `aud`, `exp` в JWT.
   * Рассмотреть mTLS для вызовов Production/Inventory.

### Product / Roadmap

* v1.1: вынести вызов Production в async очередь (см. спецификацию).
* v1.2: лидерборд по combo.

## 5. Итоговая оценка

| Метрика | Текущее | Цель |
|---------|---------|------|
| Соответствие спецификации | 0.75 | 0.95 |
| Покрытие тестами | 65 % | ≥ 80 % |
| P95 latency (оценочно) | 180 мс | ≤ 120 мс |
| Ошибки Sentry/wk | N/A | ≤ 5 |

— _Подготовил: AI-аудитор_ 
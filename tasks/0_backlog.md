# Backlog — задачи к рассмотрению

> Здесь собираются идеи, фичи и задачи, которые еще не проработаны или не приоритизированы.


---
**Формат добавления задач:**
```
## Название задачи
**Описание:** краткое описание проблемы или функции
**Приоритет:** [Высокий/Средний/Низкий]
**Оценка:** [XS/S/M/L/XL]
**Зависимости:** список зависимых задач
**Критерии готовности:** что должно быть выполнено
```



---
## D-Deck-002: Конфигурация и переменные окружения
**Описание:** Добавить пакет `internal/config` с загрузкой ENV-переменных, описанных в спецификации Deck Game Service (DATABASE_URL, PRODUCTION_INTERNAL_URL, AUTH_PUBLIC_KEY_URL и др.).
**Приоритет:** Высокий
**Оценка:** S
**Зависимости:** D-Deck-001
**Критерии готовности:**
- Значения читаются через viper/env
- `GET /config` внутренний отключен; тест unit проверяет defaults
**Ресурсы:**
- docs/specs/deck-game-service.md (ENV таблица)
- docs/architecture/architecture.md
- services/inventory-service/internal/config/*.go (пример)

---
## D-Deck-003: JWT Middleware
**Описание:** Реализовать middleware проверки JWT, копируя логику `inventory-service/internal/middleware/auth.go` (проверка RS256 подписи, Redis-revocation, запись user в контекст).
**Приоритет:** Высокий
**Оценка:** S
**Зависимости:** D-Deck-002
**Критерии готовности:**
- Middleware выдаёт 401 при невалидном токене
- Тесты используют сгенерированный RSA-ключ и мок Redis
**Ресурсы:**
- docs/specs/deck-game-service.md (раздел JWT Аутентификация)
- docs/specs/auth-service.md
- services/inventory-service/internal/middleware/auth.go

---
## D-Deck-004: Endpoint GET /deck/daily-chest/status
**Описание:** Реализовать handler, вычисляющий `expected_combo`, `finished`, `crafts_done`, `last_reward_at` по запросу к БД production (schema `production.production_tasks`).
**Приоритет:** Высокий
**Оценка:** M
**Зависимости:** D-Deck-003
**Критерии готовности:**
- Handler возвращает данные, соответствующие `StatusResponse`
- Интеграционный тест с test-containers Postgres (dev compose) возвращает ожидаемый ответ
**Ресурсы:**
- docs/specs/deck-game-service.md
- docs/specs/deck-game-service-openapi.yml
- docs/architecture/database.dbml (schema production.production_tasks)
- docs/specs/production-service-openapi.yml (для Contract)

---
## D-Deck-005: Endpoint POST /deck/daily-chest/claim
**Описание:** Реализовать бизнес-логику выдачи сундука: валидация `combo`, `chest_indices`, запуск `POST /production/factory/start`, claim результата и обогащение данных через Inventory Service `/items/details`.
**Приоритет:** Высокий
**Оценка:** L
**Зависимости:** D-Deck-004
**Критерии готовности:**
- Успешный путь и все ошибки 400/404 покрыты unit-тестами
- Статусы ошибок и поля ответа соответствуют OpenAPI
**Ресурсы:**
- docs/specs/deck-game-service.md
- docs/specs/deck-game-service-openapi.yml
- docs/specs/production-service-openapi.yml
- docs/specs/inventory-service-openapi.yml

---
## D-Deck-006: Метрики и логирование
**Описание:** Добавить Prometheus-метрики (`dgs_http_requests_total`, `dgs_daily_craft_total`, ... ) и structured-logging (slog) аналогично inventory-service.
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** D-Deck-005
**Критерии готовности:**
- `/metrics` содержит новые счётчики
- Примерное покрытие логами INFO/WARN/ERROR
**Ресурсы:**
- docs/specs/deck-game-service.md (метрики)
- services/inventory-service/internal/metrics/* (пример)

---
## D-Deck-007: Dockerfile & Dev Compose
**Описание:** Создать Dockerfile и обновить `deploy/dev/docker-compose.yml`, добавив контейнер `deck-game-service` (порт 8080 и 8090). Сеть и переменные окружения настроены как у inventory.
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** D-Deck-006
**Критерии готовности:**
- `docker compose up` поднимает сервис без ошибок
- Health-check проходит в 5 сек.
**Ресурсы:**
- deploy/dev/docker-compose.yml (пример inventory)
- services/inventory-service/Dockerfile

---
## D-Deck-008: OpenAPI генерация и CI чек
**Описание:** Автоматически публиковать `deck-game-service-openapi.yml` в aggregated `docs/architecture/openapi.yml`, валидировать в CI (spectral lint).
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** D-Deck-007
**Критерии готовности:**
- GitHub Actions (или Make target) валидирует спецификацию
- PR не проходит без валидного OpenAPI
**Ресурсы:**
- docs/specs/deck-game-service-openapi.yml
- docs/architecture/openapi.yml
- .github/workflows/* (lint examples)

---
## D-Deck-009: Unit-тесты бизнес-логики
**Описание:** Добавить unit-тесты для расчёта combo, валидации `chest_indices`, работы cooldown и редиса, без e2e/интеграции.
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** D-Deck-005
**Критерии готовности:**
- Покрытие >80 % для пакетов `service/` и `middleware/`
- `go test ./...` проходит локально и в CI
**Ресурсы:**
- docs/specs/deck-game-service.md
- services/inventory-service/internal/service/* (тесты примеры)
- docs/architecture/architecture.md
# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

*Нет задач высокого приоритета*

---

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->




## M-3: HTTP API эндпоинты для Inventory Service
**Роль:** Backend Developer
**Приоритет:** Средний  
**Статус:** [ ] Готов к выполнению (все зависимости выполнены)

**Описание:**
`docs/tasks/inventory-service-task-5-http-endpoints.md`
Реализация всех HTTP эндпоинтов согласно OpenAPI спецификации. Включает публичные, внутренние и административные эндпоинты с полной валидацией, аутентификацией и обработкой ошибок.

**Файлы для изучения:**
- `docs/specs/inventory-service-openapi.yml` - полная спецификация OpenAPI
- `docs/specs/inventory-service.md` - описание всех эндпоинтов (строки 17-196)
- `services/auth-service/internal/handlers/` - паттерны handlers  
- `services/auth-service/internal/middleware/` - JWT authentication
- `docs/tasks/inventory-service-task-5-http-endpoints.md` - детальные требования

**Критерии готовности:**

**HTTP handlers:**
- [ ] `internal/handlers/inventory.go` - структура InventoryHandler
- [ ] `internal/handlers/public.go` - GET /inventory, GET /inventory/items/{item_id}
- [ ] `internal/handlers/internal.go` - /reserve, /return-reserve, /consume-reserve, /add-items
- [ ] `internal/handlers/admin.go` - POST /admin/inventory/adjust

**Middleware:**
- [ ] `internal/middleware/auth.go` - JWT authentication
- [ ] `internal/middleware/admin.go` - admin authorization
- [ ] `internal/middleware/logging.go` - request logging
- [ ] `internal/middleware/metrics.go` - metrics collection

**Error handling:**
- [ ] `internal/handlers/errors.go` - структурированные ошибки
- [ ] ErrorResponse, InsufficientItemsError согласно OpenAPI
- [ ] Правильные HTTP статус коды
- [ ] Валидация входных данных

**Routing:**
- [ ] `internal/handlers/router.go` - настройка всех маршрутов
- [ ] Группировка по типам (public, internal, admin)
- [ ] Применение соответствующих middleware

**Проверка:**
```bash
curl -H "Authorization: Bearer <jwt>" http://localhost:8080/inventory # работает
go test ./internal/handlers/... # coverage >85%
```

**Зависимости:** ✅ M-2 (бизнес-логика), ✅ Auth Service (JWT токены)
**Оценка:** 2-3 дня

---


## Низкий приоритет
<!-- Задачи для будущих итераций -->

## L-1: Мониторинг и метрики для Inventory Service
**Роль:** DevOps/Backend Developer
**Приоритет:** Низкий
**Статус:** [ ] Частично готов (базовые метрики есть, нужен дашборд и алерты)

**Описание:**
`docs/tasks/inventory-service-task-6-monitoring-metrics.md`
Реализация комплексной системы мониторинга для inventory-service на основе Prometheus метрик и Grafana дашбордов, аналогично auth-service.

**Файлы для изучения:**
- `services/auth-service/pkg/metrics/` - структура метрик как референс
- `docs/specs/inventory-service.md` - типы операций (строки 231-265)  
- `docs/tasks/inventory-service-task-6-monitoring-metrics.md` - детальные требования
- `monitoring/grafana/dashboards/auth-service.json` - дашборд как шаблон

**Критерии готовности:**

**Бизнес метрики:**
- [ ] `inventory_operations_total` - счетчик операций по типам
- [ ] `inventory_balance_calculation_duration_seconds` - время расчета остатков
- [ ] `inventory_cache_hit_ratio` - эффективность кеширования
- [ ] `inventory_insufficient_balance_errors_total` - ошибки недостаточного баланса

**Grafana дашборд:**
- [ ] `monitoring/grafana/dashboards/inventory-service.json` - дашборд с 5 группами панелей
- [ ] Overview, HTTP Metrics, Business Metrics, Technical Metrics, Alerts Status

**Prometheus алерты:**
- [ ] `monitoring/prometheus/rules/inventory-service.yml` - 6 алертов
- [ ] High error rate, high latency, DB issues, cache problems, service down

**Зависимости:** ⏳ M-3 (HTTP API), ✅ M-4 (auth-service метрики как референс)

**Прогресс реализации:**
- ✅ Базовые метрики: `pkg/metrics/metrics.go` содержит HTTP, DB, Redis и бизнес-метрики
- ❌ Grafana дашборд: отсутствует `monitoring/grafana/dashboards/inventory-service.json`
- ❌ Prometheus алерты: отсутствует `monitoring/prometheus/rules/inventory-service.yml`
**Оценка:** 2 дня

---

## L-2: Интеграционное тестирование и развертывание для Inventory Service
**Роль:** QA/DevOps Engineer
**Приоритет:** Низкий
**Статус:** [ ] Готов к выполнению

**Описание:**
Создание комплексной системы тестирования inventory-service, включая интеграционные тесты с другими сервисами, E2E тесты, нагрузочное тестирование и настройку CI/CD pipeline.

**Файлы для изучения:**
- `docs/specs/inventory-service.md` - полные сценарии работы
- `docs/concept/game-mechanics-chests-keys-deck-minigame.md` - игровые сценарии
- `services/auth-service/` - API для JWT токенов
- `docs/tasks/inventory-service-task-7-integration-testing.md` - детальные требования

**Критерии готовности:**

**Интеграционные тесты:**
- [ ] `tests/integration/auth/` - тесты JWT аутентификации с auth-service
- [ ] `tests/e2e/` - полные пользовательские сценарии
- [ ] Тесты concurrent операций и race conditions

**Нагрузочные тесты:**
- [ ] `tests/load/k6_loadtest.js` - k6 тесты производительности
- [ ] `tests/load/artillery_loadtest.yml` - Artillery тесты
- [ ] SLA: 95% запросов <1s, error rate <10%

**CI/CD Pipeline:**
- [ ] `.github/workflows/inventory-service.yml` - автоматические тесты и деплой
- [ ] Docker образы в Container Registry
- [ ] Автоматический деплой в staging/production

**Development environment:**
- [ ] `docker-compose.dev.yml` - полный стек для разработки
- [ ] Включает inventory-service, auth-service, PostgreSQL, Redis, Prometheus, Grafana

**Проверка:**
```bash
docker-compose -f docker-compose.dev.yml up -d # полный стек
go test -tags=e2e ./tests/e2e/... # E2E тесты проходят
k6 run tests/load/loadtest.js # нагрузочные тесты
```

**Зависимости:** ⏳ L-1 (мониторинг), ✅ M-1 (модели), ✅ M-2 (бизнес-логика), ⏳ M-3 (HTTP API)

**Прогресс реализации:**
- ❌ Интеграционные тесты: отсутствует директория `tests/`
- ❌ E2E тесты: не реализованы
- ❌ CI/CD Pipeline: отсутствует `.github/workflows/inventory-service.yml`
- ❌ Development environment: нет интеграции в `deploy/dev`
**Оценка:** 2-3 дня

---
**Критерии перехода в TODO:**
- [ ] Задача четко сформулирована
- [ ] Определены критерии готовности
- [ ] Проведена оценка трудозатрат
- [ ] Нет блокирующих зависимостей
- [ ] Назначен ответственный (опционально) 
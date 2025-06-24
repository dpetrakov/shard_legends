# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

*Нет задач высокого приоритета*

---

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->







## Низкий приоритет
<!-- Задачи для будущих итераций -->

## L-1: Мониторинг и метрики для Inventory Service
**Роль:** DevOps/Backend Developer
**Приоритет:** Низкий
**Статус:** [ ] Частично готов (базовые метрики есть, нужны дополнительные метрики, дашборд и алерты)

**Описание:**
Реализация комплексной системы мониторинга для inventory-service на основе Prometheus метрик и Grafana дашбордов, аналогично auth-service. Фокус на производительности балансовых расчетов, эффективности кеширования и мониторинге бизнес-операций.

**Файлы для изучения:**
- `deploy/monitoring/grafana/dashboards/auth-service-metrics.json` - референс структуры дашборда
- `services/inventory-service/pkg/metrics/metrics.go` - текущие базовые метрики
- `docs/specs/inventory-service.md` - бизнес-операции и алгоритмы
- `docs/tasks/inventory-service-task-6-monitoring-metrics.md` - детальные требования

**Критерии готовности:**

**Дополнительные бизнес-метрики (дополнить существующие):**
- [ ] `inventory_balance_calculation_duration_seconds` - время расчета остатков с label cache_hit
- [ ] `inventory_daily_balance_created_total` - ленивое создание дневных остатков
- [ ] `inventory_cache_hit_ratio` - эффективность кеширования по типам (balances, classifiers)
- [ ] `inventory_classifier_conversions_total` - преобразования код↔UUID
- [ ] `inventory_insufficient_balance_errors_total` - ошибки недостаточного баланса с деталями
- [ ] `inventory_transaction_rollbacks_total` - откаты транзакций с причинами
- [ ] `inventory_service_up` - статус сервиса (аналогично auth-service)
- [ ] `inventory_service_start_time_seconds` - время запуска для uptime

**Grafana дашборд (аналогично auth-service):**
- [ ] `deploy/monitoring/grafana/dashboards/inventory-service-metrics.json` - основной дашборд
- [ ] 7 групп панелей: Service Overview, HTTP Metrics, Inventory Business Metrics, Cache Performance, Database Metrics, Dependencies Health, Admin Operations
- [ ] Панели: Service Status, Uptime, Memory Usage, Goroutines, HTTP rates/latency, Balance calculation performance, Cache hit ratios, DB/Redis operations

**Prometheus алерты:**
- [ ] `deploy/monitoring/prometheus/rules/inventory-service.yml` - правила алертов
- [ ] InventoryServiceDown - сервис недоступен >1 мин
- [ ] InventoryHighErrorRate - error rate >10% за 5 мин  
- [ ] InventoryHighLatency - p95 latency >2s за 5 мин
- [ ] InventoryDatabaseIssues - DB queries >5s или errors >5%
- [ ] InventoryCacheProblems - cache hit ratio <70% за 10 мин
- [ ] InventoryBalanceCalculationSlow - balance calculation >1s за 5 мин

**Интеграция в middleware:**
- [ ] Обновить `internal/middleware/metrics.go` для новых метрик
- [ ] Добавить метрики в ключевые бизнес-алгоритмы (balance_calculator, cache_manager, etc.)

**Зависимости:** ✅ M-3 (HTTP API), ✅ M-4 (auth-service как референс), ✅ M-2 (бизнес-логика)

**Прогресс реализации:**
- ✅ Базовые метрики: HTTP, DB, Redis, начальные inventory operations
- ❌ Специфичные inventory метрики: balance calculation, cache hit ratio, daily balance creation
- ❌ Grafana дашборд: нет `inventory-service-metrics.json`
- ❌ Prometheus алерты: нет rules для inventory-service
- ❌ Интеграция: метрики не подключены к бизнес-алгоритмам
**Оценка:** 2-3 дня

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
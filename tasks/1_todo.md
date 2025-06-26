# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

**D-3: Поддержка очереди производственных заданий**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
Реализовать систему очередей для производственных заданий, включая создание и мониторинг заданий.

**Критерии выполнения:**
- [ ] **TaskRepository**: Доработать методы CreateTask, GetUserTasks, UpdateTaskStatus в `internal/storage/task_repository.go`
- [ ] **Эндпоинт GET /factory/queue**: HTTP handler для получения очереди заданий пользователя
- [ ] **Эндпоинт POST /factory/start**: HTTP handler для создания нового производственного задания
- [ ] **Эндпоинт GET /factory/completed**: HTTP handler для получения завершенных заданий
- [ ] **Предрасчет результатов**: Алгоритм определения выходных предметов при создании задания
- [ ] **Интеграция с Inventory Service**: Резервирование предметов через API inventory-service
- [ ] **Управление производственными слотами**: Проверка доступности слотов и распределение заданий
- [ ] **Система модификаторов**: Применение ускорителей и бонусов при создании задания
- [ ] **Интеграция с User Service**: 
  - Вызов `GET /internal/users/{user_id}/production-slots` для получения доступных слотов
  - Вызов `GET /internal/users/{user_id}/production-modifiers` для получения пользовательских модификаторов
- [ ] **Обработка ошибок**: Корректная обработка недостаточности предметов и превышения лимитов
- [ ] **Юнит-тесты**: Покрытие основной логики создания и управления заданиями

**Зависимости:**
- ✅ D-2 (модели данных и repository для рецептов)
- Интеграция с inventory-service для резервирования предметов
- ✅ D-8 (User Service для получения слотов и модификаторов пользователя)

**Файлы для изучения:**
- [docs/specs/production-service.md](docs/specs/production-service.md)
- [docs/specs/production-service-openapi.yml](docs/specs/production-service-openapi.yml)


**D-4: Реализация системы модификаторов и предрасчета результатов**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
Внедрить систему модификаторов, поддержку автоматических, событийных и пользовательских модификаторов.

**Критерии выполнения:**
- [ ] **ModifierService**: Сервис для применения различных типов модификаторов
- [ ] **Ускорители-предметы**: Обработка модификаторов из предметов инвентаря пользователя
- [ ] **Пользовательские модификаторы**: Интеграция с user-service для получения VIP, уровня, достижений
- [ ] **Событийные модификаторы**: Интеграция с event-service для сезонных и временных бонусов
- [ ] **Автоматические модификаторы**: Клановые бонусы и серверные бафы
- [ ] **Алгоритм предрасчета**: Определение результата производства в момент создания задания
- [ ] **Случайная генерация**: Выбор предметов по вероятностям с учетом групп
- [ ] **Наследование коллекций**: Передача коллекций и качества от входных к выходным предметам
- [ ] **Фиксация результата**: Сохранение предрасчитанного результата в task_output_items
- [ ] **Логирование модификаторов**: Сохранение примененных модификаторов для аудита

**Зависимости:**
- ✅ D-2 (модели данных ProductionTask, TaskOutputItem)
- ✅ D-3 (создание заданий)
- Интеграция с user-service для модификаторов пользователя
- Интеграция с event-service для событийных модификаторов

**Файлы для изучения:**
- [docs/specs/production-service.md](docs/specs/production-service.md)
- [docs/specs/production-service-openapi.yml](docs/specs/production-service-openapi.yml)


**D-5: Исполнение Claim и завершение процессов**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
Реализовать эндпоинты для Claim завершенных задач, с сохранением результатов в инвентаре.

**Критерии выполнения:**
- [ ] **Эндпоинт POST /factory/claim**: HTTP handler для получения результатов завершенных заданий
- [ ] **Эндпоинт POST /factory/cancel**: HTTP handler для отмены заданий в статусе pending
- [ ] **ClaimService**: Бизнес-логика для claim операций с валидацией статусов
- [ ] **Интеграция с Inventory Service**: 
  - Вызов `POST /inventory/consume-reserve` для уничтожения зарезервированных материалов
  - Вызов `POST /inventory/add-items` для добавления результатов в инвентарь
- [ ] **Обновление статусов**: Изменение статуса задания с completed на claimed
- [ ] **Освобождение слотов**: Логика освобождения производственных слотов после claim
- [ ] **Атомарность операций**: Обеспечение консистентности между production-service и inventory-service
- [ ] **Обработка ошибок**: Откат операций при сбоях в inventory-service
- [ ] **Пакетный claim**: Поддержка claim всех готовых заданий одним запросом
- [ ] **Аудит операций**: Логирование всех claim и cancel операций

**Зависимости:**
- ✅ D-2 (модели данных и базовые сервисы)
- ✅ D-3 (создание и управление заданиями)
- ✅ D-4 (предрасчитанные результаты)
- Интеграция с inventory-service для операций с предметами

**Файлы для изучения:**
- [docs/specs/production-service.md](docs/specs/production-service.md)
- [docs/specs/production-service-openapi.yml](docs/specs/production-service-openapi.yml)

---

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->

### D-6: Обзор и тестирование архитектуры и бизнес-логики
**Роль:** QA Engineer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Написать юнит-тесты на все ключевые компоненты системы после реализации бизнес-логики.

**Критерии выполнения:**
- Юнит-тесты покрывают не менее 80% бизнеса-логики.

**Файлы для изучения:**
- [docs/specs/production-service.md](docs/specs/production-service.md)
- [docs/specs/production-service-openapi.yml](docs/specs/production-service-openapi.yml)


### D-7: Внедрение метрик и мониторинга
**Роль:** DevOps/Backend Developer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Интеграция мониторинга и метрик для наблюдения за работой сервиса.

**Критерии выполнения:**
- Поддержка метрик в ключевых эндпоинтах и процессы, дашборды в Prometheus и Grafana.

**Файлы для изучения:**
- [docs/specs/production-service.md](docs/specs/production-service.md)
- [docs/specs/production-service-openapi.yml](docs/specs/production-service-openapi.yml)


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
- [x] `deploy/monitoring/grafana/dashboards/inventory-service-metrics.json` - основной дашборд
- [x] 7 групп панелей: Service Overview, HTTP Metrics, Inventory Business Metrics, Cache Performance, Database Metrics, Dependencies Health, Admin Operations
- [x] Панели: Service Status, Uptime, Memory Usage, Goroutines, HTTP rates/latency, Balance calculation performance, Cache hit ratios, DB/Redis operations

**Prometheus алерты:**
- [x] `deploy/monitoring/alerts.yml` - правила алертов для inventory-service реализованы
- [x] InventoryServiceDown - сервис недоступен >2 мин
- [x] InventoryHighErrorRate - error rate >5% за 5 мин  
- [x] InventoryHighResponseTime - p95 latency >0.5s за 5 мин
- [x] InventoryDatabaseDown - PostgreSQL health check failing
- [x] InventoryRedisDown - Redis health check failing
- [x] InventoryLowCacheHitRate - cache hit ratio <50% за 10 мин
- [x] InventorySlowBalanceCalculation - balance calculation >2s за 5 мин
- [x] InventoryHighTransactionFailureRate - transaction failures >5%

**Интеграция в middleware:**
- [ ] Обновить `internal/middleware/metrics.go` для новых метрик
- [ ] Добавить метрики в ключевые бизнес-алгоритмы (balance_calculator, cache_manager, etc.)

**Зависимости:** ✅ M-3 (HTTP API), ✅ M-4 (auth-service как референс), ✅ M-2 (бизнес-логика)

**Прогресс реализации (~70% выполнено):**
- ✅ Базовые метрики: HTTP, DB, Redis, inventory operations полностью реализованы
- ✅ Grafana дашборд: `inventory-service-metrics.json` создан с 7 группами панелей
- ✅ Prometheus алерты: 8 алертов реализованы в `alerts.yml` (Service Down, Errors, Latency, DB/Redis health)
- ✅ HTTP middleware интеграция: `internal/middleware/metrics.go` работает
- ❌ Специфичные inventory метрики: balance calculation, cache hit ratio, daily balance creation (7 метрик)
- ❌ Интеграция в бизнес-логику: метрики не подключены к repositories/services
**Оценка:** 1-1.5 дня (доработка)

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
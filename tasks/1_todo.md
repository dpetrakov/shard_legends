# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

## H-1: Создание скриптов миграции базы данных для Inventory Service
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
`docs/tasks/inventory-service-task-1-database-migrations.md`
Создание SQL-скриптов миграции для инициализации схемы базы данных inventory-service. Включает создание таблиц, индексов, constraints и первичную заливку справочных данных с фиксированными UUID.

**Файлы для изучения:**
- `docs/architecture/database.dbml` - полная схема БД с индексами и constraints (строки 54-261)
- `docs/specs/inventory-service.md` - описание таблиц и модели данных (строки 198-586)
- `docs/concept/game-mechanics-chests-keys-deck-minigame.md` - игровые механики для валидации данных

**Критерии готовности:**

**Структура схемы:**
- [ ] `migrations/001_create_inventory_schema.sql` - создание всех таблиц схемы inventory
- [ ] `migrations/001_create_inventory_schema_rollback.sql` - скрипт отката структуры
- [ ] Все constraints из database.dbml реализованы
- [ ] Все индексы для производительности созданы
- [ ] Foreign key constraints настроены
- [ ] UUID default значения установлены

**Дистрибутивные данные:**
- [ ] `migrations/002_inventory_classifiers_data.sql` - все 16 классификаторов с фиксированными UUID
- [ ] `migrations/002_inventory_classifiers_data_rollback.sql` - скрипт очистки классификаторов
- [ ] ~90 элементов классификаторов с описаниями
- [ ] Возможность многократных откатов/накатов без изменения UUID

**Тестовые данные:**
- [ ] `migrations/003_inventory_test_data.sql` - тестовые предметы и операции
- [ ] Создание тестовых изображений (item_images)
- [ ] Создание тестовых операций для демонстрации
- [ ] Создание тестовых дневных остатков

**Проверка:**
```bash
psql -d shard_legends -c "SELECT count(*) FROM inventory.classifiers;" # = 16
psql -d shard_legends -c "SELECT count(*) FROM inventory.classifier_items;" # = ~90
```

**Зависимости:** PostgreSQL 17
**Оценка:** 1-2 дня

---

## H-2: Базовое Go приложение для Inventory Service
**Роль:** Backend Developer  
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
Создание основной структуры Go приложения для inventory-service с базовой конфигурацией, подключением к БД, логированием и health check эндпоинтами.

**Файлы для изучения:**
- `services/auth-service/` - структура проекта как референс
- `docs/specs/inventory-service.md` - технические характеристики (строки 7-16)
- `docs/tasks/inventory-service-task-2-core-application.md` - детальные требования

**Критерии готовности:**

**Структура проекта:**
- [ ] Go модуль инициализирован в `services/inventory-service/`
- [ ] Структура каталогов: cmd/, internal/, pkg/, migrations/, docker/
- [ ] Dockerfile с multi-stage build
- [ ] docker-compose.yml с PostgreSQL, Redis, Prometheus

**Конфигурация:**
- [ ] `internal/config/config.go` - чтение env переменных
- [ ] Поддержка DB_HOST, DB_PORT, REDIS_HOST, SERVER_PORT
- [ ] Конфигурация логирования и метрик

**Подключения:**
- [ ] `internal/database/postgres.go` - connection pool для PostgreSQL 17
- [ ] `internal/database/redis.go` - подключение к Redis 8.0.2
- [ ] Ping проверки соединений
- [ ] Graceful shutdown

**Health Checks:**
- [ ] `internal/handlers/health.go` - /health, /health/ready, /health/live
- [ ] Проверка доступности БД и Redis
- [ ] JSON ответы с timestamp и статусом сервисов

**Логирование:**
- [ ] `pkg/logger/logger.go` - structured JSON logging
- [ ] Различные уровни логирования
- [ ] Request ID для трассировки

**Проверка:**
```bash
curl http://localhost:8080/health # возвращает 200
docker-compose up -d # стек поднимается без ошибок
```

**Зависимости:** H-1 (миграции БД)
**Оценка:** 1 день

---

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->

## M-1: Модели данных и репозитории для Inventory Service
**Роль:** Backend Developer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Создание Go структур для работы с данными, реализация repository паттерна для доступа к БД и базовых операций с классификаторами и предметами.

**Файлы для изучения:**
- `docs/architecture/database.dbml` - все таблицы схемы inventory (строки 54-261)  
- `docs/specs/inventory-service.md` - описание моделей данных (строки 198-255)
- `docs/specs/inventory-service-openapi.yml` - request/response схемы
- `services/auth-service/internal/repository/` - паттерны repository
- `docs/tasks/inventory-service-task-3-data-models.md` - детальные требования

**Критерии готовности:**

**Go структуры:**
- [ ] `internal/models/classifier.go` - Classifier, ClassifierItem
- [ ] `internal/models/item.go` - Item, ItemImage  
- [ ] `internal/models/inventory.go` - DailyBalance, Operation
- [ ] `internal/models/dto.go` - API структуры с валидацией
- [ ] Правильные db и json теги для всех полей

**Repository интерфейсы:**
- [ ] `internal/repository/interfaces.go` - ClassifierRepository, ItemRepository, InventoryRepository
- [ ] Методы для CRUD операций
- [ ] Методы для кеширования и преобразования код↔UUID

**Реализация репозиториев:**
- [ ] `internal/repository/classifier_repo.go` - работа с классификаторами + Redis кеш
- [ ] `internal/repository/item_repo.go` - операции с предметами и изображениями
- [ ] `internal/repository/inventory_repo.go` - дневные остатки и операции
- [ ] Batch операции для производительности
- [ ] Транзакционные методы

**Валидация:**
- [ ] `internal/models/validation.go` - валидаторы для всех структур
- [ ] Custom validation rules
- [ ] Error handling для валидации

**Проверка:**
```bash
go test ./internal/repository/... # все тесты проходят
go test ./internal/models/... # валидация работает
```

**Зависимости:** H-2 (базовое приложение)
**Оценка:** 2-3 дня

---

## M-2: Бизнес-логика и общие алгоритмы для Inventory Service
**Роль:** Backend Developer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Реализация основных бизнес-алгоритмов inventory-service согласно спецификации. Включает расчет остатков, создание дневных балансов, преобразование кодов и валидацию операций.

**Файлы для изучения:**
- `docs/specs/inventory-service.md` - общие алгоритмы (строки 306-384)
- `docs/specs/inventory-service.md` - формула расчета остатков (строки 273-280)
- `docs/specs/inventory-service.md` - ленивое создание остатков (строки 281-287)
- `docs/tasks/inventory-service-task-4-business-logic.md` - детальные требования

**Критерии готовности:**

**Service интерфейсы:**
- [ ] `internal/service/interfaces.go` - InventoryService, ClassifierService
- [ ] Интерфейсы для всех 6 ключевых алгоритмов

**Ключевые алгоритмы:**
- [ ] `internal/service/balance_calculator.go` - CalculateCurrentBalance с кешированием
- [ ] `internal/service/daily_balance_creator.go` - CreateDailyBalance (ленивое создание)
- [ ] `internal/service/code_converter.go` - ConvertClassifierCodes (двунаправленное)
- [ ] `internal/service/balance_checker.go` - CheckSufficientBalance
- [ ] `internal/service/operation_creator.go` - CreateOperationsInTransaction
- [ ] `internal/service/cache_manager.go` - InvalidateUserCache

**Главный сервис:**
- [ ] `internal/service/inventory_service.go` - координация всех алгоритмов
- [ ] Транзакционная безопасность
- [ ] Error handling и логирование

**Кеширование:**
- [ ] Redis кеширование с TTL 1 час для остатков
- [ ] Redis кеширование с TTL 24 часа для классификаторов
- [ ] Namespace ключей: `inventory:{user_id}:...`
- [ ] Graceful degradation при недоступности Redis

**Проверка:**
```bash
go test ./internal/service/... # coverage >90%
go test -race ./internal/service/... # проверка race conditions
```

**Зависимости:** M-1 (модели и репозитории)
**Оценка:** 3-4 дня

---

## M-3: HTTP API эндпоинты для Inventory Service
**Роль:** Backend Developer
**Приоритет:** Средний  
**Статус:** [ ] Готов к выполнению

**Описание:**
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

**Зависимости:** M-2 (бизнес-логика), Auth Service (JWT токены)
**Оценка:** 2-3 дня

---

## M-4: Добавление Prometheus метрик в Auth Service  
**Роль:** DevOps/Backend Developer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Интегрировать Prometheus метрики в auth-service для мониторинга производительности, безопасности и состояния системы. Реализовать comprehensive набор метрик для всех критических операций и создать основу для alerting системы.

**Файлы для изучения:**
- `services/auth-service/` - текущая реализация сервиса
- `docs/specs/auth-service.md` - спецификация сервиса
- Prometheus Go client documentation

**Критерии готовности:**

**Базовые HTTP метрики:**
- [ ] Добавлена зависимость на `github.com/prometheus/client_golang` в go.mod
- [ ] Создан middleware для сбора HTTP метрик (request duration, status codes, request count)
- [ ] Реализован endpoint `GET /metrics` для экспорта метрик в Prometheus формате
- [ ] Настроены labels для метрик: method, endpoint, status_code

**Метрики аутентификации:**
- [ ] `auth_requests_total` - счетчик всех запросов аутентификации с labels: status, reason
- [ ] `auth_request_duration_seconds` - гистограмма времени обработки запросов
- [ ] `auth_telegram_validation_duration_seconds` - время валидации Telegram подписи
- [ ] `auth_new_users_total` - счетчик регистраций новых пользователей
- [ ] `auth_rate_limit_hits_total` - количество заблокированных запросов

**Метрики JWT токенов:**
- [ ] `jwt_tokens_generated_total` - счетчик созданных JWT токенов
- [ ] `jwt_tokens_validated_total` - счетчик валидированных токенов с labels: status
- [ ] `jwt_key_generation_duration_seconds` - время генерации RSA ключей
- [ ] `jwt_active_tokens_count` - gauge активных токенов в системе

**Интеграция и конфигурация:**
- [ ] Создан пакет `internal/metrics/` с инициализацией всех метрик
- [ ] Интегрированы метрики во все handlers, storage слои, services
- [ ] Настроен namespace для метрик: `auth_service_`
- [ ] Добавлена конфигурация метрик через переменные окружения

**Проверка:**
```bash
curl http://localhost:9090/metrics | grep auth_service # метрики видны
```

**Зависимости:** Auth Service полностью реализован
**Оценка:** 2-3 дня

---

## Низкий приоритет
<!-- Задачи для будущих итераций -->

## L-1: Мониторинг и метрики для Inventory Service
**Роль:** DevOps/Backend Developer
**Приоритет:** Низкий
**Статус:** [ ] Готов к выполнению

**Описание:**
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

**Зависимости:** M-3 (HTTP API), M-4 (auth-service метрики как референс)
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

**Зависимости:** L-1 (мониторинг), все предыдущие задачи Inventory Service
**Оценка:** 2-3 дня

---
**Критерии перехода в TODO:**
- [ ] Задача четко сформулирована
- [ ] Определены критерии готовности
- [ ] Проведена оценка трудозатрат
- [ ] Нет блокирующих зависимостей
- [ ] Назначен ответственный (опционально) 
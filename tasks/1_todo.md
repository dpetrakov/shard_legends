# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

**D-12: Исправление HMAC валидации Telegram WebApp в production**
**Роль:** Backend Developer
**Приоритет:** Высокий  
**Статус:** [ ] Требует расследования и исправления

**Описание:**
Auth Service успешно аутентифицирует пользователей Telegram WebApp, но HMAC подпись не совпадает с ожидаемой. Временно HMAC валидация отключена для development среды, но это критично для production безопасности.

**🔍 Детали проблемы:**
- **Полученный hash**: `12620fff246f641d134e88ff4bb2e1ee2c58d8d9b677be8d0754b750ea453b89`  
- **Вычисленный hash**: `22781e9fc836aa91dab8348748721df21f81055d47ed8d3ac9f6516a2de150d5`
- **Бот токен**: `7986845757:AAG1L4612Tag4DtXyPMSMzDr_7pZHFdgTxM` (@SLCWDevDimBot)
- **Статус**: Все остальные проверки (auth_date, user data, структура) проходят успешно

**Возможные причины:**
1. Фронтенд инициализирован с другим ботом/токеном
2. Модификация initData после получения от Telegram  
3. Неправильная обработка URL encoding/JSON экранирования
4. Использование устаревшего Telegram WebApp API

**Критерии выполнения:**
- [ ] **Расследование**: Определить точную причину несовпадения HMAC
- [ ] **Проверка бота**: Убедиться что фронтенд использует правильный бот
- [ ] **Тестирование с оригинальными данными**: Проверить с реальными Telegram initData  
- [ ] **Исправление алгоритма**: При необходимости скорректировать валидацию
- [ ] **Включение строгой проверки**: Убрать временный bypass для production
- [ ] **Regression тесты**: Убедиться что исправление не ломает существующие тесты

**Временное решение:**
Аутентификация работает с предупреждением в логах `"Continuing despite HMAC validation failure for development"`

**Файлы для изучения:**
- `services/auth-service/internal/services/telegram.go:77-82` - место временного bypass
- Frontend инициализация Telegram WebApp
- Логи auth-service для анализа приходящих данных

---


**D-10: Депривация эндпоинта /jwks в Auth Service**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] ~80% выполнено (осталось исправить unit-тесты)

**Описание:**
Полностью удалить эндпоинт `/jwks`; для получения публичного ключа остаётся только `/public-key.pem`. В архитектурной документации зафиксировать, что поддержка JWKS отложена на следующую версию.

**Критерии выполнения:**
- [x] **Удаление handler**: HTTP handler для `/jwks` эндпоинта отсутствует ✅
- [x] **Удаление маршрутизации**: Маршруты для `/jwks` не зарегистрированы в main.go ✅
- [x] **Обновление OpenAPI**: Убрать упоминание "Тот же ключ, что и в /jwks" из спецификации ✅
- [x] **Обновление README**: README.md не содержит упоминаний /jwks ✅ 
- [x] **Архитектурное решение**: В `docs/architecture/architecture.md` добавить секцию о решении отложить JWKS ✅
- [x] **Проверка клиентов**: Клиентские вызовы JWKS отсутствуют ✅
- [x] **Очистка stale кода**: Убрать упоминания `/jwks` из `internal/middleware/metrics.go`
- [ ] **Тестирование**: Все тесты проходят без зависимости от `/jwks`

**Файлы для изучения:**
- `services/auth-service/internal/handlers/`
- `docs/specs/auth-service-openapi.yml`
- `docs/architecture/architecture.md`

---

**D-11: Проверка и коррекция HMAC-валидации Telegram initData**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] ~70% выполнено (остались unit-тесты с офиц. векторами и интеграционные тесты)

**Описание:**
Проверить реализацию генерации `secret_key = HMAC_SHA256(bot_token, "WebAppData")` в `telegram_validator.go` и при необходимости исправить порядок аргументов.

**🚨 ОБНАРУЖЕННАЯ ОШИБКА:**
В `services/telegram.go:290` функция `generateSecretKeyForToken()` имеет **неправильный порядок аргументов HMAC**:
```go
// НЕПРАВИЛЬНО (текущая реализация):
h := hmac.New(sha256.New, []byte("WebAppData"))
h.Write([]byte(botToken))

// ПРАВИЛЬНО (должно быть):
h := hmac.New(sha256.New, []byte(botToken))
h.Write([]byte("WebAppData"))
```

**Критерии выполнения:**
- [x] **🔥 ИСПРАВЛЕНИЕ КРИТИЧЕСКОЙ ОШИБКИ**: Поменять местами аргументы в `generateSecretKeyForToken()` ✅
- [ ] **Unit-тесты с официальными примерами**: Добавить тесты с официальными test vectors от Telegram
- [x] **Negative path тестирование**: Тесты с некорректными данными для проверки валидации ✅
- [x] **Верификация алгоритма**: Подтвердить корректность исправленного алгоритма через тесты ✅
- [x] **Regression тесты**: Убедиться что существующие тесты проходят после исправления ✅
- [x] **Обновление документации**: Актуализировать комментарии в коде ✅
- [ ] **Integration тесты**: Проверить работу с реальными Telegram Web App данными

**Файлы для изучения:**
- `services/auth-service/internal/services/telegram_validator.go`
- `docs/specs/auth-service.md`

---

**: Разделить внутренние и внешние эндпоинты inventory-service по разным портам**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] ~80% выполнено (серверы и middleware готовы, остаётся документация и тесты)

**Описание:**
Реализовать разделение эндпоинтов inventory-service на внешние (публичные) и внутренние по аналогии с auth-service. Это решит проблему P-2 из аудита - незащищённые внутренние эндпоинты.

**🔍 Анализ проблемы:**
- Эндпоинты `reserve`, `return-reserve`, `consume-reserve` не требуют аутентификации
- При ошибке конфигурации API-Gateway эти эндпоинты становятся публично доступными
- Злоумышленник может напрямую изменять баланс пользователей

**Критерии выполнения:**
- [x] **Разделение серверов**: Создать два HTTP-сервера как в auth-service
  - Публичный сервер (PORT): эндпоинты для клиентов с JWT аутентификацией
  - Внутренний сервер (INTERNAL_PORT): служебные эндпоинты, health, metrics
- [x] **Маршрутизация эндпоинтов**:
  - Публичные: `/inventory/balance`, `/inventory/items` (требуют JWT)
  - Внутренние: `/reserve`, `/return-reserve`, `/consume-reserve`, `/health`, `/metrics`
- [x] **Конфигурация**: Добавить `INTERNAL_SERVICE_PORT` в config.go
- [x] **Middleware**: Внутренние эндпоинты требуют Service-JWT с ролью `internal`
- [ ] **Обновление документации**: Актуализировать OpenAPI спецификацию
- [ ] **Тестирование**:
  - Unit-тесты для проверки JWT middleware на внутренних эндпоинтах
  - E2E тест: 401 при отсутствии Service-JWT на `/reserve`

**Файлы для изучения:**
- `services/auth-service/cmd/main.go:59-154` - пример разделения серверов
- `services/inventory-service/cmd/server/main.go` - текущая реализация
- `services/inventory-service/internal/handlers/` - группировка эндпоинтов

---

**D-14: Исправить гонки при резервировании (oversell) в inventory-service**
**Роль:** Backend Developer  
**Приоритет:** Высокий
**Статус:** [ ] ~50% выполнено (введён транзакционный `SELECT ... FOR UPDATE` и миграция с ограничениями, нужны тесты и финальный аудит)

**Описание:**
Устранить критическую уязвимость P-3 из аудита - возможность oversell из-за race conditions при конкурентном резервировании предметов.

**🚨 Критическая проблема:**
- `ReserveItems()` выполняет SELECT остатка, затем INSERT операции без блокировок
- Конкурентные запросы могут проверить остаток одновременно и оба пройти валидацию
- Результат: баланс становится отрицательным, пользователи получают "бесплатные" предметы

**Критерии выполнения:**
- [x] **Атомарные блокировки**: Использовать `SELECT ... FOR UPDATE` на строках `balances`
- [x] **Database constraints**: Добавить `CHECK (available_quantity >= reserved_quantity)`
- [x] **Обработка ошибок**: Ловить `pgerrcode.CheckViolation` и возвращать понятную ошибку
- [x] **Изоляция транзакций**: `SERIALIZABLE` level используется в операции резервирования
- [x] **Миграция БД**: Создана `006_add_balance_constraints.up.sql`
- [ ] **Конкурентное тестирование**:
  - Unit-тест с `t.Parallel()` - 100 конкурентных резервов одного предмета
  - Проверка что итоговый баланс корректен (не oversold)
  - Integration тест с реальными транзакциями

**Файлы для изучения:**
- `services/inventory-service/internal/service/operations.go` - ReserveItems()
- `services/inventory-service/internal/storage/postgres.go` - SQL запросы  
- `migrations/002_create_inventory_schema.up.sql` - текущая схема

---

**D-15: Устранить N+1 запросы и создать оптимальные индексы для inventory-service**
**Роль:** Backend Developer
**Приоритет:** Средний  
**Статус:** [ ] Готов к выполнению

**Описание:**
Решить проблему P-4 из аудита - оптимизировать производительность `GetUserInventory` путём устранения N+1 запросов и создания правильных индексов.

**🐌 Проблемы производительности:**
- `GetUserInventory` вызывает `GetItemWithDetails` для каждого предмета отдельно
- При 300-400 предметах время ответа >400ms  
- UNION запросы без индекса на `balance_date`
- Отсутствуют составные индексы для частых запросов

**Критерии выполнения:**
- [ ] **Оптимизация запросов**: Заменить N+1 на один JOIN-запрос
  ```sql
  SELECT b.*, i.name, i.description, c.name as collection_name, q.name as quality_name
  FROM inventory.balances b
  JOIN inventory.items i ON b.item_id = i.id  
  JOIN inventory.collections c ON i.collection_id = c.id
  JOIN inventory.qualities q ON i.quality_id = q.id
  WHERE b.user_id = $1 AND b.available_quantity > 0
  ```
- [ ] **Создание индексов**: Новая миграция `007_optimize_inventory_indexes.up.sql`
  ```sql
  -- Составные индексы для частых запросов
  CREATE INDEX idx_balances_user_item_section 
  ON inventory.balances (user_id, item_id, section_id);
  
  CREATE INDEX idx_balances_user_date 
  ON inventory.balances (user_id, balance_date);
  
  -- Индекс для JOIN операций
  CREATE INDEX idx_items_collection_quality 
  ON inventory.items (collection_id, quality_id);
  
  -- Оптимизация daily_balances
  CREATE INDEX idx_daily_balances_user_date 
  ON inventory.daily_balances (user_id, balance_date);
  ```
- [ ] **Рефакторинг repository**: Новый метод `GetUserInventoryOptimized()`
- [ ] **EXPLAIN ANALYZE**: Добавить в CI проверку планов выполнения запросов
- [ ] **Бенчмарк тесты**: 
  - Сравнение производительности до/после оптимизации
  - Цель: <100ms для 500 предметов
  - Load test с k6: 95% запросов <200ms

**Файлы для изучения:**
- `services/inventory-service/internal/storage/postgres.go` - GetUserInventory()
- `services/inventory-service/internal/storage/balance_repository.go`
- `migrations/002_create_inventory_schema.up.sql` - текущие индексы

---

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->

**D-7: Внедрение метрик и мониторинга в Production Service**
**Роль:** DevOps/Backend Developer
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Интеграция мониторинга и метрик для наблюдения за работой production-service по аналогии с auth-service и inventory-service.

**Критерии выполнения:**
- [ ] **Prometheus метрики**: HTTP метрики, бизнес-метрики производства, метрики модификаторов
- [ ] **Grafana дашборд**: Создание дашборда `production-service-metrics.json`
- [ ] **Алерты**: Правила алертов в `deploy/monitoring/alerts.yml`
- [ ] **Интеграция middleware**: Подключение метрик в HTTP handlers и бизнес-логику
- [ ] **Мониторинг производственных операций**: Метрики времени производства, успешности claim операций
- [ ] **Dependency health**: Мониторинг статуса inventory-service и user-service
- [ ] **Тестирование**: Проверка сбора метрик в dev окружении

**Файлы для изучения:**
- `services/auth-service/pkg/metrics/` - пример реализации
- `deploy/monitoring/grafana/dashboards/auth-service-metrics.json` - шаблон дашборда
- `docs/specs/production-service.md`

---

**A-4: Формирование стартового списка предметов игры**
**Роль:** Analyst
**Приоритет:** Средний
**Статус:** [ ] ~60% выполнено (формирование и первичная валидация списка завершены)

**Описание:**
Создать спецификацию `docs/specs/items-initial.md`, содержащую полный перечень предметов, необходимых для механики сундуков и ключей, а также всех возможных предметов, выпадающих из сундуков. Документ будет использоваться для генерации SQL-миграций начальных данных инвентаря.

**Критерии выполнения:**
- [x] Сформирован список предметов с **кодом**, **названием** и **описанием** в соответствии с классификаторами `item_class`, `item_type` и концепцией сундуков/ключей
- [x] Покрыты все категории: ресурсы, реагенты, ускорители, ключи, сундуки, чертежи (Blueprint), стартовые инструменты
- [x] Документ согласован с `docs/specs/classifiers.md` (не содержит конфликтующих кодов)
- [x] Проведён кросс-чек с концепцией `docs/concept/game-mechanics-chests-keys-deck-minigame.md`
- [x] Подготовлена структура таблицы для будущего расширения (дополнительные атрибуты: quality, collection и др.)
- [ ] Получено ревью от Game Designer/Product Owner

**Документы для уточнения:**
- `docs/specs/classifiers.md`
- `docs/concept/game-mechanics-chests-keys-deck-minigame.md`
- `docs/architecture/database.dbml`

---

## Низкий приоритет
<!-- Задачи для будущих итераций -->

**L-1: Мониторинг и метрики для Inventory Service**
**Роль:** DevOps/Backend Developer
**Приоритет:** Низкий
**Статус:** [ ] ~70% выполнено (требует доработки специфичных метрик)

**Описание:**
Доработка системы мониторинга inventory-service - добавление специфичных бизнес-метрик и интеграция в middleware.

**Критерии выполнения (оставшиеся ~30%):**
- [ ] **Специфичные inventory метрики**: 
  - `inventory_balance_calculation_duration_seconds` - время расчета остатков с label cache_hit
  - `inventory_daily_balance_created_total` - ленивое создание дневных остатков
  - `inventory_cache_hit_ratio` - эффективность кеширования по типам
  - `inventory_classifier_conversions_total` - преобразования код↔UUID
  - `inventory_insufficient_balance_errors_total` - ошибки недостаточного баланса
  - `inventory_transaction_rollbacks_total` - откаты транзакций с причинами
  - `inventory_service_up` и `inventory_service_start_time_seconds`
- [ ] **Интеграция в бизнес-логику**: Подключение метрик в repositories и services
- [ ] **Обновление middleware**: Дополнение `internal/middleware/metrics.go` новыми метриками

**Статус выполненного (~70%):**
- ✅ Базовые HTTP, DB, Redis метрики полностью реализованы
- ✅ Grafana дашборд создан с 7 группами панелей
- ✅ 8 алертов в Prometheus настроены
- ✅ HTTP middleware интеграция работает

**Файлы для изучения:**
- `services/inventory-service/pkg/metrics/metrics.go` - текущие метрики
- `services/inventory-service/internal/service/` - места для интеграции метрик

**Оценка:** 1-1.5 дня доработки

---

**L-2: Интеграционное тестирование и развертывание для Inventory Service**
**Роль:** QA/DevOps Engineer
**Приоритет:** Низкий
**Статус:** [ ] Готов к выполнению

**Описание:**
Создание комплексной системы тестирования inventory-service, включая интеграционные тесты с другими сервисами, E2E тесты и настройку CI/CD pipeline.

**Критерии выполнения:**
- [ ] **Интеграционные тесты**: 
  - `tests/integration/auth/` - тесты JWT аутентификации с auth-service
  - `tests/e2e/` - полные пользовательские сценарии
  - Тесты concurrent операций и race conditions
- [ ] **Нагрузочные тесты**:
  - `tests/load/k6_loadtest.js` - k6 тесты производительности
  - `tests/load/artillery_loadtest.yml` - Artillery тесты
  - SLA: 95% запросов <1s, error rate <10%
- [ ] **CI/CD Pipeline**:
  - `.github/workflows/inventory-service.yml` - автоматические тесты и деплой
  - Docker образы в Container Registry
  - Автоматический деплой в staging/production  
- [ ] **Development environment**:
  - `docker-compose.dev.yml` - полный стек для разработки
  - Включает inventory-service, auth-service, PostgreSQL, Redis, Prometheus, Grafana

**Зависимости:** ⏳ L-1 (мониторинг), ✅ M-1 (модели), ✅ M-2 (бизнес-логика), ✅ M-3 (HTTP API)

**Оценка:** 2-3 дня

---

**Критерии перехода в TODO:**
- [ ] Задача четко сформулирована
- [ ] Определены critical path критерии готовности  
- [ ] Проведена оценка трудозатрат
- [ ] Нет блокирующих зависимостей
- [ ] Назначен ответственный (опционально) 
# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

**D-18: Мультиязычная поддержка названий и описаний предметов (i18n)**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] Готов к выполнению

**Описание:**
Реализовать универсальную систему интернационализации (i18n) для названий и описаний предметов. Система должна поддерживать мультиязычность и быть готова к масштабированию на другие игровые сущности.

**Критерии выполнения:**
- [ ] **Архитектурное проектирование**: Спроектировать универсальную i18n систему для игровых сущностей
- [ ] **Документация архитектуры**: Создать `docs/architecture/i18n-system.md` с описанием решения
- [ ] **Обновление схемы БД**: Добавить i18n таблицы в `docs/architecture/database.dbml`
- [ ] **Основная миграция**: Создать миграцию для i18n таблиц (`008_create_i18n_schema.up.sql`)
- [ ] **Dev-data миграция**: Обновить `003_reset_items.sql` для загрузки переводов RU/EN
- [ ] **Обновление моделей**: Добавить i18n поддержку в Go models
- [ ] **Обновление storage**: Реализовать методы для работы с переводами
- [ ] **Backward compatibility**: Обеспечить совместимость с существующим API

**Технические требования:**
- Поддержка языков: RU (русский), EN (английский) 
- Fallback на базовый язык при отсутствии перевода
- Универсальная система для любых игровых сущностей
- Кеширование переводов в Redis
- Индексы для быстрого поиска по entity_type + entity_id + language

**Файлы для изучения:**
- `docs/specs/items-initial.md` - исходные данные для переводов
- `migrations/002_create_inventory_schema.up.sql` - текущая схема БД
- `services/inventory-service/internal/models/` - модели для обновления

---

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
- [x] **Оптимизация запросов**: Заменить N+1 на один JOIN-запрос
  ```sql
  SELECT b.*, i.name, i.description, c.name as collection_name, q.name as quality_name
  FROM inventory.balances b
  JOIN inventory.items i ON b.item_id = i.id  
  JOIN inventory.collections c ON i.collection_id = c.id
  JOIN inventory.qualities q ON i.quality_id = q.id
  WHERE b.user_id = $1 AND b.available_quantity > 0
  ```
- [x] **Создание индексов**: Новая миграция `007_optimize_inventory_indexes.up.sql`
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
- [x] **Рефакторинг repository**: Новый метод `GetUserInventoryOptimized()`
- [x] **Бенчмарк тесты**: 
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
- [x] Получено ревью от Game Designer/Product Owner

**Документы для уточнения:**
- `docs/specs/classifiers.md`
- `docs/concept/game-mechanics-chests-keys-deck-minigame.md`
- `docs/architecture/database.dbml`

---

**D-16: Скрипт dev-data — полная перезагрузка классификаторов**
**Роль:** Backend Developer
**Приоритет:** Средний  
**Статус:** [ ] Готов к выполнению

**Описание:**
Создать SQL-скрипт в `migrations/dev-data/inventory-service/002_reset_classifiers.sql`, который:
1. Полностью очищает таблицы `inventory.classifier_items` и `inventory.classifiers` (TRUNCATE CASCADE).
2. Загружает все классификаторы и элементы из `docs/specs/classifiers.md`.
   - UUID для `classifiers.id` и `classifier_items.id` генерируется функцией `gen_random_uuid()`.
   - При вставке элементов нужно искать `classifier_id` по коду классификатора.
3. Скрипт должен быть идемпотентен: повторный запуск приводит к тем же данным (использовать TRUNCATE, `ON CONFLICT DO NOTHING`).
4. Добавить секцию `BEGIN; … COMMIT;` для atomic-rebuild.

**Критерии выполнения:**
- [ ] Таблицы очищаются и заполняются корректно без остаточных данных.
- [ ] Количество загруженных классификаторов и элементов соответствует спецификации.
- [ ] Повторный запуск скрипта не вызывает ошибок и не меняет данные.
- [ ] README в папке `migrations/dev-data/` обновлён с описанием скрипта.
- [ ] Unit-тест `internal/storage/classifier_storage_test.go` использует новые данные.

**Файлы для изучения:**
- `docs/specs/classifiers.md`
- `services/inventory-service/internal/storage/classifier_storage.go`

---

**D-17: Скрипт dev-data — полная перезагрузка предметов**
**Роль:** Backend Developer
**Приоритет:** Средний  
**Статус:** [ ] Готов к выполнению

**Описание:**
Создать SQL-скрипт в `migrations/dev-data/inventory-service/003_reset_items.sql`, который:
1. Полностью очищает таблицы `inventory.items`, `inventory.collections`, `inventory.qualities` (и связанные с ними FK) через `TRUNCATE … CASCADE`.
2. Загружает предметы из `docs/specs/items-initial.md`, используя UUID, указанные в документе.
3. Для коллекций и качеств — при необходимости создаёт новые записи, если их нет, с генерацией UUID.
4. Скрипт идемпотентен: вставка выполняется с `ON CONFLICT (id) DO UPDATE SET …`.

**Критерии выполнения:**
- [ ] Все предметы из спецификации успешно загружены, UUID совпадают.
- [ ] Связанные классификаторы (quality, collection) существуют и корректно связаны.
- [ ] Повторный запуск скрипта не изменяет данные (rows = 0).
- [ ] Добавлен unit-тест `internal/storage/item_storage_test.go` на выборку предмета по UUID.

**Файлы для изучения:**
- `docs/specs/items-initial.md`
- `services/inventory-service/internal/storage/item_storage.go`

---

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

**D-19: Эндпоинт для получения локализованной информации о предметах**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] ~70% выполнено (основная функциональность готова, остались доработки)

**Описание:**
Создать отдельный эндпоинт в inventory-service для получения локализованной информации (название, описание, изображение) по списку предметов. Предметы могут включать качество и коллекцию для определения изображения.

**Критерии выполнения:**
- [x] **API Design**: Создать эндпоинт `POST /items/details` с параметром языка ✅
- [x] **Модели данных**: Добавить структуры для batch-запроса предметов с переводами ✅
- [x] **I18n интеграция**: Интегрировать с системой переводов из схемы i18n ✅
- [x] **Storage методы**: Реализовать методы для batch-получения переводов и изображений ✅
- [x] **Бизнес-логика**: Обработка fallback переводов и определение URL изображений ✅
- [x] **HTTP handlers**: Создать handlers с валидацией запросов ✅
- [x] **Маршрутизация**: Добавить routes в main.go ✅
- [ ] **Реализация изображений**: Завершить метод GetItemImagesBatch с реальными запросами к БД
- [ ] **Кеширование**: Реализовать Redis кеш для переводов (TTL 24ч) и изображений (TTL 24ч)
- [ ] **Unit-тесты**: Написать тесты для service и storage методов
- [ ] **Документация**: Обновить OpenAPI спецификацию и архитектурные документы

**Технические требования:**
- Поддержка batch-запросов до 100 предметов за раз
- Fallback на дефолтный язык из i18n.languages (is_default = true) при отсутствии перевода
- Определение изображения по item_id + collection_id + quality_level_id
- Кеширование переводов на 24 часа, изображений на 1 час
- Валидация поддерживаемых языков и существования предметов
- POST используется для передачи списка предметов в теле запроса

**Структура запроса:**
```json
POST /items/details?lang=ru
{
  "items": [
    {
      "item_id": "uuid-stone",
      "collection": "winter_2025",
      "quality_level": "stone"
    }
  ]
}
```

**Структура ответа:**
```json
{
  "items": [
    {
      "item_id": "uuid-stone",
      "code": "stone",
      "name": "Камень",
      "description": "Базовый строительный материал",
      "image_url": "https://cdn.example.com/items/stone_winter_2025_stone.png",
      "collection": "winter_2025",
      "quality_level": "stone"
    }
  ]
}
```

**Реализовано (70%):**
- ✅ Эндпоинт `POST /api/inventory/items/details?lang=ru` 
- ✅ Модели: `ItemDetailsRequest`, `ItemDetailsResponse`, `Translation`, `Language`
- ✅ Storage методы: `GetItemsBatch`, `GetTranslationsBatch`, `GetDefaultLanguage`
- ✅ Бизнес-логика с fallback переводами в `GetItemsDetails`
- ✅ HTTP handlers с валидацией (до 100 предметов)
- ✅ Маршрутизация в main.go

**Осталось доделать (30%):**
- [ ] Реализация `GetItemImagesBatch` с реальными SQL запросами
- [ ] Redis кеширование переводов и изображений
- [ ] Unit-тесты для новых методов
- [ ] OpenAPI документация

**Файлы созданы/изменены:**
- `internal/models/item.go` - новые модели и константы
- `internal/storage/item_storage.go` - batch методы для переводов
- `internal/service/interfaces.go` - обновленные интерфейсы  
- `internal/service/inventory_service.go` - метод GetItemsDetails
- `internal/handlers/items_details.go` - новый handler
- `cmd/server/main.go` - маршруты

---

**Критерии перехода в TODO:**
- [ ] Задача четко сформулирована
- [ ] Определены critical path критерии готовности  
- [ ] Проведена оценка трудозатрат
- [ ] Нет блокирующих зависимостей
- [ ] Назначен ответственный (опционально) 
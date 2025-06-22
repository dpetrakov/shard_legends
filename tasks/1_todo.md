# TODO — готово к работе

> Задачи проработаны, приоритизированы и готовы к выполнению. Отсортированы по приоритету.

## Высокий приоритет
<!-- Критически важные задачи, блокирующие другие -->

## Средний приоритет  
<!-- Важные задачи для текущего спринта/итерации -->

## M-1: Добавление Prometheus метрик в Auth Service
**Роль:** Разработчик
**Приоритет:** Средний
**Статус:** [ ] Готов к выполнению

**Описание:**
Интегрировать Prometheus метрики в auth-service для мониторинга производительности, безопасности и состояния системы. Реализовать comprehensive набор метрик для всех критических операций и создать основу для alerting системы.

**Критерии готовности:**

**Базовые HTTP метрики:**
- [ ] Добавлена зависимость на `github.com/prometheus/client_golang` в go.mod
- [ ] Создан middleware для сбора HTTP метрик (request duration, status codes, request count)
- [ ] Реализован endpoint `GET /metrics` для экспорта метрик в Prometheus формате
- [ ] Настроены labels для метрик: method, endpoint, status_code

**Метрики аутентификации:**
- [ ] `auth_requests_total` - счетчик всех запросов аутентификации с labels: status (success/failed), reason (valid/invalid_signature/expired/rate_limited)
- [ ] `auth_request_duration_seconds` - гистограмма времени обработки запросов аутентификации
- [ ] `auth_telegram_validation_duration_seconds` - время валидации Telegram подписи
- [ ] `auth_new_users_total` - счетчик регистраций новых пользователей
- [ ] `auth_rate_limit_hits_total` - количество заблокированных запросов по rate limiting с label: ip

**Метрики JWT токенов:**
- [ ] `jwt_tokens_generated_total` - счетчик созданных JWT токенов
- [ ] `jwt_tokens_validated_total` - счетчик валидированных токенов с labels: status (valid/invalid/expired)
- [ ] `jwt_key_generation_duration_seconds` - время генерации RSA ключей
- [ ] `jwt_active_tokens_count` - gauge активных токенов в системе
- [ ] `jwt_tokens_per_user_histogram` - распределение количества токенов на пользователя

**Метрики Redis операций:**
- [ ] `redis_operations_total` - счетчик операций с Redis с labels: operation (get/set/del/exists), status (success/error)
- [ ] `redis_operation_duration_seconds` - время выполнения операций с Redis
- [ ] `redis_connection_pool_active` - активные соединения в pool
- [ ] `redis_connection_pool_idle` - idle соединения в pool
- [ ] `redis_token_cleanup_duration_seconds` - время выполнения cleanup операций
- [ ] `redis_expired_tokens_cleaned_total` - количество удаленных просроченных токенов
- [ ] `redis_cleanup_processed_users_total` - количество пользователей обработанных при cleanup

**Метрики PostgreSQL:**
- [ ] `postgres_operations_total` - счетчик операций с БД с labels: operation (select/insert/update/delete), table, status
- [ ] `postgres_operation_duration_seconds` - время выполнения SQL запросов
- [ ] `postgres_connection_pool_active` - активные соединения к БД
- [ ] `postgres_connection_pool_idle` - idle соединения к БД
- [ ] `postgres_connection_pool_max` - максимальное количество соединений

**Метрики здоровья системы:**
- [ ] `auth_service_up` - gauge доступности сервиса (1 = up, 0 = down)
- [ ] `auth_service_start_time_seconds` - время запуска сервиса (unix timestamp)
- [ ] `auth_dependencies_healthy` - gauge здоровья зависимостей с labels: dependency (postgres/redis/jwt_keys)
- [ ] `auth_memory_usage_bytes` - использование памяти сервисом
- [ ] `auth_goroutines_count` - количество активных goroutines

**Метрики админских операций:**
- [ ] `admin_operations_total` - счетчик админских операций с labels: operation (get_stats/revoke_token/cleanup), status
- [ ] `admin_token_revocations_total` - счетчик отозванных токенов с labels: method (single/user_all/manual)
- [ ] `admin_cleanup_operations_total` - счетчик ручных cleanup операций

**Интеграция и конфигурация:**
- [ ] Создан пакет `internal/metrics/` с инициализацией всех метрик
- [ ] Интегрированы метрики во все handlers (auth, health, admin)
- [ ] Интегрированы метрики в storage слои (postgres, redis)
- [ ] Интегрированы метрики в services (jwt, telegram)
- [ ] Добавлена конфигурация метрик через переменные окружения
- [ ] Настроен namespace для метрик: `auth_service_`

**Тестирование и документация:**
- [ ] Unit тесты для metrics middleware с проверкой корректности labels
- [ ] Integration тесты проверяющие сбор метрик в реальных сценариях
- [ ] Тесты проверки формата метрик соответствующего Prometheus стандартам
- [ ] Документация всех метрик в README с описанием назначения
- [ ] Примеры Prometheus queries для типичных мониторинговых задач
- [ ] Рекомендации по alerting rules для критических метрик

**Production готовность:**
- [ ] Endpoint `/metrics` защищен от публичного доступа (только для Prometheus)
- [ ] Метрики не влияют на производительность основных операций (async сбор где возможно)
- [ ] Обработка ошибок сбора метрик не влияет на бизнес логику
- [ ] Настроена ротация и cleanup старых метрик для предотвращения memory leaks
- [ ] Добавлено логирование ошибок инициализации metrics системы

**Интеграция с мониторингом:**
- [ ] Создана конфигурация Prometheus для scraping auth-service метрик
- [ ] Обновлен docker-compose.yml для экспозиции metrics endpoint
- [ ] Подготовлены базовые Grafana дашборды для визуализации метрик
- [ ] Документированы рекомендуемые alert правила для production

**Ссылки:**
- Prometheus Go client: https://github.com/prometheus/client_golang
- Метрики в спецификации: `docs/specs/auth-service.md` (раздел "Метрики")
- Мониторинг стратегия: `docs/architecture/logging-monitoring-strategy.md`

**Зависимости:** Auth Service должен быть полностью реализован и протестирован
**Оценка:** 2-3 дня

---

## Низкий приоритет
<!-- Задачи для будущих итераций -->

---
**Критерии перехода в TODO:**
- [ ] Задача четко сформулирована
- [ ] Определены критерии готовности
- [ ] Проведена оценка трудозатрат
- [ ] Нет блокирующих зависимостей
- [ ] Назначен ответственный (опционально) 
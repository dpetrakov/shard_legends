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
# In Progress — в работе

> Задачи, которые активно выполняются. Максимум 3-5 задач одновременно для фокуса.

---

**Лимит WIP:** Максимум 5 задач в этом разделе одновременно.


**D-11: Проверка и коррекция HMAC-валидации Telegram initData**
**Роль:** Backend Developer
**Приоритет:** Высокий
**Статус:** [ ] ~70% выполнено (остались unit-тесты с офиц. векторами)

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

**Файлы для изучения:**
- `services/auth-service/internal/services/telegram_validator.go`
- `docs/specs/auth-service.md`

---

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
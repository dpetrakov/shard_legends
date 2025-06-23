# Задача 4: Бизнес-логика и общие алгоритмы

## Описание

Реализация основных бизнес-алгоритмов inventory-service согласно спецификации. Включает расчет остатков, создание дневных балансов, преобразование кодов и валидацию операций.

## Цели

1. Реализовать 6 ключевых алгоритмов из спецификации
2. Создать service слой для бизнес-логики
3. Добавить кеширование для производительности
4. Реализовать транзакционную безопасность
5. Покрыть тестами все бизнес-сценарии

## Подзадачи

### 4.1. Service интерфейсы
**Файл**: `internal/service/interfaces.go`

**Интерфейсы**:
```go
type InventoryService interface {
    // Основные операции
    CalculateCurrentBalance(ctx context.Context, req BalanceRequest) (int64, error)
    CreateDailyBalance(ctx context.Context, req DailyBalanceRequest) (*models.DailyBalance, error)
    CheckSufficientBalance(ctx context.Context, req SufficientBalanceRequest) error
    CreateOperationsInTransaction(ctx context.Context, operations []*models.Operation) ([]uuid.UUID, error)
    
    // Утилитарные операции
    ConvertClassifierCodes(ctx context.Context, req CodeConversionRequest) (*CodeConversionResponse, error)
    InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
}

type ClassifierService interface {
    GetClassifierMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
    GetReverseClassifierMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
    RefreshClassifierCache(ctx context.Context, classifierCode string) error
}
```

### 4.2. Алгоритм расчета текущего остатка
**Файл**: `internal/service/balance_calculator.go`

**Функция**: `CalculateCurrentBalance`

**Входные данные**:
```go
type BalanceRequest struct {
    UserID         uuid.UUID
    SectionID      uuid.UUID
    ItemID         uuid.UUID
    CollectionID   uuid.UUID
    QualityLevelID uuid.UUID
}
```

**Алгоритм**:
1. Проверить кеш Redis по ключу `inventory:{user_id}:{section_id}:{item_id}:{collection_id}:{quality_level_id}`
2. Если в кеше есть - вернуть значение
3. Если нет:
   - Найти последний дневной остаток из `daily_balances`
   - Если дневного остатка нет - вызвать `CreateDailyBalance` для его создания
   - Суммировать операции из `operations` от даты остатка до текущего момента
   - Вычислить: current = daily_balance + sum(quantity_change)
   - Сохранить в кеш на 1 час
4. Вернуть результат

### 4.3. Алгоритм создания дневного остатка
**Файл**: `internal/service/daily_balance_creator.go`

**Функция**: `CreateDailyBalance`

**Входные данные**:
```go
type DailyBalanceRequest struct {
    UserID         uuid.UUID
    SectionID      uuid.UUID
    ItemID         uuid.UUID
    CollectionID   uuid.UUID
    QualityLevelID uuid.UUID
    TargetDate     time.Time // обычно вчерашняя дата
}
```

**Алгоритм**:
1. Проверить, что остаток за target_date еще не существует
2. Найти последний существующий дневной остаток до target_date (может быть давно)
3. Если не найден - начальный остаток = 0
4. Суммировать ВСЕ операции от даты найденного остатка до конца target_date
5. Вычислить: new_balance = previous_balance + sum(operations)
6. Создать запись в `daily_balances` ТОЛЬКО для target_date
7. Вернуть созданный остаток

**Важно**: Создается остаток только за target_date, промежуточные дни НЕ заполняются

### 4.4. Алгоритм преобразования кодов
**Файл**: `internal/service/code_converter.go`

**Функция**: `ConvertClassifierCodes`

**Входные данные**:
```go
type CodeConversionRequest struct {
    Direction string // "toUUID" или "fromUUID"
    Data      map[string]interface{} // объект с кодами/UUID
}

type CodeConversionResponse struct {
    Data map[string]interface{} // преобразованный объект
}
```

**Алгоритм**:
1. Загрузить маппинг код<->UUID из кеша
2. Если нет в кеше - загрузить из БД и закешировать на 24 часа
3. Преобразовать все коды в UUID (или обратно)
4. Обработать отсутствующие поля как стандартные значения
5. Вернуть преобразованный объект

### 4.5. Алгоритм проверки достаточности остатков
**Файл**: `internal/service/balance_checker.go`

**Функция**: `CheckSufficientBalance`

**Входные данные**:
```go
type SufficientBalanceRequest struct {
    UserID    uuid.UUID
    SectionID uuid.UUID
    Items     []ItemQuantityCheck
}

type ItemQuantityCheck struct {
    ItemID         uuid.UUID
    CollectionID   uuid.UUID
    QualityLevelID uuid.UUID
    RequiredQty    int64
}
```

**Алгоритм**:
1. Для каждого предмета:
   - Вызвать `CalculateCurrentBalance`
   - Сравнить с требуемым количеством
   - Если недостаточно - добавить в список недостающих
2. Если список недостающих не пуст - вернуть ошибку с деталями
3. Иначе вернуть успех

### 4.6. Алгоритм создания операций в транзакции
**Файл**: `internal/service/operation_creator.go`

**Функция**: `CreateOperationsInTransaction`

**Входные данные**:
```go
type OperationBatch struct {
    Operations []models.Operation
    UserID     uuid.UUID
}
```

**Алгоритм**:
1. Валидировать все операции (quantity_change != 0)
2. Преобразовать коды классификаторов в UUID через `ConvertClassifierCodes`
3. Для каждой операции:
   - Установить created_at = now()
   - Вставить запись в `operations`
4. После успешной вставки всех операций - вызвать `InvalidateUserCache`
5. Вернуть ID созданных операций

### 4.7. Алгоритм сброса кеша пользователя
**Файл**: `internal/service/cache_manager.go`

**Функция**: `InvalidateUserCache`

**Входные данные**:
```go
userID uuid.UUID
```

**Алгоритм**:
1. Найти все ключи кеша по паттерну `inventory:{user_id}:*`
2. Удалить найденные ключи из Redis
3. Опционально: логировать количество удаленных ключей

### 4.8. Реализация InventoryService
**Файл**: `internal/service/inventory_service.go`

**Содержание**:
- Главная реализация всех интерфейсов
- Координация между репозиториями
- Транзакционная логика
- Error handling и логирование

### 4.9. Тесты бизнес-логики
**Файлы**: `internal/service/*_test.go`

**Типы тестов**:
- Unit тесты для каждого алгоритма
- Integration тесты с БД
- Test cases для edge cases
- Performance тесты для критичных операций

**Важные сценарии**:
- Расчет остатка для нового пользователя
- Создание дневного остатка при отсутствии предыдущих
- Преобразование кодов со стандартными значениями
- Проверка недостаточных остатков
- Rollback транзакций при ошибках

## Критерии готовности

### Функциональные
- [ ] Все 6 алгоритмов работают согласно спецификации
- [ ] Кеширование функционирует корректно
- [ ] Транзакции rollback'ятся при ошибках
- [ ] Ленивое создание дневных остатков работает
- [ ] Преобразование кодов обрабатывает edge cases

### Технические
- [ ] Coverage тестов > 90%
- [ ] Нет race conditions в concurrent операциях
- [ ] Memory usage оптимален
- [ ] Логирование информативно

### Проверочные
- [ ] Можно рассчитать остаток для любого предмета
- [ ] Создание дневного остатка не создает дубликатов
- [ ] Недостаточные остатки корректно обрабатываются
- [ ] Batch операции работают атомарно
- [ ] Кеш инвалидируется после операций

## Методы тестирования

### 1. Unit тесты
```bash
# Тесты алгоритмов
go test ./internal/service/...

# Тесты с мокированными зависимостями
go test -tags=unit ./internal/service/...
```

### 2. Integration тесты
```bash
# Тесты с реальной БД и Redis
go test -tags=integration ./internal/service/...

# Тесты concurrent операций
go test -race ./internal/service/...
```

### 3. Performance тесты
```bash
# Benchmark критичных операций
go test -bench=BenchmarkCalculateBalance ./internal/service/...
go test -bench=BenchmarkCreateOperations ./internal/service/...
```

### 4. Load тесты
```bash
# Симуляция высокой нагрузки
go test -tags=load ./internal/service/...
```

## Зависимости

### Входящие
- Repository слой (Задача 3)
- Модели данных (Задача 3)
- Конфигурация кеширования (Задача 2)

### Исходящие
- Готовые сервисы для HTTP handlers
- Бизнес-логика для API endpoints
- Переиспользуемые алгоритмы

## Go зависимости

```go
// Дополнительные зависимости
github.com/pkg/errors              // Error wrapping
github.com/google/uuid             // UUID operations
go.uber.org/zap                    // Structured logging
github.com/stretchr/testify        // Testing assertions
```

## Заметки по реализации

### Паттерны проектирования
- Strategy pattern для различных типов операций
- Chain of responsibility для валидации
- Observer pattern для cache invalidation

### Производительность
- Lazy loading для дневных остатков
- Batch операции для множественных изменений
- Optimistic locking для concurrent updates

### Надежность
- Graceful degradation при недоступности кеша
- Retry logic для временных сбоев БД
- Circuit breaker для внешних зависимостей

## Риски и ограничения

- **Риск**: Race conditions при concurrent операциях
  **Митигация**: Proper locking и тестирование с -race flag

- **Риск**: Inconsistent состояние кеша и БД
  **Митигация**: Transactional cache updates

- **Ограничение**: Производительность при большом количестве операций
  **Решение**: Batch processing и асинхронные обновления

## Метрики для мониторинга

- Время выполнения каждого алгоритма
- Hit/miss ratio кеша
- Количество созданных дневных остатков
- Частота invalidation кеша
- Количество rollback'ов транзакций
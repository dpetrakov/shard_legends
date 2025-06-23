# Задача 3: Модели данных и репозитории

## Описание

Создание Go структур для работы с данными, реализация repository паттерна для доступа к БД и базовых операций с классификаторами и предметами.

## Цели

1. Создать Go структуры для всех таблиц БД
2. Реализовать repository интерфейсы
3. Создать базовые CRUD операции
4. Реализовать кеширование справочных данных
5. Добавить валидацию данных

## Подзадачи

### 3.1. Модели данных
**Файл**: `internal/models/classifier.go`

**Структуры**:
```go
type Classifier struct {
    ID          uuid.UUID `json:"id" db:"id"`
    Code        string    `json:"code" db:"code"`
    Description *string   `json:"description" db:"description"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ClassifierItem struct {
    ID           uuid.UUID `json:"id" db:"id"`
    ClassifierID uuid.UUID `json:"classifier_id" db:"classifier_id"`
    Code         string    `json:"code" db:"code"`
    Description  *string   `json:"description" db:"description"`
    IsActive     bool      `json:"is_active" db:"is_active"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
```

**Файл**: `internal/models/item.go`

**Структуры**:
```go
type Item struct {
    ID                      uuid.UUID `json:"id" db:"id"`
    ItemClassID             uuid.UUID `json:"item_class_id" db:"item_class_id"`
    ItemTypeID              uuid.UUID `json:"item_type_id" db:"item_type_id"`
    QualityLevelsClassifierID uuid.UUID `json:"quality_levels_classifier_id" db:"quality_levels_classifier_id"`
    CollectionsClassifierID uuid.UUID `json:"collections_classifier_id" db:"collections_classifier_id"`
    CreatedAt               time.Time `json:"created_at" db:"created_at"`
    UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

type ItemImage struct {
    ItemID         uuid.UUID `json:"item_id" db:"item_id"`
    CollectionID   uuid.UUID `json:"collection_id" db:"collection_id"`
    QualityLevelID uuid.UUID `json:"quality_level_id" db:"quality_level_id"`
    ImageURL       string    `json:"image_url" db:"image_url"`
    IsActive       bool      `json:"is_active" db:"is_active"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
```

**Файл**: `internal/models/inventory.go`

**Структуры**:
```go
type DailyBalance struct {
    UserID         uuid.UUID `json:"user_id" db:"user_id"`
    SectionID      uuid.UUID `json:"section_id" db:"section_id"`
    ItemID         uuid.UUID `json:"item_id" db:"item_id"`
    CollectionID   uuid.UUID `json:"collection_id" db:"collection_id"`
    QualityLevelID uuid.UUID `json:"quality_level_id" db:"quality_level_id"`
    BalanceDate    time.Time `json:"balance_date" db:"balance_date"`
    Quantity       int64     `json:"quantity" db:"quantity"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Operation struct {
    ID              uuid.UUID  `json:"id" db:"id"`
    UserID          uuid.UUID  `json:"user_id" db:"user_id"`
    SectionID       uuid.UUID  `json:"section_id" db:"section_id"`
    ItemID          uuid.UUID  `json:"item_id" db:"item_id"`
    CollectionID    uuid.UUID  `json:"collection_id" db:"collection_id"`
    QualityLevelID  uuid.UUID  `json:"quality_level_id" db:"quality_level_id"`
    QuantityChange  int64      `json:"quantity_change" db:"quantity_change"`
    OperationTypeID uuid.UUID  `json:"operation_type_id" db:"operation_type_id"`
    OperationID     *uuid.UUID `json:"operation_id" db:"operation_id"`
    RecipeID        *uuid.UUID `json:"recipe_id" db:"recipe_id"`
    Comment         *string    `json:"comment" db:"comment"`
    CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}
```

### 3.2. DTO структуры для API
**Файл**: `internal/models/dto.go`

**Структуры**:
```go
// Входящие данные (с кодами)
type ItemQuantityRequest struct {
    ItemID       uuid.UUID `json:"item_id" validate:"required"`
    Collection   *string   `json:"collection,omitempty"`
    QualityLevel *string   `json:"quality_level,omitempty"`
    Quantity     int64     `json:"quantity" validate:"required,min=1"`
}

// Исходящие данные (с кодами)
type InventoryItemResponse struct {
    ItemID       uuid.UUID `json:"item_id"`
    ItemClass    string    `json:"item_class"`
    ItemType     string    `json:"item_type"`
    Collection   *string   `json:"collection,omitempty"`
    QualityLevel *string   `json:"quality_level,omitempty"`
    Quantity     int64     `json:"quantity"`
}

// Внутренняя структура для расчетов
type ItemBalance struct {
    UserID         uuid.UUID
    SectionID      uuid.UUID
    ItemID         uuid.UUID
    CollectionID   uuid.UUID
    QualityLevelID uuid.UUID
    CurrentBalance int64
}
```

### 3.3. Repository интерфейсы
**Файл**: `internal/repository/interfaces.go`

**Интерфейсы**:
```go
type ClassifierRepository interface {
    GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error)
    GetClassifierItems(ctx context.Context, classifierID uuid.UUID) ([]*models.ClassifierItem, error)
    GetClassifierItemByCode(ctx context.Context, classifierID uuid.UUID, code string) (*models.ClassifierItem, error)
    GetAllClassifiersWithItems(ctx context.Context) (map[string][]*models.ClassifierItem, error)
}

type ItemRepository interface {
    GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
    GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error)
    GetItemImage(ctx context.Context, itemID, collectionID, qualityLevelID uuid.UUID) (*models.ItemImage, error)
}

type InventoryRepository interface {
    GetDailyBalance(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, date time.Time) (*models.DailyBalance, error)
    CreateDailyBalance(ctx context.Context, balance *models.DailyBalance) error
    GetOperations(ctx context.Context, userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID, fromDate time.Time) ([]*models.Operation, error)
    CreateOperation(ctx context.Context, operation *models.Operation) error
    CreateOperations(ctx context.Context, operations []*models.Operation) error
    GetOperationsByExternalID(ctx context.Context, operationID uuid.UUID) ([]*models.Operation, error)
}
```

### 3.4. Реализация ClassifierRepository
**Файл**: `internal/repository/classifier_repo.go`

**Содержание**:
- CRUD операции с классификаторами
- Кеширование справочных данных в Redis
- Методы для преобразования код ↔ UUID
- Batch загрузка всех классификаторов

**Ключевые методы**:
```go
func (r *classifierRepo) GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error)
func (r *classifierRepo) GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error)
func (r *classifierRepo) InvalidateCache(ctx context.Context, classifierCode string) error
```

### 3.5. Реализация ItemRepository
**Файл**: `internal/repository/item_repo.go`

**Содержание**:
- CRUD операции с предметами
- Получение информации о доступных коллекциях/качестве
- Работа с изображениями предметов
- Joining с классификаторами

### 3.6. Реализация InventoryRepository
**Файл**: `internal/repository/inventory_repo.go`

**Содержание**:
- Операции с дневными остатками
- CRUD операции с операциями инвентаря
- Batch операции для производительности
- Транзакционные методы

### 3.7. Валидация данных
**Файл**: `internal/models/validation.go`

**Содержание**:
- Валидаторы для всех структур
- Custom validation rules
- Error handling для валидации

**Примеры валидаторов**:
```go
func ValidateItemQuantityRequest(req *ItemQuantityRequest) error
func ValidateOperation(op *Operation) error
func ValidateUUID(id uuid.UUID) error
```

### 3.8. Тесты репозиториев
**Файлы**: `internal/repository/*_test.go`

**Содержание**:
- Unit тесты для всех repository методов
- Integration тесты с тестовой БД
- Mock'и для зависимостей
- Тестовые данные и fixtures

## Критерии готовности

### Функциональные
- [ ] Все модели корректно маппятся на БД таблицы
- [ ] Repository методы работают с реальной БД
- [ ] Кеширование классификаторов функционирует
- [ ] Валидация отклоняет некорректные данные
- [ ] Преобразование код ↔ UUID работает

### Технические
- [ ] Все тесты проходят успешно
- [ ] Coverage тестов > 80%
- [ ] Нет memory leaks в длительных операциях
- [ ] Производительность запросов приемлемая

### Проверочные
- [ ] Можно получить классификатор по коду
- [ ] Можно получить все элементы классификатора
- [ ] Можно создать операцию в БД
- [ ] Можно получить дневной остаток
- [ ] Кеш инвалидируется корректно

## Методы тестирования

### 1. Unit тесты
```bash
# Тесты моделей
go test ./internal/models/...

# Тесты репозиториев
go test ./internal/repository/...

# Coverage отчет
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2. Integration тесты
```bash
# Тесты с реальной БД
TEST_DB_URL=postgres://... go test -tags=integration ./...

# Тесты с Redis
REDIS_URL=redis://... go test -tags=integration ./...
```

### 3. Benchmark тесты
```bash
# Производительность repository операций
go test -bench=. ./internal/repository/...

# Производительность кеширования
go test -bench=BenchmarkCache ./...
```

## Зависимости

### Входящие
- Готовая БД структура (Задача 1)
- Базовое приложение (Задача 2)
- PostgreSQL и Redis подключения

### Исходящие
- Готовые модели для service слоя
- Repository интерфейсы для mocking
- Валидация для HTTP handlers

## Go зависимости

```go
// Дополнительные зависимости
github.com/jmoiron/sqlx              // SQL extensions
github.com/go-playground/validator/v10 // Validation
github.com/go-redis/cache/v8          // Redis caching
github.com/golang/mock                // Mocking
```

## Заметки по реализации

### Паттерны проектирования
- Repository pattern для абстракции БД
- DTO pattern для API boundaries
- Builder pattern для сложных запросов

### Производительность
- Batch операции где возможно
- Prepared statements для частых запросов
- Connection pooling оптимизация

### Кеширование
- TTL = 24 часа для классификаторов
- Namespace ключей: `inventory:classifier:{code}`
- Graceful degradation при недоступности Redis

## Риски и ограничения

- **Риск**: Несоответствие моделей и БД схемы
  **Митигация**: Автогенерация из schema или строгий review

- **Риск**: N+1 проблемы в запросах
  **Митигация**: Eager loading и batch запросы

- **Ограничение**: Зависимость от Redis для кеша
  **Решение**: Fallback на БД при недоступности кеша
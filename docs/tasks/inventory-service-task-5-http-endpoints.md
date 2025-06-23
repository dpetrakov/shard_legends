# Задача 5: HTTP API эндпоинты

## Описание

Реализация всех HTTP эндпоинтов согласно OpenAPI спецификации. Включает публичные, внутренние и административные эндпоинты с полной валидацией, аутентификацией и обработкой ошибок.

## Цели

1. Реализовать все эндпоинты из OpenAPI спецификации
2. Добавить JWT аутентификацию и авторизацию
3. Реализовать валидацию входных данных
4. Создать middleware для логирования и метрик
5. Обеспечить корректную обработку ошибок

## Подзадачи

### 5.1. HTTP handlers структура
**Файл**: `internal/handlers/inventory.go`

**Структура**:
```go
type InventoryHandler struct {
    inventoryService service.InventoryService
    classifierService service.ClassifierService
    logger           *zap.Logger
    validator        *validator.Validate
}

func NewInventoryHandler(
    inventoryService service.InventoryService,
    classifierService service.ClassifierService,
    logger *zap.Logger,
) *InventoryHandler
```

### 5.2. Публичные эндпоинты
**Файл**: `internal/handlers/public.go`

#### GET /inventory
```go
func (h *InventoryHandler) GetUserInventory(c *gin.Context) {
    // 1. Извлечь user_id из JWT токена
    // 2. Валидировать query параметры (section)
    // 3. Получить список уникальных комбинаций предметов
    // 4. Для каждой комбинации вызвать CalculateCurrentBalance
    // 5. Преобразовать UUID в коды через ConvertClassifierCodes
    // 6. Отфильтровать предметы с нулевым остатком
    // 7. Вернуть InventoryResponse
}
```

**Request**:
```
GET /inventory?section=main
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "items": [
    {
      "item_id": "550e8400-e29b-41d4-a716-446655440000",
      "item_class": "tools",
      "item_type": "axe",
      "collection": "winter_2024",
      "quality_level": "wooden",
      "quantity": 5
    }
  ]
}
```

#### GET /inventory/items/{item_id}
```go
func (h *InventoryHandler) GetItemInfo(c *gin.Context) {
    // 1. Валидировать item_id UUID
    // 2. Получить информацию о предмете из ItemRepository
    // 3. Получить доступные коллекции и уровни качества
    // 4. Преобразовать UUID в коды
    // 5. Вернуть ItemInfoResponse
}
```

### 5.3. Внутренние эндпоинты
**Файл**: `internal/handlers/internal.go`

#### POST /inventory/reserve
```go
func (h *InventoryHandler) ReserveItems(c *gin.Context) {
    // 1. Парсинг ReserveItemsRequest
    // 2. Валидация входных данных
    // 3. Преобразование кодов в UUID
    // 4. Вызов CheckSufficientBalance
    // 5. Подготовка операций резервирования (2 на предмет)
    // 6. Вызов CreateOperationsInTransaction
    // 7. Возврат OperationResponse
}
```

**Request**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "operation_id": "550e8400-e29b-41d4-a716-446655440001",
  "items": [
    {
      "item_id": "550e8400-e29b-41d4-a716-446655440002",
      "collection": "winter_2024",
      "quality_level": "wooden",
      "quantity": 10
    }
  ]
}
```

#### POST /inventory/return-reserve
```go
func (h *InventoryHandler) ReturnReservedItems(c *gin.Context) {
    // 1. Парсинг ReturnReserveRequest
    // 2. Найти операции резервирования по operation_id
    // 3. Создать операции возврата (обратные резервированию)
    // 4. Вызов CreateOperationsInTransaction
    // 5. Сброс кеша пользователя
    // 6. Возврат OperationResponse
}
```

#### POST /inventory/consume-reserve
```go
func (h *InventoryHandler) ConsumeReservedItems(c *gin.Context) {
    // 1. Парсинг ConsumeReserveRequest
    // 2. Найти операции резервирования по operation_id
    // 3. Создать операции потребления (списание из factory)
    // 4. Вызов CreateOperationsInTransaction
    // 5. Сброс кеша пользователя
    // 6. Возврат OperationResponse
}
```

#### POST /inventory/add-items
```go
func (h *InventoryHandler) AddItems(c *gin.Context) {
    // 1. Парсинг AddItemsRequest
    // 2. Валидация данных
    // 3. Преобразование кодов в UUID
    // 4. Подготовка операций добавления
    // 5. Вызов CreateOperationsInTransaction
    // 6. Возврат OperationResponse
}
```

### 5.4. Административные эндпоинты
**Файл**: `internal/handlers/admin.go`

#### POST /admin/inventory/adjust
```go
func (h *InventoryHandler) AdjustInventory(c *gin.Context) {
    // 1. Проверка административных прав доступа
    // 2. Парсинг AdjustInventoryRequest
    // 3. Валидация данных и прав
    // 4. Для отрицательных изменений: CheckSufficientBalance
    // 5. Подготовка операций корректировки
    // 6. Вызов CreateOperationsInTransaction
    // 7. Запись в аудит-лог
    // 8. Расчет итоговых остатков
    // 9. Возврат AdjustInventoryResponse
}
```

### 5.5. Middleware
**Файл**: `internal/middleware/auth.go`

#### JWT Authentication
```go
func JWTAuthMiddleware(authService AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Извлечь JWT токен из Authorization header
        // 2. Валидировать токен через Auth Service
        // 3. Извлечь user_id и сохранить в context
        // 4. Продолжить выполнение или вернуть 401
    }
}
```

#### Admin Authorization
```go
func AdminAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Проверить роль пользователя
        // 2. Убедиться в наличии admin прав
        // 3. Продолжить или вернуть 403
    }
}
```

**Файл**: `internal/middleware/logging.go`

#### Request Logging
```go
func RequestLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Логировать входящий запрос
        // 2. Измерить время выполнения
        // 3. Логировать ответ с метриками
    }
}
```

**Файл**: `internal/middleware/metrics.go`

#### Metrics Collection
```go
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Начать измерение времени
        // 2. Инкрементировать счетчик запросов
        // 3. Записать метрики после обработки
    }
}
```

### 5.6. Error handling
**Файл**: `internal/handlers/errors.go`

**Структуры ошибок**:
```go
type ErrorResponse struct {
    Error   string      `json:"error"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type InsufficientItemsError struct {
    Error        string                    `json:"error"`
    Message      string                    `json:"message"`
    MissingItems []InsufficientItemDetail `json:"missing_items"`
}

type InsufficientItemDetail struct {
    ItemID       uuid.UUID `json:"item_id"`
    Collection   *string   `json:"collection,omitempty"`
    QualityLevel *string   `json:"quality_level,omitempty"`
    Required     int64     `json:"required"`
    Available    int64     `json:"available"`
}
```

**Error handlers**:
```go
func HandleValidationError(c *gin.Context, err error)
func HandleBusinessLogicError(c *gin.Context, err error)
func HandleInternalError(c *gin.Context, err error)
func HandleInsufficientItemsError(c *gin.Context, missing []InsufficientItemDetail)
```

### 5.7. Валидация
**Файл**: `internal/handlers/validation.go`

**Валидаторы**:
```go
func ValidateUUID(fl validator.FieldLevel) bool
func ValidateInventorySection(fl validator.FieldLevel) bool
func ValidateOperationType(fl validator.FieldLevel) bool
func ValidateQuantity(fl validator.FieldLevel) bool
```

**Request валидация**:
```go
func (h *InventoryHandler) validateRequest(req interface{}) error {
    if err := h.validator.Struct(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

### 5.8. Routing setup
**Файл**: `internal/handlers/router.go`

```go
func SetupRoutes(
    r *gin.Engine,
    inventoryHandler *InventoryHandler,
    authService AuthService,
    metrics *Metrics,
    logger *zap.Logger,
) {
    // Middleware
    r.Use(RequestLoggingMiddleware(logger))
    r.Use(MetricsMiddleware(metrics))
    r.Use(gin.Recovery())

    // Public routes with JWT auth
    public := r.Group("/inventory")
    public.Use(JWTAuthMiddleware(authService))
    {
        public.GET("", inventoryHandler.GetUserInventory)
        public.GET("/items/:item_id", inventoryHandler.GetItemInfo)
    }

    // Internal routes (service-to-service)
    internal := r.Group("/inventory")
    // internal.Use(ServiceAuthMiddleware()) // если нужно
    {
        internal.POST("/reserve", inventoryHandler.ReserveItems)
        internal.POST("/return-reserve", inventoryHandler.ReturnReservedItems)
        internal.POST("/consume-reserve", inventoryHandler.ConsumeReservedItems)
        internal.POST("/add-items", inventoryHandler.AddItems)
    }

    // Admin routes
    admin := r.Group("/admin/inventory")
    admin.Use(JWTAuthMiddleware(authService))
    admin.Use(AdminAuthMiddleware())
    {
        admin.POST("/adjust", inventoryHandler.AdjustInventory)
    }
}
```

### 5.9. Тесты HTTP handlers
**Файлы**: `internal/handlers/*_test.go`

**Типы тестов**:
- Unit тесты для каждого handler'а
- Integration тесты с полным стеком
- Тесты авторизации и аутентификации
- Тесты валидации входных данных
- Тесты error handling

**Тестовые сценарии**:
```go
func TestGetUserInventory_Success(t *testing.T)
func TestGetUserInventory_InvalidJWT(t *testing.T)
func TestReserveItems_InsufficientBalance(t *testing.T)
func TestAdjustInventory_AdminOnly(t *testing.T)
func TestAddItems_ValidationError(t *testing.T)
```

## Критерии готовности

### Функциональные
- [ ] Все эндпоинты работают согласно OpenAPI spec
- [ ] JWT аутентификация функционирует
- [ ] Валидация отклоняет некорректные данные
- [ ] Ошибки возвращаются в правильном формате
- [ ] Авторизация admin эндпоинтов работает

### Технические
- [ ] Тесты покрывают >85% кода handlers
- [ ] Производительность эндпоинтов приемлемая
- [ ] Логирование информативно
- [ ] Метрики собираются корректно

### Проверочные
- [ ] curl запросы к эндпоинтам работают
- [ ] Swagger UI отображает правильную документацию
- [ ] Postman коллекция выполняется успешно
- [ ] Load тесты показывают стабильность

## Методы тестирования

### 1. Unit тесты
```bash
# Тесты handlers
go test ./internal/handlers/...

# Тесты middleware
go test ./internal/middleware/...
```

### 2. Integration тесты
```bash
# HTTP тесты с тестовой БД
go test -tags=integration ./internal/handlers/...

# Тесты с реальным JWT сервисом
go test -tags=auth ./internal/handlers/...
```

### 3. API тесты
```bash
# Postman/Newman тесты
newman run inventory-service-collection.json

# curl скрипты
./scripts/test-api.sh
```

### 4. Load тесты
```bash
# Artillery/k6 нагрузочные тесты
k6 run loadtest.js
```

## Зависимости

### Входящие
- Service слой (Задача 4)
- Модели и DTO (Задача 3)
- JWT аутентификация (Auth Service)
- OpenAPI спецификация

### Исходящие
- Готовый HTTP API
- Документированные эндпоинты
- Метрики для мониторинга

## Go зависимости

```go
// HTTP framework и middleware
github.com/gin-gonic/gin               // HTTP framework
github.com/gin-contrib/cors            // CORS middleware
github.com/gin-contrib/requestid       // Request ID
github.com/dgrijalva/jwt-go            // JWT parsing

// Валидация
github.com/go-playground/validator/v10 // Validation

// Тестирование
github.com/stretchr/testify            // Testing
github.com/gavv/httpexpect/v2          // HTTP testing
```

## Заметки по реализации

### RESTful принципы
- Правильные HTTP статус коды
- Consistent error responses
- Resource-oriented URLs
- Idempotent operations где возможно

### Безопасность
- Input validation для всех параметров
- SQL injection protection
- Rate limiting (в будущих версиях)
- Audit logging для admin операций

### Производительность
- Connection pooling
- Response caching где appropriate
- Pagination для больших списков
- Compression для больших ответов

## Риски и ограничения

- **Риск**: JWT token hijacking
  **Митигация**: Short-lived tokens, refresh mechanism

- **Риск**: DoS через massive requests
  **Митигация**: Rate limiting, request size limits

- **Ограничение**: Нет real-time уведомлений
  **Решение**: Polling или WebSocket в будущих версиях

## Метрики для API

- Request latency по эндпоинтам
- Request count по статус кодам
- Error rate по типам ошибок
- Authentication failures
- Authorization failures
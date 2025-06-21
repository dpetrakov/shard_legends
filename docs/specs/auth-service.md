# Спецификация сервиса авторизации (Auth Service)

## Обзор

Сервис авторизации обеспечивает аутентификацию и авторизацию пользователей Telegram Mini App для Shard Legends: Clan Wars. Сервис интегрирован в существующую микросервисную архитектуру через API Gateway и использует данные Telegram Web App для валидации пользователей.

## Архитектурная интеграция

### Позиционирование в системе
- **Тип**: Микросервис на Golang
- **Порт**: 8080 (внутренний)
- **Доступ**: Через API Gateway (nginx) на маршруте `/api/auth`
- **Зависимости**: PostgreSQL, Redis, API Gateway

### Интеграция с существующими компонентами
```mermaid
graph TB
    TMA[Telegram Mini App] --> Gateway[API Gateway nginx:9000]
    Gateway --> AuthService[Auth Service:8080]
    AuthService --> DB[(PostgreSQL 17)]
    AuthService --> Cache[(Redis 8.0.2)]
    
    subgraph "Проверка токенов другими сервисами"
        GameAPI[Game API] --> Cache
        UserService[User Service] --> Cache
        OtherServices[Other Services] --> Cache
    end
    
    AuthService -.->|JWT signing| OtherServices
```

## Функциональные требования

### 1. Аутентификация через Telegram Web App

#### Вход через эндпоинт `/auth`
**Метод**: `POST /auth`

**Заголовки запроса**:
- `X-Telegram-Init-Data`: Полные данные инициализации от Telegram Web App (включает user, auth_date, hash и другие поля)

**Структура Telegram Web App initData**:
Данные передаются как URL-encoded строка и содержат следующие поля:
- `user`: JSON-объект с данными пользователя (обязательное поле)
- `auth_date`: Unix timestamp открытия формы (обязательное поле)
- `hash`: HMAC-SHA256 хеш для проверки целостности (обязательное поле)
- `query_id`: Уникальный идентификатор сессии (опционально)
- `start_param`: Параметр запуска из attachment menu (опционально)
- `chat`: Данные чата, если запущено из чата (опционально)
- `signature`: Ed25519 подпись для третьих лиц (опционально, новое)

**Структура объекта User в initData**:
```json
{
  "id": 123456789,
  "is_bot": false,
  "first_name": "John",
  "last_name": "Doe",
  "username": "john_doe",
  "language_code": "en",
  "is_premium": true,
  "photo_url": "https://t.me/i/userpic/320/abc123.jpg"
}
```

**Алгоритм валидации Telegram данных (HMAC-SHA256)**:
1. **Извлечение данных из заголовка**:
   - Парсинг URL-encoded строки `X-Telegram-Init-Data`
   - Извлечение `hash` и остальных параметров
   
2. **Создание data-check-string**:
   - Сортировка всех полей (кроме `hash`) в алфавитном порядке
   - Формат: `key=value` для каждого поля
   - Разделение полей символом новой строки (`\n`)
   - Пример: `auth_date=1672531200\nquery_id=ABC123\nuser={"id":123456789,"first_name":"John"}`

3. **Генерация секретного ключа**:
   ```
   secret_key = HMAC-SHA256(bot_token, "WebAppData")
   ```

4. **Проверка подписи**:
   ```
   calculated_hash = HMAC-SHA256(data_check_string, secret_key)
   if (hex(calculated_hash) == received_hash) {
       // Данные валидны
   }
   ```

5. **Дополнительные проверки безопасности**:
   - `auth_date` не старше 24 часов (защита от replay-атак)
   - Валидация JSON структуры объекта `user`
   - Проверка обязательных полей (`id`, `first_name`)

**Логика аутентификации**:
1. **Проверка обязательных заголовков**
   - Отсутствует `X-Telegram-Init-Data` → HTTP 400 Bad Request
   
2. **Валидация Telegram данных**
   - Парсинг initData и извлечение полей
   - HMAC-SHA256 валидация с bot token
   - Проверка актуальности `auth_date` (не старше 24 часов)
   - Валидация структуры и обязательных полей объекта user
   
3. **Обработка пользователя**
   - Если пользователь существует → генерация нового JWT токена
   - Если пользователь новый → регистрация + генерация JWT токена
   - Отзыв всех существующих активных токенов пользователя

**Edge Cases и обработка ошибок**:
- Пользователь без `username` (поле может отсутствовать)
- Пользователь без `last_name` (поле может отсутствовать)
- Пользователь без `photo_url` (поле может отсутствовать)
- Некорректный формат JSON в поле `user`
- Просроченный `auth_date` (старше 24 часов)
- Невалидная подпись HMAC
- Отсутствие обязательных полей `id` или `first_name`

#### Формат ответа
**Успешная авторизация** (HTTP 200):
```json
{
  "success": true,
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-12-22T10:30:00Z",
  "user": {
    "id": "uuid-internal-user-id",
    "telegram_id": 123456789,
    "username": "user_telegram",
    "first_name": "John",
    "last_name": "Doe",
    "is_new_user": false
  }
}
```

**Ошибка авторизации** (HTTP 401):
```json
{
  "success": false,
  "error": "invalid_telegram_data",
  "message": "Telegram authentication failed"
}
```

### 2. JWT токены

#### Структура JWT токена
```json
{
  "iss": "shard-legends-auth",
  "sub": "uuid-internal-user-id",
  "telegram_id": 123456789,
  "iat": 1703243400,
  "exp": 1703329800,
  "jti": "uuid-token-id"
}
```

#### Характеристики токена
- **Алгоритм**: RS256 (RSA подпись)
- **Время жизни**: 24 часа
- **Приватный ключ**: Известен только Auth Service
- **Публичный ключ**: Доступен другим сервисам для валидации
- **JTI (JWT ID)**: UUID для уникальной идентификации каждого токена

### 3. Управление токенами в Redis

#### Активные токены (для валидации)
**Ключ**: `active:{token_jti}`
**Значение**: JSON с метаданными токена
**TTL**: 24 часа (время жизни токена)

```json
{
  "user_id": "uuid-internal-user-id",
  "telegram_id": 123456789,
  "issued_at": "2024-12-21T10:30:00Z",
  "expires_at": "2024-12-22T10:30:00Z"
}
```

#### Отозванные токены (для быстрой проверки)
**Ключ**: `revoked:{token_jti}`
**Значение**: Причина отзыва и timestamp
**TTL**: 24 часа (до истечения срока действия)

```json
{
  "reason": "user_reauth",
  "revoked_at": "2024-12-21T11:00:00Z",
  "user_id": "uuid-internal-user-id"
}
```

#### Пользовательские токены (Redis SET для быстрого поиска)
**Ключ**: `user_tokens:{user_id}`
**Тип**: Redis SET
**Значения**: token_jti (каждый jti как элемент множества)
**TTL**: Без TTL (управляется вручную при отзыве)

#### Алгоритм отзыва всех токенов пользователя при новой авторизации:

```redis
# 1. Получить все активные токены пользователя
SMEMBERS user_tokens:{user_id}
# Результат: ["jti1", "jti2", "jti3"]

# 2. Переместить все токены в отозванные (pipeline/transaction)
MULTI
  # Для каждого jti из результата шага 1:
  DEL active:{jti1}
  DEL active:{jti2}  
  DEL active:{jti3}
  SET revoked:{jti1} "{"reason":"user_reauth","revoked_at":"2024-12-21T11:00:00Z"}" EX 86400
  SET revoked:{jti2} "{"reason":"user_reauth","revoked_at":"2024-12-21T11:00:00Z"}" EX 86400
  SET revoked:{jti3} "{"reason":"user_reauth","revoked_at":"2024-12-21T11:00:00Z"}" EX 86400
  DEL user_tokens:{user_id}
EXEC

# 3. Добавить новый токен
MULTI
  SET active:{new_jti} "{...token_metadata...}" EX 86400
  SADD user_tokens:{user_id} {new_jti}
EXEC
```

#### Операции с токенами

**Выдача нового токена**:
```redis
MULTI
  SET active:{jti} "{...metadata...}" EX 86400
  SADD user_tokens:{user_id} {jti}
EXEC
```

**Проверка токена (используется другими сервисами)**:
```redis
# 1. Проверить, не отозван ли токен
EXISTS revoked:{jti}
# Если EXISTS возвращает 1 → токен отозван

# 2. Опционально: получить метаданные активного токена
GET active:{jti}
```

**Отзыв конкретного токена**:
```redis
MULTI
  DEL active:{jti}
  SET revoked:{jti} "{"reason":"manual_revoke"}" EX 86400
  SREM user_tokens:{user_id} {jti}
EXEC
```

#### Очистка просроченных токенов

**Проблема**: `user_tokens:{user_id}` SET не имеет TTL и может накапливать просроченные JTI, которые уже удалились из `active:{jti}` и `revoked:{jti}` по TTL.

**Решение**: Встроенная очистка в auth-service каждый час

**Алгоритм очистки** (запускается по cron каждый час):
```redis
# 1. Получить всех пользователей, у которых есть токены
SCAN 0 MATCH user_tokens:* COUNT 100

# 2. Для каждого user_tokens:{user_id}:
SMEMBERS user_tokens:{user_id}
# Результат: ["jti1", "jti2", "jti3", "jti4"]

# 3. Проверить, какие токены еще существуют
MGET active:{jti1} active:{jti2} active:{jti3} active:{jti4}
# Результат: ["{metadata}", null, "{metadata}", null]
# null означает, что токен просрочен и удален Redis'ом

# 4. Удалить просроченные JTI из SET'а пользователя
MULTI
  SREM user_tokens:{user_id} {jti2} {jti4}  # Удаляем просроченные
EXEC

# 5. Если SET стал пустым - удалить его полностью
SCARD user_tokens:{user_id}
# Если результат 0:
DEL user_tokens:{user_id}
```

**Реализация в коде**:
```go
// CleanupExpiredTokens очищает просроченные токены из пользовательских SET'ов
func (r *RedisTokenStorage) CleanupExpiredTokens(ctx context.Context) error {
    iter := r.client.Scan(ctx, 0, "user_tokens:*", 100).Iterator()
    
    for iter.Next(ctx) {
        userTokensKey := iter.Val()
        
        // Получить все JTI пользователя
        jtis, err := r.client.SMembers(ctx, userTokensKey).Result()
        if err != nil {
            continue
        }
        
        if len(jtis) == 0 {
            r.client.Del(ctx, userTokensKey)
            continue
        }
        
        // Проверить, какие токены еще активны
        activeKeys := make([]string, len(jtis))
        for i, jti := range jtis {
            activeKeys[i] = "active:" + jti
        }
        
        results, err := r.client.MGet(ctx, activeKeys...).Result()
        if err != nil {
            continue
        }
        
        // Найти просроченные токены
        expiredJTIs := make([]interface{}, 0)
        for i, result := range results {
            if result == nil { // токен просрочен
                expiredJTIs = append(expiredJTIs, jtis[i])
            }
        }
        
        // Удалить просроченные JTI из SET'а
        if len(expiredJTIs) > 0 {
            r.client.SRem(ctx, userTokensKey, expiredJTIs...)
        }
        
        // Проверить, остались ли токены
        remaining := r.client.SCard(ctx, userTokensKey).Val()
        if remaining == 0 {
            r.client.Del(ctx, userTokensKey)
        }
    }
    
    return iter.Err()
}
```

**Настройка cron очистки в auth-service**:
```go
// В main.go или отдельном пакете scheduler
func startTokenCleanupScheduler(tokenStorage TokenStorage) {
    ticker := time.NewTicker(1 * time.Hour)
    
    go func() {
        for {
            select {
            case <-ticker.C:
                ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
                
                start := time.Now()
                err := tokenStorage.CleanupExpiredTokens(ctx)
                duration := time.Since(start)
                
                if err != nil {
                    log.Error("Token cleanup failed", "error", err, "duration", duration)
                } else {
                    log.Info("Token cleanup completed", "duration", duration)
                }
                
                cancel()
            }
        }
    }()
}
```

**Преимущества встроенной очистки в auth-service**:

1. **Простота архитектуры** - нет дополнительного сервиса
2. **Низкие накладные расходы** - работает только с пользователями, у которых есть токены
3. **Эффективность** - использует Redis SCAN для постепенной обработки
4. **Безопасность** - только auth-service имеет доступ к токенам
5. **Мониторинг** - легко добавить метрики очистки

**Альтернативные варианты** (не рекомендуются):

1. **Отдельный сервис очистки**: Избыточная сложность для простой задачи
2. **Redis Keyspace Notifications**: Не подходит, так как нужно очищать SET'ы, а не отдельные ключи
3. **TTL на user_tokens**: Не подходит, так как активные пользователи должны сохранять свои SET'ы

#### Преимущества новой схемы:

1. **O(1) проверка отзыва** токена через `EXISTS revoked:{jti}`
2. **O(1) получение всех токенов пользователя** через `SMEMBERS user_tokens:{user_id}`
3. **Атомарный отзыв всех токенов** пользователя через Redis MULTI/EXEC
4. **Эффективное управление памятью** - автоматическое удаление через TTL
5. **Простое добавление/удаление токенов** из пользовательского множества

### 4. Регистрация новых пользователей

#### Алгоритм регистрации
1. **Извлечение данных из Telegram Web App initData**:
   - `id` (telegram_id) - BIGINT, обязательное поле
   - `first_name` - VARCHAR(100), обязательное поле
   - `last_name` - VARCHAR(100), опциональное поле (может отсутствовать)
   - `username` - VARCHAR(100), опциональное поле (может отсутствовать)
   - `language_code` - VARCHAR(10), опциональное поле (например: "en", "ru", "es")
   - `is_premium` - BOOLEAN, статус Telegram Premium (по умолчанию false)
   - `photo_url` - TEXT, ссылка на фото профиля (опциональное, может отсутствовать)

2. **Создание записи пользователя**:
   - Генерация внутреннего UUID как первичного ключа
   - Запись в таблицу `users` PostgreSQL со всеми доступными данными
   - Установка `created_at` и `last_login_at` в текущее время
   - Установка `is_active = true` по умолчанию

3. **Генерация JWT токена** для нового пользователя

#### Примеры данных пользователей для разных случаев

**Полные данные пользователя с Premium**:
```json
{
  "id": 123456789,
  "is_bot": false,
  "first_name": "John",
  "last_name": "Doe", 
  "username": "john_doe",
  "language_code": "en",
  "is_premium": true,
  "photo_url": "https://t.me/i/userpic/320/abc123.jpg"
}
```

**Минимальные данные пользователя** (без username, last_name, photo):
```json
{
  "id": 987654321,
  "is_bot": false,
  "first_name": "Maria",
  "language_code": "ru",
  "is_premium": false
}
```

**Пользователь без языка** (старые аккаунты):
```json
{
  "id": 555666777,
  "is_bot": false,
  "first_name": "Ahmed",
  "last_name": "Al-Rashid",
  "username": "ahmed_ar",
  "is_premium": false
}
```

#### Обработка edge cases при регистрации

1. **Отсутствующий username**:
   - Сохранить как NULL в базе данных
   - Не влияет на процесс авторизации
   - Отображение в интерфейсе по `first_name + last_name`

2. **Отсутствующий last_name**:
   - Сохранить как NULL в базе данных
   - Отображение только по `first_name`

3. **Отсутствующий language_code**:
   - Сохранить как NULL, использовать "en" по умолчанию в приложении
   - Не критично для функциональности авторизации

4. **Отсутствующий photo_url**:
   - Сохранить как NULL
   - Использовать дефолтный аватар в интерфейсе

5. **Очень длинные имена**:
   - Обрезка до лимитов VARCHAR при сохранении
   - Логирование предупреждений о превышении лимитов

#### Структура таблицы users (PostgreSQL) - обновленная
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    language_code VARCHAR(10),
    is_premium BOOLEAN DEFAULT FALSE,
    photo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    CONSTRAINT users_telegram_id_positive CHECK (telegram_id > 0),
    CONSTRAINT users_first_name_not_empty CHECK (length(trim(first_name)) > 0)
);

-- Индексы для производительности
CREATE INDEX idx_users_telegram_id ON users(telegram_id);
CREATE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;

-- Комментарии для документации
COMMENT ON TABLE users IS 'Базовые данные пользователей из Telegram Web App для авторизации';
COMMENT ON COLUMN users.telegram_id IS 'Уникальный ID пользователя в Telegram';
COMMENT ON COLUMN users.first_name IS 'Имя пользователя в Telegram (обязательное поле)';
COMMENT ON COLUMN users.is_premium IS 'Статус Telegram Premium пользователя';
COMMENT ON COLUMN users.photo_url IS 'URL фото профиля пользователя в Telegram';
```

### 5. Валидация токенов другими сервисами

#### Процесс проверки токена
1. **Валидация JWT подписи** (используя публичный ключ)
2. **Проверка времени действия** токена
3. **Проверка в Redis** на отзыв (`revoked_token:{token_jti}`)
4. **Извлечение данных пользователя** из токена

#### Middleware для других сервисов
```go
// Пример middleware для проверки авторизации (обновленная версия)
func AuthMiddleware(publicKey *rsa.PublicKey, redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "missing_token"})
            c.Abort()
            return
        }
        
        // Удаление префикса "Bearer "
        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = tokenString[7:]
        }
        
        // Валидация JWT
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return publicKey, nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "invalid_token"})
            c.Abort()
            return
        }
        
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(401, gin.H{"error": "invalid_claims"})
            c.Abort()
            return
        }
        
        // ОБНОВЛЕНО: Проверка отзыва токена по новой схеме
        jti := claims["jti"].(string)
        isRevoked := redisClient.Exists(ctx, "revoked:"+jti).Val()
        if isRevoked > 0 {
            c.JSON(401, gin.H{"error": "token_revoked"})
            c.Abort()
            return
        }
        
        // Сохранение данных пользователя в контексте
        c.Set("user_id", claims["sub"].(string))
        c.Set("telegram_id", int64(claims["telegram_id"].(float64)))
        c.Next()
    }
}
```

### Интерфейс TokenStorage для Redis

```go
// TokenStorage определяет интерфейс для управления JWT токенами в Redis
type TokenStorage interface {
    // Выдача нового токена пользователю (с отзывом всех предыдущих)
    IssueNewToken(ctx context.Context, userID, jti string, metadata TokenMetadata, ttl time.Duration) error
    
    // Проверка, отозван ли токен
    IsTokenRevoked(ctx context.Context, jti string) (bool, error)
    
    // Получение метаданных активного токена
    GetActiveToken(ctx context.Context, jti string) (*TokenMetadata, error)
    
    // Отзыв конкретного токена
    RevokeToken(ctx context.Context, userID, jti string, reason string) error
    
    // Получение всех активных токенов пользователя
    GetUserTokens(ctx context.Context, userID string) ([]string, error)
    
    // Отзыв всех токенов пользователя (используется при новой авторизации)
    RevokeAllUserTokens(ctx context.Context, userID string, reason string) error
    
    // Очистка просроченных токенов из пользовательских SET'ов (запускается по cron)
    CleanupExpiredTokens(ctx context.Context) error
}

type TokenMetadata struct {
    UserID      string    `json:"user_id"`
    TelegramID  int64     `json:"telegram_id"`
    IssuedAt    time.Time `json:"issued_at"`
    ExpiresAt   time.Time `json:"expires_at"`
}

// Реализация интерфейса для Redis
type RedisTokenStorage struct {
    client *redis.Client
}

func (r *RedisTokenStorage) IssueNewToken(ctx context.Context, userID, jti string, metadata TokenMetadata, ttl time.Duration) error {
    // Атомарная операция: отозвать все старые токены + выдать новый
    pipe := r.client.TxPipeline()
    
    // 1. Получить все активные токены пользователя
    userTokens := pipe.SMembers(ctx, "user_tokens:"+userID)
    
    // Выполнить первую часть транзакции
    _, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }
    
    // 2. Отозвать все старые токены
    oldTokens, err := userTokens.Result()
    if err == nil && len(oldTokens) > 0 {
        pipe = r.client.TxPipeline()
        
        for _, oldJTI := range oldTokens {
            pipe.Del(ctx, "active:"+oldJTI)
            pipe.Set(ctx, "revoked:"+oldJTI, 
                fmt.Sprintf(`{"reason":"user_reauth","revoked_at":"%s","user_id":"%s"}`, 
                    time.Now().Format(time.RFC3339), userID), ttl)
        }
        pipe.Del(ctx, "user_tokens:"+userID)
        
        _, err = pipe.Exec(ctx)
        if err != nil {
            return err
        }
    }
    
    // 3. Добавить новый токен
    pipe = r.client.TxPipeline()
    
    metadataJSON, _ := json.Marshal(metadata)
    pipe.Set(ctx, "active:"+jti, string(metadataJSON), ttl)
    pipe.SAdd(ctx, "user_tokens:"+userID, jti)
    
    _, err = pipe.Exec(ctx)
    return err
}

func (r *RedisTokenStorage) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
    result := r.client.Exists(ctx, "revoked:"+jti)
    exists, err := result.Result()
    return exists > 0, err
}

func (r *RedisTokenStorage) GetActiveToken(ctx context.Context, jti string) (*TokenMetadata, error) {
    result := r.client.Get(ctx, "active:"+jti)
    data, err := result.Result()
    if err != nil {
        if err == redis.Nil {
            return nil, nil // Токен не найден
        }
        return nil, err
    }
    
    var metadata TokenMetadata
    err = json.Unmarshal([]byte(data), &metadata)
    return &metadata, err
}

func (r *RedisTokenStorage) RevokeToken(ctx context.Context, userID, jti string, reason string) error {
    pipe := r.client.TxPipeline()
    
    pipe.Del(ctx, "active:"+jti)
    pipe.Set(ctx, "revoked:"+jti, 
        fmt.Sprintf(`{"reason":"%s","revoked_at":"%s","user_id":"%s"}`, 
            reason, time.Now().Format(time.RFC3339), userID), 24*time.Hour)
    pipe.SRem(ctx, "user_tokens:"+userID, jti)
    
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisTokenStorage) GetUserTokens(ctx context.Context, userID string) ([]string, error) {
    result := r.client.SMembers(ctx, "user_tokens:"+userID)
    return result.Result()
}

func (r *RedisTokenStorage) RevokeAllUserTokens(ctx context.Context, userID string, reason string) error {
    // Получить все токены пользователя
    tokens, err := r.GetUserTokens(ctx, userID)
    if err != nil {
        return err
    }
    
    if len(tokens) == 0 {
        return nil // Нет токенов для отзыва
    }
    
    // Атомарно отозвать все токены
    pipe := r.client.TxPipeline()
    
    for _, jti := range tokens {
        pipe.Del(ctx, "active:"+jti)
        pipe.Set(ctx, "revoked:"+jti, 
            fmt.Sprintf(`{"reason":"%s","revoked_at":"%s","user_id":"%s"}`, 
                reason, time.Now().Format(time.RFC3339), userID), 24*time.Hour)
    }
    pipe.Del(ctx, "user_tokens:"+userID)
    
    _, err = pipe.Exec(ctx)
    return err
}
```

### Оптимизация для основного кейса использования

**Сценарий**: Пользователь X авторизуется → нужно отозвать все его активные токены + выдать новый

**Эффективность операций**:

1. **Поиск токенов пользователя**: `SMEMBERS user_tokens:{user_id}` — **O(N)**, где N = количество токенов пользователя (обычно 1-3)

2. **Отзыв токенов**: 
   - `DEL active:{jti}` для каждого токена — **O(1)** 
   - `SET revoked:{jti}` для каждого токена — **O(1)**

3. **Проверка отзыва** (используется другими сервисами): `EXISTS revoked:{jti}` — **O(1)**

4. **Атомарность**: Все операции в Redis MULTI/EXEC транзакции

**Производительность**:
- **Типичный случай** (1 токен): 4 операции Redis (~1-2ms)
- **Множественные токены** (5 токенов): 12 операций Redis (~3-5ms)  
- **Проверка отзыва**: 1 операция Redis (~0.1ms)

**Масштабируемость**:
- Линейная сложность относительно количества токенов у пользователя
- Независимость от общего количества пользователей в системе
- Автоматическая очистка памяти через TTL + hourly cleanup

**Производительность очистки**:
- **Частота**: Каждый час (настраивается через переменную окружения)
- **Время выполнения**: ~1-5 секунд для 10,000 пользователей с токенами
- **Нагрузка на Redis**: Минимальная, использует SCAN с COUNT=100
- **Память**: Освобождает просроченные JTI из user_tokens SET'ов

## Нефункциональные требования

### 1. Производительность
- **Время ответа**: < 200ms для аутентификации
- **Пропускная способность**: до 1000 запросов/сек
- **Concurrent connections**: до 500 одновременных подключений

### 2. Безопасность
- **RSA ключи**: 2048 бит минимум
- **Rate limiting**: 10 запросов/минуту на IP для `/auth`
- **Telegram данные**: проверка подписи обязательна
- **JWT токены**: защита от replay атак через jti

### 3. Надежность
- **Availability**: 99.9%
- **Health check**: `/health` эндпоинт
- **Graceful shutdown**: корректное завершение работы
- **Circuit breaker**: для подключений к PostgreSQL и Redis

### 4. Мониторинг
- **Метрики**: количество успешных/неуспешных авторизаций
- **Логирование**: все операции с токенами
- **Alerts**: при превышении error rate 5%

#### Метрики очистки токенов
- `auth_token_cleanup_duration_seconds` - время выполнения очистки
- `auth_token_cleanup_expired_tokens_total` - количество удаленных просроченных токенов
- `auth_token_cleanup_processed_users_total` - количество обработанных пользователей
- `auth_token_cleanup_errors_total` - количество ошибок при очистке
- `auth_token_cleanup_last_run_timestamp` - timestamp последней успешной очистки

## API Эндпоинты

**📋 Полная OpenAPI 3.0 спецификация**: [`auth-service-openapi.yml`](./auth-service-openapi.yml)

### Основные эндпоинты

#### 1. POST /auth - Аутентификация
**Описание**: Основной эндпоинт для авторизации пользователей через Telegram Web App  
**Rate Limiting**: 10 запросов/минуту на IP  
**Заголовки**: `X-Telegram-Init-Data` (обязательный)

**Коды ответов**:
- `200` - Успешная авторизация с JWT токеном
- `400` - Некорректный запрос (отсутствуют данные, невалидный формат)
- `401` - Ошибка авторизации (невалидная подпись Telegram)
- `429` - Превышен rate limit
- `500` - Внутренняя ошибка сервера

#### 2. GET /health - Health Check
**Описание**: Проверка состояния сервиса и его зависимостей

**Пример ответа**:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-21T10:30:00Z",
  "version": "1.0.0",
  "dependencies": {
    "postgresql": "healthy",
    "redis": "healthy",
    "jwt_keys": "loaded"
  }
}
```

#### 3. POST /refresh - Обновление токена (опционально)
**Описание**: Возможность обновления токена без полной переавторизации  
**Статус**: Планируется для будущих версий

#### 4. POST /revoke - Отзыв токена (опционально)  
**Описание**: Принудительный отзыв токена (logout)  
**Статус**: Планируется для будущих версий

### Модели данных

**AuthResponse**, **User**, **ErrorResponse**, **HealthResponse** - подробные схемы описаны в OpenAPI спецификации.

### Интеграция с другими сервисами

Для валидации JWT токенов в других сервисах используйте:
- **Публичный ключ** RSA для проверки подписи
- **Redis проверку** на отзыв токена: `EXISTS revoked:{jti}`
- **Middleware** из основной спецификации

## Структура проекта сервиса

```
services/auth-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   ├── auth.go
│   │   └── health.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   └── user.go
│   ├── services/
│   │   ├── auth.go
│   │   ├── jwt.go
│   │   └── telegram.go
│   └── storage/
│       ├── postgres.go
│       └── redis.go
├── migrations/
│   └── 001_create_users_table.sql
├── pkg/
│   └── utils/
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Интеграция с миграциями

### Создание миграций
Все изменения схемы БД создаются как SQL скрипты в `/migrations/auth-service/`:

```sql
-- 001_create_users_table.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    language_code VARCHAR(10),
    is_premium BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

### Применение миграций
Миграции применяются одноразовыми контейнерами при развертывании:

```yaml
# docker-compose.yml
services:
  auth-migration:
    build: ../migrations
    volumes:
      - ../migrations/auth-service:/migrations
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/shard_legends
      - MIGRATION_PATH=/migrations
    depends_on:
      - postgres
    command: ["migrate", "up"]
```

## Конфигурация

### Переменные окружения
```bash
# Основные настройки
AUTH_SERVICE_PORT=8080
AUTH_SERVICE_HOST=0.0.0.0

# База данных
DATABASE_URL=postgresql://user:pass@postgres:5432/shard_legends
DATABASE_MAX_CONNECTIONS=10

# Redis
REDIS_URL=redis://redis:6379/0
REDIS_MAX_CONNECTIONS=10
REDIS_PERSISTENCE_ENABLED=true
REDIS_SAVE_INTERVAL=60
REDIS_AOF_ENABLED=true

# JWT
JWT_PRIVATE_KEY_PATH=/etc/auth/private_key.pem
JWT_PUBLIC_KEY_PATH=/etc/auth/public_key.pem
JWT_ISSUER=shard-legends-auth
JWT_EXPIRY_HOURS=24

# Telegram
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# Безопасность
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW=60s

# Очистка токенов
TOKEN_CLEANUP_INTERVAL_HOURS=1
TOKEN_CLEANUP_TIMEOUT_MINUTES=5
```

## Тестирование

### Unit тесты
- Покрытие кода ≥ 80%
- Тесты для всех handlers, services, models
- Мокирование внешних зависимостей (PostgreSQL, Redis)

### Integration тесты
- Тестирование с реальными PostgreSQL и Redis
- Проверка полного цикла аутентификации
- Тестирование отзыва токенов

### E2E тесты
- Тестирование через API Gateway
- Проверка интеграции с Telegram Web App
- Тестирование rate limiting

## Развертывание

### Docker конфигурация
```yaml
# docker-compose.yml
auth-service:
  build: ../../services/auth-service
  container_name: slcw-auth-service-dev
  environment:
    - DATABASE_URL=postgresql://postgres:password@postgres:5432/shard_legends
    - REDIS_URL=redis://redis:6379/0
    - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
  volumes:
    - ./auth-keys:/etc/auth:ro
  networks:
    - slcw-dev
  depends_on:
    - postgres
    - redis
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 30s
    timeout: 10s
    retries: 3
```

### API Gateway интеграция
```nginx
# api-gateway/nginx.conf
upstream auth_service {
    server auth-service:8080;
}

location /auth {
    proxy_pass http://auth_service/auth;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Требования к конфигурации Redis

### Персистентность данных

**Критическая важность**: Redis ДОЛЖЕН сохранять данные между перезапусками, так как содержит:
- Активные JWT токены пользователей
- Отозванные токены (для безопасности)
- Пользовательские сессии

**Обязательная конфигурация Redis для dev/stage/prod:**

```conf
# redis.conf
# Персистентность данных
save 60 1000          # RDB snapshot каждые 60 сек если изменилось ≥1000 ключей
save 300 100          # RDB snapshot каждые 5 мин если изменилось ≥100 ключей
save 900 1            # RDB snapshot каждые 15 мин если изменился ≥1 ключ

# AOF (Append Only File) для максимальной надежности
appendonly yes
appendfsync everysec  # Синхронизация каждую секунду (баланс производительность/надежность)

# КРИТИЧНО: Автоматическая перезапись AOF для предотвращения бесконечного роста
auto-aof-rewrite-percentage 100  # Перезапись при удвоении размера AOF
auto-aof-rewrite-min-size 64mb   # Минимальный размер для запуска перезаписи
no-appendfsync-on-rewrite yes    # Отключать fsync во время перезаписи для производительности

# Дополнительные настройки AOF
aof-load-truncated yes           # Загружать усеченные AOF файлы при восстановлении
aof-use-rdb-preamble yes         # Использовать RDB формат в начале AOF для ускорения

# Настройки памяти
maxmemory 512mb       # Ограничение памяти для предотвращения OOM
maxmemory-policy allkeys-lru  # Удаление старых ключей при нехватке памяти

# Оптимизация производительности
tcp-keepalive 300
timeout 0
tcp-backlog 511
```

**Docker volumes для персистентности:**
```yaml
# docker-compose.yml
redis:
  image: redis:8.0.2
  volumes:
    - redis_data:/data
    - ./redis.conf:/usr/local/etc/redis/redis.conf
  command: redis-server /usr/local/etc/redis/redis.conf

volumes:
  redis_data:
    driver: local
```

**Управление размером AOF и очисткой токенов:**

**Как работает очистка с AOF:**
1. **TTL операции** записываются в AOF как обычно
2. **DEL команды** (от нашей hourly cleanup) тоже записываются
3. **AOF rewrite** автоматически удаляет записи об удаленных ключах
4. **Результат**: размер AOF остается управляемым

**Оптимизация для токенов с TTL:**
```conf
# Пример размеров для 10,000 активных пользователей:
# - Активные токены: ~10,000 ключей (~2MB в Redis)
# - Отозванные токены: ~5,000 ключей (~1MB в Redis) 
# - User tokens SETs: ~10,000 ключей (~500KB в Redis)
# - Итого в памяти: ~3.5MB
# - AOF файл после rewrite: ~7-10MB (с метаданными)

# При автоматической перезаписи каждые 64MB:
# AOF файл никогда не превысит ~128MB (64MB * 2)
```

**Дополнительная оптимизация - более частая AOF перезапись:**
```conf
auto-aof-rewrite-min-size 32mb   # Уменьшить до 32MB для чаще перезаписи
auto-aof-rewrite-percentage 50   # Перезапись при росте на 50% вместо 100%
```

**Мониторинг размера AOF:**
```bash
# Команды для мониторинга
INFO persistence                 # Размер AOF файла
BGREWRITEAOF                     # Принудительная перезапись AOF
CONFIG GET auto-aof-rewrite-*    # Текущие настройки перезаписи
```

**Риски отсутствия персистентности:**
1. **Массовый logout** - все пользователи потеряют сессии
2. **Уязвимость безопасности** - отозванные токены станут валидными
3. **Потеря данных очистки** - просроченные токены не будут корректно отслеживаться
4. **Негативный UX** - пользователи будут вынуждены постоянно переавторизовываться

## Критерии готовности

### Обязательные функции
- [ ] Аутентификация через Telegram Web App данные
- [ ] Генерация JWT токенов с RSA подписью
- [ ] Сохранение активных токенов в Redis с персистентностью
- [ ] Управление отозванными токенами с персистентностью
- [ ] Регистрация новых пользователей в PostgreSQL
- [ ] Health check эндпоинт
- [ ] Rate limiting для защиты от злоупотреблений
- [ ] Конфигурация Redis с RDB + AOF персистентностью

### Качество кода
- [ ] Unit тесты с покрытием ≥ 80%
- [ ] Integration тесты с PostgreSQL и Redis
- [ ] Линтинг без ошибок (golint, go vet)
- [ ] Structured logging на английском языке
- [ ] Graceful shutdown

### Интеграция
- [ ] Интеграция с API Gateway (nginx конфигурация)
- [ ] Docker контейнер с health check
- [ ] Миграции БД в `/migrations/auth-service/`
- [ ] Документация API в OpenAPI формате
- [ ] Middleware для других сервисов

### Развертывание
- [ ] Работа в dev/stage/prod окружениях
- [ ] Конфигурация через переменные окружения
- [ ] Мониторинг и метрики (Prometheus совместимые)
- [ ] Логирование в централизованную систему

Данная спецификация обеспечивает полное техническое описание для реализации сервиса авторизации ИИ-агентом разработчиком.
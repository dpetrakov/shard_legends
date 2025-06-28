# Система интернационализации (i18n) для Shard Legends: Clan Wars

## Обзор

Данный документ описывает архитектуру системы интернационализации (i18n) для игры Shard Legends: Clan Wars, обеспечивающую поддержку мультиязычности для названий и описаний игровых сущностей.

## Принципы проектирования

### 1. Универсальность
- Единая система для всех типов игровых сущностей (предметы, классификаторы, достижения и т.д.)
- Возможность легкого добавления новых типов сущностей без изменения схемы БД

### 2. Производительность  
- Кеширование переводов в Redis для быстрого доступа
- Оптимальные индексы для поиска по entity_type + entity_id + language
- Batch операции для загрузки переводов нескольких сущностей

### 3. Надежность
- Fallback на базовый язык при отсутствии перевода
- Валидация поддерживаемых языков
- Транзакционность операций с переводами

### 4. Масштабируемость
- Поддержка произвольного количества языков
- Возможность добавления новых полей для перевода
- Разделение по сервисам (каждый сервис управляет переводами своих сущностей)

## Архитектура БД

### Схема таблиц

```sql
-- Поддерживаемые языки
CREATE TABLE i18n.languages (
    code VARCHAR(5) PRIMARY KEY,          -- 'ru', 'en', 'zh-CN' 
    name VARCHAR(100) NOT NULL,           -- 'Русский', 'English'
    is_default BOOLEAN DEFAULT FALSE,     -- Базовый язык для fallback
    is_active BOOLEAN DEFAULT TRUE,       -- Активен ли язык
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Универсальная таблица переводов
CREATE TABLE i18n.translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,     -- 'item', 'classifier', 'achievement'
    entity_id UUID NOT NULL,              -- ID сущности
    field_name VARCHAR(50) NOT NULL,      -- 'name', 'description', 'tooltip'
    language_code VARCHAR(5) NOT NULL REFERENCES i18n.languages(code),
    content TEXT NOT NULL,                -- Переведенный текст
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(entity_type, entity_id, field_name, language_code)
);
```

### Индексы для производительности

```sql
-- Основной индекс для поиска переводов
CREATE INDEX idx_translations_lookup 
ON i18n.translations (entity_type, entity_id, field_name, language_code);

-- Индекс для поиска по языку (для админки)
CREATE INDEX idx_translations_language 
ON i18n.translations (language_code, entity_type);

-- Индекс для поиска сущностей без переводов
CREATE INDEX idx_translations_entity 
ON i18n.translations (entity_type, entity_id);
```

## Логика работы

### 1. Получение перевода

```
GET /api/v1/items/{id}?lang=ru

1. Проверить кеш Redis: "i18n:item:{id}:ru"
2. Если в кеше нет:
   a. Запросить из БД translations для entity_type='item', entity_id={id}, language_code='ru'
   b. Если перевода нет - fallback на дефолтный язык (en)
   c. Сохранить в кеш на 24 часа
3. Объединить базовые данные сущности с переводами
4. Вернуть клиенту
```

### 2. Кеширование

**Ключи Redis:**
- `i18n:translations:{entity_type}:{entity_id}:{language}` - переводы конкретной сущности
- `i18n:supported_languages` - список поддерживаемых языков
- `i18n:default_language` - код базового языка

**Стратегия инвалидации:**
- TTL: 24 часа для переводов, 1 час для мета-информации
- Manual: при обновлении переводов через админку
- Bulk: при применении миграций с новыми переводами

### 3. API контракт

**Входящий запрос:**
```json
GET /api/v1/items/1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60?lang=ru
```

**Ответ с переводами:**
```json
{
  "id": "1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60",
  "code": "stone",
  "item_class": "resources",
  "item_type": "stone",
  "name": "Камень",
  "description": "Базовый строительный материал, используется в большинстве рецептов и заданий.",
  "created_at": "2025-06-28T10:00:00Z",
  "updated_at": "2025-06-28T10:00:00Z"
}
```

**Fallback при отсутствии перевода:**
```json
{
  "id": "1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60", 
  "code": "stone",
  "item_class": "resources",
  "item_type": "stone",
  "name": "Stone",                    // fallback на английский
  "description": "Basic building material used in most recipes and quests.",
  "created_at": "2025-06-28T10:00:00Z",
  "updated_at": "2025-06-28T10:00:00Z"
}
```

## Типы сущностей

### Поддерживаемые entity_type

| Entity Type | Описание | Переводимые поля |
|-------------|----------|------------------|
| `item` | Игровые предметы | `name`, `description` |
| `classifier` | Классификаторы | `description` |
| `classifier_item` | Элементы классификаторов | `description` |
| `achievement` | Достижения | `name`, `description`, `tooltip` |
| `recipe` | Рецепты производства | `name`, `description` |
| `user_message` | Пользовательские сообщения | `title`, `content` |

### Расширение на новые типы

Для добавления нового типа сущности:
1. Добавить новый `entity_type` в enum (если используется)
2. Обновить документацию в этом файле
3. Создать dev-data миграцию с переводами
4. Обновить методы в storage слое

## Реализация в Go

### Модели

```go
// Translation представляет перевод сущности
type Translation struct {
    ID           uuid.UUID `json:"id" db:"id"`
    EntityType   string    `json:"entity_type" db:"entity_type"`
    EntityID     uuid.UUID `json:"entity_id" db:"entity_id"`
    FieldName    string    `json:"field_name" db:"field_name"`
    LanguageCode string    `json:"language_code" db:"language_code"`
    Content      string    `json:"content" db:"content"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Language представляет поддерживаемый язык
type Language struct {
    Code      string    `json:"code" db:"code"`
    Name      string    `json:"name" db:"name"`
    IsDefault bool      `json:"is_default" db:"is_default"`
    IsActive  bool      `json:"is_active" db:"is_active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ItemWithTranslations расширяет базовую модель Item переводами
type ItemWithTranslations struct {
    Item
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

### Storage интерфейс

```go
type I18nStorage interface {
    // Получение переводов
    GetTranslations(ctx context.Context, entityType string, entityID uuid.UUID, languageCode string) (map[string]string, error)
    GetTranslationsBatch(ctx context.Context, entityType string, entityIDs []uuid.UUID, languageCode string) (map[uuid.UUID]map[string]string, error)
    
    // Управление переводами
    SetTranslation(ctx context.Context, translation *Translation) error
    SetTranslationsBatch(ctx context.Context, translations []*Translation) error
    DeleteTranslations(ctx context.Context, entityType string, entityID uuid.UUID) error
    
    // Управление языками
    GetSupportedLanguages(ctx context.Context) ([]*Language, error)
    GetDefaultLanguage(ctx context.Context) (*Language, error)
    
    // Кеш
    InvalidateCache(ctx context.Context, entityType string, entityID uuid.UUID) error
}
```

## Миграции и dev-data

### Порядок применения миграций

1. `008_create_i18n_schema.up.sql` - создание i18n схемы и таблиц
2. `009_populate_languages.up.sql` - загрузка поддерживаемых языков
3. `dev-data/.../004_load_translations.sql` - загрузка переводов для тестовых данных

### Структура dev-data переводов

```sql
-- Примеры переводов предметов
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Камень
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'name', 'ru', 'Камень'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'name', 'en', 'Stone'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'description', 'ru', 'Базовый строительный материал'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'description', 'en', 'Basic building material');
```

## Мониторинг и метрики

### Ключевые метрики

- `i18n_translation_cache_hits_total` - попадания в кеш переводов
- `i18n_translation_cache_misses_total` - промахи кеша
- `i18n_fallback_used_total` - использование fallback языка
- `i18n_translation_load_duration_seconds` - время загрузки переводов из БД

### Алерты

- Высокий процент fallback (>10%) - возможно отсутствуют переводы
- Низкий hit rate кеша (<80%) - проблемы с кешированием
- Медленные запросы переводов (>100ms) - проблемы с производительностью БД

## Ограничения и соображения

### Ограничения

1. **Размер контента**: максимум 10KB на перевод (поле TEXT)
2. **Количество языков**: практически неограничено, но рекомендуется <50
3. **Версионирование**: система не поддерживает версионирование переводов

### Соображения производительности

1. **Batch операции**: для загрузки переводов нескольких сущностей используйте batch методы
2. **Кеширование**: обязательно для production, TTL = 24 часа
3. **Индексы**: критически важны для производительности поиска
4. **Connection pooling**: используйте пулы соединений для БД

### Безопасность

1. **Валидация входных данных**: проверка поддерживаемых языков и типов сущностей
2. **Санитизация**: очистка HTML/JS в переводах (если поддерживается rich text)
3. **Права доступа**: ограничение на изменение переводов только для администраторов

## Roadmap

### Версия 1.0 (MVP)
- [x] Базовая схема БД
- [x] Поддержка RU/EN языков  
- [x] Переводы для предметов
- [ ] Кеширование в Redis
- [ ] REST API для получения переводов

### Версия 1.1
- [ ] Админ-панель для управления переводами
- [ ] Batch API для массовых операций
- [ ] Метрики и мониторинг

### Версия 1.2
- [ ] Поддержка rich text (Markdown)
- [ ] Версионирование переводов
- [ ] Экспорт/импорт переводов для переводчиков

### Версия 2.0
- [ ] Автоматический перевод через API (Google Translate, DeepL)
- [ ] A/B тестирование переводов
- [ ] Интеграция с внешними TMS системами

Данная архитектура обеспечивает гибкое и масштабируемое решение для интернационализации игры, готовое к расширению и оптимизации в будущем.
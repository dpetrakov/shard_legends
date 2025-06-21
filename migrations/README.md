# Database Migrations for Shard Legends: Clan Wars

## Простая архитектура миграций

Мы используем **простую линейную структуру миграций** для управления схемой PostgreSQL:

- **Одна база данных** для всех микросервисов
- **Отдельные схемы** для каждого сервиса (auth, game, clan)
- **Единый набор миграций** в правильном порядке
- **Простое управление** без избыточной сложности

### Структура миграций

```
migrations/
├── Dockerfile                         # Контейнер для выполнения миграций
├── README.md                         # Документация
├── 000_init_schemas.up.sql           # Инициализация всех схем и расширений
├── 000_init_schemas.down.sql         # Откат инициализации
├── 001_create_users_table.up.sql     # Auth: таблица пользователей
├── 001_create_users_table.down.sql   # Auth: откат таблицы
├── 002_create_game_profiles.up.sql   # Game: игровые профили (планируется)
└── 002_create_game_profiles.down.sql # Game: откат профилей (планируется)
```

## Соглашения по нумерации

### 000-099: Общие компоненты
- `000_*` - Инициализация схем, расширений, прав доступа
- `001-099` - Общие функции, триггеры, типы данных

### 100-199: Auth Service
- `100_*` - Таблицы пользователей и авторизации
- `101_*` - Индексы и оптимизации

### 200-299: Game Service  
- `200_*` - Игровая логика и прогресс
- `201_*` - Match-3 доски и сессии

### 300-399: Clan Service
- `300_*` - Кланы и участники
- `301_*` - Клановые войны

## Использование

### Применение всех миграций
```bash
# Применение всех миграций в правильном порядке
docker-compose --profile migrations run --rm migrate up

# Проверка статуса миграций
docker-compose --profile migrations run --rm migrate version
```

### Откат миграций
```bash
# Откат последней миграции
docker-compose --profile migrations run --rm migrate down 1

# Откат до конкретной версии
docker-compose --profile migrations run --rm migrate goto 0

# Принудительная установка версии (при ошибках)
docker-compose --profile migrations run --rm migrate force 1
```

### Ручное выполнение с переменными окружения
```bash
# Прямое выполнение миграций через docker run
docker run --rm --network slcw-dev \
  -e DATABASE_URL="postgresql://slcw_user:dev_password_2024@slcw-postgres-dev:5432/shard_legends_dev?sslmode=disable" \
  -v $(pwd)/migrations:/migrations \
  dev-migrate:latest \
  up
```

## Преимущества простого подхода

### ✅ Плюсы:
1. **Простота** - нет сложной структуры папок
2. **Единый порядок** - все миграции в одном месте
3. **Нет конфликтов** - четкая нумерация и зависимости
4. **Легко понять** - линейная последовательность изменений
5. **Простое развертывание** - один контейнер, одна команда

### 🎯 Принципы:
1. **Каждый сервис использует свою схему** (`auth.*`, `game.*`, `clan.*`)
2. **Миграции нумеруются по порядку добавления** (000, 001, 002...)
3. **Зависимости указываются в комментариях** миграций
4. **Общие компоненты создаются первыми** (схемы, расширения)

## Создание новой миграции

### 1. Определить номер
```bash
# Найти последний номер миграции
ls -1 migrations/*.up.sql | sort | tail -1
# Результат: 001_create_users_table.up.sql

# Следующий номер: 002
```

### 2. Создать файлы миграции
```bash
# UP миграция
touch migrations/002_add_user_preferences.up.sql
# DOWN миграция  
touch migrations/002_add_user_preferences.down.sql
```

### 3. Заполнить содержимое
```sql
-- UP миграция
-- Migration UP: 002_add_user_preferences.up.sql
-- Description: Добавление пользовательских настроек
-- Service: auth-service
-- Depends: 001_create_users_table.up.sql

ALTER TABLE auth.users ADD COLUMN preferences JSONB DEFAULT '{}';
CREATE INDEX idx_users_preferences ON auth.users USING GIN (preferences);
```

```sql
-- DOWN миграция
-- Migration DOWN: 002_add_user_preferences.down.sql
-- Description: Откат добавления пользовательских настроек

DROP INDEX IF EXISTS auth.idx_users_preferences;
ALTER TABLE auth.users DROP COLUMN IF EXISTS preferences;
```

## Troubleshooting

### Проблема: Конфликт номеров миграций
**Решение:** Координируйтесь с командой при создании новых миграций

### Проблема: Миграция не применяется
**Решение:** Проверьте статус через `migrate version` и логи контейнера

### Проблема: Нужно изменить существующую структуру
**Решение:** Создайте новую миграцию с изменениями, не изменяйте существующие

### Проблема: Откат не работает
**Решение:** Убедитесь, что DOWN-миграция корректно описывает откат всех изменений UP-миграции

## Текущие миграции

- `000_init_schemas.*` - Создание схем auth, game, clan + PostgreSQL расширения
- `001_create_users_table.*` - Таблица пользователей для auth-service

Следующие миграции будут добавляться по мере развития проекта.
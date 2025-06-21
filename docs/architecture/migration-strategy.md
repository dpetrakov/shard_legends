# Стратегия управления миграциями БД

## Обзор

Проект Shard Legends: Clan Wars использует декларативный подход к управлению схемой базы данных PostgreSQL. Все изменения схемы применяются через SQL миграции, организованные по сервисам и выполняемые одноразовыми контейнерами при развертывании.

## Принципы управления миграциями

### 1. Декларативный подход
- Схема БД описывается через SQL миграции
- Каждая миграция - отдельный SQL файл
- Миграции применяются последовательно в алфавитном порядке
- Идемпотентность: повторное применение не изменяет результат

### 2. Сервисно-ориентированная организация
```
migrations/
├── Dockerfile                              # Контейнер для выполнения миграций
├── auth-service/
│   ├── 001_create_users_table.up.sql      # Создание таблицы пользователей
│   ├── 001_create_users_table.down.sql    # Откат создания таблицы
│   ├── 002_add_user_indexes.up.sql        # Добавление индексов
│   └── 002_add_user_indexes.down.sql      # Откат индексов
├── game-service/
│   ├── 001_create_game_sessions.up.sql    # Игровые сессии
│   ├── 001_create_game_sessions.down.sql  # Откат сессий
│   ├── 002_create_match3_boards.up.sql    # Match-3 доски
│   └── 002_create_match3_boards.down.sql  # Откат досок
└── shared/
    ├── 001_create_extensions.up.sql       # PostgreSQL расширения
    ├── 001_create_extensions.down.sql     # Откат расширений
    ├── 002_create_common_functions.up.sql # Общие функции
    └── 002_create_common_functions.down.sql # Откат функций
```

### 3. Версионирование
- **Формат UP**: `{номер}_{описание}.up.sql` (применение миграции)
- **Формат DOWN**: `{номер}_{описание}.down.sql` (откат миграции)
- **Номер**: 3-значный (001, 002, 003...)
- **Описание**: краткое описание на английском с подчеркиваниями
- **Примеры**: 
  - `001_create_users_table.up.sql` / `001_create_users_table.down.sql`
  - `002_add_telegram_id_index.up.sql` / `002_add_telegram_id_index.down.sql`

## Инструменты для миграций

### Migrate CLI
Используется [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) для выполнения миграций:

```dockerfile
# migrations/Dockerfile
FROM alpine:3.19

# Установка migrate CLI и PostgreSQL client
RUN apk add --no-cache \
    curl \
    postgresql-client && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-arm64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR /migrations
COPY . /migrations/

ENTRYPOINT ["migrate"]
CMD ["-help"]
```

### Docker контейнер для миграций
```yaml
# docker-compose.yml
services:
  migrate-auth:
    build: ../migrations
    volumes:
      - ../migrations/auth-service:/migrations
    environment:
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/shard_legends
    depends_on:
      - postgres
    command: ["migrate", "-path", "/migrations", "-database", "${DATABASE_URL}", "up"]
```

## Процесс применения миграций

### 1. Создание новой миграции

```bash
# Создание миграции для auth-service
cd migrations/auth-service/
# Найти следующий номер
ls -1 *.sql | sort | tail -1  # получить последний файл
# Создать новую миграцию
touch 003_add_user_preferences.sql
```

### 2. Содержимое миграции
```sql
-- 001_create_users_table.sql
-- Создание таблицы пользователей для auth-service

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

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Комментарии для документации
COMMENT ON TABLE users IS 'Базовые данные пользователей для авторизации';
COMMENT ON COLUMN users.telegram_id IS 'Уникальный ID пользователя в Telegram';
COMMENT ON COLUMN users.is_premium IS 'Статус Telegram Premium пользователя';
```

### 3. Применение в разных средах

#### Development
```bash
cd deploy/dev/
docker-compose up migrate-auth
```

#### Staging  
```bash
cd deploy/stage/
docker-compose up migrate-auth
```

#### Production
```bash
cd deploy/prod/
docker-compose up migrate-auth
```

### 4. Проверка результата
```bash
# Проверка применения миграций
docker-compose exec postgres psql -U postgres -d shard_legends -c "\dt"
docker-compose exec postgres psql -U postgres -d shard_legends -c "SELECT * FROM schema_migrations;"
```

## Структура миграций по сервисам

### Auth Service (`migrations/auth-service/`)
```sql
-- 001_create_users_table.sql - базовая таблица пользователей
-- 002_add_user_indexes.sql - индексы для производительности
-- 003_add_last_login_tracking.sql - отслеживание последнего входа
```

### Game Service (`migrations/game-service/`)
```sql
-- 001_create_game_sessions.sql - игровые сессии
-- 002_create_match3_boards.sql - состояния игровых досок
-- 003_create_user_progress.sql - прогресс пользователей
```

### Shared (`migrations/shared/`)
```sql
-- 001_create_extensions.sql - необходимые PostgreSQL расширения
-- 002_create_common_functions.sql - общие функции
-- 003_create_audit_triggers.sql - триггеры для аудита
```

## Правила создания миграций

### 1. Обязательные элементы
- **Комментарий с описанием** в начале файла
- **IF NOT EXISTS** для CREATE операций
- **Индексы** для часто запрашиваемых колонок
- **Комментарии** для таблиц и важных колонок

### 2. Именование
- **Таблицы**: множественное число, snake_case (`users`, `game_sessions`)
- **Колонки**: snake_case (`telegram_id`, `created_at`)
- **Индексы**: `idx_{table}_{column}` (`idx_users_telegram_id`)
- **Constraints**: `{table}_{column}_{type}` (`users_telegram_id_unique`)

### 3. Типы данных
- **ID**: UUID с `gen_random_uuid()` по умолчанию
- **Внешние ID**: BIGINT для Telegram ID
- **Timestamps**: `TIMESTAMP WITH TIME ZONE` с `DEFAULT NOW()`
- **Строки**: VARCHAR с явным указанием размера
- **Булевы**: BOOLEAN с DEFAULT значением

### 4. Безопасность миграций
```sql
-- ✅ Безопасно: добавление колонки
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- ✅ Безопасно: создание индекса
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_phone ON users(phone);

-- ⚠️ Осторожно: удаление колонки (только после проверки использования)
-- ALTER TABLE users DROP COLUMN IF EXISTS old_column;

-- ❌ Опасно: изменение типа данных (требует downtime)
-- ALTER TABLE users ALTER COLUMN telegram_id TYPE VARCHAR(50);
```

## CI/CD интеграция

### 1. GitHub Actions workflow
```yaml
# .github/workflows/migrations.yml
name: Database Migrations

on:
  push:
    paths:
      - 'migrations/**'
      - 'deploy/**'

jobs:
  test-migrations:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: shard_legends_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    
    - name: Run migrations
      run: |
        cd migrations
        docker build -t migrations .
        docker run --rm \
          --network host \
          migrations \
          -database "postgresql://postgres:test@localhost:5432/shard_legends_test?sslmode=disable" \
          up
```

### 2. Автоматическое применение в staging
```bash
# deploy/stage/deploy.sh
#!/bin/bash

# Применение миграций перед развертыванием сервисов
echo "Applying database migrations..."
docker-compose up migrate-auth migrate-game migrate-shared

# Проверка успешности миграций
if [ $? -eq 0 ]; then
    echo "Migrations applied successfully"
    docker-compose up -d auth-service game-service
else
    echo "Migration failed, aborting deployment"
    exit 1
fi
```

## Мониторинг и отладка

### 1. Логирование миграций
```sql
-- Таблица для отслеживания миграций (создается автоматически migrate CLI)
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE
);

-- Просмотр примененных миграций
SELECT version, dirty FROM schema_migrations ORDER BY version;
```

### 2. Откат миграций
```bash
# Откат последней миграции
docker-compose run migrate-auth migrate \
  -path /migrations \
  -database $DATABASE_URL \
  down 1

# Откат до конкретной версии
docker-compose run migrate-auth migrate \
  -path /migrations \
  -database $DATABASE_URL \
  goto 2
```

### 3. Проверка состояния
```bash
# Проверка текущей версии
docker-compose run migrate-auth migrate \
  -path /migrations \
  -database $DATABASE_URL \
  version

# Проверка ожидающих миграций
docker-compose run migrate-auth migrate \
  -path /migrations \
  -database $DATABASE_URL \
  up -dry-run
```

## Примеры миграций

### Auth Service миграции
```sql
-- migrations/auth-service/001_create_users_table.sql
-- Создание базовой таблицы пользователей для системы авторизации

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    language_code VARCHAR(10),
    is_premium BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    CONSTRAINT users_telegram_id_positive CHECK (telegram_id > 0)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = TRUE;

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии для документации
COMMENT ON TABLE users IS 'Базовые данные пользователей для системы авторизации';
COMMENT ON COLUMN users.id IS 'Внутренний UUID пользователя в системе';
COMMENT ON COLUMN users.telegram_id IS 'Уникальный ID пользователя в Telegram';
COMMENT ON COLUMN users.username IS 'Username в Telegram (может отсутствовать)';
COMMENT ON COLUMN users.is_premium IS 'Статус Telegram Premium пользователя';
COMMENT ON COLUMN users.last_login_at IS 'Время последней авторизации пользователя';
```

Эта стратегия обеспечивает контролируемое и безопасное управление схемой базы данных для всех микросервисов проекта.
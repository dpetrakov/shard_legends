# Стратегия управления миграциями БД

## Обзор

Проект Shard Legends: Clan Wars использует **простую линейную архитектуру миграций** для управления схемой базы данных PostgreSQL. Все изменения схемы применяются через SQL миграции в едином порядке с использованием golang-migrate/migrate.

## Принципы управления миграциями

### 1. Простая линейная структура
- **Одна база данных** для всех микросервисов
- **Отдельные схемы** для каждого сервиса (auth, game, clan, inventory)
- **Единый набор миграций** в правильном порядке
- **Простое управление** без избыточной сложности

### 2. Организация миграций
```
migrations/
├── Dockerfile                            # Контейнер для выполнения миграций
├── README.md                            # Документация  
├── 000_init_schemas.up.sql              # Инициализация всех схем и расширений
├── 000_init_schemas.down.sql            # Откат инициализации
├── 001_create_users_table.up.sql        # Auth: таблица пользователей
├── 001_create_users_table.down.sql      # Auth: откат таблицы
├── 002_create_inventory_schema.up.sql   # Inventory: структура схемы
├── 002_create_inventory_schema.down.sql # Inventory: откат структуры
├── 003_populate_classifiers.up.sql      # Inventory: дистрибутивные данные
├── 003_populate_classifiers.down.sql    # Inventory: откат данных
└── dev-data/                            # Тестовые данные только для dev
    └── inventory-service/
        └── 001_test_items_and_operations.sql
```

### 3. Соглашения по нумерации

**000-099: Общие компоненты**
- `000_*` - Инициализация схем, расширений, прав доступа
- `001-099` - Общие функции, триггеры, типы данных

**100-199: Auth Service**  
- `100_*` - Таблицы пользователей и авторизации
- `101_*` - Индексы и оптимизации

**200-299: Game Service**
- `200_*` - Игровая логика и прогресс  
- `201_*` - Match-3 доски и сессии

**300-399: Clan Service**
- `300_*` - Кланы и участники
- `301_*` - Клановые войны

**400-499: Inventory Service**
- `400_*` - Структура инвентаря
- `401_*` - Классификаторы и данные

### 4. Версионирование
- **Формат UP**: `{номер}_{описание}.up.sql` (применение миграции)
- **Формат DOWN**: `{номер}_{описание}.down.sql` (откат миграции)
- **Номер**: 3-значный (000, 001, 002...)
- **Описание**: краткое описание на английском с подчеркиваниями
- **Примеры**: 
  - `000_init_schemas.up.sql` / `000_init_schemas.down.sql`
  - `001_create_users_table.up.sql` / `001_create_users_table.down.sql`
  - `002_create_inventory_schema.up.sql` / `002_create_inventory_schema.down.sql`

## Инструменты для миграций

### Migrate CLI
Используется [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) для выполнения миграций:

```dockerfile
# migrations/Dockerfile
FROM alpine:3.19

# Установка migrate CLI и PostgreSQL client с поддержкой архитектур
RUN apk add --no-cache curl postgresql-client && \
    ARCH=$(case $(uname -m) in \
        x86_64) echo "amd64" ;; \
        aarch64) echo "arm64" ;; \
        armv7l) echo "armv7" ;; \
        *) echo "amd64" ;; \
    esac) && \
    curl -L "https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-${ARCH}.tar.gz" | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR /migrations
COPY . /migrations/

# Wrapper скрипт для работы с переменными окружения
RUN echo '#!/bin/sh' > /migrate.sh && \
    echo 'exec migrate -path /migrations -database "$DATABASE_URL" "$@"' >> /migrate.sh && \
    chmod +x /migrate.sh

ENTRYPOINT ["/migrate.sh"]
CMD ["up"]
```

### Docker контейнер для миграций
```yaml
# deploy/dev/docker-compose.yml
services:
  migrate:
    build: ../../migrations
    container_name: slcw-migrate-dev
    environment:
      - DATABASE_URL=postgresql://slcw_user:dev_password_2024@postgres:5432/shard_legends_dev?sslmode=disable
    volumes:
      - ../../migrations:/migrations
    networks:
      - slcw-dev
    depends_on:
      postgres:
        condition: service_healthy
    profiles:
      - migrations
```

## Процесс применения миграций

### 1. Создание новой миграции

```bash
# Перейти в папку миграций
cd migrations/

# Найти следующий номер
ls -1 *.up.sql | sort | tail -1  # получить последний файл
# Результат: 001_create_users_table.up.sql

# Создать новую миграцию (следующий номер: 002)
touch 002_create_inventory_schema.up.sql
touch 002_create_inventory_schema.down.sql
```

### 2. Содержимое миграции
```sql
-- 002_create_inventory_schema.up.sql
-- Migration UP: Создание схемы inventory и таблиц
-- Service: inventory-service
-- Depends: 000_init_schemas.up.sql

CREATE TABLE IF NOT EXISTS inventory.classifiers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(100) UNIQUE NOT NULL,
    description text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL
);

-- Индексы и комментарии...
```

### 3. Применение в разных средах

#### Development (применение всех миграций)
```bash
cd deploy/dev/
docker-compose --profile migrations run --rm migrate up
```

#### Staging  
```bash
cd deploy/stage/
docker-compose --profile migrations run --rm migrate up
```

#### Production
```bash
cd deploy/prod/
docker-compose --profile migrations run --rm migrate up
```

### 4. Проверка результата
```bash
# Проверка статуса миграций
docker-compose --profile migrations run --rm migrate version

# Проверка схем и таблиц в БД
docker-compose exec postgres psql -U slcw_user -d shard_legends_dev -c "\dt auth.*"
docker-compose exec postgres psql -U slcw_user -d shard_legends_dev -c "\dt inventory.*"

# Проверка таблицы миграций
docker-compose exec postgres psql -U slcw_user -d shard_legends_dev -c "SELECT * FROM schema_migrations ORDER BY version;"
```

## Структура миграций по сервисам

### Общие компоненты (000-099)
```sql
-- 000_init_schemas.up.sql - инициализация схем auth, game, clan, inventory + расширения
-- 001-099 - резерв для общих функций, триггеров, типов данных
```

### Auth Service (100-199)
```sql  
-- 001_create_users_table.up.sql - базовая таблица пользователей
-- 100-199 - будущие миграции auth-service
```

### Inventory Service (002-003 сейчас, 400-499 в будущем)
```sql
-- 002_create_inventory_schema.up.sql - структура схемы inventory
-- 003_populate_classifiers.up.sql - дистрибутивные данные классификаторов
-- 400-499 - будущие миграции inventory-service
```

### Game Service (200-299) 
```sql
-- 200-299 - планируемые миграции game-service
-- 200_create_game_sessions.up.sql - игровые сессии
-- 201_create_match3_boards.up.sql - состояния игровых досок
```

### Clan Service (300-399)
```sql  
-- 300-399 - планируемые миграции clan-service
-- 300_create_clans.up.sql - структура кланов
-- 301_create_clan_wars.up.sql - клановые войны
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

### Пример inventory-service миграции
```sql
-- migrations/002_create_inventory_schema.up.sql
-- Migration UP: Создание схемы inventory и всех таблиц
-- Service: inventory-service  
-- Depends: 000_init_schemas.up.sql

-- Общий классификатор для справочных данных
CREATE TABLE inventory.classifiers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(100) UNIQUE NOT NULL,
    description text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    
    CONSTRAINT chk_classifiers_code_not_empty CHECK (length(trim(code)) > 0)
);

-- Индекс для быстрого поиска по коду классификатора
CREATE UNIQUE INDEX idx_classifiers_code ON inventory.classifiers (code);

-- Элементы классификаторов
CREATE TABLE inventory.classifier_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    classifier_id uuid NOT NULL REFERENCES inventory.classifiers(id),
    code varchar(100) NOT NULL,
    description text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    
    CONSTRAINT chk_classifier_items_code_not_empty CHECK (length(trim(code)) > 0)
);

-- Остальные таблицы inventory...

-- Комментарии для документации
COMMENT ON TABLE inventory.classifiers IS 'Общий классификатор для всех справочных данных системы';
COMMENT ON TABLE inventory.classifier_items IS 'Элементы классификаторов - конкретные значения внутри каждого классификатора';
```

## Тестовые данные для dev среды

Тестовые данные размещаются отдельно от основных миграций в папке `dev-data/` и применяются только в dev среде:

```bash
# Применение тестовых данных вручную в dev среде
docker-compose exec postgres psql -U slcw_user -d shard_legends_dev -f /migrations/dev-data/inventory-service/001_test_items_and_operations.sql
```

Эта простая стратегия обеспечивает контролируемое и безопасное управление схемой базы данных для всех микросервисов проекта.
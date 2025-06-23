# Задача 7: Интеграционное тестирование и развертывание

## Описание

Создание комплексной системы тестирования inventory-service, включая интеграционные тесты с другими сервисами, E2E тесты, нагрузочное тестирование и настройку CI/CD pipeline для автоматического развертывания.

## Цели

1. Создать интеграционные тесты с auth-service
2. Реализовать E2E тесты полных сценариев
3. Добавить нагрузочное тестирование
4. Настроить Docker Compose для локальной разработки
5. Создать CI/CD pipeline
6. Подготовить документацию для развертывания

## Подзадачи

### 7.1. Интеграционные тесты с auth-service
**Директория**: `tests/integration/auth/`

**Файл**: `tests/integration/auth/jwt_test.go`

**Сценарии**:
```go
func TestJWTAuthentication(t *testing.T) {
    // 1. Получить валидный JWT токен от auth-service
    // 2. Отправить запрос к inventory с токеном
    // 3. Проверить успешную аутентификацию
    // 4. Проверить извлечение user_id из токена
}

func TestJWTExpiredToken(t *testing.T) {
    // 1. Использовать просроченный JWT токен
    // 2. Отправить запрос к inventory
    // 3. Проверить получение 401 ошибки
}

func TestJWTInvalidToken(t *testing.T) {
    // 1. Использовать поврежденный JWT токен
    // 2. Отправить запрос к inventory
    // 3. Проверить получение 401 ошибки
}

func TestJWTMissingToken(t *testing.T) {
    // 1. Отправить запрос без Authorization header
    // 2. Проверить получение 401 ошибки
}
```

### 7.2. E2E тесты полных сценариев
**Директория**: `tests/e2e/`

**Файл**: `tests/e2e/inventory_scenarios_test.go`

**Сценарии**:
```go
func TestFullInventoryWorkflow(t *testing.T) {
    // 1. Аутентификация пользователя
    // 2. Добавление предметов в инвентарь
    // 3. Проверка остатков
    // 4. Резервирование предметов
    // 5. Возврат/потребление резерва
    // 6. Финальная проверка остатков
}

func TestInsufficientBalanceScenario(t *testing.T) {
    // 1. Создать пользователя с минимальным инвентарем
    // 2. Попытаться зарезервировать больше предметов чем есть
    // 3. Проверить корректную ошибку insufficient_items
    // 4. Проверить что инвентарь не изменился
}

func TestConcurrentOperations(t *testing.T) {
    // 1. Создать пользователя с инвентарем
    // 2. Запустить параллельные операции резервирования
    // 3. Проверить корректность финального состояния
    // 4. Убедиться в отсутствии race conditions
}

func TestDailyBalanceCreation(t *testing.T) {
    // 1. Создать операции за несколько дней
    // 2. Запросить остатки (должны создаться daily_balances)
    // 3. Проверить корректность созданных остатков
    // 4. Проверить производительность последующих запросов
}

func TestAdminOperations(t *testing.T) {
    // 1. Аутентификация администратора
    // 2. Корректировка инвентаря пользователя
    // 3. Проверка изменений
    // 4. Проверка audit log
}
```

### 7.3. Нагрузочное тестирование
**Директория**: `tests/load/`

**Файл**: `tests/load/k6_loadtest.js`

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp up
    { duration: '1m', target: 100 },   // Stay at 100 users
    { duration: '30s', target: 200 },  // Ramp up to 200
    { duration: '2m', target: 200 },   // Stay at 200 users
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% requests under 1s
    http_req_failed: ['rate<0.1'],     // Error rate under 10%
  },
};

const BASE_URL = 'http://localhost:8080';

export function setup() {
  // Get JWT token for testing
  let authResponse = http.post('http://localhost:8081/auth/login', {
    telegram_id: 123456789,
    username: 'testuser',
  });
  
  return { token: authResponse.json('token') };
}

export default function (data) {
  let headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // Test GET /inventory
  let getInventoryRes = http.get(`${BASE_URL}/inventory`, { headers });
  check(getInventoryRes, {
    'GET /inventory status is 200': (r) => r.status === 200,
    'GET /inventory response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Test POST /inventory/add-items
  let addItemsPayload = JSON.stringify({
    user_id: '550e8400-e29b-41d4-a716-446655440000',
    section: 'main',
    operation_type: 'system_reward',
    operation_id: '550e8400-e29b-41d4-a716-446655440001',
    items: [
      {
        item_id: '550e8400-e29b-41d4-a716-446655440002',
        quantity: 10
      }
    ]
  });

  let addItemsRes = http.post(`${BASE_URL}/inventory/add-items`, addItemsPayload, { headers });
  check(addItemsRes, {
    'POST /inventory/add-items status is 200': (r) => r.status === 200,
    'POST /inventory/add-items response time < 1s': (r) => r.timings.duration < 1000,
  });

  sleep(1);
}
```

**Файл**: `tests/load/artillery_loadtest.yml`

```yaml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
    - duration: 120
      arrivalRate: 50
    - duration: 60
      arrivalRate: 100
  defaults:
    headers:
      Authorization: 'Bearer {{token}}'

scenarios:
  - name: 'Inventory Operations'
    weight: 100
    flow:
      - get:
          url: '/inventory'
          capture:
            - json: '$.items[0].item_id'
              as: 'item_id'
      - post:
          url: '/inventory/reserve'
          json:
            user_id: '550e8400-e29b-41d4-a716-446655440000'
            operation_id: '550e8400-e29b-41d4-a716-446655440001'
            items:
              - item_id: '{{item_id}}'
                quantity: 1
      - post:
          url: '/inventory/return-reserve'
          json:
            user_id: '550e8400-e29b-41d4-a716-446655440000'
            operation_id: '550e8400-e29b-41d4-a716-446655440001'
```

### 7.4. Docker Compose для разработки
**Файл**: `docker-compose.dev.yml`

```yaml
version: '3.8'

services:
  inventory-service:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
      - "9090:9090"  # metrics
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=shard_legends
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LOG_LEVEL=debug
    depends_on:
      - postgres
      - redis
      - auth-service
    volumes:
      - ./migrations:/app/migrations
    networks:
      - shard-legends

  auth-service:
    image: shard-legends/auth-service:latest
    ports:
      - "8081:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=shard_legends
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis
    networks:
      - shard-legends

  postgres:
    image: postgres:17
    environment:
      - POSTGRES_DB=shard_legends
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - shard-legends

  redis:
    image: redis:8.0.2-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - shard-legends

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus:/etc/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - shard-legends

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./monitoring/grafana:/etc/grafana/provisioning
      - grafana_data:/var/lib/grafana
    networks:
      - shard-legends

volumes:
  postgres_data:
  redis_data:
  grafana_data:

networks:
  shard-legends:
    driver: bridge
```

### 7.5. CI/CD Pipeline
**Файл**: `.github/workflows/inventory-service.yml`

```yaml
name: Inventory Service CI/CD

on:
  push:
    branches: [ main, develop ]
    paths: [ 'services/inventory-service/**' ]
  pull_request:
    branches: [ main ]
    paths: [ 'services/inventory-service/**' ]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:8.0.2-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

    - name: Install dependencies
      working-directory: ./services/inventory-service
      run: go mod download

    - name: Run database migrations
      working-directory: ./services/inventory-service
      run: |
        go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable" up

    - name: Run unit tests
      working-directory: ./services/inventory-service
      run: go test -v ./...

    - name: Run integration tests
      working-directory: ./services/inventory-service
      run: go test -tags=integration -v ./tests/integration/...
      env:
        TEST_DB_URL: postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable
        TEST_REDIS_URL: redis://localhost:6379

    - name: Generate coverage report
      working-directory: ./services/inventory-service
      run: |
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./services/inventory-service/coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: ./services/inventory-service
        push: true
        tags: |
          ghcr.io/${{ github.repository }}/inventory-service:${{ github.sha }}
          ghcr.io/${{ github.repository }}/inventory-service:latest
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    steps:
    - name: Deploy to staging
      run: |
        echo "Deploying to staging environment"
        # Add actual deployment commands here

  deploy-production:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to production
      run: |
        echo "Deploying to production environment"
        # Add actual deployment commands here
```

### 7.6. Тестовые утилиты
**Файл**: `tests/utils/test_helpers.go`

```go
package utils

import (
    "context"
    "database/sql"
    "fmt"
    "testing"
    "time"
    
    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

// TestDatabase provides database utilities for testing
type TestDatabase struct {
    DB *sql.DB
}

func NewTestDatabase(t *testing.T) *TestDatabase {
    db, err := sql.Open("postgres", getTestDatabaseURL())
    require.NoError(t, err)
    
    return &TestDatabase{DB: db}
}

func (td *TestDatabase) CleanupTables(t *testing.T) {
    tables := []string{
        "inventory.operations",
        "inventory.daily_balances",
        "inventory.item_images",
        "inventory.items",
        "inventory.classifier_items",
        "inventory.classifiers",
    }
    
    for _, table := range tables {
        _, err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
        require.NoError(t, err)
    }
}

func (td *TestDatabase) SeedClassifiers(t *testing.T) {
    // Insert test classifiers and items
    // This would contain the same data as migrations
}

func (td *TestDatabase) CreateTestUser(t *testing.T) uuid.UUID {
    userID := uuid.New()
    
    _, err := td.DB.Exec(`
        INSERT INTO auth.users (id, telegram_id, first_name, username) 
        VALUES ($1, $2, $3, $4)
    `, userID, 123456789, "Test User", "testuser")
    require.NoError(t, err)
    
    return userID
}

func (td *TestDatabase) AddTestItems(t *testing.T, userID uuid.UUID, items []TestItem) {
    for _, item := range items {
        _, err := td.DB.Exec(`
            INSERT INTO inventory.operations (
                user_id, section_id, item_id, collection_id, quality_level_id,
                quantity_change, operation_type_id, operation_id
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, userID, item.SectionID, item.ItemID, item.CollectionID, 
           item.QualityLevelID, item.Quantity, item.OperationTypeID, uuid.New())
        require.NoError(t, err)
    }
}

type TestItem struct {
    ItemID         uuid.UUID
    SectionID      uuid.UUID
    CollectionID   uuid.UUID
    QualityLevelID uuid.UUID
    OperationTypeID uuid.UUID
    Quantity       int64
}

// JWT utilities for testing
func GenerateTestJWT(t *testing.T, userID uuid.UUID) string {
    // Implementation depends on auth-service JWT generation
    // This should match the auth-service implementation
    return "test-jwt-token"
}

func GetTestAuthHeaders(t *testing.T, userID uuid.UUID) map[string]string {
    token := GenerateTestJWT(t, userID)
    return map[string]string{
        "Authorization": fmt.Sprintf("Bearer %s", token),
        "Content-Type":  "application/json",
    }
}
```

### 7.7. Performance тесты
**Файл**: `tests/performance/benchmark_test.go`

```go
package performance

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
)

func BenchmarkGetInventory(b *testing.B) {
    // Setup test environment
    service := setupTestService(b)
    userID := createTestUser(b)
    addTestItems(b, userID, 1000) // Add 1000 items
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.GetUserInventory(context.Background(), userID, "main")
        require.NoError(b, err)
    }
}

func BenchmarkCalculateBalance(b *testing.B) {
    service := setupTestService(b)
    userID := createTestUser(b)
    
    // Create many operations to test performance
    for i := 0; i < 1000; i++ {
        addTestOperation(b, userID)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.CalculateCurrentBalance(context.Background(), balanceRequest)
        require.NoError(b, err)
    }
}

func BenchmarkConcurrentOperations(b *testing.B) {
    service := setupTestService(b)
    userID := createTestUser(b)
    
    b.ResetTimer()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := service.AddItems(context.Background(), addItemsRequest)
            require.NoError(b, err)
        }
    })
}
```

### 7.8. Документация развертывания
**Файл**: `docs/deployment/README.md`

```markdown
# Inventory Service Deployment Guide

## Prerequisites

- Docker and Docker Compose
- PostgreSQL 17
- Redis 8.0.2
- Auth Service running

## Local Development

1. Start all services:
```bash
docker-compose -f docker-compose.dev.yml up -d
```

2. Run migrations:
```bash
make migrate-up
```

3. Verify deployment:
```bash
curl http://localhost:8080/health
```

## Production Deployment

### Environment Variables

Required:
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_NAME` - Database name
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port

Optional:
- `LOG_LEVEL` - Logging level (default: info)
- `SERVER_PORT` - HTTP server port (default: 8080)
- `METRICS_PORT` - Metrics port (default: 9090)

### Health Checks

- `/health` - Basic health check
- `/health/ready` - Readiness probe
- `/health/live` - Liveness probe

### Monitoring

- Metrics available at `:9090/metrics`
- Grafana dashboard: "Inventory Service"
- Prometheus alerts configured
```

## Критерии готовности

### Функциональные
- [ ] Все интеграционные тесты проходят
- [ ] E2E сценарии выполняются успешно
- [ ] Нагрузочные тесты показывают приемлемую производительность
- [ ] Docker Compose поднимает полный стек
- [ ] CI/CD pipeline работает автоматически

### Технические
- [ ] Coverage тестов > 80%
- [ ] Performance тесты в пределах SLA
- [ ] Load тесты не показывают memory leaks
- [ ] Deployment scripts работают без ошибок

### Операционные
- [ ] Мониторинг функционирует
- [ ] Логи структурированы и информативны
- [ ] Метрики корректно собираются
- [ ] Алерты срабатывают при проблемах

## Зависимости

### Входящие
- Все предыдущие задачи (1-6)
- auth-service для интеграционных тестов
- Инфраструктура (PostgreSQL, Redis, Prometheus, Grafana)

### Исходящие
- Production-ready сервис
- Полная тестовая автоматизация
- CI/CD pipeline
- Операционная документация

## Риски и ограничения

- **Риск**: Flaky тесты в CI/CD
  **Митигация**: Retry механизмы, stable test data

- **Риск**: Performance деградация под нагрузкой
  **Митигация**: Regular load testing, profiling

- **Ограничение**: Зависимость от внешних сервисов
  **Решение**: Mock'и для изолированного тестирования
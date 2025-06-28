# Auth Service

Authentication and authorization service for Shard Legends: Clan Wars. Provides JWT token-based authentication using Telegram Web App data validation.

## Features

- Telegram Web App authentication
- JWT token generation and validation with RSA-2048 signing
- Automatic RSA key generation and management
- Public key export for other microservices
- Redis-based token storage and revocation
- PostgreSQL user storage
- Rate limiting
- Health checks
- Graceful shutdown
- Prometheus metrics for monitoring and observability

## Configuration

The service is configured through environment variables:

### Required Variables

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string  
- `TELEGRAM_BOT_TOKEN` - Telegram bot token for validation

### Optional Variables

- `AUTH_SERVICE_HOST` - Service host (default: 0.0.0.0)
- `AUTH_SERVICE_PORT` - Public API port (default: 8080)
- `AUTH_INTERNAL_SERVICE_PORT` - Internal API port for health, metrics, admin endpoints (default: 8090)
- `DATABASE_MAX_CONNECTIONS` - Max DB connections (default: 10)
- `REDIS_MAX_CONNECTIONS` - Max Redis connections (default: 10)
- `JWT_PRIVATE_KEY_PATH` - JWT private key path (default: /etc/auth/private_key.pem)
- `JWT_PUBLIC_KEY_PATH` - JWT public key path (default: /etc/auth/public_key.pem)
- `JWT_ISSUER` - JWT issuer (default: shard-legends-auth)
- `JWT_EXPIRY_HOURS` - JWT expiry time in hours (default: 24)
- `RATE_LIMIT_REQUESTS` - Rate limit requests per window (default: 10)
- `RATE_LIMIT_WINDOW` - Rate limit window duration (default: 60s)
- `TOKEN_CLEANUP_INTERVAL_HOURS` - Token cleanup interval (default: 1)
- `TOKEN_CLEANUP_TIMEOUT_MINUTES` - Token cleanup timeout (default: 5)

## API Endpoints

The auth-service exposes two separate API interfaces:

### Public API (Port 8080)
- `/auth` - Authentication endpoint

### Internal API (Port 8090)
- `/health` - Health check endpoint
- `/metrics` - Prometheus metrics
- `/public-key.pem` - JWT public key in PEM format (for other microservices)
- `/admin/*` - Admin endpoints for token management

### GET /health

Health check endpoint that returns service status and dependencies (Internal API only).

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-12-21T10:30:00Z",
  "version": "1.0.0",
  "dependencies": {
    "postgresql": "not_configured",
    "redis": "not_configured",
    "jwt_keys": "not_configured"
  }
}
```


### GET /public-key.pem

Returns JWT public key in PEM format for simple integration (Internal API only).

**Response:**
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----
```

### GET /metrics

Prometheus metrics endpoint for monitoring and observability (Internal API only).

**Response:** Prometheus format metrics including HTTP requests, authentication metrics, JWT metrics, Redis/PostgreSQL operations, and system health indicators.

## Prometheus Metrics

The auth-service exposes comprehensive metrics for monitoring via the `/metrics` endpoint. All metrics use the `auth_service` namespace.

### HTTP Metrics
- `auth_service_http_requests_total` - Total HTTP requests (labels: method, endpoint, status_code)
- `auth_service_http_request_duration_seconds` - HTTP request duration histogram (labels: method, endpoint)
- `auth_service_http_requests_in_flight` - Current number of HTTP requests being processed

### Authentication Metrics
- `auth_service_auth_requests_total` - Total authentication requests (labels: status, reason)
- `auth_service_auth_request_duration_seconds` - Authentication request processing time (labels: status)
- `auth_service_auth_telegram_validation_duration_seconds` - Telegram signature validation time
- `auth_service_auth_new_users_total` - Total new user registrations
- `auth_service_auth_rate_limit_hits_total` - Rate limit violations (labels: ip)

### JWT Metrics
- `auth_service_jwt_tokens_generated_total` - Total JWT tokens generated
- `auth_service_jwt_tokens_validated_total` - Total JWT tokens validated (labels: status)
- `auth_service_jwt_key_generation_duration_seconds` - RSA key generation time
- `auth_service_jwt_active_tokens_count` - Current number of active tokens
- `auth_service_jwt_tokens_per_user` - Distribution of tokens per user

### Redis Metrics
- `auth_service_redis_operations_total` - Total Redis operations (labels: operation, status)
- `auth_service_redis_operation_duration_seconds` - Redis operation duration (labels: operation)
- `auth_service_redis_connection_pool_active` - Active Redis connections
- `auth_service_redis_connection_pool_idle` - Idle Redis connections
- `auth_service_redis_token_cleanup_duration_seconds` - Token cleanup operation duration
- `auth_service_redis_expired_tokens_cleaned_total` - Total expired tokens cleaned
- `auth_service_redis_cleanup_processed_users_total` - Total users processed during cleanup

### PostgreSQL Metrics
- `auth_service_postgres_operations_total` - Total PostgreSQL operations (labels: operation, table, status)
- `auth_service_postgres_operation_duration_seconds` - PostgreSQL operation duration (labels: operation, table)
- `auth_service_postgres_connection_pool_active` - Active PostgreSQL connections
- `auth_service_postgres_connection_pool_idle` - Idle PostgreSQL connections
- `auth_service_postgres_connection_pool_max` - Maximum PostgreSQL connections

### System Health Metrics
- `auth_service_service_up` - Service availability (1 = up, 0 = down)
- `auth_service_service_start_time_seconds` - Service start time as unix timestamp
- `auth_service_dependencies_healthy` - Dependency health status (labels: dependency)
- `auth_service_memory_usage_bytes` - Memory usage in bytes
- `auth_service_goroutines_count` - Number of active goroutines

### Admin Metrics
- `auth_service_admin_operations_total` - Total admin operations (labels: operation, status)
- `auth_service_admin_token_revocations_total` - Token revocations (labels: method)
- `auth_service_admin_cleanup_operations_total` - Manual cleanup operations

### Example Prometheus Queries

**Request Rate:**
```promql
rate(auth_service_http_requests_total[5m])
```

**Authentication Success Rate:**
```promql
rate(auth_service_auth_requests_total{status="success"}[5m]) / rate(auth_service_auth_requests_total[5m])
```

**P95 Response Time:**
```promql
histogram_quantile(0.95, rate(auth_service_http_request_duration_seconds_bucket[5m]))
```

**Active Token Count:**
```promql
auth_service_jwt_active_tokens_count
```

**Redis Connection Pool Usage:**
```promql
auth_service_redis_connection_pool_active / (auth_service_redis_connection_pool_active + auth_service_redis_connection_pool_idle)
```

### Recommended Alerting Rules

```yaml
groups:
- name: auth-service
  rules:
  - alert: AuthServiceDown
    expr: auth_service_service_up == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Auth service is down"

  - alert: HighAuthFailureRate
    expr: rate(auth_service_auth_requests_total{status="failed"}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High authentication failure rate"

  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(auth_service_http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High response time"

  - alert: DependencyUnhealthy
    expr: auth_service_dependencies_healthy == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "{{ $labels.dependency }} dependency is unhealthy"
```

## Development

### Prerequisites

- Go 1.22 or later
- PostgreSQL 17
- Redis 8.0.2

### Running Locally

1. Set environment variables:
```bash
export DATABASE_URL="postgresql://user:pass@localhost:5432/shard_legends"
export REDIS_URL="redis://localhost:6379/0"
export TELEGRAM_BOT_TOKEN="your_bot_token"
```

2. Run the service:
```bash
go run cmd/main.go
```

### Building

```bash
go build -o bin/auth-service ./cmd/main.go
```

### Testing

```bash
go test ./...
go vet ./...
go fmt ./...
```

## Docker

### Building

```bash
docker build -t auth-service .
```

### Running

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@postgres:5432/shard_legends" \
  -e REDIS_URL="redis://redis:6379/0" \
  -e TELEGRAM_BOT_TOKEN="your_bot_token" \
  -v auth_jwt_keys:/etc/auth \
  auth-service
```

**Note:** The volume mount is crucial for persistent JWT keys across container restarts.

## Project Structure

```
services/auth-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── handlers/
│   │   ├── auth.go          # Authentication handlers
│   │   └── health.go        # Health check handlers
│   ├── middleware/
│   │   └── jwt_public_key.go # JWT public key export middleware
│   ├── models/              # Data models (future)
│   ├── services/
│   │   ├── jwt.go           # JWT token service
│   │   ├── jwt_test.go      # JWT service tests
│   │   ├── telegram.go      # Telegram validation service
│   │   └── telegram_test.go # Telegram service tests
│   └── storage/             # Data access layer (future)
├── pkg/
│   └── utils/
│       └── logger.go        # Logging utilities
├── Dockerfile               # Container configuration
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
└── README.md               # This file
```

## Dependencies

For full API specification, see [auth-service-openapi.yml](../../docs/specs/auth-service-openapi.yml).

## License

Copyright (c) 2024 Shard Legends: Clan Wars
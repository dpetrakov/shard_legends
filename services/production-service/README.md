# Production Service

Production Service manages production recipes and manufacturing tasks in the Shard Legends game.

## Features

- Production recipe management
- Manufacturing task queue system
- Pre-calculation of production results with modifiers
- Integration with Inventory Service for item management
- JWT authentication via Auth Service

## Architecture

The service follows a layered architecture:
- **Handlers**: HTTP request handling (public, internal, admin endpoints)
- **Service**: Business logic layer
- **Storage**: Database operations
- **Models**: Data structures and DTOs

## Configuration

The Production Service uses a unified configuration system based on the [viper](https://github.com/spf13/viper) library. Configuration can be provided via:

1. **Environment variables** (recommended for production)
2. **YAML configuration file** (optional, see `config.sample.yaml`)

All environment variables use the `PROD_SVC_` prefix for consistency.

### Required Environment Variables

These environment variables must be set for the service to start:

- `PROD_SVC_DATABASE_URL`: PostgreSQL connection string
- `PROD_SVC_REDIS_URL`: Redis connection string for general caching  
- `PROD_SVC_REDIS_AUTH_URL`: Redis connection string for JWT revocation (typically database 0)
- `PROD_SVC_SERVER_PORT`: Public API port 
- `PROD_SVC_SERVER_INTERNAL_PORT`: Internal API port for health/metrics/admin
- `PROD_SVC_AUTH_PUBLIC_KEY_URL`: URL to fetch JWT public key from Auth Service
- `PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL`: Inventory Service base URL
- `PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL`: User Service base URL

### Optional Environment Variables

These have sensible defaults but can be overridden:

**Server Configuration:**
- `PROD_SVC_SERVER_HOST`: Service host (default: `0.0.0.0`)
- `PROD_SVC_SERVER_READ_TIMEOUT`: HTTP read timeout (default: `15s`)
- `PROD_SVC_SERVER_WRITE_TIMEOUT`: HTTP write timeout (default: `15s`)
- `PROD_SVC_SERVER_IDLE_TIMEOUT`: HTTP idle timeout (default: `60s`)

**Database Configuration:**
- `PROD_SVC_DATABASE_MAX_CONNECTIONS`: Max DB connections (default: `25`)
- `PROD_SVC_DATABASE_MAX_IDLE_TIME`: Max connection idle time (default: `5m`)
- `PROD_SVC_DATABASE_HEALTH_CHECK_PERIOD`: Health check period (default: `1m`)
- `PROD_SVC_DATABASE_PING_TIMEOUT`: DB ping timeout (default: `5s`)

**Redis Configuration:**
- `PROD_SVC_REDIS_MAX_CONNECTIONS`: Max Redis connections (default: `10`)
- `PROD_SVC_REDIS_READ_TIMEOUT`: Redis read timeout (default: `3s`)
- `PROD_SVC_REDIS_WRITE_TIMEOUT`: Redis write timeout (default: `3s`)
- `PROD_SVC_REDIS_MAX_RETRIES`: Redis max retries (default: `3`)
- `PROD_SVC_REDIS_PING_TIMEOUT`: Redis ping timeout (default: `5s`)

**Authentication Configuration:**
- `PROD_SVC_AUTH_CACHE_TTL`: JWT key cache TTL (default: `1h`)
- `PROD_SVC_AUTH_REFRESH_INTERVAL`: JWT key refresh interval (default: `24h`)

**Logging Configuration:**
- `PROD_SVC_LOGGING_LEVEL`: Logging level - debug, info, warn, error (default: `info`)

**External Service Timeouts:**
- `PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_TIMEOUT`: Inventory Service timeout (default: `10s`)
- `PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_TIMEOUT`: User Service timeout (default: `5s`)

**Application Timeouts:**
- `PROD_SVC_TIMEOUTS_HTTP_MIDDLEWARE`: HTTP middleware timeout (default: `60s`)
- `PROD_SVC_TIMEOUTS_JWT_VALIDATOR_CLIENT`: JWT validator timeout (default: `10s`)
- `PROD_SVC_TIMEOUTS_GRACEFUL_SHUTDOWN`: Graceful shutdown timeout (default: `30s`)
- `PROD_SVC_TIMEOUTS_DATABASE_HEALTH`: Database health check timeout (default: `2s`)
- `PROD_SVC_TIMEOUTS_REDIS_HEALTH`: Redis health check timeout (default: `2s`)

**Background Cleanup Configuration:**
- `PROD_SVC_CLEANUP_ORPHANED_TASK_TIMEOUT`: Timeout for orphaned DRAFT tasks (default: `5m`)
- `PROD_SVC_CLEANUP_CLEANUP_INTERVAL`: Cleanup process interval (default: `5m`)
- `PROD_SVC_CLEANUP_CLEANUP_TIMEOUT`: Cleanup operation timeout (default: `5m`)

**Metrics Configuration:**
- `PROD_SVC_METRICS_UPDATE_INTERVAL`: Metrics update interval (default: `10s`)

### Configuration Files

You can optionally provide configuration via YAML files. The service searches for configuration files in:

1. Current directory (`./config.yaml`)
2. `./configs/` directory
3. `/etc/production-service/` directory

Environment variables always take precedence over configuration file values.

Example configuration file (`config.yaml`):

```yaml
# See config.sample.yaml for a complete example
server:
  port: "8080"
  internal_port: "8081"
  host: "0.0.0.0"
  
database:
  url: "postgres://user:pass@localhost:5432/db"
  max_connections: 25
  
redis:
  url: "redis://localhost:6379/1"
  auth_url: "redis://localhost:6379/0"
  max_connections: 10

logging:
  level: "info"
```

### Fail-Fast Validation

The service implements fail-fast validation during startup. If any required configuration is missing or invalid, the service will exit immediately with a descriptive error message indicating exactly which configuration field needs to be set.

## API Endpoints

See `docs/specs/production-service-openapi.yml` for complete API documentation.

### Public Endpoints (Public Port, JWT required)
- `GET /production/recipes` - Get available recipes
- `GET /production/factory/queue` - Get user's production queue
- `POST /production/factory/start` - Start production task
- `POST /production/factory/claim` - Claim completed task

### Internal Endpoints (Internal Port)
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/internal/task/{taskId}` - Get task details
- `GET /api/v1/internal/recipe/{recipeId}` - Get recipe details

### Admin Endpoints (Internal Port, JWT with admin role)
- `GET /api/v1/admin/tasks` - List all tasks
- `GET /api/v1/admin/stats` - Production statistics

## Development

```bash
# Run locally (set required environment variables first)
export PROD_SVC_DATABASE_URL="postgres://user:pass@localhost:5432/shard_legends_dev"
export PROD_SVC_REDIS_URL="redis://localhost:6379/2"
export PROD_SVC_REDIS_AUTH_URL="redis://localhost:6379/0"
export PROD_SVC_SERVER_PORT="8080"
export PROD_SVC_SERVER_INTERNAL_PORT="8081"
export PROD_SVC_AUTH_PUBLIC_KEY_URL="http://localhost:8080/public-key.pem"
export PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL="http://localhost:8081"
export PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL="http://localhost:8082"

go run cmd/server/main.go

# Or use the provided .env.dev file
cp .env.dev .env
# Edit .env with your specific configuration
source .env
go run cmd/server/main.go

# Test endpoints
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/production/recipes
curl http://localhost:8081/health
curl http://localhost:8081/metrics

# Run tests
go test ./...

# Run configuration tests specifically
go test ./internal/config -v

# Run with Docker (using environment file)
docker build -t production-service .
docker run -p 8080:8080 -p 8081:8081 --env-file .env production-service
```

## Dependencies

- Auth Service: JWT token validation
- Inventory Service: Item management
- PostgreSQL: Main database
- Redis: Caching and JWT revocation
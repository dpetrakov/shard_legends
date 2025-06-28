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

Environment variables:
- `PRODUCTION_SERVICE_HOST`: Service host (default: 0.0.0.0)
- `PROD_SVC_PUBLIC_PORT`: Public API port (required, no default)
- `PROD_SVC_INTERNAL_PORT`: Internal API port for health/metrics/admin (required, no default)
- `DATABASE_URL`: PostgreSQL connection string (required)
- `REDIS_URL`: Redis connection string (required)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `DATABASE_MAX_CONNECTIONS`: Max DB connections (default: 25)
- `REDIS_MAX_CONNECTIONS`: Max Redis connections (default: 10)

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
# Run locally
PROD_SVC_PUBLIC_PORT=8082 PROD_SVC_INTERNAL_PORT=8091 go run cmd/server/main.go

# Test endpoints
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8082/production/recipes
curl http://localhost:8091/health
curl http://localhost:8091/metrics

# Run tests
go test ./...

# Run with Docker
docker build -t production-service .
docker run -p 8082:8082 -p 8091:8091 \
  -e PROD_SVC_PUBLIC_PORT=8082 \
  -e PROD_SVC_INTERNAL_PORT=8091 \
  production-service
```

## Dependencies

- Auth Service: JWT token validation
- Inventory Service: Item management
- PostgreSQL: Main database
- Redis: Caching and JWT revocation
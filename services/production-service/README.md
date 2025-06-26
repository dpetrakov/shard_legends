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
- `PRODUCTION_SERVICE_PORT`: Service port (default: 8082)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `DATABASE_MAX_CONNECTIONS`: Max DB connections (default: 25)
- `REDIS_MAX_CONNECTIONS`: Max Redis connections (default: 10)

## API Endpoints

See `docs/specs/production-service-openapi.yml` for complete API documentation.

### Public Endpoints (JWT required)
- `GET /recipes` - Get available recipes
- `GET /factory/queue` - Get user's production queue
- `POST /factory/start` - Start production task
- `POST /factory/claim` - Claim completed task

### Internal Endpoints
- `GET /internal/task/{taskId}` - Get task details
- `GET /internal/recipe/{recipeId}` - Get recipe details

### Admin Endpoints (JWT with admin role)
- `GET /admin/tasks` - List all tasks
- `GET /admin/stats` - Production statistics

## Development

```bash
# Run locally
go run cmd/server/main.go

# Run tests
go test ./...

# Run with Docker
docker build -t production-service .
docker run -p 8082:8082 production-service
```

## Dependencies

- Auth Service: JWT token validation
- Inventory Service: Item management
- PostgreSQL: Main database
- Redis: Caching and JWT revocation
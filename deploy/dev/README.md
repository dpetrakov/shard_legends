# Development Environment Setup

This directory contains the Docker Compose configuration for the development environment of Shard Legends: Clan Wars.

## Services

### Infrastructure Services

#### PostgreSQL 17
- **Container**: `slcw-postgres-dev`
- **Image**: `postgres:17-alpine`
- **Database**: `shard_legends_dev`
- **User**: `slcw_user`
- **Network**: Internal only (no external ports)
- **Schemas**: `auth`, `game`, `clan`

#### Redis 8.0.2
- **Container**: `slcw-redis-dev`
- **Image**: `redis:8.0.2-alpine`
- **Network**: Internal only (no external ports)
- **Persistence**: RDB + AOF enabled for JWT token reliability
- **Memory**: 256MB limit with TTL-based eviction policy

### Application Services

#### API Gateway
- **Container**: `slcw-api-gateway-dev`
- **Port**: `127.0.0.1:9000`
- **Role**: Routes requests to microservices

#### Frontend
- **Container**: `slcw-frontend-dev`
- **Port**: `127.0.0.1:8092`
- **Framework**: Next.js

#### Microservices
- **Ping Service**: Test microservice
- **Telegram Bot Service**: Main bot service
- **Telegram Bot Service (Forly)**: Optional secondary bot

## Quick Start

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Start infrastructure services:
   ```bash
   docker-compose up -d postgres redis
   ```

3. Start all services:
   ```bash
   docker-compose up -d
   ```

4. Access:
   - Frontend: http://localhost:8092
   - API Gateway: http://localhost:9000
   - Domain: https://dev.slcw.dimlight.online

## Database Connection

### PostgreSQL
```bash
# Connect to PostgreSQL
docker exec -it slcw-postgres-dev psql -U slcw_user -d shard_legends_dev

# Environment variables
POSTGRES_DB=shard_legends_dev
POSTGRES_USER=slcw_user
POSTGRES_PASSWORD=dev_password_2024
```

### Redis
```bash
# Connect to Redis
docker exec -it slcw-redis-dev redis-cli

# Test connection
docker exec slcw-redis-dev redis-cli ping
```

## Management Commands

### Infrastructure Only
```bash
# Start PostgreSQL and Redis only
docker-compose up -d postgres redis

# Check status
docker-compose ps postgres redis
```

### All Services
```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f [service-name]

# Stop services
docker-compose down
```

### Health Checks
All services include health checks:
- PostgreSQL: `pg_isready` check
- Redis: `redis-cli ping` check
- Other services: HTTP endpoint checks

## Persistence

### PostgreSQL
- Data persisted in named volume: `slcw-postgres-dev`
- Initialization scripts in `./postgres/init/`

### Redis
- Data persisted in named volume: `slcw-redis-dev`
- Configuration: `./redis/redis.conf`
- RDB snapshots: Every 900s/1 key, 300s/10 keys, 60s/10000 keys
- AOF enabled with `everysec` sync for JWT token reliability

## Network

All services run on the internal network `slcw-dev`. Only the API Gateway and Frontend expose external ports for development access.

## Environment Variables

Configuration is managed through `.env` file. Key variables:

- `POSTGRES_DB`: Database name (default: shard_legends_dev)
- `POSTGRES_USER`: Database user (default: slcw_user)  
- `POSTGRES_PASSWORD`: Database password
- `TELEGRAM_BOT_TOKEN`: Telegram bot token
- `WEBAPP_BASE_URL`: Base URL for web app
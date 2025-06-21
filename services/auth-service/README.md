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

## Configuration

The service is configured through environment variables:

### Required Variables

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string  
- `TELEGRAM_BOT_TOKEN` - Telegram bot token for validation

### Optional Variables

- `AUTH_SERVICE_HOST` - Service host (default: 0.0.0.0)
- `AUTH_SERVICE_PORT` - Service port (default: 8080)
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

### GET /health

Health check endpoint that returns service status and dependencies.

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

### GET /jwks

Returns JWT public key in JWKS (JSON Web Key Set) format for other services.

**Response:**
```json
{
  "keys": [{
    "kty": "RSA",
    "use": "sig",
    "alg": "RS256", 
    "kid": "key_fingerprint",
    "pem": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
  }]
}
```

### GET /public-key.pem

Returns JWT public key in PEM format for simple integration.

**Response:**
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----
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
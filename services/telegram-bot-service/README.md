# Telegram Bot Service

This service provides Telegram bot functionality for the Shard Legends: Clan Wars game.

## Configuration

The service is configured via environment variables:

### Required Variables
- `TELEGRAM_BOT_TOKEN` - Bot token from BotFather
- `WEBAPP_BASE_URL` - Base URL of the web application

### Mode-specific Variables

#### Webhook Mode (`TELEGRAM_BOT_MODE=webhook`)
- `TELEGRAM_WEBHOOK_URL` - Public HTTPS URL for webhook endpoint
- `SERVICE_PORT` - Port for HTTP server (default: 8080)

#### Long Polling Mode (`TELEGRAM_BOT_MODE=longpoll` or not specified)
- `TELEGRAM_POLL_TIMEOUT` - Timeout for getUpdates API calls in seconds (default: 30, max: 60)

## Running the Service

### Development
```bash
# Create .env file with your configuration
cp .env.example .env

# Run the service
go run main.go
```

### Testing
```bash
# Run all tests with coverage
go test ./... -cover -v

# Run specific package tests
go test ./internal/config -cover -v
```

### Building
```bash
# Build binary
go build -o telegram-bot-service .

# Run binary
./telegram-bot-service
```

## Project Structure
```
.
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── internal/
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handlers/          # HTTP handlers (webhook mode)
│   ├── telegram/          # Telegram bot logic
│   └── models/            # Data models
└── README.md
```

## Modes of Operation

### Webhook Mode
- Service runs an HTTP server listening on `SERVICE_PORT`
- Telegram sends updates to `TELEGRAM_WEBHOOK_URL`
- Webhook is automatically set on startup
- Ideal for production environments

### Long Polling Mode
- Service actively polls Telegram for updates
- No public endpoint required
- Webhook is automatically removed on startup
- Ideal for development and testing

## TODO
- [ ] **ВАЖНО**: Протестировать webhook режим работы с реальным Telegram API
- [ ] Реализовать команды бота (/start, deep links, echo mode)
- [ ] Добавить интеграцию с Telegram Bot API
- [ ] Провести полное интеграционное тестирование
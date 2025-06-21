# Telegram Bot Service - Спецификация для разработки

## Сервис-дескриптор
- **Название**: telegram-bot-service
- **Язык**: Golang  
- **Порт**: 8080 (внутренний контейнер)
- **Тип**: HTTP сервер + Telegram Webhook
- **Развертывание**: Docker контейнер в составе микросервисной архитектуры

## Окружения и именование ботов

### Стенды и соответствующие боты:
- **dev**: `SLCWDevBot` - основной dev стенд
- **dev-dim**: `SLCWDevDimBot` - персональный dev стенд Dim  
- **dev-forly**: `SLCWDevForlyBot` - персональный dev стенд Forly
- **stage**: `SLCWStageBot` - staging окружение
- **prod**: `SLCWBot` - продуктивный бот

### Схема переменных окружения:
```
TELEGRAM_BOT_TOKEN=bot_token_here
TELEGRAM_BOT_MODE=webhook_или_longpoll
TELEGRAM_POLL_TIMEOUT=30  // только для longpoll режима
WEBAPP_BASE_URL=webapp_url_here
TELEGRAM_WEBHOOK_URL=webhook_url_here  // только для webhook режима  
TELEGRAM_SECRET_TOKEN=secret_token_here  // только для webhook режима (опционально)
TELEGRAM_ALLOWED_USERS=username1,username2  // список разрешенных username (опционально)
SERVICE_PORT=8080  // только для webhook режима
```

## Основная функциональность

### Core Features:
1. **Update Processor** - обработка входящих Telegram обновлений (webhook/longpoll)
2. **User Access Control** - ограничение доступа к боту по списку разрешенных username
3. **Echo Handler** - отправка обратно всех полученных сообщений
4. **Basic Commands** - обработка команды /start
5. **Mini App Launcher** - запуск веб-приложения через команды

### API Endpoints:
- `POST /webhook` - Telegram webhook endpoint (только в webhook режиме)
- `GET /health` - health check

## Telegram Commands

### Базовые команды:
- `/start` - запуск мини-приложения с приветственным сообщением

### Deep linking параметры:
- `/start game` - запуск игры (передается параметр в веб-приложение)

### Echo Mode:
- Все остальные сообщения (текст, стикеры, медиа) отправляются обратно пользователю

## Интеграция с Backend API

**На текущем этапе интеграция с backend не реализуется**

Бот работает в режиме эхо без обращений к внешним API.

## Конфигурация по окружениям

### Port mapping (согласно nginx конфигурации):
- **dev**: 9003 → 8080
- **dev-dim**: 10083 → 8080  
- **dev-forly**: 9003 → 8080 (shared)
- **stage**: 7003 → 8080
- **prod**: 5003 → 8080

### Webhook URLs (при TELEGRAM_BOT_MODE=webhook):
- **dev**: `https://dev.slcw.dimlight.online/api/telegram-bot/webhook`
- **dev-dim**: `https://dev-dim.slcw.dimlight.online/api/telegram-bot/webhook`
- **dev-forly**: `https://dev-forly.slcw.dimlight.online/api/telegram-bot/webhook`
- **stage**: `https://stage.slcw.dimlight.online/api/telegram-bot/webhook`
- **prod**: `https://slcw.dimlight.online/api/telegram-bot/webhook`

### WebApp URLs:
- **dev**: `https://dev.slcw.dimlight.online`
- **dev-dim**: `https://dev-dim.slcw.dimlight.online`
- **dev-forly**: `https://dev-forly.slcw.dimlight.online`
- **stage**: `https://stage.slcw.dimlight.online`
- **prod**: `https://slcw.dimlight.online`

## Структура проекта

```
services/telegram-bot-service/
├── main.go
├── Dockerfile
├── go.mod
├── go.sum
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   ├── webhook.go
│   │   └── health.go
│   ├── telegram/
│   │   ├── bot.go
│   │   ├── commands.go
│   │   ├── echo.go
│   │   ├── webhook.go
│   │   ├── longpoll.go
│   │   └── webapp.go
│   └── models/
│       └── update.go
└── README.md
```

## Зависимости

### Required Go packages:
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/gorilla/mux` - HTTP router (только для webhook режима)
- `github.com/joho/godotenv` - environment variables
- Standard library: net/http, encoding/json, log, os, strings

## Error Handling

### Telegram API errors:
- Rate limiting - exponential backoff
- Invalid token - service shutdown с error log
- Webhook validation fail - 400 response
- Network errors - retry mechanism

### Backend API errors:  
- Connection timeout - fallback response
- 5xx errors - retry with backoff
- Authentication errors - user notification
- Validation errors - user-friendly message

## Monitoring Requirements

### Metrics to track:
- Webhook requests per minute
- Command usage statistics  
- User registration rate
- API response times
- Error rates по типам

### Logs format:
- JSON structured logging
- Request/response correlation IDs
- User actions tracking
- Error context с stack traces

## Security

### Webhook validation:
- Telegram secret token verification через X-Telegram-Bot-Api-Secret-Token header
- Request content validation
- IP whitelist (опционально для дополнительной защиты)

### User access control:
- Username whitelist через TELEGRAM_ALLOWED_USERS переменную окружения
- Если переменная не задана - доступ разрешен всем пользователям
- Если переменная задана - доступ только указанным username (разделение запятыми)
- Неавторизованные пользователи получают сообщение о недоступности бота

### Data protection:
- No storage of sensitive user data
- Telegram user ID as primary identifier
- Secure token передача to backend
- No logging of tokens/credentials

## Deployment

### Docker configuration:
- Multi-stage build (builder + runtime)
- Non-root user execution
- Health check endpoint
- Graceful shutdown handling

### Environment-specific settings:
- Отдельные bot tokens для каждого стенда
- Соответствующие webhook URLs
- Environment-specific logging levels
- Feature flags для dev/prod различий

## Режимы работы

### Webhook Mode (TELEGRAM_BOT_MODE=webhook):
- HTTP сервер принимает webhook от Telegram
- Требует публичный HTTPS endpoint  
- Мгновенная доставка обновлений
- Используется в production и staging

### Long Polling Mode (TELEGRAM_BOT_MODE=longpoll):
- Активное получение обновлений через getUpdates API
- Не требует публичного endpoint
- Задержка до POLL_TIMEOUT секунд
- Используется для локальной разработки и отладки

### Автоматическое переключение:
- При запуске в webhook режиме автоматически устанавливается webhook
- При запуске в longpoll режиме автоматически удаляется webhook
- Graceful shutdown корректно очищает ресурсы

## Performance Requirements

### Response times:
- Webhook processing: < 100ms
- Command responses: < 200ms  
- Health check: < 50ms
- Long polling timeout: 30-60s (настраивается)

### Scalability:
- Single instance per environment (current)
- Stateless design для будущего масштабирования
- Connection pooling для backend API
- Rate limiting compliance с Telegram limits
# Telegram Bot Service

Telegram бот для игры Shard Legends: Clan Wars. Поддерживает работу в webhook и longpoll режимах с системой контроля доступа пользователей.

## Особенности

- **Два режима работы**: webhook (продуктив) и longpoll (разработка)
- **Контроль доступа**: whitelist пользователей по username
- **Безопасность**: поддержка Telegram Secret Token для webhook
- **Echo режим**: отправка обратно всех сообщений
- **WebApp интеграция**: запуск мини-приложения через команды
- **Health check**: мониторинг состояния сервиса

## Конфигурация

### Обязательные переменные
```bash
TELEGRAM_BOT_TOKEN=bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
WEBAPP_BASE_URL=https://example.com
```

### Режим работы
```bash
# Webhook режим (по умолчанию для продуктива)
TELEGRAM_BOT_MODE=webhook
TELEGRAM_WEBHOOK_URL=https://example.com/api/webhook
SERVICE_PORT=8080

# Long polling режим (по умолчанию для разработки)  
TELEGRAM_BOT_MODE=longpoll
TELEGRAM_POLL_TIMEOUT=30
```

### Безопасность (опционально)
```bash
# Secret token для webhook (рекомендуется)
TELEGRAM_SECRET_TOKEN=your_secret_token_here

# Whitelist пользователей (если не задано - доступ всем)
TELEGRAM_ALLOWED_USERS=username1,username2,username3
```

## Быстрый старт

## Команды бота

### /start
Запускает мини-приложение с приветственным сообщением и inline кнопкой WebApp.

```
/start          # Обычный запуск
/start game     # Deep link параметр передается в WebApp
```

### Echo режим
Все остальные сообщения (текст, стикеры, медиа) отправляются обратно пользователю.

## API Endpoints

- `POST /webhook` - Telegram webhook (только в webhook режиме)
- `GET /health` - Health check для мониторинга

## Структура проекта

```
telegram-bot-service/
├── main.go                    # Точка входа
├── Dockerfile                 # Docker образ
├── go.mod, go.sum            # Go зависимости
├── internal/
│   ├── config/               # Управление конфигурацией
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handlers/             # HTTP обработчики
│   │   ├── health.go         # Health check
│   │   ├── webhook.go        # Webhook endpoint
│   │   └── webhook_test.go
│   └── telegram/             # Telegram бот логика
│       ├── bot.go            # Основной бот
│       ├── bot_test.go
│       ├── commands.go       # Команды (/start)
│       ├── commands_test.go
│       ├── echo.go           # Echo обработчик
│       └── echo_test.go
└── README.md
```

## Режимы работы

### Webhook Mode
- Telegram отправляет обновления на HTTP endpoint
- Требует публичный HTTPS URL
- Мгновенная доставка сообщений
- Автоматическая установка/удаление webhook
- Поддержка Secret Token для безопасности

### Long Polling Mode  
- Активное получение обновлений через API
- Не требует публичного endpoint
- Подходит для разработки и отладки
- Настраиваемый timeout (0-60 секунд)

## Контроль доступа

### Без ограничений
Если `TELEGRAM_ALLOWED_USERS` не задана - бот доступен всем пользователям.

### С whitelist
```bash
TELEGRAM_ALLOWED_USERS=developer1,admin,tester
```
Доступ только указанным username. Пользователи без username получают отказ.

### Сообщения при отказе
Неавторизованные пользователи получают:
> "Извините, доступ к этому боту ограничен. Обратитесь к администратору."

## Развертывание по окружениям

## Мониторинг

### Health Check
```bash
curl http://localhost:8080/health
# Ответ: {"status": "healthy"}
```

### Логирование
- Структурированные логи с timestamp
- Отслеживание команд и пользователей
- Ошибки с контекстом
- Безопасность: токены не логируются

### Метрики (планируется)
- Количество webhook запросов
- Статистика использования команд
- Время ответа API
- Количество ошибок по типам

## Безопасность

### Webhook защита
- Проверка Secret Token (X-Telegram-Bot-Api-Secret-Token header)
- Валидация содержимого запроса
- Возврат 401 при неверном токене

### Данные пользователей
- Не сохраняются персональные данные
- Используется Telegram user ID как идентификатор
- Логи не содержат чувствительной информации

## Устранение неполадок

### Проблемы с webhook
```bash
# Проверить статус webhook
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"

# Удалить webhook принудительно
curl -X POST "https://api.telegram.org/bot<TOKEN>/deleteWebhook"
```

### Проблемы с доступом
1. Проверить `TELEGRAM_ALLOWED_USERS` в логах
2. Убедиться что у пользователя есть username
3. Проверить регистр символов в username

### Отладка
```bash
# Включить подробное логирование
export LOG_LEVEL=debug

# Запустить в longpoll для отладки
export TELEGRAM_BOT_MODE=longpoll
go run main.go
```

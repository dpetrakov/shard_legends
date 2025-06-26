# User Service (Temporary Mock Version)

**⚠️ ВРЕМЕННАЯ ВЕРСИЯ**: Этот сервис возвращает частично заглушенные данные и не имеет собственной базы данных. Предназначен для обеспечения базовой интеграции с Production Service.

## Обзор

User Service — микросервис для управления пользователями в игре Shard Legends: Clan Wars. В текущей версии предоставляет моковые данные для профиля пользователя и производственных слотов.

## Возможности

### Публичные эндпоинты (с JWT аутентификацией)
- `GET /profile` - Данные профиля пользователя
- `GET /production-slots` - Информация о производственных слотах

### Внутренние эндпоинты (для других сервисов)
- `GET /internal/users/{user_id}/production-slots` - Слоты для Production Service
- `GET /internal/users/{user_id}/production-modifiers` - Модификаторы для Production Service

### Системные эндпоинты
- `GET /health` - Health check

## Конфигурация

### Переменные окружения

```bash
# Сервер
USER_SERVICE_PORT=8080
USER_SERVICE_HOST=0.0.0.0

# Аутентификация
AUTH_SERVICE_PUBLIC_KEY_URL=http://auth-service:8080/public-key.pem
REDIS_URL=redis://redis:6379/0
```

## Особенности временной версии

1. **Отсутствие собственной БД** - все данные генерируются в runtime
2. **VIP статус** - случайная генерация (30% вероятность активного VIP)
3. **Производственные слоты** - фиксированно 2 универсальных слота
4. **Модификаторы** - всегда нулевые значения
5. **Профильные данные** - статические заглушки

## Запуск

### Docker (рекомендуется)
```bash
docker build -t user-service .
docker run -p 8080:8080 user-service
```

### Локальная разработка
```bash
go mod download
go run ./cmd/server/main.go
```

## Интеграция

### JWT Аутентификация
Сервис использует публичный ключ от Auth Service для валидации JWT токенов:
- Получение ключа: `GET http://auth-service:8080/public-key.pem`
- Проверка отзыва токенов в Redis: `EXISTS revoked:{jti}`

### Зависимости
- **Auth Service** - для валидации JWT токенов
- **Redis** - для проверки отозванных токенов

## Примеры запросов

### Получение профиля
```bash
curl -H "Authorization: Bearer <jwt_token>" \
     http://user-service:8080/profile
```

### Получение слотов (внутренний API)
```bash
curl http://user-service:8080/internal/users/{user_id}/production-slots
```

## Health Check

```bash
curl http://user-service:8080/health
```

Ответ:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-26T10:30:00Z",
  "version": "1.0.0",
  "service": "user-service"
}
```

## Логирование

Сервис использует structured JSON логирование:
- Все HTTP запросы логируются
- JWT аутентификация логируется
- Ошибки детально документируются

## Планируемые улучшения

В будущих версиях будут добавлены:
- PostgreSQL для персистентного хранения
- Реальная VIP система
- Система уровней и опыта
- Достижения и награды
- Клановая система
- Метрики и мониторинг

---

**Версия**: 1.0.0 (временная)  
**Статус**: Готов для интеграции с Production Service
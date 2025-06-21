# API Gateway Architecture

## Обзор

Создан внутренний API Gateway для объединения всех микросервисов под единым `/api/*` эндпоинтом.

## Архитектура

### До оптимизации:
```
External Nginx → ping-service:9001/ping
               → telegram-bot:9003/webhook
               → auth-service:9002/auth
               → game-service:9005/game
```

### После оптимизации:
```
External Nginx /api/ping    → API Gateway:9000 /ping     → ping-service:8080/ping
               /api/webhook → API Gateway:9000 /webhook  → telegram-bot:8080/webhook
               /api/auth    → API Gateway:9000 /auth     → auth-service:8080/auth
               /api/*       → API Gateway:9000 /*        → [services]:8080/*
```

**Важно:** External nginx убирает префикс `/api` при проксировании в API Gateway.

## Преимущества

1. **Единая точка входа** - один порт (9000) вместо множества
2. **Упрощение внешнего nginx** - нет сложных map-based роутингов
3. **Централизованное логирование** API запросов
4. **Легкое масштабирование** - добавление новых сервисов только в API Gateway
5. **Безопасность** - микросервисы не доступны извне напрямую

## Доступные эндпоинты

- `/api/ping` → ping-service
- `/api/webhook` → telegram-bot-service

**Примечание:** `/health` эндпоинты используются только для внутренних Docker health check'ов и не выставляются наружу через API Gateway.

## Миграция

### 1. Обновить docker-compose.yml:
```bash
cd deploy/dev
docker-compose down
docker-compose up -d
```

### 2. Обновить внешний nginx:
Заменить `deploy/nginx/slcw.conf` на `deploy/nginx/slcw-optimized.conf`

### 3. Проверить работу:
```bash
# Через API Gateway
curl https://dev.slcw.dimlight.online/api/ping

# Прямое обращение к API Gateway
curl http://localhost:9000/api/ping
curl http://localhost:9000/health  # Health check самого gateway
```

## Добавление новых микросервисов

1. Добавить upstream в `api-gateway/nginx.conf`:
```nginx
upstream new_service {
    server new-service:8080;
}
```

2. Добавить location:
```nginx
location /api/newservice {
    proxy_pass http://new_service/newservice;
}
```

3. Пересобрать API Gateway:
```bash
docker-compose build api-gateway
docker-compose up -d api-gateway
```

## Логи

```bash
# API Gateway логи
docker logs slcw-api-gateway-dev

# Логи микросервисов (теперь только внутренние)
docker logs slcw-ping-service-dev
docker logs slcw-telegram-bot-dev
```

## Масштабирование

Для high-load можно запустить несколько инстансов API Gateway:
```yaml
api-gateway-1:
  # ... config
  ports: ["127.0.0.1:9000:8080"]
api-gateway-2:
  # ... config  
  ports: ["127.0.0.1:9001:8080"]
```

И настроить балансировку в внешнем nginx.
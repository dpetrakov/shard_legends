# SLCW Frontend - Development Environment

Этот каталог содержит конфигурацию Docker для запуска фронтенда в dev окружении.

## Быстрый старт

1. Скопируйте `.env.example` в `.env` и заполните переменные окружения:
   ```bash
   cp .env.example .env
   ```

2. Запустите фронтенд:
   ```bash
   ./start.sh
   ```

3. Доступ к приложению:
   - Напрямую: http://localhost:8092
   - Через nginx: https://dev.slcw.dimlight.online

## Доступные скрипты

- `./start.sh` - запуск контейнеров
- `./stop.sh` - остановка контейнеров
- `./logs.sh` - просмотр логов

## Структура

- `frontend.Dockerfile` - Dockerfile для сборки фронтенда
- `docker-compose.yml` - конфигурация Docker Compose
- `.env.example` - пример файла с переменными окружения

## Особенности dev окружения

- Hot reload включен через volume монтирование
- Порт 8092 (соответствует nginx конфигурации для dev)
- API endpoint: http://localhost:8082/api
- Автоматический рестарт при падении
- Health check для мониторинга состояния

## Интеграция с nginx

Конфигурация соответствует `deploy/nginx/slcw.conf`:
- Frontend dev: порт 8092
- API dev: порт 8082
- Домен: dev.slcw.dimlight.online
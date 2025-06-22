# Настройка логирования для Auth Service в Grafana

## Обзор

Эта конфигурация предоставляет удобный интерфейс для мониторинга и анализа логов auth-service через Grafana Loki.

## Основные возможности

### 1. Главный дашборд Auth Service Logs Explorer

Содержит следующие панели:

1. **Auth Service Logs Explorer** - Основной просмотр логов с фильтрацией
   - Фильтр по тексту (переменная `$filter`)
   - Фильтр по уровню логов (DEBUG, INFO, WARN, ERROR)
   - JSON парсинг для структурированных логов

2. **Log Levels Distribution** - График распределения уровней логов во времени
   - Визуализация трендов ошибок
   - Цветовая кодировка: ERROR (красный), WARN (желтый), INFO (зеленый), DEBUG (синий)

3. **Top 10 Errors** - Таблица с наиболее частыми ошибками
   - Автоматическая группировка по сообщению об ошибке
   - Счетчик количества вхождений

4. **JWT Operations** - Специализированная панель для JWT операций
   - Отслеживание генерации токенов
   - Валидация токенов
   - Ошибки JWT

5. **Telegram Authentication** - Мониторинг Telegram авторизации
   - Успешные и неуспешные попытки
   - Валидация подписей
   - Информация о пользователях

6. **Database Operations** - Операции с PostgreSQL
   - CRUD операции с пользователями
   - Ошибки подключения
   - Производительность запросов

7. **HTTP Requests** - HTTP эндпоинты
   - Методы и пути запросов
   - Статус-коды ответов
   - Время обработки

## Использование

### Базовые запросы LogQL

1. **Все логи auth-service:**
   ```logql
   {container="slcw-auth-service"}
   ```

2. **Только ошибки:**
   ```logql
   {container="slcw-auth-service"} | json | level="ERROR"
   ```

3. **JWT операции:**
   ```logql
   {container="slcw-auth-service"} | json | msg =~ ".*JWT.*"
   ```

4. **Поиск по user_id:**
   ```logql
   {container="slcw-auth-service"} | json | user_id="specific-uuid"
   ```

5. **Telegram авторизация:**
   ```logql
   {container="slcw-auth-service"} | json | msg =~ ".*Telegram.*|.*auth.*"
   ```

### Расширенные запросы

1. **Ошибки JWT валидации:**
   ```logql
   {container="slcw-auth-service"} 
   | json 
   | msg="JWT token claims validation failed" 
   | line_format "{{.time}} {{.error}} user_id={{.user_id}}"
   ```

2. **Неуспешные попытки авторизации:**
   ```logql
   {container="slcw-auth-service"} 
   | json 
   | msg="Telegram data validation failed" 
   | line_format "{{.time}} error={{.error}}"
   ```

3. **База данных - создание пользователей:**
   ```logql
   {container="slcw-auth-service"} 
   | json 
   | msg="User created successfully"
   | line_format "{{.time}} user_id={{.user_id}} telegram_id={{.telegram_id}}"
   ```

4. **Статистика по операциям (rate):**
   ```logql
   sum by (operation) (
     rate({container="slcw-auth-service"} 
     | json 
     | regexp "(?P<operation>JWT|Telegram|PostgreSQL)" 
     | __error__="" [$__interval])
   )
   ```

## Настройка алертов

### Пример алерта на критические ошибки:

```yaml
groups:
  - name: auth_service_alerts
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate({container="slcw-auth-service"} | json | level="ERROR" [5m])) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in auth-service"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: JWTValidationFailures
        expr: |
          sum(rate({container="slcw-auth-service"} 
          | json 
          | msg="JWT token claims validation failed" [5m])) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High JWT validation failure rate"
          description: "JWT validation failures: {{ $value }} failures/sec"
```

## Оптимизация производительности

### 1. Использование индексов Loki
Для часто используемых запросов можно создать статические labels:

```yaml
- labels:
    operation_type: 
      source: msg
      regex: "^(JWT|Telegram|Database|HTTP).*"
```

### 2. Ограничение времени запроса
Используйте временные ограничения для больших запросов:

```logql
{container="slcw-auth-service"} | json | level="ERROR" 
  | __error__="" 
  | line_format "{{.msg}}" 
  | pattern "<_> error=<error>" 
  | error != ""
```

### 3. Использование LogQL pattern parser
Для быстрого извлечения данных без полного JSON парсинга:

```logql
{container="slcw-auth-service"} 
  | pattern "<_> level=<level> <_> msg=<msg> <_>"
  | level="ERROR"
```

## Интеграция с другими сервисами

### Корреляция с метриками Prometheus:

1. **Сопоставление ошибок с нагрузкой:**
   - Используйте временные метки для сопоставления всплесков ошибок с метриками CPU/Memory
   - Панель "Mixed" в Grafana для объединения логов и метрик

2. **Trace ID для распределенной трассировки:**
   ```logql
   {container=~"slcw-.*"} | json | trace_id="specific-trace-id"
   ```

## Troubleshooting

### Проблема: Логи не появляются
1. Проверьте, что контейнер запущен: `docker ps | grep auth-service`
2. Проверьте Promtail: `docker logs slcw-promtail`
3. Проверьте Loki: `curl http://localhost:15100/ready`

### Проблема: JSON парсинг не работает
1. Убедитесь, что логи в правильном формате slog JSON
2. Проверьте наличие `| json` в запросе
3. Используйте `| __error__=""` для фильтрации ошибок парсинга

### Проблема: Медленные запросы
1. Ограничьте временной диапазон
2. Используйте более специфичные селекторы
3. Добавьте индексы через labels в Promtail конфигурации
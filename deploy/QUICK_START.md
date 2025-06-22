# Быстрый старт мониторинга

## 🚀 Полное развертывание за 2 минуты

### 1. Запуск мониторинга

```bash
cd deploy/monitoring
docker compose up -d
```

### 2. Перезапуск основных сервисов

```bash
cd ../dev
docker compose down
docker compose up -d
```

### 3. Проверка результата

```bash
# Статус всех контейнеров
docker ps --format "table {{.Names}}\t{{.Status}}"

# Открыть Grafana
open http://localhost:15000
```

## 🎯 Что получили

### Мониторинг доступен по адресам:
- **Grafana**: http://localhost:15000 (admin/admin)
- **Prometheus**: http://localhost:15090
- **Loki**: http://localhost:15100
- **cAdvisor**: http://localhost:15081

### Сервисы интегрированы:
- ✅ auth-service
- ✅ ping-service  
- ✅ telegram-bot-service
- ✅ api-gateway
- ✅ frontend

## 📊 Что мониторится

### Автоматически собираются:
- 📈 **Метрики контейнеров** (CPU, память, сеть)
- 📋 **Логи всех сервисов** в JSON формате
- ❤️ **Health checks** всех сервисов

### Нужно добавить вручную:
- 🎯 **Бизнес-метрики** в сервисах (`/metrics` эндпоинты)
- 📊 **Кастомные дашборды** в Grafana

## 🔧 Следующие шаги

1. **Добавить метрики в сервисы**:
   ```go
   // В Go сервисах
   http.Handle("/metrics", promhttp.Handler())
   ```

2. **Создать дашборды в Grafana**:
   - Импортировать готовые: Docker Container & Host Metrics (ID: 10619)
   - Создать кастомные для бизнес-метрик

3. **Настроить алерты**:
   - Service Down
   - High Error Rate
   - High Memory/CPU Usage

## 🆘 Troubleshooting

### Проблема: Порты заняты
```bash
# Проверить занятые порты
for port in 15000 15090 15100 15081; do 
  lsof -i :$port && echo "Порт $port занят"
done

# Освободить порт (осторожно!)
sudo lsof -ti:15000 | xargs kill -9
```

### Проблема: Сервисы не видны в Prometheus
```bash
# Проверить сеть
docker network inspect slcw-monitoring

# Проверить что метрики доступны
curl http://localhost:15090/api/v1/targets
```

### Проблема: Логи не поступают в Loki
```bash
# Проверить Promtail
docker logs slcw-promtail

# Проверить формат логов
docker logs slcw-auth-service-dev | head -1 | jq .
```

## 📖 Подробная документация

- [Стратегия мониторинга](../docs/architecture/logging-monitoring-strategy.md)
- [Распределение портов](../docs/architecture/ports-allocation.md)
- [Интеграция сервисов](./monitoring/integrate-services.md)
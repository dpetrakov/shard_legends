
# nginx

```bash
nano /etc/nginx/conf.d/slcw.conf
sudo nginx -t
systemctl restart nginx.service
```

# ssh туннель 
```bash
ssh -R 0.0.0.0:10081:localhost:9000 root@209.145.52.169 -N -p 2222 -o ServerAliveInterval=10 -o ServerAliveCountMax=5
```

# Docker
```bash
docker compose down
docker compose build --no-cache
docker compose up -d --build ping-service
docker compose logs ping-service
docker compose up -d

# запуск с профилем
docker compose --profile forly up -d
```

# Генерация токена для TG
```bash
openssl rand -hex 32
```

# Управление миграциями
```bash
  # Применение миграций
  docker compose --profile migrations run --rm migrate up

  # Откат миграций  
  docker compose --profile migrations run --rm migrate down 1

  # Проверка версии
  docker compose --profile migrations run --rm migrate version
```

Тестирование
```bash
go test ./...
```

  1. Откройте Grafana: http://localhost:15000
  2. Войдите: admin / 
  3. Проверьте логи:
    - Перейдите в Explore (компас слева)
    - Выберите datasource "Loki"
    - Используйте запрос: {job="dockerlogs"}
    - Нажмите "Run query"
  4. Проверьте метрики:
    - Выберите datasource "Prometheus"
    - Используйте запрос: up
    - Должны увидеть метрики от prometheus и cadvisor



  1. Prometheus: http://localhost:15090
    - Проверьте статус таргетов: Status → Targets
    - Найдите ping-service в списке (должен быть UP)
    - Выполните запросы метрик:
        - ping_requests_total - общее количество ping запросов
      - http_requests_total - все HTTP запросы
      - service_uptime_seconds - время работы сервиса
  2. Grafana: http://localhost:15000
    - Логин: admin / 
    - Datasources → Prometheus → должен быть зеленый статус
    - Explore → выберите Prometheus → введите метрику ping_requests_total

  Метрики ping-service успешно добавлены и работают! Вы можете видеть:
  - ping_requests_total = 3 (количество выполненных ping запросов)
  - http_requests_total с разбивкой по endpoint, method, status
  - service_uptime_seconds - время работы сервиса
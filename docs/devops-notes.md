
# nginx

```bash
rm /etc/nginx/conf.d/slcw.conf
cp ../nginx/slcw.conf /etc/nginx/conf.d/slcw.conf

nano /etc/nginx/conf.d/slcw.conf
sudo nginx -t
sudo nginx -s reload
systemctl restart nginx
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

# Тестирование
```bash
go test ./...
```

# Создание токена dpetrkov78
```bash
cd scripts/jwt_gen
go run .
```

Запуск Gemini
```bash
npx https://github.com/google-gemini/gemini-cli
```

Запрос в БД - список рецептов
```bash
docker compose -f deploy/dev/docker-compose.yml exec -T postgres \
  psql -U slcw_user -d shard_legends_dev \
  -c "SELECT id, code, operation_class_code, is_active FROM production.recipes;"
```

```bash
docker compose -f deploy/dev/docker-compose.yml exec -T postgres \
  psql -U slcw_user -d shard_legends_dev -v ON_ERROR_STOP=1 --echo-queries \
  -c "UPDATE i18n.languages SET is_default = FALSE; UPDATE i18n.languages SET is_default = TRUE WHERE code = 'ru';"
```

## Скрипт для сброса дейли сундуков
```bash
docker compose -f deploy/dev/docker-compose.yml exec -T postgres \
  psql -U slcw_user -d shard_legends_dev -v ON_ERROR_STOP=1 --echo-queries \
  -c "UPDATE production.production_tasks \
        SET created_at = created_at - INTERVAL '1 day' \
      WHERE recipe_id = '9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2' \
        AND status = 'claimed' \
        AND created_at::date = CURRENT_DATE;"
```

## Перезалить скрипты инвентаря
```bash
    for file in ../../migrations/dev-data/inventory-service/*.sql; do
      if [ -f "$file" ]; then
        echo "Applying script: $file"
        cat "$file" | docker-compose exec -T postgres psql -U slcw_user -d shard_legends_dev
      else
        echo "Warning: No SQL scripts found in migrations/dev-data/inventory-service/"
        break
      fi
    done
```

``` bash
    cat ../../migrations/dev-data/clean_orphan_items.sql | docker compose exec -T postgres psql -U slcw_user -d shard_legends_dev
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
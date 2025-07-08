
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

Сбросить весь кеш Redis
```bash
# перейти в каталог deploy/dev
cd deploy/dev

# выполнить команду в запущенном контейнере Redis
docker compose exec redis redis-cli FLUSHALL        # удалить данные из всех БД
# или точечно, например очистить только DB 1 (Inventory Service)
docker compose exec redis redis-cli -n 1 FLUSHDB
```

Пролить предметы
```bash
cd scripts/item_loader/
go mod tidy  

go run . --files ../../game-data/items/resources.yaml
go run . --all ../../game-data/items/
go run . --all ../../game-data/classifiers/
```



Давай сделаем в services/deck-game-service еще один публичный метод – покупка предметов. На вход он должен принимать  либо идентификатор рецепта, либо код покупаемого товара.

Если передан код подкупаемого товара с определенным качеством и серией, то ищется рецепт с operation_class_code: trade_purchase и с output_items равным указанному покупаемому товару с учетом качества и серии. Если найден ровно один рецепт, то берется его идентификатор и продолжается выполнение. Если не найден рецепт или найдено более одного рецепта, по котормоу можно купить этот товар – возвращается соответствующая ошибка.

Далее ставиться задание на покупку в сервис production-service (/factory/start) c  указанным количеством. В случае если задание принято успешно сразу же делается клайм купленных товаров /factory/claim 

Добавь детальную задачу для релазиации этого функционала в tasks/0_backlog.md, доработай спецификацию docs/specs/deck-game-service.md и docs/specs/deck-game-service-openapi.yml
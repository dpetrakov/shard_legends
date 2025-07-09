# Backlog — задачи к рассмотрению

> Здесь собираются идеи, фичи и задачи, которые еще не проработаны или не приоритизированы.


---
**Формат добавления задач:**
```
## Название задачи
**Описание:** краткое описание проблемы или функции
**Приоритет:** [Высокий/Средний/Низкий]
**Оценка:** [XS/S/M/L/XL]
**Зависимости:** список зависимых задач
**Критерии готовности:** что должно быть выполнено
```

## A-R-1 Рецепт открытия resource_chest_s
- **Описание:** Разработать и загрузить рецепт `resource_chest_s_opening` (operation_class_code `chest_opening`), который открывает малый ресурсный сундук. Рецепт потребляет 1× `resource_chest_s` и 1× `key` качества `small`, возвращает фиксированный набор ресурсов согласно бизнес-требованиям.
- **Пошаговый план:**
  1. Изучить бизнес-документ [`game-mechanics-chests-keys-deck-minigame.md`](docs/concept/game-mechanics-chests-keys-deck-minigame.md) — раздел «Ресурсные сундуки» (таблица количеств 100/40/40/15/5).
  2. Проверить наличие классификаторов и уровней качества в [`002_reset_classifiers.sql`](migrations/dev-data/inventory-service/002_reset_classifiers.sql).
  3. Определить UUID предметов `resource_chest_s`, `key` и ресурсов (`stone`, `wood`, `ore`, `diamond`) в [`003_reset_items.sql`](migrations/dev-data/inventory-service/003_reset_items.sql).
  4. Ознакомиться с примером SQL-скрипта [`004_insert_reward_chest_recipe.sql`](migrations/dev-data/inventory-service/004_insert_reward_chest_recipe.sql) и YAML-форматом в [`production-recipes-initial.md`](docs/specs/production-recipes-initial.md) + [`reward-chest-recipe.yaml`](docs/specs/recipes/reward-chest-recipe.yaml).
  5. Сформировать YAML-рецепт `resource-chest-open-recipes.yaml` (папка `docs/specs/recipes/`) со структурой:
     - `id` — новый фиксированный UUID
     - `code` — `resource_chest_s_open`
     - `input_items` — сундук + ключ
     - `output_items` — количество зависит от типа ресурса: stone 40, wood 40, ore 15, diamond 5; вероятности 40/40/15/5 % в сумме дают 100 %
  6. На основе YAML создать миграцию `005_insert_resource_chest_open_recipes.sql` (dev-data) по образцу п. 3.
  7. Добавить RU/EN переводы названия/описания через `i18n.translations`.
  8. Протестировать скрипт локально: запустить проливку миграций через контейнер migrate:
      ```bash
      docker compose -f deploy/dev/docker-compose.yml --profile migrations run --rm \
        --entrypoint /bin/sh migrate -c \
        'psql "$DATABASE_URL" -f /migrations/dev-data/inventory-service/005_insert_resource_chest_open_recipes.sql'
      ```
      Убедиться в отсутствии ошибок и корректной сумме вероятностей 100 %.
- **Приоритет:** Средний
- **Оценка:** M
- **Зависимости:** 002_reset_classifiers, 003_reset_items
- **Критерии готовности:**
  - [x] YAML-спецификация добавлена и прошла CI-валидацию
  - [x] SQL-миграция выполняется без ошибок на чистой dev БД
  - [x] Все item_id резолвятся, foreign keys валидны
  - [x] Выходные количества ресурсов соответствуют 40/40/15/5 (итого 100)
  - [x] Добавлены переводы RU/EN для name/description
  - [ ] Unit-тест пред-загрузки рецепта (fixture test) зелёный

## A-R-2 Рецепт открытия resource_chest_m
- **Описание:** Аналогично `resource_chest_s_opening`, но для среднего сундука (`resource_chest_m`) и ключа `key` качества `medium`. Количество ресурсов масштабируется до 3 500: stone 1400, wood 1400, ore 525, diamond 175.
- **Пошаговый план:**
  1. Повторить шаги 1-4, используя данные для размера **M** (таблица в концепте).
  2. Дописать рецепт в существующий YAML-файл [`resource-chest-open-recipes.yaml`](docs/specs/recipes/resource-chest-open-recipes.yaml):
      - `output_items` имеют фиксированное количество **3 500** ед.; вероятности 40/40/15/5 % (stone/wood/ore/diamond).
      - `fixed_quality_level_code: medium`.
  3. Расширить SQL-скрипт [`005_insert_resource_chest_open_recipes.sql`](migrations/dev-data/inventory-service/005_insert_resource_chest_open_recipes.sql): добавить новые INSERT для input/output/translation блока рецепта `resource_chest_m_open`, гарантировать фикс. количество 3 500 и вероятности 40/40/15/5 %.
  4. Пролить изменения тем же образом, что и для A-R-1 (см. bash команду выше).
- **Приоритет:** Средний
- **Оценка:** M
- **Зависимости:** docs/specs/recipes/resource-chest-open-recipes.yaml, migrations/dev-data/inventory-service/005_insert_resource_chest_open_recipes.sql
- **Критерии готовности:**
  - [x] YAML-файл и миграция добавлены
  - [x] Количество ресурсов 3 500, распределение 1400/1400/525/175
  - [x] Ключ качества `medium` корректно проверяется
  - [x] Переводы RU/EN присутствуют

## A-R-3 Рецепт открытия resource_chest_l
- **Описание:** Аналогично предыдущим, но для большого сундука (`resource_chest_l`) и ключа `key` качества `large`. Количество ресурсов 47 000: stone 18 800, wood 18 800, ore 7 050, diamond 2 350.
- **Пошаговый план:**
  1. Повторить шаги 1-4, используя данные для размера **L**.
  2. Дописать рецепт в тот же YAML-файл [`resource-chest-open-recipes.yaml`](docs/specs/recipes/resource-chest-open-recipes.yaml):
      - `output_items` имеют фиксированное количество **47 000** ед.; вероятности 40/40/15/5 %.
      - `fixed_quality_level_code: large`.
  3. Расширить SQL-скрипт [`005_insert_resource_chest_open_recipes.sql`](migrations/dev-data/inventory-service/005_insert_resource_chest_open_recipes.sql): добавить INSERT-ы для `resource_chest_l_open`, фикс. количество 47 000, вероятности 40/40/15/5 %.
  4. Пролить изменения той же docker compose командой.
- **Приоритет:** Средний
- **Оценка:** M
- **Зависимости:** docs/specs/recipes/resource-chest-open-recipes.yaml, migrations/dev-data/inventory-service/005_insert_resource_chest_open_recipes.sql
- **Критерии готовности:**
  - [x] YAML-файл и миграция добавлены
  - [x] Количество ресурсов 47 000, распределение 18 800/18 800/7 050/2 350
  - [x] Ключ качества `large` корректно проверяется
  - [x] Переводы RU/EN присутствуют

## Рецепт открытия reagent_chest_s
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `reagent_chest_s` (входящий предмет: `reagent_chest_s`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия reagent_chest_m
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `reagent_chest_m` (входящий предмет: `reagent_chest_m`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия reagent_chest_l
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `reagent_chest_l` (входящий предмет: `reagent_chest_l`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия booster_chest_s
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `booster_chest_s` (входящий предмет: `booster_chest_s`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия booster_chest_m
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `booster_chest_m` (входящий предмет: `booster_chest_m`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия booster_chest_l
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `booster_chest_l` (входящий предмет: `booster_chest_l`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD

## Рецепт открытия blueprint_chest
**Описание:** Создать миграционный скрипт для рецепта открытия сундука `blueprint_chest` (входящий предмет: `blueprint_chest`)
**Приоритет:** Средний
**Оценка:** S
**Зависимости:** 
**Критерии готовности:** TBD


## D-O-1 Реализация эндпоинта открытия сундуков (/deck/chest/open) [ЗАВЕРШЕНО]
- **Описание:** Добавить в Deck Game Service публичный POST-эндпоинт `/deck/chest/open`, который принимает тип/качество сундука и количество, ищет подходящий `chest_opening`-рецепт, запускает `Production Service → /production/factory/start`, сразу выполняет `claim`, агрегирует полученные предметы и возвращает их клиенту. Бизнес-правила идентичны сценарию ежедневных сундуков (cooldown = 0, лимитов нет).
- **Статус:** ✅ РЕАЛИЗОВАНО
- **Реализованные компоненты:**
  - ✅ Модели `OpenChestRequest` и `OpenChestResponse` в `internal/models/open_chest.go`
  - ✅ Расширение интерфейса `InventoryClient` с методом `GetItemQuantity`
  - ✅ Маппинг сундуков в рецепты в `internal/service/chest_mapping.go`
  - ✅ Реализация метода `OpenChest` в `internal/service/deck_game_service.go`
  - ✅ Хэндлер `OpenChest` в `internal/handlers/deck_game_handler.go`
  - ✅ Регистрация маршрута `/deck/chest/open` в `cmd/server/main.go`
  - ✅ Полное покрытие unit-тестами (handlers и service)
  - ✅ Валидация взаимоисключающих полей `quantity` и `open_all`
  - ✅ Обработка ошибок: `invalid_input`, `recipe_not_found`, `insufficient_chests`
- **Тестирование:** Все тесты проходят успешно, сборка без ошибок

- [ ] D-O-2: Публичный метод покупки предметов в deck-game-service
**Описание:**
Реализовать публичный endpoint для покупки предметов через deck-game-service. Метод должен принимать либо идентификатор рецепта, либо код покупаемого товара (с качеством и серией). Если передан код товара, сервис ищет подходящий рецепт с operation_class_code: trade_purchase и нужным output_items. Если найден ровно один рецепт — используется его id, иначе возвращается ошибка. Далее сервис ставит задание на покупку в production-service (/factory/start) с нужным количеством, после чего сразу делает claim через /factory/claim. В случае успеха возвращает купленные предметы.

**Критерии готовности:**
- [ ] Новый endpoint доступен публично
- [ ] Поиск рецепта по коду товара, качеству и серии
- [ ] Корректная обработка ошибок (нет рецепта, несколько рецептов)
- [ ] Интеграция с production-service (start + claim)
- [ ] Тесты на все ветки логики
- [ ] Документация и OpenAPI обновлены

**Связанные файлы:**
- docs/specs/deck-game-service.md
- docs/specs/deck-game-service-openapi.yml

**Референсная реализация:**
- Метод `/deck/chest/open` в этом же сервисе deck-game-service (см. реализацию поиска рецепта, интеграции с production-service и обработки claim)

**См. спецификацию:**
- [docs/specs/deck-game-service.md](../docs/specs/deck-game-service.md)
- [docs/specs/deck-game-service-openapi.yml](../docs/specs/deck-game-service-openapi.yml)

## D-O-3 Публичный метод получения списка товаров в магазине за сапфиры
**Описание:**
Реализовать публичный endpoint для получения списка товаров, доступных для покупки за сапфиры через deck-game-service. Метод возвращает список идентификаторов рецептов, у которых на вход только сапфиры, а на выход только один предмет. Для каждого рецепта возвращаются идентификаторы, код, редкость и коллекция входящих и исходящих предметов. Редкость и коллекция берутся из fixed_collection_code, fixed_quality_level_code (если null — считаем, что это base и в выходной набор параметров можно не включать).
**Приоритет:** Средний
**Оценка:** M
**Зависимости:** production.recipes, production.recipe_input_items, production.recipe_output_items
**Критерии готовности:**
- [ ] Новый публичный endpoint доступен по спецификации
- [ ] Корректно фильтруются рецепты: вход — только сапфиры, выход — только один предмет
- [ ] В ответе присутствуют все необходимые идентификаторы, коды, коллекции и редкости
- [ ] Корректно обрабатываются значения fixed_collection_code, fixed_quality_level_code (null → base)
- [ ] Покрытие unit-тестами не менее 80%
- [ ] Документация и OpenAPI обновлены

**Пошаговый план:**
1. Изучить структуру production.recipes, recipe_input_items, recipe_output_items
2. Реализовать фильтрацию рецептов по условиям (вход — только сапфиры, выход — один предмет)
3. Собрать для каждого рецепта нужные данные (id, code, входящие/исходящие предметы, коллекция, редкость)
4. Реализовать публичный endpoint (GET /deck/shop/sapphires)
5. Покрыть логику unit-тестами
6. Обновить спецификацию и OpenAPI

**Связанные файлы:**
- docs/specs/deck-game-service.md
- docs/specs/deck-game-service-openapi.yml

**Референсная реализация:**
- Логика поиска и фильтрации рецептов аналогична реализации поиска рецепта для покупки предмета (см. buyItem endpoint, D-O-2)
- Использовать подходы из методов поиска и агрегации рецептов в Deck Game Service

**См. спецификацию:**
- [docs/specs/deck-game-service.md](../docs/specs/deck-game-service.md)
- [docs/specs/deck-game-service-openapi.yml](../docs/specs/deck-game-service-openapi.yml)




# Спецификация начального набора производственных рецептов

> Документ формируется на основании концепции «Сундуки, ключи и мини-игра "Дека"» (`docs/concept/game-mechanics-chests-keys-deck-minigame.md`) и существующих классификаторов/предметов (`docs/specs/items-initial.md`, `docs/specs/classifiers.md`).
> Цель — подготовить данные, которые будут загружены в **Production Service** при инициализации системы.

## 📌 Пошаговый план подготовки

| № | Шаг | Описание | Статус |
|---|-----|----------|--------|
| 1 | Сбор требований и определение категорий рецептов | Изучить концепт, выделить группы рецептов, утвердить объём первой итерации | ✅ Выполнен |
| 2 | Описание структуры рецепта | Определить формат JSON/SQL для загрузки, маппинг полей на модель `production_service.internal.models.recipe` | ✅ |
| 3 | Таблица рецептов «Переплавка сырья» | Сформировать детальные записи для 4 базовых рецептов переплавки (планка, блок, слиток, бриллиант) | ✅ |
| 4 | Таблица рецептов «Крафт инструментов» | Подготовить рецепты крафта инструментов (лопата, серп, топор, кирка) для 4 уровней качества | ✅ |
| 5 | Таблица рецептов «Открытие сундуков» | Подготовить рецепты открытия сундуков S/M/L и Blueprint; потребляют соответствующий ключ и сундук, возвращают награды | ⬜ |
| 6 | Таблица рецептов «Выдача сундуков (reward)» | Рецепты без входных параметров, которые выдают случайный сундук с дневными лимитами (мини-игра «Дека») | ✅ |
| 7 | Параметры лимитов и слотов | Установить лимиты запуска и требования к слотам для каждой категории рецептов | ⬜ |
| 8 | Валидация с PO | Передать черновик на проверку продукт-оунеру, собрать фидбек | ⬜ |
| 9 | Финализация и миграции | Создать SQL/fixtures и Pull Request | ⬜ |

## Шаг 1 — Результаты сбора требований ✅

На основании раздела «Система кузницы» в концепте выделены **две ключевые категории** производственных рецептов, которые требуется загрузить в систему на старте:

1. **Переплавка сырья** (`smelting`)
   - 4 рецепта: деревянный брусок, каменный блок, металлический слиток, бриллиант
   - Позволяют превратить базовые ресурсы в полу-фабрикаты/усовершенствованные материалы
   - Используют реагенты для катализирования процесса
   - Время выполнения: 1 час за партию (N × часов при масштабе)
2. **Крафт инструментов** (`crafting`)
   - Базируется на чертежах из Blueprint-сундуков
   - 4 типа инструментов × 4 уровня качества = 16 рецептов
   - Требуют 4 заготовки соответствующего качества
   - Время выполнения: 12 часов за инструмент

3. **Открытие сундуков** (`chest_opening`)
   - 4 рецепта для сундуков: ресурсный (`resource_chest_*`), реагентный (`reagent_chest_*`), ускорительный (`booster_chest_*`) и Blueprint (`blueprint_chest`)
   - Потребляют 1 сундук + 1 ключ соответствующего размера/типа
   - Время выполнения: мгновенно (0 сек), но учитывается задержка claim логики

4. **Выдача сундуков (reward)** (`chest_reward`)
   - Рецепты без входных параметров, инициируемые системой (мини-игра «Дека», ежедневные задания)
   - Производят один сундук определённого типа/размера
   - Лимиты: дневные/недельные в зависимости от источника

Дополнительно рассмотрены, но **не входят** в текущий объём:
- Специальные сезонные или эвент-рецепты — запланированы на будущие итерации

Следующим шагом будет формализация **структуры объекта рецепта** (Шаг 2). После её утверждения будут заполнены детальные таблицы рецептов. 

## Шаг 2 — Формат объекта рецепта ✅

Ниже приведён универсальный YAML-формат, который будет использоваться для генерации SQL-фикстур или прямой загрузки через внутренний админ-API Production Service. Формат однозначно маппится на модель `ProductionRecipe` и связанные структуры (`RecipeInputItem`, `RecipeOutputItem`, `RecipeLimit`).

```yaml
recipes:
  - id: 4b885302-6574-48f3-b533-1d2a8b2da55a   # UUID рецепта (фиксируется для миграций)
    operation_class_code: smelting            # ↔ models.OperationClassSmelting
    is_active: true
    production_time_seconds: 3600            # 1 час

    translations:
      - language_code: en
        field_name: name
        content: "Wooden Plank Smelting"
      - language_code: ru
        field_name: name
        content: "Переплавка: Деревянный брусок"
      - language_code: en
        field_name: description
        content: "Convert 100 wood and 4 discs into 1 wooden plank"
      - language_code: ru
        field_name: description
        content: "Переплавить 100 древесины и 4 диска в 1 деревянный брусок"

    input_items:
      - item_code: wood                      # ↔ docs/specs/items-initial.md → `wood`
        quantity: 100
      - item_code: disc                      # реагент
        quantity: 4

    output_items:
      - item_code: wooden_plank              # результирующий предмет (должен быть в items-initial v2)
        min_quantity: 1                      # min и max одинаковы → фиксированное количество
        max_quantity: 1
        probability_percent: 100             # 100% вероятность

    limits:
      - limit_type: per_day                  # ↔ CL014 `per_day`
        limit_object: recipe_execution       # ↔ CL015 `recipe_execution`
        limit_quantity: 20                   # не более 20 запусков в день

```

### Правила маппинга полей

1. **`id`** — фиксированный UUID для обеспечения идемпотентных миграций.
2. **`operation_class_code`** — строка из классификатора `production_operation_class` (см. `models` constants).
3. **Переводы** (`translations`) грузятся в `i18n.translations` с `entity_type = 'recipe'`; таблица `production.recipes` не содержит `name`/`description`.
4. **`input_items[].item_code`** и **`output_items[].item_code`** указывают `code` предмета из `items-initial.md`. Преобразуются в `item_id` через lookup.
5. **Коллекции/качество** (необязательные поля):
   - `collection_code` / `fixed_collection_code`
   - `quality_level_code` / `fixed_quality_level_code`
   При отсутствии значения берётся `base` коллекция/качество.
6. **`output_items`** поддерживает вероятностные выпадения. Если `probability_percent < 100`, система использует roll-магазин при pre-calculate.
7. **`limits`** описывают пользовательские или глобальные лимиты; если список пуст, лимитов нет.

### Реализация миграций

1. YAML парсится утилитой загрузки (будет добавлена в `scripts/`), которая создаёт данные в таблицах `production.recipes`, `production.recipe_input_items`, `production.recipe_output_items`, `production.recipe_limits`.
2. Переводы из блока `translations` импортируются в `i18n.translations` в единой транзакции.
3. Проверка консистентности выполняется в unit-тестах (дубликаты UUID, корректность ссылок, наличие переводов для базового языка).

---

Следующий шаг — составление таблицы рецептов «Переплавка сырья» (Шаг 3).

### 📖 Пример рецепта с расширенными возможностями

Ниже демонстрационный рецепт **«Открытие сундука чертежей»**, включающий все поддерживаемые поля.

```yaml
recipes:
  - id: 3b99d82d-0c77-4da9-9c6c-fb8b3b07c4e4
    operation_class_code: chest_opening
    is_active: true
    production_time_seconds: 0          # мгновенно

    translations:
      - language_code: en
        field_name: name
        content: "Blueprint Chest Opening"
      - language_code: ru
        field_name: name
        content: "Открытие сундука чертежей"
      - language_code: en
        field_name: description
        content: "Open a blueprint chest with a blueprint key to obtain a random seasonal blueprint."
      - language_code: ru
        field_name: description
        content: "Открыть сундук чертежей при помощи ключа, чтобы получить случайный сезонный чертеж."

    # INPUTS
    input_items:
      # 0 — сам сундук, коллекция (сезон) закодирована в chest item
      - item_code: blueprint_chest
        quantity: 1

      # 1 — ключ, фиксированное качество small/medium/large здесь не используется
      - item_code: blueprint_key
        quantity: 1

    # OUTPUTS
    output_items:
      # Группа "main" — гарантированный чертеж (probability 100)
      - item_code: blueprint_shovel
        min_quantity: 1
        max_quantity: 1
        probability_percent: 25          # 25% шанс
        output_group: main               # группа для взаимно-исключающего расчёта
        collection_source_input_index: 0 # наследуем коллекцию (сезон) из сундука
        fixed_quality_level_code: base   # качество фиксированное

      - item_code: blueprint_pickaxe
        min_quantity: 1
        max_quantity: 1
        probability_percent: 25
        output_group: main
        collection_source_input_index: 0
        fixed_quality_level_code: base

      - item_code: blueprint_axe
        min_quantity: 1
        max_quantity: 1
        probability_percent: 25
        output_group: main
        collection_source_input_index: 0
        fixed_quality_level_code: base

      - item_code: blueprint_sickle
        min_quantity: 1
        max_quantity: 1
        probability_percent: 25
        output_group: main
        collection_source_input_index: 0
        fixed_quality_level_code: base

      # Группа "bonus" — редкий бонус-сундук с 5% шансом
      - item_code: resource_chest_s
        min_quantity: 1
        max_quantity: 1
        probability_percent: 5
        output_group: bonus              # отдельная группа => рассчитывается отдельно от "main"
        fixed_collection_code: base
        fixed_quality_level_code: small

    # LIMITS
    limits:
      # Ограничение на открытие 20 сундуков чертежей в день
      - limit_type: per_day
        limit_object: recipe_execution
        limit_quantity: 20

      # Глобальный лимит на получение конкретного чертежа за сезон (пример продвинутого кейса)
      - limit_type: per_season
        limit_object: item_receipt
        target_item_code: blueprint_pickaxe   # будет преобразован в UUID
        limit_quantity: 3
```

Ключевые моменты примера:

1. **`output_group`** определяет набор результатов, из которых 
   выбирается ровно _один_ (или ни одного) предмет. Для группы `main` суммарная вероятность должна быть ≤100 %; система выполняет roll. Для независимых групп (`bonus`) roll выполняется отдельно, т.е. игрок может получить предмет из каждой группы.
2. **`collection_source_input_index`** указывает, что коллекция (например, сезон) берётся из предмета-входа с индексом 0 (сундук). Это упрощает поддержку сезонных дроп-таблиц без дублирования рецептов.
3. **`fixed_quality_level_code`** принудительно задаёт качество выходного предмета, независимо от входов.
4. В блоке `limits` показан пример как ограничить как количество запусков рецепта, так и количество получений конкретного предмета в рамках сезона.

✅ Шаг 3 выполнен: см. файл `docs/specs/recipes/smelting-recipes.yaml`.
✅ Шаг 4 выполнен: см. файл `docs/specs/recipes/crafting-recipes.yaml`.
✅ Шаг 6 выполнен: см. файл `docs/specs/recipes/reward-chest-recipe.yaml`.

--- 
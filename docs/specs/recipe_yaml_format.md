# Формат YAML для описания рецептов

Этот документ описывает единый формат `*.yaml`-файлов для описания производственных рецептов. Он будет использоваться для автоматической генерации SQL-миграций и загрузки данных.

## 1. Общая структура файла

```yaml
recipes:
  - id: <uuid>                           # ОБЯЗАТЕЛЬНО. Постоянный UUID v4 для рецепта.
    code: <string>                       # ОБЯЗАТЕЛЬНО. Уникальный код рецепта (snake_case).
    operation_class_code: <string>       # ОБЯЗАТЕЛЬНО. Код из классификатора operation_class.
    is_active: <boolean>                 # ОБЯЗАТЕЛЬНО. Активность рецепта.
    production_time_seconds: <integer>   # ОБЯЗАТЕЛЬНО. Время производства в секундах.

    translations:  # СПИСОК ПЕРЕВОДОВ
      - language_code: <string>           # Примеры: 'en', 'ru'
        field_name: <string>             # Примеры: 'name', 'description'
        content: <string>                # Текст перевода.

    input_items:  # ВХОДНЫЕ ПРЕДМЕТЫ
      - item_code: <string>               # Код предмета из items.
        quantity: <integer>              # Количество.
        fixed_quality_level_code: <string> # (опционально) Фиксированный уровень качества.

    output_items:  # ВЫХОДНЫЕ ПРЕДМЕТЫ
      - item_code: <string>               # Код предмета.
        min_quantity: <integer>          # Минимальное количество.
        max_quantity: <integer>          # Максимальное количество.
        probability_percent: <float>     # Вероятность в процентах.
        output_group: <string>           # Группа взаимно-исключающих выходов.
        fixed_quality_level_code: <string> # (опционально) Фиксированный уровень качества.

    limits:  # ЛИМИТЫ НА ИСПОЛЬЗОВАНИЕ
      - limit_type: <string>             # Примеры: 'daily', 'weekly'.
        max_uses: <integer>              # Максимальное количество использований.
```

## 2. Описание полей

| Поле | Тип | Обяз. | Примечание |
|------|-----|-------|------------|
| `id` | UUID v4 | ✔ | Уникальный идентификатор рецепта |
| `code` | string | ✔ | Уникальный код рецепта |
| `operation_class_code` | string | ✔ | Операционный класс рецепта |
| `is_active` | boolean | ✔ | Определяет, активен ли рецепт |
| `production_time_seconds` | integer | ✔ | Время выполнения рецепта в секундах |
| `translations` | list | ✔ | Переводы для имени и описания рецепта |
| `input_items` | list | ✔ | Входные предметы, используемые в рецепте |
| `output_items` | list | ✔ | Выходные предметы, произведенные рецептом |
| `limits` | list | ✖ | Лимиты использования рецепта |

## 3. Пример полного рецепта

```yaml
recipes:
  - id: 9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2
    code: resource_chest_s_open
    operation_class_code: chest_opening
    is_active: true
    production_time_seconds: 0

    translations:
      - language_code: en
        field_name: name
        content: "Small Resource Chest Opening"
      - language_code: ru
        field_name: name
        content: "Открытие малого ресурсного сундука"

    input_items:
      - item_code: resource_chest_s
        quantity: 1
        fixed_quality_level_code: small

    output_items:
      - item_code: stone
        min_quantity: 40
        max_quantity: 40
        probability_percent: 40.0
        output_group: main
        fixed_quality_level_code: base

    limits:
      - limit_type: daily
        max_uses: 10
```

## Загрузка рецептов в БД

Утилита загрузки данных должна:
1. Удалять предыдущие версии рецептов перед вставкой (кроме `recipes`).
2. Использовать UPSERT для вставки данных в таблицы `production.recipes`, `production.recipe_input_items`, `production.recipe_output_items`, `production.recipe_limits`.
3. Загружать переводы в `i18n.translations` с `entity_type='recipe'`.

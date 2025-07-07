# Формат YAML для описания игровых предметов

Ниже приведена спецификация **единого** формата `*.yaml`-файлов, из которых будет автоматически генерироваться SQL-загрузка данных в сервис *inventory*.

---

## 1. Общая структура файла

```yaml
items:               # ОБЯЗАТЕЛЬНО. Массив предметов.
  - id: <uuid>       # ОБЯЗАТЕЛЬНО. Постоянный UUID v4 предмета.
    code: <string>   # ОБЯЗАТЕЛЬНО. Уникальный код предмета (snake_case).
    class: <string>  # ОБЯЗАТЕЛЬНО. Код classifier_item из классификатора `item_class` (napr. resources, reagents).
    type: <string>   # ОБЯЗАТЕЛЬНО. Код classifier_item из *производного* классификатора (resource_type, reagent_type ...)
    quality_levels_classifier: <string>  # НЕОБЯЗАТЕЛЬНО. Код классификатора, который будет хранить уровни качества
                                          # для данного предмета. По умолчанию: `quality_level`. Для инструментов – `tool_quality_levels`.
    collections_classifier: <string>     # НЕОБЯЗАТЕЛЬНО. Классификатор, в котором лежит коллекция.
                                          # По умолчанию: `base`.

    # --- I18N ---
    translations:     # ОБЯЗАТЕЛЬНО. Локализации имени и описания.
      <lang_code>:    # ISO-код языка (ru, en, zh-CN ...)
        name: <string>
        description: <string>
      # ... дополнительные языки

    # --- Изображения ---
    images:           # ОПЦИОНАЛЬНО. Массив изображений предмета.
      - collection: <string>         # Код classifier_item из указанного выше collections_classifier
        quality_level: <string>      # Код classifier_item из указанного выше quality_levels_classifier
        url: <string>                # URL изображения (абсолютный или /statics/...)
      # ... дополнительные изображения
```

---

## 2. Описание полей

| Поле | Тип | Обяз. | Примечание |
|------|-----|-------|------------|
| `id` | UUID v4 | ✔ | Должен совпадать со всеми ссылками в других файлах (рецепты, квесты).
| `code` | string | ✔ | Уникальный во всём наборе предметов. Используется как внешний ключ в BALANCE файлах.
| `class` | string | ✔ | Один из кодов классификатора `item_class`:<br>`resources`, `reagents`, `boosters`, `keys`, `chests`, `blueprints`, `tools`.
| `type` | string | ✔ | Код элемента *подчинённого* классификатора. Примеры:<br>• `stone`, `wood`, `ore`, `diamond` для ресурсов (classifier `resource_type`)<br>• `abrasive`, `disc`, `inductor`, `paste` для реагентов (classifier `reagent_type`)<br>• `repair_tool` … для ускорителей (classifier `booster_type`) |
| `quality_levels_classifier` | string | ✖ | Если не указано – `quality_level`. Для предметов, у которых своя шкала (например инструменты), нужно указать другой классификатор.
| `collections_classifier` | string | ✖ | Если не указано – `collection`.
| `translations` | map | ✔ | Ключ – код языка, значение – объект `{name, description}`.
| `images` | list | ✖ | Для UI. Каждая запись определяет изображение под конкретное сочетание `collection` + `quality_level`. |

---

## 3. Минимальный пример (ресурс «Камень»)

```yaml
items:
  - id: 1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60
    code: stone
    class: resources
    type: stone

    translations:
      ru:
        name: Камень
        description: Базовый строительный материал, используется в большинстве рецептов и заданий.
      en:
        name: Stone
        description: Basic building material used in most recipes and quests.
```

---

## 4. Расширенный пример (сундук с изображениями)

```yaml
items:
  - id: 9421cc9f-a56e-4c7d-b636-4c8fdfef7166
    code: resource_chest
    class: chests
    type: resource_chest

    translations:
      ru:
        name: Ресурсный сундук
        description: Содержит случайные ресурсы. Размер и количество зависят от качества сундука.
      en:
        name: Resource Chest
        description: Contains random resources. Size and quantity depend on chest quality.

    images:
      - collection: base      # базовая коллекция
        quality_level: small
        url: /statics/images/items/small-chess-res.png
      - collection: base
        quality_level: medium
        url: /statics/images/items/medium-chess-res.png
      - collection: base
        quality_level: large
        url: /statics/images/items/big-chess-res.png
```

---

## 5. Правила валидации

1. **Уникальность**: комбинация `id` и `code` должна быть уникальна во всём репозитории.
2. **Ссылочная целостность**:
   * `class`, `type`, `quality_levels_classifier`, `collections_classifier`, `collection`, `quality_level` должны существовать в файле `docs/specs/classifiers.md`.
3. **Обязательные переводы**: для каждого предмета должны быть хотя бы `ru` и `en`-локализации.
4. **Изображения**: если указаны, то для каждой записи должна быть пара `collection` + `quality_level` + `url`.
5. **Формат UUID**: v4, 36 символов, дефисы на позициях 9-14-19-24.

---

## 6. Разбиение по группам

Для удобства редакторов каждый YAML будет содержать **одну логическую группу** предметов
(например `resources.yaml`, `reagents.yaml`).

---

## 7. Путь и именование файлов

Файлы располагаются в каталоге `game-data/items/`.

Шаблон имени: `<item_class>.yaml` или `<item_class>-<subgroup>.yaml` (в зависимости от количества элементов - если эелементов не много - группируем в общие файлы).

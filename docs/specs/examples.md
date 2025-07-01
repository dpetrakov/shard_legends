
## Users

 ## Запрос статуса по дейли-сундукам

 ```bash
 curl -L 'https://dev-forly.slcw.dimlight.online//api/deck/daily-chest/status' -H 'Authorization: Bearer {{jwt}}'
 ```

### Пример ответа
```json
{
    "expected_combo": 7,
    "finished": false,
    "crafts_done": 2,
    "last_reward_at": "2025-07-01T16:08:18.69328Z"
}
```

## Запрос дейли-сундука

 ```bash
 curl -L 'https://dev-forly.slcw.dimlight.online//api/deck/daily-chest/claim' -H 'Content-Type: application/json' -H 'Authorization: Bearer {{jwt}}' -d '{"combo":6,"chest_indices":[1]}'
 ```

### Пример ответа
```json
{
    "items": [
        {
            "item_id": "9421cc9f-a56e-4c7d-b636-4c8fdfef7166",
            "item_class": "chests",
            "item_type": "resource_chest",
            "name": "Маленький ресурсный сундук",
            "description": "Содержит небольшое количество базовых ресурсов.",
            "image_url": "/images/items/default.png",
            "quantity": 1
        }
    ],
    "next_expected_combo": 7,
    "crafts_done": 2
}
```

----


Получние профиля пользователя:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/user/profile' -H 'Authorization: Bearer {{jwt}}'
```

Получение списка доступных производственных слотов:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/user/production-slots' -H 'Authorization: Bearer {{jwt}}'
```

## Auth

Авторизация:
```bash
curl -L -X POST 'https://dev.slcw.dimlight.online/api/auth' -H 'X-Telegram-Init-Data: user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22John%22%2C%22username%22%3A%22john_doe%22%7D&auth_date=1672531200&hash=a1b2c3d4e5f6...' -H 'Accept: application/json' -H 'Authorization: Bearer {{jwt}}'
```

## Inventory

Инвентарь:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/inventory/items' -H 'Authorization: Bearer {{jwt}}' -d ''
```

Получение детальной информации по предметам:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/inventory/items/details' -H 'Content-Type: application/json' -H 'Authorization: Bearer {{jwt}}' -d '{"items":[{"item_id":"1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60"},{"item_id":"aa58eb38-5e91-47f0-bd4e-6ed02cb059b1"}]}'
```

## Production

Список доступных рецептов:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/production/recipes' -H 'Authorization: Bearer {{jwt}}'
```

Задачи в производстве и очереди:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/production/factory/queue' -H 'Authorization: Bearer {{jwt}}'
```

Выполненные задачи, доступные для клайма:
```bash
curl -L 'https://dev.slcw.dimlight.online/api/production/factory/completed' -H 'Authorization: Bearer {{jwt}}'
```

 Клайм всех готовых заказов:
 ```bash
 curl -L 'https://dev.slcw.dimlight.online/api/production/factory/claim' -H 'Content-Type: application/json' -H 'Authorization: Bearer {{jwt}}' -d '{}'
 ```

 Запуск производства рецепта:
 ```bash
 curl -L 'https://dev.slcw.dimlight.online/api/production/factory/start' -H 'Content-Type: application/json' -H 'Authorization: Bearer {{jwt}}' -d '{"recipe_id":"9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2","execution_count":1}'
 ```



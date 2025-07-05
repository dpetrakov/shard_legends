-- clean_orphan_items.sql
-- Description: Удаляет предметы из каталога inventory.items, которые ссылаются на несуществующие
-- записи в классификаторах (classifier_items). Это "предметы-сироты".

-- Включаем вывод информации о ходе выполнения
\set ON_ERROR_STOP on
\timing on

DO $$
DECLARE
    deleted_count integer;
BEGIN
    RAISE NOTICE 'Starting cleanup of orphan items in inventory.items...';

    -- Создаем временную таблицу для хранения ID "предметов-сирот"
    CREATE TEMP TABLE items_to_delete AS
    SELECT id FROM inventory.items
    WHERE
        -- Проверяем, существует ли item_class_id в classifier_items
        NOT EXISTS (
            SELECT 1
            FROM inventory.classifier_items ci
            WHERE ci.id = inventory.items.item_class_id
        ) OR
        -- Проверяем, существует ли item_type_id в classifier_items
        NOT EXISTS (
            SELECT 1
            FROM inventory.classifier_items ci
            WHERE ci.id = inventory.items.item_type_id
        );

    -- Получаем количество предметов для удаления
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RAISE NOTICE '% orphan items found to be deleted.', deleted_count;

    IF deleted_count > 0 THEN
        -- Сначала удаляем связанные данные, чтобы избежать ошибок внешних ключей (foreign key violations)
        -- ВАЖНО: Добавьте сюда удаление из других таблиц, если они ссылаются на inventory.items
        -- Например, если есть таблицы, которые ссылаются на inventory.items
        -- ВАЖНО: Добавьте сюда удаление из других таблиц, если они ссылаются на inventory.items
        -- Пример: DELETE FROM inventory.item_images WHERE item_id IN (SELECT id FROM items_to_delete);
        -- Пример: DELETE FROM inventory.operations WHERE item_id IN (SELECT id FROM items_to_delete);
        -- Сейчас я закомментирую эти строки, так как не знаю всех зависимостей в вашей системе.
        -- Раскомментируйте и адаптируйте их при необходимости.

        -- RAISE NOTICE 'Deleting related data from other tables...';
        -- DELETE FROM inventory.item_images WHERE item_id IN (SELECT id FROM items_to_delete);
        -- DELETE FROM inventory.operations WHERE item_id IN (SELECT id FROM items_to_delete);
        -- DELETE FROM inventory.daily_balances WHERE item_id IN (SELECT id FROM items_to_delete);

        -- Удаляем сами "предметы-сироты"
        RAISE NOTICE 'Deleting orphan items from inventory.items...';
        DELETE FROM inventory.items WHERE id IN (SELECT id FROM items_to_delete);

        GET DIAGNOSTICS deleted_count = ROW_COUNT;
        RAISE NOTICE 'Successfully deleted % orphan items.', deleted_count;
    ELSE
        RAISE NOTICE 'No orphan items found. Database is clean.';
    END IF;

    -- Удаляем временную таблицу
    DROP TABLE items_to_delete;

    RAISE NOTICE 'Cleanup script finished.';
END;
$$;

\timing off 
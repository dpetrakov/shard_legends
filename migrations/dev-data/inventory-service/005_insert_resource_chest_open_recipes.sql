-- =============================================================================
-- DEV DATA Script: 005_insert_resource_chest_open_recipes.sql
-- Description : Загрузка рецепта Small Resource Chest Opening (resource_chest_s_open)
-- Service     : production-service (dev only)
-- Depends     : 004_insert_reward_chest_recipe.sql, 003_reset_items.sql
-- Created     : 2025-07-05
-- =============================================================================

\echo '>>> Loading Small Resource Chest Opening recipe'

BEGIN;

-- -----------------------------------------------------------------------------
-- Параметры
-- -----------------------------------------------------------------------------
DO $$
DECLARE
    v_recipe_id CONSTANT UUID := '7d0afba0-985e-4d74-b027-3b2a32bb2760';
    v_recipe_code CONSTANT TEXT := 'resource_chest_s_open';
BEGIN
    -- -------------------------------------------------------------------------
    -- Удаляем старую версию рецепта (если была)
    -- -------------------------------------------------------------------------
    DELETE FROM i18n.translations WHERE entity_type = 'recipe' AND entity_id = v_recipe_id;
    DELETE FROM production.recipe_limits WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_output_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_input_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipes WHERE id = v_recipe_id;

    -- -------------------------------------------------------------------------
    -- Вставляем основную запись рецепта
    -- -------------------------------------------------------------------------
    INSERT INTO production.recipes (
        id,
        code,
        operation_class_code,
        is_active,
        production_time_seconds
    ) VALUES (
        v_recipe_id,
        v_recipe_code,
        'chest_opening',
        TRUE,
        0
    );

    -- -------------------------------------------------------------------------
    -- Входные предметы: только сундук малого размера (без ключа)
    -- -------------------------------------------------------------------------
    INSERT INTO production.recipe_input_items (
        recipe_id, item_id, quality_level_code, quantity
    ) VALUES 
        -- сундук малого размера с правильным quality_level_code
        (v_recipe_id, '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 'small', 1); -- resource_chest_s

    -- -------------------------------------------------------------------------
    -- Выходные предметы (stone / wood / ore / diamond) – 100 ед., распределение 40/40/15/5
    -- -------------------------------------------------------------------------
    INSERT INTO production.recipe_output_items (
        recipe_id, item_id, min_quantity, max_quantity, probability_percent,
        output_group, fixed_quality_level_code
    ) VALUES
        (v_recipe_id, '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 40, 40, 40.0, 'main', 'base'),   -- stone
        (v_recipe_id, '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 40, 40, 40.0, 'main', 'base'),   -- wood
        (v_recipe_id, 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 15, 15, 15.0, 'main', 'base'),   -- ore
        (v_recipe_id, '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 5, 5,  5.0, 'main', 'base');  -- diamond

    -- -------------------------------------------------------------------------
    -- (Лимитов нет, открытие ограничено предметами)
    -- -------------------------------------------------------------------------

    -- -------------------------------------------------------------------------
    -- Переводы
    -- -------------------------------------------------------------------------
    INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content)
    VALUES
        ('recipe', v_recipe_id, 'name', 'en', 'Small Resource Chest Opening'),
        ('recipe', v_recipe_id, 'name', 'ru', 'Открытие малого ресурсного сундука'),
        ('recipe', v_recipe_id, 'description', 'en', 'Open a small resource chest to obtain 100 resources (stone, wood, ore, diamond).'),
        ('recipe', v_recipe_id, 'description', 'ru', 'Откройте малый ресурсный сундук и получите 100 ресурсов (камень, дерево, руда, алмаз).');

END $$;

-- =============================================================================
-- Рецепт: Medium Resource Chest Opening
-- =============================================================================
DO $$
DECLARE
    v_recipe_id CONSTANT UUID := '4a5026a1-b851-48e9-88d1-156d2e3aa8b9';
    v_recipe_code CONSTANT TEXT := 'resource_chest_m_open';
BEGIN
    -- Remove existing if any
    DELETE FROM i18n.translations WHERE entity_type = 'recipe' AND entity_id = v_recipe_id;
    DELETE FROM production.recipe_limits WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_output_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_input_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipes WHERE id = v_recipe_id;

    -- Insert recipe
    INSERT INTO production.recipes (id, code, operation_class_code, is_active, production_time_seconds)
    VALUES (v_recipe_id, v_recipe_code, 'chest_opening', TRUE, 0);

    -- Inputs: только средний сундук (без ключа)
    INSERT INTO production.recipe_input_items (recipe_id, item_id, quality_level_code, quantity) VALUES
        (v_recipe_id, '6c0f7fd6-4a6e-4d42-b596-a1a2b775cdbc', 'medium', 1); -- resource_chest_m

    -- Outputs quantities 3 500
    INSERT INTO production.recipe_output_items (
        recipe_id, item_id, min_quantity, max_quantity, probability_percent, output_group, fixed_quality_level_code
    ) VALUES
        (v_recipe_id, '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 1400, 1400, 40.0, 'main', 'base'), -- stone
        (v_recipe_id, '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 1400, 1400, 40.0, 'main', 'base'), -- wood
        (v_recipe_id, 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 525, 525, 15.0, 'main', 'base'),   -- ore
        (v_recipe_id, '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 175, 175, 5.0, 'main', 'base');   -- diamond

    -- Translations
    INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
        ('recipe', v_recipe_id, 'name', 'en', 'Medium Resource Chest Opening'),
        ('recipe', v_recipe_id, 'name', 'ru', 'Открытие среднего ресурсного сундука'),
        ('recipe', v_recipe_id, 'description', 'en', 'Open a medium resource chest to obtain 3,500 resources.'),
        ('recipe', v_recipe_id, 'description', 'ru', 'Откройте средний ресурсный сундук и получите 3 500 ресурсов.');
END $$;

-- =============================================================================
-- Рецепт: Large Resource Chest Opening
-- =============================================================================
DO $$
DECLARE
    v_recipe_id CONSTANT UUID := '9f4c3d36-b4e1-4a61-b4f1-0ed2a8b16c77';
    v_recipe_code CONSTANT TEXT := 'resource_chest_l_open';
BEGIN
    DELETE FROM i18n.translations WHERE entity_type = 'recipe' AND entity_id = v_recipe_id;
    DELETE FROM production.recipe_limits WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_output_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipe_input_items WHERE recipe_id = v_recipe_id;
    DELETE FROM production.recipes WHERE id = v_recipe_id;

    INSERT INTO production.recipes (id, code, operation_class_code, is_active, production_time_seconds)
    VALUES (v_recipe_id, v_recipe_code, 'chest_opening', TRUE, 0);

    -- Inputs: только большой сундук (без ключа)
    INSERT INTO production.recipe_input_items (recipe_id, item_id, quality_level_code, quantity) VALUES
        (v_recipe_id, '0f8aa2c1-25b8-4aed-9d6b-8c1e927bf71f', 'large', 1); -- resource_chest_l

    INSERT INTO production.recipe_output_items (
        recipe_id, item_id, min_quantity, max_quantity, probability_percent, output_group, fixed_quality_level_code
    ) VALUES
        (v_recipe_id, '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 18800, 18800, 40.0, 'main', 'base'), -- stone
        (v_recipe_id, '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 18800, 18800, 40.0, 'main', 'base'), -- wood
        (v_recipe_id, 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 7050, 7050, 15.0, 'main', 'base'),   -- ore
        (v_recipe_id, '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 2350, 2350, 5.0, 'main', 'base');   -- diamond

    INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
        ('recipe', v_recipe_id, 'name', 'en', 'Large Resource Chest Opening'),
        ('recipe', v_recipe_id, 'name', 'ru', 'Открытие большого ресурсного сундука'),
        ('recipe', v_recipe_id, 'description', 'en', 'Open a large resource chest to obtain 47,000 resources.'),
        ('recipe', v_recipe_id, 'description', 'ru', 'Откройте большой ресурсный сундук и получите 47 000 ресурсов.');
END $$;

COMMIT;

\echo '<<< Small Resource Chest Opening recipe loaded' 
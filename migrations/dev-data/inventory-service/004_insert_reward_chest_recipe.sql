-- =============================================================================
-- DEV DATA Script: 004_insert_reward_chest_recipe.sql
-- Description : Загрузка рецепта Daily Chest Reward из specs/recipes/reward-chest-recipe.yaml
-- Service     : production-service (dev only)
-- Depends     : 003_reset_items.sql, 008_create_i18n_schema.up.sql, 009_populate_languages.up.sql
-- Created     : 2025-06-28
-- =============================================================================

\echo '>>> Loading Daily Chest Reward recipe'

BEGIN;

-- -----------------------------------------------------------------------------
-- Параметры
-- -----------------------------------------------------------------------------
DO $$
DECLARE
    v_recipe_id CONSTANT UUID := '9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2';
    v_recipe_code CONSTANT TEXT := 'daily_chest_reward';
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
    -- Вставляем выходные предметы (используем UUID из docs/specs/items-initial.md)
    -- -------------------------------------------------------------------------
    INSERT INTO production.recipe_output_items (
        recipe_id, item_id, min_quantity, max_quantity, probability_percent,
        output_group, fixed_quality_level_code
    ) VALUES
        -- Resource chests (30.50 + 15.25 + 4.75 = 50.50%)
        (v_recipe_id, '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 1, 1, 30.50, 'main', 'small'),  -- resource_chest_s
        (v_recipe_id, '6c0f7fd6-4a6e-4d42-b596-a1a2b775cdbc', 1, 1, 15.25, 'main', 'medium'), -- resource_chest_m
        (v_recipe_id, '0f8aa2c1-25b8-4aed-9d6b-8c1e927bf71f', 1, 1,  4.75, 'main', 'large'),  -- resource_chest_l
        -- Reagent chests (18.50 + 8.75 + 2.25 = 29.50%)
        (v_recipe_id, 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 1, 1, 18.50, 'main', 'small'),  -- reagent_chest_s
        (v_recipe_id, 'b6dde60a-6530-4fa3-836b-415520d05f37', 1, 1,  8.75, 'main', 'medium'), -- reagent_chest_m
        (v_recipe_id, '359e86d5-d094-4b2b-b96e-6114e3c66d6b', 1, 1,  2.25, 'main', 'large'),  -- reagent_chest_l
        -- Booster chests (6.50 + 2.75 + 0.75 = 10.00%)
        (v_recipe_id, '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 1, 1,  6.50, 'main', 'small'),  -- booster_chest_s
        (v_recipe_id, 'd9a3e79a-50d3-4ab5-be86-8137145c34e3', 1, 1,  2.75, 'main', 'medium'), -- booster_chest_m
        (v_recipe_id, 'aa58eb38-5e91-47f0-bd4e-6ed02cb059b1', 1, 1,  0.75, 'main', 'large'),  -- booster_chest_l
        -- Blueprint chest (10.00%)
        (v_recipe_id, '012d9076-a37d-4e9d-a49a-fbc7a07e5bd9', 1, 1, 10.00, 'main', NULL);     -- blueprint_chest
        -- Итого: 100.00%

    -- -------------------------------------------------------------------------
    -- Вставляем лимиты (10 запусков в день)
    -- -------------------------------------------------------------------------
    INSERT INTO production.recipe_limits (
        recipe_id, limit_type, max_uses
    ) VALUES (
        v_recipe_id, 'daily', 10
    );

    -- -------------------------------------------------------------------------
    -- Вставляем переводы (RU/EN) через универсальную таблицу
    -- -------------------------------------------------------------------------
    INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content)
    VALUES
        ('recipe', v_recipe_id, 'name', 'en', 'Daily Chest Reward'),
        ('recipe', v_recipe_id, 'name', 'ru', 'Дневная награда: Сундук'),
        ('recipe', v_recipe_id, 'description', 'en', 'Receive a random chest as a daily reward from the Deck minigame.'),
        ('recipe', v_recipe_id, 'description', 'ru', 'Получите случайный сундук как дневную награду из мини-игры "Дека".');

END $$;

COMMIT;

\echo '<<< Daily Chest Reward recipe loaded' 
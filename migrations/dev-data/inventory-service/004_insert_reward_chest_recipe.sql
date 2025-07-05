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
        -- Resource chests (один ID с разными качествами: 30.50 + 15.25 + 4.75 = 50.50%)
        (v_recipe_id, '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 1, 1, 30.50, 'main', 'small'),  -- resource_chest small
        (v_recipe_id, '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 1, 1, 15.25, 'main', 'medium'), -- resource_chest medium
        (v_recipe_id, '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 1, 1,  4.75, 'main', 'large'),  -- resource_chest large
        -- Reagent chests (один ID с разными качествами: 18.50 + 8.75 + 2.25 = 29.50%)
        (v_recipe_id, 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 1, 1, 18.50, 'main', 'small'),  -- reagent_chest small
        (v_recipe_id, 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 1, 1,  8.75, 'main', 'medium'), -- reagent_chest medium
        (v_recipe_id, 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 1, 1,  2.25, 'main', 'large'),  -- reagent_chest large
        -- Booster chests (один ID с разными качествами: 6.50 + 2.75 + 0.75 = 10.00%)
        (v_recipe_id, '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 1, 1,  6.50, 'main', 'small'),  -- booster_chest small
        (v_recipe_id, '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 1, 1,  2.75, 'main', 'medium'), -- booster_chest medium
        (v_recipe_id, '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 1, 1,  0.75, 'main', 'large'),  -- booster_chest large
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
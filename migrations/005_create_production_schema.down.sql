-- Migration DOWN: Drop production schema and tables
-- Service: production-service

-- Drop indexes first
DROP INDEX IF EXISTS production.idx_production_tasks_user_status_created;
DROP INDEX IF EXISTS production.idx_production_tasks_user_recipe_active;
DROP INDEX IF EXISTS production.idx_production_tasks_draft_cleanup;
DROP INDEX IF EXISTS production.idx_task_output_items_item_id;
DROP INDEX IF EXISTS production.idx_task_output_items_task_id;
DROP INDEX IF EXISTS production.idx_production_tasks_recipe_id;
DROP INDEX IF EXISTS production.idx_production_tasks_user_slot;
DROP INDEX IF EXISTS production.idx_production_tasks_completion_time;
DROP INDEX IF EXISTS production.idx_production_tasks_status;
DROP INDEX IF EXISTS production.idx_production_tasks_user_id;
DROP INDEX IF EXISTS production.idx_recipe_limits_type;
DROP INDEX IF EXISTS production.idx_recipe_limits_recipe_id;
DROP INDEX IF EXISTS production.idx_recipe_output_items_item_id;
DROP INDEX IF EXISTS production.idx_recipe_output_items_recipe_id;
DROP INDEX IF EXISTS production.idx_recipe_input_items_item_id;
DROP INDEX IF EXISTS production.idx_recipe_input_items_recipe_id;
DROP INDEX IF EXISTS production.idx_recipes_active;
DROP INDEX IF EXISTS production.idx_recipes_operation_class;
DROP INDEX IF EXISTS production.idx_recipes_code;

-- Drop triggers
DROP TRIGGER IF EXISTS update_production_tasks_updated_at ON production.production_tasks;
DROP TRIGGER IF EXISTS update_recipes_updated_at ON production.recipes;

-- Drop trigger function
DROP FUNCTION IF EXISTS production.update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS production.task_output_items;
DROP TABLE IF EXISTS production.production_tasks;
DROP TABLE IF EXISTS production.recipe_limits;
DROP TABLE IF EXISTS production.recipe_output_items;
DROP TABLE IF EXISTS production.recipe_input_items;
DROP TABLE IF EXISTS production.recipes;

-- Drop the production schema
DROP SCHEMA IF EXISTS production;
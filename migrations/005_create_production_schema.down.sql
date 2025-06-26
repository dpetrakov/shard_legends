-- Migration DOWN: Drop production schema and tables
-- Service: production-service

-- Drop triggers first
DROP TRIGGER IF EXISTS update_recipes_updated_at ON production.recipes;
DROP TRIGGER IF EXISTS update_production_tasks_updated_at ON production.production_tasks;

-- Drop trigger function
DROP FUNCTION IF EXISTS production.update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS production.task_output_items;
DROP TABLE IF EXISTS production.production_tasks;
DROP TABLE IF EXISTS production.recipe_limits;
DROP TABLE IF EXISTS production.recipe_outputs;
DROP TABLE IF EXISTS production.recipe_inputs;
DROP TABLE IF EXISTS production.recipes;

-- Drop the production schema only if it's empty
DROP SCHEMA IF EXISTS production;
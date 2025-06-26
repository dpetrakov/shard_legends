-- Migration UP: Create production schema and tables
-- Service: production-service
-- Depends: 000_init_schemas.up.sql

-- Create production schema
CREATE SCHEMA IF NOT EXISTS production;

-- Comment on schema
COMMENT ON SCHEMA production IS 'Схема для сервиса производства и крафта предметов';

-- Grant privileges to production schema
GRANT ALL PRIVILEGES ON SCHEMA production TO slcw_user;

-- Set default privileges for future objects in production schema
ALTER DEFAULT PRIVILEGES IN SCHEMA production GRANT ALL ON TABLES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA production GRANT ALL ON SEQUENCES TO slcw_user;

-- Create recipes table
CREATE TABLE IF NOT EXISTS production.recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    operation_class_code VARCHAR(50) NOT NULL,
    base_time_seconds INTEGER NOT NULL CHECK (base_time_seconds > 0),
    unlock_conditions JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_recipes_code_not_empty CHECK (length(trim(code)) > 0),
    CONSTRAINT chk_recipes_name_not_empty CHECK (length(trim(name)) > 0)
);

-- Create recipe_inputs table
CREATE TABLE IF NOT EXISTS production.recipe_inputs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production.recipes(id) ON DELETE CASCADE,
    item_code VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_recipe_inputs_item_code_not_empty CHECK (length(trim(item_code)) > 0)
);

-- Create recipe_outputs table
CREATE TABLE IF NOT EXISTS production.recipe_outputs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production.recipes(id) ON DELETE CASCADE,
    item_code VARCHAR(50) NOT NULL,
    min_quantity INTEGER NOT NULL CHECK (min_quantity >= 0),
    max_quantity INTEGER NOT NULL CHECK (max_quantity >= min_quantity),
    probability DECIMAL(5,4) NOT NULL CHECK (probability >= 0 AND probability <= 1),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_recipe_outputs_item_code_not_empty CHECK (length(trim(item_code)) > 0)
);

-- Create recipe_limits table
CREATE TABLE IF NOT EXISTS production.recipe_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production.recipes(id) ON DELETE CASCADE,
    limit_type VARCHAR(50) NOT NULL, -- 'daily', 'weekly', 'total'
    max_uses INTEGER NOT NULL CHECK (max_uses > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(recipe_id, limit_type),
    CONSTRAINT chk_recipe_limits_limit_type CHECK (limit_type IN ('daily', 'weekly', 'total'))
);

-- Create production_tasks table
CREATE TABLE IF NOT EXISTS production.production_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    recipe_id UUID NOT NULL REFERENCES production.recipes(id),
    slot_number INTEGER NOT NULL CHECK (slot_number >= 1),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'completed', 'claimed', 'cancelled'
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completion_time TIMESTAMP WITH TIME ZONE NOT NULL,
    claimed_at TIMESTAMP WITH TIME ZONE,
    pre_calculated_results JSONB NOT NULL,
    modifiers_applied JSONB,
    reservation_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_production_tasks_user_id_not_empty CHECK (length(trim(user_id)) > 0),
    CONSTRAINT chk_production_tasks_status CHECK (status IN ('active', 'completed', 'claimed', 'cancelled')),
    CONSTRAINT chk_production_tasks_completion_time CHECK (completion_time > started_at)
);

-- Create task_output_items table for pre-calculated results
CREATE TABLE IF NOT EXISTS production.task_output_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES production.production_tasks(id) ON DELETE CASCADE,
    item_code VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_task_output_items_item_code_not_empty CHECK (length(trim(item_code)) > 0)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_recipes_code ON production.recipes(code);
CREATE INDEX IF NOT EXISTS idx_recipes_operation_class ON production.recipes(operation_class_code);
CREATE INDEX IF NOT EXISTS idx_recipes_active ON production.recipes(is_active);

CREATE INDEX IF NOT EXISTS idx_recipe_inputs_recipe_id ON production.recipe_inputs(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_inputs_item_code ON production.recipe_inputs(item_code);

CREATE INDEX IF NOT EXISTS idx_recipe_outputs_recipe_id ON production.recipe_outputs(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_outputs_item_code ON production.recipe_outputs(item_code);

CREATE INDEX IF NOT EXISTS idx_recipe_limits_recipe_id ON production.recipe_limits(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_limits_type ON production.recipe_limits(limit_type);

CREATE INDEX IF NOT EXISTS idx_production_tasks_user_id ON production.production_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_production_tasks_status ON production.production_tasks(status);
CREATE INDEX IF NOT EXISTS idx_production_tasks_completion_time ON production.production_tasks(completion_time);
CREATE INDEX IF NOT EXISTS idx_production_tasks_user_slot ON production.production_tasks(user_id, slot_number);
CREATE INDEX IF NOT EXISTS idx_production_tasks_recipe_id ON production.production_tasks(recipe_id);

CREATE INDEX IF NOT EXISTS idx_task_output_items_task_id ON production.task_output_items(task_id);
CREATE INDEX IF NOT EXISTS idx_task_output_items_item_code ON production.task_output_items(item_code);

-- Create updated_at trigger function if not exists
CREATE OR REPLACE FUNCTION production.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_recipes_updated_at ON production.recipes;
CREATE TRIGGER update_recipes_updated_at 
    BEFORE UPDATE ON production.recipes
    FOR EACH ROW EXECUTE FUNCTION production.update_updated_at_column();

DROP TRIGGER IF EXISTS update_production_tasks_updated_at ON production.production_tasks;
CREATE TRIGGER update_production_tasks_updated_at 
    BEFORE UPDATE ON production.production_tasks
    FOR EACH ROW EXECUTE FUNCTION production.update_updated_at_column();

-- Add table comments for documentation
COMMENT ON TABLE production.recipes IS 'Production recipes defining how to create items';
COMMENT ON TABLE production.recipe_inputs IS 'Required input items for each recipe';
COMMENT ON TABLE production.recipe_outputs IS 'Possible output items from each recipe with probabilities';
COMMENT ON TABLE production.recipe_limits IS 'Usage limits for recipes (daily, weekly, total)';
COMMENT ON TABLE production.production_tasks IS 'Active and completed production tasks for users';
COMMENT ON TABLE production.task_output_items IS 'Pre-calculated output items for each production task';

-- Add column comments for important fields
COMMENT ON COLUMN production.recipes.operation_class_code IS 'Code from classifier items defining the operation class';
COMMENT ON COLUMN production.recipes.base_time_seconds IS 'Base production time in seconds before modifiers';
COMMENT ON COLUMN production.recipes.unlock_conditions IS 'JSON conditions required to unlock this recipe';
COMMENT ON COLUMN production.production_tasks.pre_calculated_results IS 'JSON with pre-calculated results including modifiers';
COMMENT ON COLUMN production.production_tasks.modifiers_applied IS 'JSON with list of modifiers applied to this task';
COMMENT ON COLUMN production.production_tasks.reservation_id IS 'UUID of inventory reservation for input items';
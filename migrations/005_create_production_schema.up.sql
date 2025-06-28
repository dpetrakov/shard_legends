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
    operation_class_code VARCHAR(50) NOT NULL,
    production_time_seconds INTEGER NOT NULL CHECK (production_time_seconds >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_recipes_code_not_empty CHECK (length(trim(code)) > 0)
);

-- Create recipe_input_items table
CREATE TABLE IF NOT EXISTS production.recipe_input_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production.recipes(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory.items(id),
    collection_code VARCHAR(50),
    quality_level_code VARCHAR(50),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create recipe_output_items table
CREATE TABLE IF NOT EXISTS production.recipe_output_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production.recipes(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory.items(id),
    min_quantity INTEGER NOT NULL CHECK (min_quantity >= 0),
    max_quantity INTEGER NOT NULL CHECK (max_quantity >= min_quantity),
    probability_percent NUMERIC(5,2) NOT NULL CHECK (probability_percent >= 0 AND probability_percent <= 100),
    output_group VARCHAR(50),
    collection_source_input_index INTEGER,
    quality_source_input_index INTEGER,
    fixed_collection_code VARCHAR(50),
    fixed_quality_level_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    user_id UUID NOT NULL,
    recipe_id UUID NOT NULL REFERENCES production.recipes(id),
    slot_number INTEGER NOT NULL CHECK (slot_number >= 1),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    started_at TIMESTAMP WITH TIME ZONE,
    completion_time TIMESTAMP WITH TIME ZONE,
    claimed_at TIMESTAMP WITH TIME ZONE,
    pre_calculated_results JSONB,
    modifiers_applied JSONB,
    reservation_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_production_tasks_status CHECK (status IN ('draft', 'pending', 'in_progress', 'completed', 'claimed', 'cancelled', 'failed'))
);

-- Create task_output_items table for pre-calculated results
CREATE TABLE IF NOT EXISTS production.task_output_items (
    task_id UUID NOT NULL REFERENCES production.production_tasks(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory.items(id),
    collection_id UUID,
    quality_level_id UUID,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    
    PRIMARY KEY (task_id, item_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_recipes_code ON production.recipes(code);
CREATE INDEX IF NOT EXISTS idx_recipes_operation_class ON production.recipes(operation_class_code);
CREATE INDEX IF NOT EXISTS idx_recipes_active ON production.recipes(is_active);

CREATE INDEX IF NOT EXISTS idx_recipe_input_items_recipe_id ON production.recipe_input_items(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_input_items_item_id ON production.recipe_input_items(item_id);

CREATE INDEX IF NOT EXISTS idx_recipe_output_items_recipe_id ON production.recipe_output_items(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_output_items_item_id ON production.recipe_output_items(item_id);

CREATE INDEX IF NOT EXISTS idx_recipe_limits_recipe_id ON production.recipe_limits(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_limits_type ON production.recipe_limits(limit_type);

CREATE INDEX IF NOT EXISTS idx_production_tasks_user_id ON production.production_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_production_tasks_status ON production.production_tasks(status);
CREATE INDEX IF NOT EXISTS idx_production_tasks_completion_time ON production.production_tasks(completion_time);
CREATE INDEX IF NOT EXISTS idx_production_tasks_user_slot ON production.production_tasks(user_id, slot_number);
CREATE INDEX IF NOT EXISTS idx_production_tasks_recipe_id ON production.production_tasks(recipe_id);

-- Index for draft task cleanup operations
CREATE INDEX IF NOT EXISTS idx_production_tasks_draft_cleanup 
ON production.production_tasks(status, created_at) 
WHERE status = 'draft';

-- Unique constraint for idempotency (user_id + recipe_id for active tasks)
CREATE UNIQUE INDEX IF NOT EXISTS idx_production_tasks_user_recipe_active
ON production.production_tasks(user_id, recipe_id)
WHERE status IN ('draft', 'pending', 'in_progress');

-- Composite index for user status and created_at for efficient querying
CREATE INDEX IF NOT EXISTS idx_production_tasks_user_status_created
ON production.production_tasks(user_id, status, created_at);

CREATE INDEX IF NOT EXISTS idx_task_output_items_task_id ON production.task_output_items(task_id);
CREATE INDEX IF NOT EXISTS idx_task_output_items_item_id ON production.task_output_items(item_id);

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
COMMENT ON TABLE production.recipe_input_items IS 'Required input items for each recipe';
COMMENT ON TABLE production.recipe_output_items IS 'Possible output items from each recipe with probabilities';
COMMENT ON TABLE production.recipe_limits IS 'Usage limits for recipes (daily, weekly, total)';
COMMENT ON TABLE production.production_tasks IS 'Active and completed production tasks for users';
COMMENT ON TABLE production.task_output_items IS 'Pre-calculated output items for each production task';

-- Add column comments for important fields
COMMENT ON COLUMN production.recipes.operation_class_code IS 'Code from classifier items defining the operation class';
COMMENT ON COLUMN production.recipes.production_time_seconds IS 'Base production time in seconds before modifiers (0 for instant recipes)';
COMMENT ON COLUMN production.recipes.is_active IS 'Indicates whether the recipe is active';
COMMENT ON COLUMN production.production_tasks.status IS 'Task status: draft (created but inventory not reserved), pending (ready to start), in_progress (running), completed (finished), claimed (results taken), cancelled (user cancelled), failed (system error)';
COMMENT ON COLUMN production.production_tasks.pre_calculated_results IS 'JSON with pre-calculated results including modifiers';
COMMENT ON COLUMN production.production_tasks.modifiers_applied IS 'JSON with list of modifiers applied to this task';
COMMENT ON COLUMN production.production_tasks.reservation_id IS 'UUID of inventory reservation for input items';
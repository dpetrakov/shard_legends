-- Add execution_count field to production_tasks table
ALTER TABLE production.production_tasks 
ADD COLUMN execution_count INTEGER NOT NULL DEFAULT 1 CHECK (execution_count >= 1);

-- Update existing tasks to have execution_count = 1 (default value)
UPDATE production.production_tasks 
SET execution_count = 1 
WHERE execution_count IS NULL; 
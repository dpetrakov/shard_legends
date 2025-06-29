-- Remove execution_count field from production_tasks table
ALTER TABLE production.production_tasks 
DROP COLUMN execution_count; 
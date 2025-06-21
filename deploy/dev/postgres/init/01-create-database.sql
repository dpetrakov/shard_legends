-- PostgreSQL initialization script for development environment
-- This script runs automatically when PostgreSQL container starts for the first time
-- IMPORTANT: Only basic database setup here, all schemas and tables via migrations!

-- Ensure database exists (should already be created by POSTGRES_DB env var)
-- SELECT 'Database already created by container initialization';

-- Grant additional privileges to main user for migration management
GRANT CREATE ON DATABASE shard_legends_dev TO slcw_user;

-- Enable necessary extensions that migrations might need
-- Extensions are created in migrations, but we ensure user can create them
ALTER USER slcw_user CREATEDB;

-- Basic logging setup
\echo 'PostgreSQL dev database initialized successfully'
\echo 'Schemas and tables will be created via migrations'
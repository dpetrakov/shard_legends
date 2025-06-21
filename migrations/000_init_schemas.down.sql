-- Migration DOWN: 000_init_schemas.down.sql
-- Description: Откат инициализации схем для всех микросервисов
-- Created: 2024-12-21

-- Удаление расширений
DROP EXTENSION IF EXISTS "pg_stat_statements";
DROP EXTENSION IF EXISTS "uuid-ossp";

-- Удаление схем (только если они пустые)
-- ВНИМАНИЕ: Эти команды не выполнятся, если в схемах есть объекты
DROP SCHEMA IF EXISTS clan;
DROP SCHEMA IF EXISTS game;
DROP SCHEMA IF EXISTS auth;
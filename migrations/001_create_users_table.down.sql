-- Migration DOWN: 001_create_users_table.down.sql
-- Description: Откат создания таблицы пользователей для сервиса авторизации
-- Service: auth-service
-- Created: 2024-12-21

-- Удаление триггера
DROP TRIGGER IF EXISTS update_users_updated_at ON auth.users;

-- Удаление функции
DROP FUNCTION IF EXISTS auth.update_updated_at_column();

-- Удаление индексов (будут удалены автоматически с таблицей, но указываем явно для ясности)
DROP INDEX IF EXISTS auth.idx_users_active;
DROP INDEX IF EXISTS auth.idx_users_created_at;
DROP INDEX IF EXISTS auth.idx_users_username;
DROP INDEX IF EXISTS auth.idx_users_telegram_id;

-- Удаление таблицы пользователей
DROP TABLE IF EXISTS auth.users;

-- Удаление схемы auth (только если она пустая)
-- ВНИМАНИЕ: Эта команда не выполнится, если в схеме есть другие объекты
DROP SCHEMA IF EXISTS auth;
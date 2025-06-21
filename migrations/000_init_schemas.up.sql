-- Migration UP: 000_init_schemas.up.sql
-- Description: Инициализация схем для всех микросервисов
-- Created: 2024-12-21

-- Создание схем для микросервисов
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS game; 
CREATE SCHEMA IF NOT EXISTS clan;

-- Комментарии к схемам
COMMENT ON SCHEMA auth IS 'Схема для сервиса авторизации и управления пользователями';
COMMENT ON SCHEMA game IS 'Схема для игровой логики и прогресса пользователей';
COMMENT ON SCHEMA clan IS 'Схема для системы кланов и гильдий';

-- Создание общих расширений PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Комментарии к расширениям  
COMMENT ON EXTENSION "uuid-ossp" IS 'Генерация UUID для первичных ключей';
COMMENT ON EXTENSION "pg_stat_statements" IS 'Статистика выполнения SQL запросов для мониторинга';

-- Настройка прав для схем
GRANT ALL PRIVILEGES ON SCHEMA auth TO slcw_user;
GRANT ALL PRIVILEGES ON SCHEMA game TO slcw_user;
GRANT ALL PRIVILEGES ON SCHEMA clan TO slcw_user;

-- Установка прав по умолчанию для будущих объектов
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON TABLES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game GRANT ALL ON TABLES TO slcw_user; 
ALTER DEFAULT PRIVILEGES IN SCHEMA clan GRANT ALL ON TABLES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON SEQUENCES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game GRANT ALL ON SEQUENCES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA clan GRANT ALL ON SEQUENCES TO slcw_user;
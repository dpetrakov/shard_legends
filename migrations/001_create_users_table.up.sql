-- Migration UP: 001_create_users_table.up.sql
-- Description: Создание таблицы пользователей для сервиса авторизации
-- Service: auth-service
-- Created: 2024-12-21
-- Depends: 000_init_schemas.up.sql (схема auth должна существовать)

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    language_code VARCHAR(10),
    is_premium BOOLEAN DEFAULT FALSE,
    photo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Constraints для валидации данных
    CONSTRAINT users_telegram_id_positive CHECK (telegram_id > 0),
    CONSTRAINT users_first_name_not_empty CHECK (length(trim(first_name)) > 0)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON auth.users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON auth.users(username) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON auth.users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_active ON auth.users(is_active) WHERE is_active = TRUE;

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Триггер для автоматического обновления updated_at при изменении записи
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_updated_at_column();

-- Комментарии к таблице и столбцам
COMMENT ON TABLE auth.users IS 'Базовые данные пользователей из Telegram Web App для авторизации';
COMMENT ON COLUMN auth.users.id IS 'Внутренний UUID пользователя для системы (первичный ключ)';
COMMENT ON COLUMN auth.users.telegram_id IS 'Уникальный ID пользователя в Telegram (обязательное поле)';
COMMENT ON COLUMN auth.users.username IS 'Username пользователя в Telegram (может отсутствовать)';
COMMENT ON COLUMN auth.users.first_name IS 'Имя пользователя в Telegram (обязательное поле)';
COMMENT ON COLUMN auth.users.last_name IS 'Фамилия пользователя в Telegram (может отсутствовать)';
COMMENT ON COLUMN auth.users.language_code IS 'Код языка пользователя (en, ru, es и т.д.)';
COMMENT ON COLUMN auth.users.is_premium IS 'Статус Telegram Premium пользователя';
COMMENT ON COLUMN auth.users.photo_url IS 'URL фото профиля пользователя в Telegram';
COMMENT ON COLUMN auth.users.created_at IS 'Время создания записи пользователя';
COMMENT ON COLUMN auth.users.updated_at IS 'Время последнего обновления записи';
COMMENT ON COLUMN auth.users.last_login_at IS 'Время последней авторизации пользователя';
COMMENT ON COLUMN auth.users.is_active IS 'Активность пользователя (для мягкого удаления)';

-- Комментарии к функциям и триггерам
COMMENT ON FUNCTION auth.update_updated_at_column() IS 'Функция для автоматического обновления поля updated_at';
COMMENT ON TRIGGER update_users_updated_at ON auth.users IS 'Триггер для автоматического обновления updated_at при изменении записи пользователя';
-- Migration UP: 009_populate_languages.up.sql
-- Description: Загрузка поддерживаемых языков системы
-- Service: inventory-service (и другие сервисы)
-- Depends: 008_create_i18n_schema.up.sql
-- Created: 2025-06-28

-- =====================================================
-- Загрузка поддерживаемых языков
-- =====================================================

-- Загружаем базовые языки
INSERT INTO i18n.languages (code, name, is_default, is_active) VALUES
    ('en', 'English', TRUE, TRUE),
    ('ru', 'Русский', FALSE, TRUE)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    is_default = EXCLUDED.is_default,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =====================================================
-- Проверка загруженных данных
-- =====================================================
DO $$
DECLARE
    v_languages_count INTEGER;
    v_default_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_languages_count FROM i18n.languages;
    SELECT COUNT(*) INTO v_default_count FROM i18n.languages WHERE is_default = TRUE;
    
    RAISE NOTICE 'Загружено языков: %', v_languages_count;
    RAISE NOTICE 'Языков по умолчанию: %', v_default_count;
    
    IF v_default_count != 1 THEN
        RAISE EXCEPTION 'Должен быть ровно один язык по умолчанию, найдено: %', v_default_count;
    END IF;
    
    IF v_languages_count < 2 THEN
        RAISE WARNING 'Загружено мало языков (ожидается минимум 2): %', v_languages_count;
    END IF;
END $$;

-- =====================================================
-- Завершение миграции
-- =====================================================
-- Миграция успешно завершена
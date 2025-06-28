-- Migration UP: 008_create_i18n_schema.up.sql
-- Description: Создание схемы i18n для системы интернационализации
-- Service: inventory-service (и другие сервисы)
-- Depends: 000_init_schemas.up.sql
-- Created: 2025-06-28

-- =====================================================
-- Создание схемы i18n
-- =====================================================
CREATE SCHEMA IF NOT EXISTS i18n;

-- Комментарий к схеме
COMMENT ON SCHEMA i18n IS 'Схема для системы интернационализации (i18n) - переводы названий и описаний игровых сущностей';

-- Настройка прав для схемы
GRANT ALL PRIVILEGES ON SCHEMA i18n TO slcw_user;

-- Установка прав по умолчанию для будущих объектов
ALTER DEFAULT PRIVILEGES IN SCHEMA i18n GRANT ALL ON TABLES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA i18n GRANT ALL ON SEQUENCES TO slcw_user;

-- =====================================================
-- Поддерживаемые языки
-- =====================================================
CREATE TABLE i18n.languages (
    code VARCHAR(5) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    CONSTRAINT chk_languages_code_not_empty CHECK (LENGTH(TRIM(code)) > 0),
    CONSTRAINT chk_languages_name_not_empty CHECK (LENGTH(TRIM(name)) > 0)
);

-- Индексы для языков
CREATE INDEX idx_languages_default ON i18n.languages (is_default) WHERE is_default = TRUE;
CREATE INDEX idx_languages_active ON i18n.languages (is_active) WHERE is_active = TRUE;

-- Комментарии к таблице и полям
COMMENT ON TABLE i18n.languages IS 'Справочник поддерживаемых языков системы. Обязательно должен содержать один язык с is_default = true.';
COMMENT ON COLUMN i18n.languages.code IS 'Код языка в формате ISO 639-1 или BCP 47 (ru, en, zh-CN)';
COMMENT ON COLUMN i18n.languages.name IS 'Название языка на этом же языке (Русский, English, 中文)';
COMMENT ON COLUMN i18n.languages.is_default IS 'Базовый язык для fallback при отсутствии перевода';
COMMENT ON COLUMN i18n.languages.is_active IS 'Активен ли язык (неактивные языки скрыты в UI)';
COMMENT ON COLUMN i18n.languages.created_at IS 'Время создания записи';
COMMENT ON COLUMN i18n.languages.updated_at IS 'Время последнего обновления';

-- =====================================================
-- Универсальная таблица переводов
-- =====================================================
CREATE TABLE i18n.translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    field_name VARCHAR(50) NOT NULL,
    language_code VARCHAR(5) NOT NULL REFERENCES i18n.languages(code) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    CONSTRAINT chk_translations_entity_type_not_empty CHECK (LENGTH(TRIM(entity_type)) > 0),
    CONSTRAINT chk_translations_field_name_not_empty CHECK (LENGTH(TRIM(field_name)) > 0),
    CONSTRAINT chk_translations_content_not_empty CHECK (LENGTH(TRIM(content)) > 0),
    UNIQUE(entity_type, entity_id, field_name, language_code)
);

-- Основные индексы для производительности
CREATE INDEX idx_translations_lookup 
ON i18n.translations (entity_type, entity_id, field_name, language_code);

CREATE INDEX idx_translations_language 
ON i18n.translations (language_code, entity_type);

CREATE INDEX idx_translations_entity 
ON i18n.translations (entity_type, entity_id);

-- Комментарии к таблице и полям
COMMENT ON TABLE i18n.translations IS 'Универсальная таблица переводов для всех типов игровых сущностей';
COMMENT ON COLUMN i18n.translations.id IS 'UUID записи перевода';
COMMENT ON COLUMN i18n.translations.entity_type IS 'Тип сущности: item, classifier, classifier_item, achievement, recipe, user_message';
COMMENT ON COLUMN i18n.translations.entity_id IS 'UUID сущности из соответствующей таблицы';
COMMENT ON COLUMN i18n.translations.field_name IS 'Имя переводимого поля: name, description, tooltip, title, content';
COMMENT ON COLUMN i18n.translations.language_code IS 'Код языка из справочника languages';
COMMENT ON COLUMN i18n.translations.content IS 'Переведенный текст (максимум ~1GB для TEXT)';
COMMENT ON COLUMN i18n.translations.created_at IS 'Время создания перевода';
COMMENT ON COLUMN i18n.translations.updated_at IS 'Время последнего обновления перевода';

-- =====================================================
-- Функции для автоматического обновления updated_at
-- =====================================================
CREATE OR REPLACE FUNCTION i18n.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_languages_updated_at 
    BEFORE UPDATE ON i18n.languages 
    FOR EACH ROW EXECUTE FUNCTION i18n.update_updated_at_column();

CREATE TRIGGER update_translations_updated_at 
    BEFORE UPDATE ON i18n.translations 
    FOR EACH ROW EXECUTE FUNCTION i18n.update_updated_at_column();

-- =====================================================
-- Ограничения уникальности для языков по умолчанию
-- =====================================================
-- Создаем уникальный индекс для обеспечения только одного языка по умолчанию
CREATE UNIQUE INDEX idx_languages_single_default 
ON i18n.languages (is_default) WHERE is_default = TRUE;

-- =====================================================
-- Функции для работы с переводами
-- =====================================================

-- Функция для получения перевода с fallback
CREATE OR REPLACE FUNCTION i18n.get_translation(
    p_entity_type VARCHAR(50),
    p_entity_id UUID,
    p_field_name VARCHAR(50),
    p_language_code VARCHAR(5) DEFAULT NULL
)
RETURNS TEXT AS $$
DECLARE
    v_translation TEXT;
    v_default_language VARCHAR(5);
BEGIN
    -- Если язык не указан, используем язык по умолчанию
    IF p_language_code IS NULL THEN
        SELECT code INTO v_default_language 
        FROM i18n.languages 
        WHERE is_default = TRUE AND is_active = TRUE
        LIMIT 1;
        
        p_language_code := COALESCE(v_default_language, 'en');
    END IF;
    
    -- Пытаемся найти перевод на запрошенном языке
    SELECT content INTO v_translation
    FROM i18n.translations
    WHERE entity_type = p_entity_type
      AND entity_id = p_entity_id
      AND field_name = p_field_name
      AND language_code = p_language_code;
    
    -- Если перевод найден, возвращаем его
    IF v_translation IS NOT NULL THEN
        RETURN v_translation;
    END IF;
    
    -- Иначе ищем fallback на языке по умолчанию
    SELECT code INTO v_default_language 
    FROM i18n.languages 
    WHERE is_default = TRUE AND is_active = TRUE
    LIMIT 1;
    
    IF v_default_language IS NOT NULL AND v_default_language != p_language_code THEN
        SELECT content INTO v_translation
        FROM i18n.translations
        WHERE entity_type = p_entity_type
          AND entity_id = p_entity_id
          AND field_name = p_field_name
          AND language_code = v_default_language;
    END IF;
    
    RETURN v_translation;
END;
$$ LANGUAGE plpgsql;

-- Комментарий к функции
COMMENT ON FUNCTION i18n.get_translation(VARCHAR(50), UUID, VARCHAR(50), VARCHAR(5)) IS 
'Получение перевода с автоматическим fallback на язык по умолчанию при отсутствии перевода на запрошенном языке';

-- =====================================================
-- Завершение миграции
-- =====================================================
-- Миграция успешно завершена
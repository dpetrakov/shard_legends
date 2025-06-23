-- Migration UP: 002_create_inventory_schema.up.sql
-- Description: Создание схемы inventory и всех таблиц
-- Service: inventory-service
-- Depends: 000_init_schemas.up.sql
-- Created: 2025-06-23

-- =====================================================
-- Создание схемы inventory
-- =====================================================
CREATE SCHEMA IF NOT EXISTS inventory;

-- Комментарий к схеме
COMMENT ON SCHEMA inventory IS 'Схема для системы инвентаря и управления предметами';

-- Настройка прав для схемы
GRANT ALL PRIVILEGES ON SCHEMA inventory TO slcw_user;

-- Установка прав по умолчанию для будущих объектов
ALTER DEFAULT PRIVILEGES IN SCHEMA inventory GRANT ALL ON TABLES TO slcw_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA inventory GRANT ALL ON SEQUENCES TO slcw_user;

-- =====================================================
-- Общий классификатор для справочных данных
-- =====================================================
CREATE TABLE inventory.classifiers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(100) UNIQUE NOT NULL,
    description text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    
    CONSTRAINT chk_classifiers_code_not_empty CHECK (length(trim(code)) > 0)
);

-- Индекс для быстрого поиска по коду классификатора
CREATE UNIQUE INDEX idx_classifiers_code ON inventory.classifiers (code);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.classifiers IS 'Общий классификатор для всех справочных данных системы. Содержит категории: классы предметов, типы предметов, уровни качества, типы операций и др.';
COMMENT ON COLUMN inventory.classifiers.id IS 'UUID классификатора';
COMMENT ON COLUMN inventory.classifiers.code IS 'Код классификатора (item_class, item_type, quality_level, operation_type)';
COMMENT ON COLUMN inventory.classifiers.description IS 'Описание назначения классификатора';
COMMENT ON COLUMN inventory.classifiers.created_at IS 'Время создания';
COMMENT ON COLUMN inventory.classifiers.updated_at IS 'Время последнего обновления';

-- =====================================================
-- Элементы классификаторов
-- =====================================================
CREATE TABLE inventory.classifier_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    classifier_id uuid NOT NULL REFERENCES inventory.classifiers(id),
    code varchar(100) NOT NULL,
    description text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    
    CONSTRAINT chk_classifier_items_code_not_empty CHECK (length(trim(code)) > 0)
);

-- Индексы для поиска элементов классификаторов
CREATE UNIQUE INDEX idx_classifier_items_classifier_code ON inventory.classifier_items (classifier_id, code);
CREATE INDEX idx_classifier_items_classifier ON inventory.classifier_items (classifier_id);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.classifier_items IS 'Элементы классификаторов - конкретные значения внутри каждого классификатора.';
COMMENT ON COLUMN inventory.classifier_items.id IS 'UUID элемента классификатора';
COMMENT ON COLUMN inventory.classifier_items.classifier_id IS 'Ссылка на классификатор';
COMMENT ON COLUMN inventory.classifier_items.code IS 'Код элемента в рамках классификатора';
COMMENT ON COLUMN inventory.classifier_items.description IS 'Описание элемента';
COMMENT ON COLUMN inventory.classifier_items.created_at IS 'Время создания';
COMMENT ON COLUMN inventory.classifier_items.updated_at IS 'Время последнего обновления';

-- =====================================================
-- Предметы
-- =====================================================
CREATE TABLE inventory.items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    item_class_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    item_type_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    quality_levels_classifier_id uuid NOT NULL REFERENCES inventory.classifiers(id),
    collections_classifier_id uuid NOT NULL REFERENCES inventory.classifiers(id),
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL
);

-- Индексы для поиска предметов
CREATE INDEX idx_items_class ON inventory.items (item_class_id);
CREATE INDEX idx_items_type ON inventory.items (item_type_id);
CREATE INDEX idx_items_quality_classifier ON inventory.items (quality_levels_classifier_id);
CREATE INDEX idx_items_collections_classifier ON inventory.items (collections_classifier_id);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.items IS 'Каталог всех игровых предметов. Каждый предмет принадлежит к определенному классу и типу из классификаторов.';
COMMENT ON COLUMN inventory.items.id IS 'UUID предмета';
COMMENT ON COLUMN inventory.items.item_class_id IS 'Класс предмета (ссылка на элемент классификатора item_class)';
COMMENT ON COLUMN inventory.items.item_type_id IS 'Тип предмета (ссылка на элемент классификатора item_type)';
COMMENT ON COLUMN inventory.items.quality_levels_classifier_id IS 'Ссылка на классификатор уровней качества для данного предмета';
COMMENT ON COLUMN inventory.items.collections_classifier_id IS 'Ссылка на классификатор коллекций для данного предмета';
COMMENT ON COLUMN inventory.items.created_at IS 'Время создания';
COMMENT ON COLUMN inventory.items.updated_at IS 'Время последнего обновления';

-- =====================================================
-- Изображения предметов
-- =====================================================
CREATE TABLE inventory.item_images (
    item_id uuid NOT NULL REFERENCES inventory.items(id),
    collection_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    quality_level_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    image_url text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    
    PRIMARY KEY (item_id, collection_id, quality_level_id),
    CONSTRAINT chk_item_images_url_not_empty CHECK (length(trim(image_url)) > 0)
);

-- Индексы для поиска изображений
CREATE INDEX idx_item_images_item ON inventory.item_images (item_id);
CREATE INDEX idx_item_images_collection_quality ON inventory.item_images (collection_id, quality_level_id);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.item_images IS 'Хранит URL изображений для каждой уникальной комбинации предмет+коллекция+качество.';
COMMENT ON COLUMN inventory.item_images.item_id IS 'UUID предмета';
COMMENT ON COLUMN inventory.item_images.collection_id IS 'Коллекция предмета';
COMMENT ON COLUMN inventory.item_images.quality_level_id IS 'Уровень качества предмета';
COMMENT ON COLUMN inventory.item_images.image_url IS 'URL изображения для данной комбинации предмет+коллекция+качество';
COMMENT ON COLUMN inventory.item_images.created_at IS 'Время создания';
COMMENT ON COLUMN inventory.item_images.updated_at IS 'Время последнего обновления';

-- =====================================================
-- Дневные остатки (снимки на конец дня)
-- =====================================================
CREATE TABLE inventory.daily_balances (
    user_id uuid NOT NULL,
    section_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    item_id uuid NOT NULL REFERENCES inventory.items(id),
    collection_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    quality_level_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    balance_date date NOT NULL,
    quantity bigint NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    
    PRIMARY KEY (user_id, section_id, item_id, collection_id, quality_level_id, balance_date),
    CONSTRAINT chk_daily_balances_quantity_positive CHECK (quantity >= 0)
);

-- Индексы для поиска остатков
CREATE INDEX idx_daily_balances_date ON inventory.daily_balances (balance_date);
CREATE INDEX idx_daily_balances_user_date ON inventory.daily_balances (user_id, balance_date);
CREATE INDEX idx_daily_balances_created ON inventory.daily_balances (created_at);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.daily_balances IS 'Дневные остатки - основа для расчета текущих остатков инвентаря. Текущий остаток рассчитывается динамически по формуле: текущий_остаток = дневной_остаток + сумма_операций_за_текущий_день';
COMMENT ON COLUMN inventory.daily_balances.user_id IS 'UUID пользователя';
COMMENT ON COLUMN inventory.daily_balances.section_id IS 'Раздел инвентаря';
COMMENT ON COLUMN inventory.daily_balances.item_id IS 'UUID предмета';
COMMENT ON COLUMN inventory.daily_balances.collection_id IS 'Коллекция предмета';
COMMENT ON COLUMN inventory.daily_balances.quality_level_id IS 'Уровень качества предмета';
COMMENT ON COLUMN inventory.daily_balances.balance_date IS 'Дата остатка (конец дня в UTC)';
COMMENT ON COLUMN inventory.daily_balances.quantity IS 'Количество предметов на конец дня';
COMMENT ON COLUMN inventory.daily_balances.created_at IS 'Время создания записи';

-- =====================================================
-- Операции с инвентарем (протокол всех изменений)
-- =====================================================
CREATE TABLE inventory.operations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    section_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    item_id uuid NOT NULL REFERENCES inventory.items(id),
    collection_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    quality_level_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    quantity_change bigint NOT NULL,
    operation_type_id uuid NOT NULL REFERENCES inventory.classifier_items(id),
    operation_id uuid,
    recipe_id uuid,
    comment text,
    created_at timestamptz DEFAULT now() NOT NULL,
    
    CONSTRAINT chk_operations_quantity_change_not_zero CHECK (quantity_change != 0)
);

-- Индексы для поиска операций
CREATE INDEX idx_operations_user ON inventory.operations (user_id);
CREATE INDEX idx_operations_user_time ON inventory.operations (user_id, created_at);
CREATE INDEX idx_operations_item_time ON inventory.operations (user_id, section_id, item_id, collection_id, quality_level_id, created_at);
CREATE INDEX idx_operations_type ON inventory.operations (operation_type_id);
CREATE INDEX idx_operations_external ON inventory.operations (operation_id);
CREATE INDEX idx_operations_recipe ON inventory.operations (recipe_id);
CREATE INDEX idx_operations_time ON inventory.operations (created_at);

-- Комментарии к таблице и полям
COMMENT ON TABLE inventory.operations IS 'Полный протокол всех операций с инвентарем пользователей. Каждое изменение количества предметов должно сопровождаться записью в этой таблице.';
COMMENT ON COLUMN inventory.operations.id IS 'UUID операции';
COMMENT ON COLUMN inventory.operations.user_id IS 'UUID пользователя';
COMMENT ON COLUMN inventory.operations.section_id IS 'Раздел инвентаря';
COMMENT ON COLUMN inventory.operations.item_id IS 'UUID предмета';
COMMENT ON COLUMN inventory.operations.collection_id IS 'Коллекция предмета';
COMMENT ON COLUMN inventory.operations.quality_level_id IS 'Уровень качества предмета';
COMMENT ON COLUMN inventory.operations.quantity_change IS 'Изменение количества (+ поступление, - расход)';
COMMENT ON COLUMN inventory.operations.operation_type_id IS 'Тип операции из классификатора';
COMMENT ON COLUMN inventory.operations.operation_id IS 'Внешний UUID операции для связи с другими системами';
COMMENT ON COLUMN inventory.operations.recipe_id IS 'UUID рецепта для контроля лимитов (опционально)';
COMMENT ON COLUMN inventory.operations.comment IS 'Комментарий к операции';
COMMENT ON COLUMN inventory.operations.created_at IS 'Время выполнения операции';

-- =====================================================
-- Функции для автоматического обновления updated_at
-- =====================================================
CREATE OR REPLACE FUNCTION inventory.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_classifiers_updated_at 
    BEFORE UPDATE ON inventory.classifiers 
    FOR EACH ROW EXECUTE FUNCTION inventory.update_updated_at_column();

CREATE TRIGGER update_classifier_items_updated_at 
    BEFORE UPDATE ON inventory.classifier_items 
    FOR EACH ROW EXECUTE FUNCTION inventory.update_updated_at_column();

CREATE TRIGGER update_items_updated_at 
    BEFORE UPDATE ON inventory.items 
    FOR EACH ROW EXECUTE FUNCTION inventory.update_updated_at_column();

CREATE TRIGGER update_item_images_updated_at 
    BEFORE UPDATE ON inventory.item_images 
    FOR EACH ROW EXECUTE FUNCTION inventory.update_updated_at_column();

-- =====================================================
-- Завершение миграции
-- =====================================================
-- Миграция успешно завершена
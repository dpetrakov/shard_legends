-- =============================================================================
-- DEV DATA Script: 003_reset_items.sql  
-- Description: Полная перезагрузка предметов из спецификации items-initial.md с i18n поддержкой
-- Service: inventory-service
-- Created: 2025-06-28
-- Depends: 002_reset_classifiers.sql, 008_create_i18n_schema.up.sql, 009_populate_languages.up.sql
--
-- ВНИМАНИЕ: Этот скрипт предназначен только для dev окружения!
-- Он полностью очищает и перезагружает все предметы и их переводы.
-- =============================================================================

BEGIN;

-- =============================================================================
-- Шаг 2: Очистка существующих данных предметов и переводов
-- =============================================================================
DELETE FROM i18n.translations WHERE entity_type = 'item';
TRUNCATE TABLE inventory.item_images CASCADE;
TRUNCATE TABLE inventory.items CASCADE;

-- =============================================================================
-- Шаг 3: Загрузка предметов из спецификации (без name и description)
-- =============================================================================

-- 1. Ресурсы (resources)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    (
        '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60',
        (SELECT id FROM inventory.classifier_items WHERE code = 'resources' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'stone' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'resource_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '65ab1f83-e2c0-48b4-9173-1010a81d13ad',
        (SELECT id FROM inventory.classifier_items WHERE code = 'resources' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'wood' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'resource_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        'df5231f1-6cb8-4a5d-9a70-fd3214966c89',
        (SELECT id FROM inventory.classifier_items WHERE code = 'resources' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'ore' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'resource_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '3b6529fd-9083-4bc5-8d71-493eb6a30d46',
        (SELECT id FROM inventory.classifier_items WHERE code = 'resources' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'diamond' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'resource_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    )
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 2. Реагенты (reagents)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    (
        '0eb6c9d4-38af-49ac-b6e9-9c78d7f7d11a',
        (SELECT id FROM inventory.classifier_items WHERE code = 'reagents' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'abrasive' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '321becd3-d06c-4523-ae16-41d6f56c7cc8',
        (SELECT id FROM inventory.classifier_items WHERE code = 'reagents' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'disc' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        'd0b2ed5d-6e3d-45e8-9f5f-07fd8dcbc2a5',
        (SELECT id FROM inventory.classifier_items WHERE code = 'reagents' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'inductor' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '18669d60-2e69-4e3f-9ba8-0ce299f1a827',
        (SELECT id FROM inventory.classifier_items WHERE code = 'reagents' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'paste' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    )
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 3. Ускорители (boosters)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    (
        'e9a746e0-8a40-4ff5-b582-3a1b7e3db3d6',
        (SELECT id FROM inventory.classifier_items WHERE code = 'boosters' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'repair_tool' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'booster_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '6a8fa441-2de9-4044-9d04-8461aa1bea2d',
        (SELECT id FROM inventory.classifier_items WHERE code = 'boosters' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'speed_processing' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'booster_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '3cd406c4-8eaf-4ec7-9c45-0020211d2b8d',
        (SELECT id FROM inventory.classifier_items WHERE code = 'boosters' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'speed_crafting' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'booster_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    )
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 4. Ключи (keys)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    (
        '7e1d9e48-49cd-4ef2-b93e-1e32a0cb9a18',
        (SELECT id FROM inventory.classifier_items WHERE code = 'keys' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'key' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'key_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    ),
    (
        '4e25a5dc-a814-4102-b5cb-54c39fce969f',
        (SELECT id FROM inventory.classifier_items WHERE code = 'keys' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
        (SELECT id FROM inventory.classifier_items WHERE code = 'blueprint_key' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'key_type')),
        (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
        (SELECT id FROM inventory.classifiers WHERE code = 'collection')
    )
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 5. Сундуки (chests) - создаем новый тип классификатора для chest_type
-- Сначала добавим классификатор chest_type, если его нет
INSERT INTO inventory.classifiers (id, code, description) VALUES
    (gen_random_uuid(), 'chest_type', 'Типы сундуков для класса chests')
ON CONFLICT (code) DO NOTHING;

-- Добавим элементы классификатора chest_type
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'chest_type'), 'resource_chest', 'Ресурсный сундук'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'chest_type'), 'reagent_chest', 'Реагентный сундук'), 
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'chest_type'), 'booster_chest', 'Сундук ускорителей'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'chest_type'), 'blueprint_chest', 'Сундук чертежей')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- Добавим сундуки (4 типа по одному ID)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    -- Один ресурсный сундук (разные качества будут через quality_level_id в операциях)
    ('9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 
     (SELECT id FROM inventory.classifier_items WHERE code = 'chests' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'resource_chest' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'chest_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    -- Один реагентный сундук
    ('a2e20668-380d-43eb-87db-cb19e4fed0ab',
     (SELECT id FROM inventory.classifier_items WHERE code = 'chests' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'reagent_chest' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'chest_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    -- Один сундук ускорителей
    ('3b5c8322-c00d-44e2-875e-d5bd9097d1c4',
     (SELECT id FROM inventory.classifier_items WHERE code = 'chests' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'booster_chest' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'chest_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    -- Один сундук чертежей
    ('012d9076-a37d-4e9d-a49a-fbc7a07e5bd9',
     (SELECT id FROM inventory.classifier_items WHERE code = 'chests' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'blueprint_chest' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'chest_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection'))
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 6. Чертежи (blueprints - 4 штуки)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    ('4fe26d9e-a4ac-47e6-a7de-4925ae53c0b0',
     (SELECT id FROM inventory.classifier_items WHERE code = 'blueprints' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'shovel' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('bf0a5e59-c463-4a8a-a55f-e32401238ac2',
     (SELECT id FROM inventory.classifier_items WHERE code = 'blueprints' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'sickle' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('ed22053e-569b-4f67-9c65-3986fcfafafb',
     (SELECT id FROM inventory.classifier_items WHERE code = 'blueprints' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'axe' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('8ca31e76-9969-4f69-9bba-df3f3beaa7d4',
     (SELECT id FROM inventory.classifier_items WHERE code = 'blueprints' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'pickaxe' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), (SELECT id FROM inventory.classifiers WHERE code = 'collection'))
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- 7. Инструменты (tools - 4 штуки)
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
    ('6bb18b46-8f3c-45e9-b483-7edaf512a74c',
     (SELECT id FROM inventory.classifier_items WHERE code = 'tools' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'shovel' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('3677ad6c-cd73-4a7b-8c62-6d1ad6bc26ab',
     (SELECT id FROM inventory.classifier_items WHERE code = 'tools' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'sickle' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('f347b613-70f2-4744-b2e8-1005c6614c29',
     (SELECT id FROM inventory.classifier_items WHERE code = 'tools' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'axe' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), (SELECT id FROM inventory.classifiers WHERE code = 'collection')),
    ('9cdbee72-713d-4e0e-9719-c9bc8bcb0533',
     (SELECT id FROM inventory.classifier_items WHERE code = 'tools' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'item_class')),
     (SELECT id FROM inventory.classifier_items WHERE code = 'pickaxe' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'tool_type')),
     (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), (SELECT id FROM inventory.classifiers WHERE code = 'collection'))
ON CONFLICT (id) DO UPDATE SET
    item_class_id = EXCLUDED.item_class_id,
    item_type_id = EXCLUDED.item_type_id,
    quality_levels_classifier_id = EXCLUDED.quality_levels_classifier_id,
    collections_classifier_id = EXCLUDED.collections_classifier_id,
    updated_at = now();

-- =============================================================================
-- Шаг 4: Загрузка переводов предметов на русском и английском языках
-- =============================================================================

-- Переводы ресурсов
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Камень
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'name', 'ru', 'Камень'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'name', 'en', 'Stone'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'description', 'ru', 'Базовый строительный материал, используется в большинстве рецептов и заданий.'),
    ('item', '1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60', 'description', 'en', 'Basic building material used in most recipes and quests.'),
    
    -- Дерево
    ('item', '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 'name', 'ru', 'Дерево'),
    ('item', '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 'name', 'en', 'Wood'),
    ('item', '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 'description', 'ru', 'Универсальный ресурс для строительства и крафта базовых предметов.'),
    ('item', '65ab1f83-e2c0-48b4-9173-1010a81d13ad', 'description', 'en', 'Universal resource for construction and crafting basic items.'),
    
    -- Руда
    ('item', 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 'name', 'ru', 'Руда'),
    ('item', 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 'name', 'en', 'Ore'),
    ('item', 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 'description', 'ru', 'Сырьё для переплавки в металлические слитки.'),
    ('item', 'df5231f1-6cb8-4a5d-9a70-fd3214966c89', 'description', 'en', 'Raw material for smelting into metal ingots.'),
    
    -- Алмаз
    ('item', '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 'name', 'ru', 'Алмаз'),
    ('item', '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 'name', 'en', 'Diamond'),
    ('item', '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 'description', 'ru', 'Редкий и ценный ресурс, применяется для создания бриллиантовых инструментов и особых предметов.'),
    ('item', '3b6529fd-9083-4bc5-8d71-493eb6a30d46', 'description', 'en', 'Rare and valuable resource used to create diamond tools and special items.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- Переводы реагентов
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Абразив
    ('item', '0eb6c9d4-38af-49ac-b6e9-9c78d7f7d11a', 'name', 'ru', 'Абразив'),
    ('item', '0eb6c9d4-38af-49ac-b6e9-9c78d7f7d11a', 'name', 'en', 'Abrasive'),
    ('item', '0eb6c9d4-38af-49ac-b6e9-9c78d7f7d11a', 'description', 'ru', 'Используется для шлифовки и обработки материалов.'),
    ('item', '0eb6c9d4-38af-49ac-b6e9-9c78d7f7d11a', 'description', 'en', 'Used for grinding and processing materials.'),
    
    -- Диск
    ('item', '321becd3-d06c-4523-ae16-41d6f56c7cc8', 'name', 'ru', 'Диск'),
    ('item', '321becd3-d06c-4523-ae16-41d6f56c7cc8', 'name', 'en', 'Disc'),
    ('item', '321becd3-d06c-4523-ae16-41d6f56c7cc8', 'description', 'ru', 'Режущий компонент для дерево- и камнеобработки.'),
    ('item', '321becd3-d06c-4523-ae16-41d6f56c7cc8', 'description', 'en', 'Cutting component for woodworking and stone processing.'),
    
    -- Индуктор
    ('item', 'd0b2ed5d-6e3d-45e8-9f5f-07fd8dcbc2a5', 'name', 'ru', 'Индуктор'),
    ('item', 'd0b2ed5d-6e3d-45e8-9f5f-07fd8dcbc2a5', 'name', 'en', 'Inductor'),
    ('item', 'd0b2ed5d-6e3d-45e8-9f5f-07fd8dcbc2a5', 'description', 'ru', 'Электромагнитный элемент для высокотемпературной переплавки.'),
    ('item', 'd0b2ed5d-6e3d-45e8-9f5f-07fd8dcbc2a5', 'description', 'en', 'Electromagnetic element for high-temperature smelting.'),
    
    -- Паста
    ('item', '18669d60-2e69-4e3f-9ba8-0ce299f1a827', 'name', 'ru', 'Паста'),
    ('item', '18669d60-2e69-4e3f-9ba8-0ce299f1a827', 'name', 'en', 'Paste'),
    ('item', '18669d60-2e69-4e3f-9ba8-0ce299f1a827', 'description', 'ru', 'Связующий реагент для улучшения качества материалов.'),
    ('item', '18669d60-2e69-4e3f-9ba8-0ce299f1a827', 'description', 'en', 'Binding reagent for improving material quality.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- Переводы ускорителей
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Набор ремонта
    ('item', 'e9a746e0-8a40-4ff5-b582-3a1b7e3db3d6', 'name', 'ru', 'Набор ремонта'),
    ('item', 'e9a746e0-8a40-4ff5-b582-3a1b7e3db3d6', 'name', 'en', 'Repair Kit'),
    ('item', 'e9a746e0-8a40-4ff5-b582-3a1b7e3db3d6', 'description', 'ru', 'Полностью восстанавливает прочность выбранного инструмента.'),
    ('item', 'e9a746e0-8a40-4ff5-b582-3a1b7e3db3d6', 'description', 'en', 'Fully restores the durability of the selected tool.'),
    
    -- Катализатор переплавки
    ('item', '6a8fa441-2de9-4044-9d04-8461aa1bea2d', 'name', 'ru', 'Катализатор переплавки'),
    ('item', '6a8fa441-2de9-4044-9d04-8461aa1bea2d', 'name', 'en', 'Smelting Catalyst'),
    ('item', '6a8fa441-2de9-4044-9d04-8461aa1bea2d', 'description', 'ru', 'Сокращает время переплавки сырья.'),
    ('item', '6a8fa441-2de9-4044-9d04-8461aa1bea2d', 'description', 'en', 'Reduces smelting time for raw materials.'),
    
    -- Смазка сборщика
    ('item', '3cd406c4-8eaf-4ec7-9c45-0020211d2b8d', 'name', 'ru', 'Смазка сборщика'),
    ('item', '3cd406c4-8eaf-4ec7-9c45-0020211d2b8d', 'name', 'en', 'Assembly Lubricant'),
    ('item', '3cd406c4-8eaf-4ec7-9c45-0020211d2b8d', 'description', 'ru', 'Уменьшает время крафта инструментов.'),
    ('item', '3cd406c4-8eaf-4ec7-9c45-0020211d2b8d', 'description', 'en', 'Reduces tool crafting time.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- Переводы ключей
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Ключ
    ('item', '7e1d9e48-49cd-4ef2-b93e-1e32a0cb9a18', 'name', 'ru', 'Ключ'),
    ('item', '7e1d9e48-49cd-4ef2-b93e-1e32a0cb9a18', 'name', 'en', 'Key'),
    ('item', '7e1d9e48-49cd-4ef2-b93e-1e32a0cb9a18', 'description', 'ru', 'Открывает сундуки размеров S/M/L; фактический размер определяется параметром size.'),
    ('item', '7e1d9e48-49cd-4ef2-b93e-1e32a0cb9a18', 'description', 'en', 'Opens chests of sizes S/M/L; actual size is determined by the size parameter.'),
    
    -- Ключ чертежей
    ('item', '4e25a5dc-a814-4102-b5cb-54c39fce969f', 'name', 'ru', 'Ключ чертежей'),
    ('item', '4e25a5dc-a814-4102-b5cb-54c39fce969f', 'name', 'en', 'Blueprint Key'),
    ('item', '4e25a5dc-a814-4102-b5cb-54c39fce969f', 'description', 'ru', 'Используется для открытия сундука с чертежами (Blueprint Chest).'),
    ('item', '4e25a5dc-a814-4102-b5cb-54c39fce969f', 'description', 'en', 'Used to open a Blueprint Chest.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- Переводы сундуков
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Ресурсный сундук (один предмет с разными качествами)
    ('item', '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 'name', 'ru', 'Ресурсный сундук'),
    ('item', '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 'name', 'en', 'Resource Chest'),
    ('item', '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 'description', 'ru', 'Содержит случайные ресурсы. Размер и количество зависят от качества сундука.'),
    ('item', '9421cc9f-a56e-4c7d-b636-4c8fdfef7166', 'description', 'en', 'Contains random resources. Size and quantity depend on chest quality.'),
    
    -- Реагентный сундук
    ('item', 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 'name', 'ru', 'Реагентный сундук'),
    ('item', 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 'name', 'en', 'Reagent Chest'),
    ('item', 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 'description', 'ru', 'Содержит случайные реагенты для крафта. Размер зависит от качества.'),
    ('item', 'a2e20668-380d-43eb-87db-cb19e4fed0ab', 'description', 'en', 'Contains random reagents for crafting. Size depends on quality.'),
    
    -- Сундук ускорителей
    ('item', '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 'name', 'ru', 'Сундук ускорителей'),
    ('item', '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 'name', 'en', 'Booster Chest'),
    ('item', '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 'description', 'ru', 'Содержит случайные ускорители производства и крафта. Размер зависит от качества.'),
    ('item', '3b5c8322-c00d-44e2-875e-d5bd9097d1c4', 'description', 'en', 'Contains random production and crafting boosters. Size depends on quality.'),
    
    -- Сундук чертежей
    ('item', '012d9076-a37d-4e9d-a49a-fbc7a07e5bd9', 'name', 'ru', 'Сундук чертежей'),
    ('item', '012d9076-a37d-4e9d-a49a-fbc7a07e5bd9', 'name', 'en', 'Blueprint Chest'),
    ('item', '012d9076-a37d-4e9d-a49a-fbc7a07e5bd9', 'description', 'ru', 'Содержит случайный сезонный чертеж инструмента.'),
    ('item', '012d9076-a37d-4e9d-a49a-fbc7a07e5bd9', 'description', 'en', 'Contains a random seasonal tool blueprint.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();



-- Переводы чертежей
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Чертеж: Лопата
    ('item', '4fe26d9e-a4ac-47e6-a7de-4925ae53c0b0', 'name', 'ru', 'Чертеж: Лопата'),
    ('item', '4fe26d9e-a4ac-47e6-a7de-4925ae53c0b0', 'name', 'en', 'Blueprint: Shovel'),
    ('item', '4fe26d9e-a4ac-47e6-a7de-4925ae53c0b0', 'description', 'ru', 'Позволяет крафтить лопату выбранного качества.'),
    ('item', '4fe26d9e-a4ac-47e6-a7de-4925ae53c0b0', 'description', 'en', 'Allows crafting shovels of selected quality.'),
    
    -- Чертеж: Серп
    ('item', 'bf0a5e59-c463-4a8a-a55f-e32401238ac2', 'name', 'ru', 'Чертеж: Серп'),
    ('item', 'bf0a5e59-c463-4a8a-a55f-e32401238ac2', 'name', 'en', 'Blueprint: Sickle'),
    ('item', 'bf0a5e59-c463-4a8a-a55f-e32401238ac2', 'description', 'ru', 'Позволяет крафтить серп выбранного качества.'),
    ('item', 'bf0a5e59-c463-4a8a-a55f-e32401238ac2', 'description', 'en', 'Allows crafting sickles of selected quality.'),
    
    -- Чертеж: Топор
    ('item', 'ed22053e-569b-4f67-9c65-3986fcfafafb', 'name', 'ru', 'Чертеж: Топор'),
    ('item', 'ed22053e-569b-4f67-9c65-3986fcfafafb', 'name', 'en', 'Blueprint: Axe'),
    ('item', 'ed22053e-569b-4f67-9c65-3986fcfafafb', 'description', 'ru', 'Позволяет крафтить топор выбранного качества.'),
    ('item', 'ed22053e-569b-4f67-9c65-3986fcfafafb', 'description', 'en', 'Allows crafting axes of selected quality.'),
    
    -- Чертеж: Кирка
    ('item', '8ca31e76-9969-4f69-9bba-df3f3beaa7d4', 'name', 'ru', 'Чертеж: Кирка'),
    ('item', '8ca31e76-9969-4f69-9bba-df3f3beaa7d4', 'name', 'en', 'Blueprint: Pickaxe'),
    ('item', '8ca31e76-9969-4f69-9bba-df3f3beaa7d4', 'description', 'ru', 'Позволяет крафтить кирку выбранного качества.'),
    ('item', '8ca31e76-9969-4f69-9bba-df3f3beaa7d4', 'description', 'en', 'Allows crafting pickaxes of selected quality.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- Переводы инструментов
INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content) VALUES
    -- Лопата
    ('item', '6bb18b46-8f3c-45e9-b483-7edaf512a74c', 'name', 'ru', 'Лопата'),
    ('item', '6bb18b46-8f3c-45e9-b483-7edaf512a74c', 'name', 'en', 'Shovel'),
    ('item', '6bb18b46-8f3c-45e9-b483-7edaf512a74c', 'description', 'ru', 'Универсальный инструмент для земляных работ; качество и серия задаются отдельными атрибутами.'),
    ('item', '6bb18b46-8f3c-45e9-b483-7edaf512a74c', 'description', 'en', 'Universal tool for earthworks; quality and series are set by separate attributes.'),
    
    -- Серп
    ('item', '3677ad6c-cd73-4a7b-8c62-6d1ad6bc26ab', 'name', 'ru', 'Серп'),
    ('item', '3677ad6c-cd73-4a7b-8c62-6d1ad6bc26ab', 'name', 'en', 'Sickle'),
    ('item', '3677ad6c-cd73-4a7b-8c62-6d1ad6bc26ab', 'description', 'ru', 'Инструмент для сбора урожая и трав; модификации через атрибуты качества/серии.'),
    ('item', '3677ad6c-cd73-4a7b-8c62-6d1ad6bc26ab', 'description', 'en', 'Tool for harvesting crops and herbs; modifications through quality/series attributes.'),
    
    -- Топор
    ('item', 'f347b613-70f2-4744-b2e8-1005c6614c29', 'name', 'ru', 'Топор'),
    ('item', 'f347b613-70f2-4744-b2e8-1005c6614c29', 'name', 'en', 'Axe'),
    ('item', 'f347b613-70f2-4744-b2e8-1005c6614c29', 'description', 'ru', 'Инструмент для рубки древесины; характеристики зависят от атрибута качества.'),
    ('item', 'f347b613-70f2-4744-b2e8-1005c6614c29', 'description', 'en', 'Tool for chopping wood; characteristics depend on quality attribute.'),
    
    -- Кирка
    ('item', '9cdbee72-713d-4e0e-9719-c9bc8bcb0533', 'name', 'ru', 'Кирка'),
    ('item', '9cdbee72-713d-4e0e-9719-c9bc8bcb0533', 'name', 'en', 'Pickaxe'),
    ('item', '9cdbee72-713d-4e0e-9719-c9bc8bcb0533', 'description', 'ru', 'Инструмент для добычи камня и руды; модификации задаются качеством и серией.'),
    ('item', '9cdbee72-713d-4e0e-9719-c9bc8bcb0533', 'description', 'en', 'Tool for mining stone and ore; modifications are set by quality and series.')
ON CONFLICT (entity_type, entity_id, field_name, language_code) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now();

-- =============================================================================
-- Шаг 4: Загрузка изображений для предметов
-- =============================================================================
-- В этом разделе мы связываем предметы с их изображениями.
-- Для сундуков изображения зависят от качества (small, medium, large).

DO $$
DECLARE
    v_collection_id uuid;
    v_quality_small_id uuid;
    v_quality_medium_id uuid;
    v_quality_large_id uuid;
    v_quality_base_id uuid;
    v_resource_chest_id uuid;
    v_reagent_chest_id uuid;
    v_blueprint_chest_id uuid;
BEGIN
    -- Получаем ID классификаторов, которые будем использовать многократно
    v_collection_id      := (SELECT id FROM inventory.classifier_items WHERE code = 'base' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'collection'));
    v_quality_small_id   := (SELECT id FROM inventory.classifier_items WHERE code = 'small' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'));
    v_quality_medium_id  := (SELECT id FROM inventory.classifier_items WHERE code = 'medium' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'));
    v_quality_large_id   := (SELECT id FROM inventory.classifier_items WHERE code = 'large' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'));
    v_quality_base_id    := (SELECT id FROM inventory.classifier_items WHERE code = 'base' AND classifier_id = (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'));

    -- Получаем ID предметов-сундуков
    v_resource_chest_id  := (SELECT id FROM inventory.items WHERE item_type_id = (SELECT id FROM inventory.classifier_items WHERE code = 'resource_chest'));
    v_reagent_chest_id   := (SELECT id FROM inventory.items WHERE item_type_id = (SELECT id FROM inventory.classifier_items WHERE code = 'reagent_chest'));
    v_blueprint_chest_id := (SELECT id FROM inventory.items WHERE item_type_id = (SELECT id FROM inventory.classifier_items WHERE code = 'blueprint_chest'));

    RAISE NOTICE 'Inserting chest images...';

    -- 1. Изображения для ресурсных сундуков (Resource Chests)
    INSERT INTO inventory.item_images (item_id, collection_id, quality_level_id, image_url) VALUES
        (v_resource_chest_id, v_collection_id, v_quality_small_id, '/statics/images/items/small-chess-res.png'),
        (v_resource_chest_id, v_collection_id, v_quality_medium_id, '/statics/images/items/medium-chess-res.png'),
        (v_resource_chest_id, v_collection_id, v_quality_large_id, '/statics/images/items/big-chess-res.png')
    ON CONFLICT (item_id, collection_id, quality_level_id) DO UPDATE SET image_url = EXCLUDED.image_url;

    -- 2. Изображения для реагентных сундуков (Reagent Chests)
    INSERT INTO inventory.item_images (item_id, collection_id, quality_level_id, image_url) VALUES
        (v_reagent_chest_id, v_collection_id, v_quality_small_id, '/statics/images/items/small-chess-ing.png'),
        (v_reagent_chest_id, v_collection_id, v_quality_medium_id, '/statics/images/items/medium-chess-ing.png'),
        (v_reagent_chest_id, v_collection_id, v_quality_large_id, '/statics/images/items/big-chess-ing.png')
    ON CONFLICT (item_id, collection_id, quality_level_id) DO UPDATE SET image_url = EXCLUDED.image_url;
    
    -- 3. Изображение для сундука с чертежами (Blueprint Chest)
    -- У этого сундука нет уровней качества, поэтому используем 'base'
    INSERT INTO inventory.item_images (item_id, collection_id, quality_level_id, image_url) VALUES
        (v_blueprint_chest_id, v_collection_id, v_quality_base_id, '/statics/images/items/chess-blueprint.png')
    ON CONFLICT (item_id, collection_id, quality_level_id) DO UPDATE SET image_url = EXCLUDED.image_url;

    RAISE NOTICE 'Chest images inserted successfully.';
END $$;


-- =============================================================================
-- Шаг 5: Загрузка переводов для всех предметов
-- =============================================================================

DO $$
DECLARE
    v_items_count INTEGER;
    v_translations_count INTEGER;
    v_expected_items INTEGER := 25; -- Ожидаемое количество предметов (убрали 6 дублирующих сундуков)
    v_expected_translations INTEGER := 100; -- Ожидаемое количество переводов (25 items * 2 fields * 2 languages)
BEGIN
    SELECT COUNT(*) INTO v_items_count FROM inventory.items;
    SELECT COUNT(*) INTO v_translations_count FROM i18n.translations WHERE entity_type = 'item';
    
    RAISE NOTICE 'Загружено предметов: %', v_items_count;
    RAISE NOTICE 'Загружено переводов предметов: %', v_translations_count;
    
    IF v_items_count < v_expected_items THEN
        RAISE WARNING 'Загружено меньше предметов, чем ожидалось (ожидается: %, фактически: %)', 
                      v_expected_items, v_items_count;
    END IF;
    
    IF v_translations_count < v_expected_translations THEN
        RAISE WARNING 'Загружено меньше переводов, чем ожидалось (ожидается: %, фактически: %)', 
                      v_expected_translations, v_translations_count;
    END IF;
    
    -- Проверим, что все категории предметов загружены
    RAISE NOTICE 'Ресурсы: %', (SELECT COUNT(*) FROM inventory.items i 
                                JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                WHERE ci.code = 'resources');
    RAISE NOTICE 'Реагенты: %', (SELECT COUNT(*) FROM inventory.items i 
                                 JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                 WHERE ci.code = 'reagents');
    RAISE NOTICE 'Ускорители: %', (SELECT COUNT(*) FROM inventory.items i 
                                   JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                   WHERE ci.code = 'boosters');
    RAISE NOTICE 'Ключи: %', (SELECT COUNT(*) FROM inventory.items i 
                              JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                              WHERE ci.code = 'keys');
    RAISE NOTICE 'Сундуки: %', (SELECT COUNT(*) FROM inventory.items i 
                                JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                WHERE ci.code = 'chests');
    RAISE NOTICE 'Чертежи: %', (SELECT COUNT(*) FROM inventory.items i 
                                JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                WHERE ci.code = 'blueprints');
    RAISE NOTICE 'Инструменты: %', (SELECT COUNT(*) FROM inventory.items i 
                                    JOIN inventory.classifier_items ci ON i.item_class_id = ci.id 
                                    WHERE ci.code = 'tools');
                                    
    -- Проверим переводы по языкам
    RAISE NOTICE 'Переводы на русском: %', (SELECT COUNT(*) FROM i18n.translations WHERE entity_type = 'item' AND language_code = 'ru');
    RAISE NOTICE 'Переводы на английском: %', (SELECT COUNT(*) FROM i18n.translations WHERE entity_type = 'item' AND language_code = 'en');
END $$;

COMMIT;

-- =============================================================================
-- Скрипт успешно завершен
-- =============================================================================
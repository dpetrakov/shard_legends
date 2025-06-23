-- Migration UP: 003_populate_classifiers.up.sql
-- Description: Дистрибутивные данные классификаторов с фиксированными UUID
-- Service: inventory-service
-- Depends: 002_create_inventory_schema.up.sql
-- Created: 2025-06-23

-- ВНИМАНИЕ: UUID в этом файле фиксированы и не должны изменяться!
-- Это гарантирует стабильность ссылок при многократных накатах/откатах

-- =====================================================
-- Классификаторы (16 основных классификаторов)
-- =====================================================

INSERT INTO inventory.classifiers (id, code, description) VALUES
-- Основные классификаторы
('01234567-8901-4abc-def0-123456789000', 'item_class', 'Классы предметов в игре'),
('01234567-8901-4abc-def0-123456789001', 'quality_level', 'Уровни качества предметов'),
('01234567-8901-4abc-def0-123456789002', 'collection', 'Коллекции предметов (сезонные и базовые)'),
('01234567-8901-4abc-def0-123456789003', 'inventory_section', 'Разделы инвентаря пользователя'),
('01234567-8901-4abc-def0-123456789004', 'operation_type', 'Типы операций с инвентарем'),

-- Типы предметов по классам
('01234567-8901-4abc-def0-123456789005', 'resource_type', 'Подтипы ресурсов для класса resources'),
('01234567-8901-4abc-def0-123456789006', 'reagent_type', 'Подтипы реагентов для класса reagents'),
('01234567-8901-4abc-def0-123456789007', 'booster_type', 'Подтипы ускорителей для класса boosters'),
('01234567-8901-4abc-def0-123456789008', 'tool_type', 'Подтипы инструментов для класса tools'),
('01234567-8901-4abc-def0-123456789009', 'key_type', 'Подтипы ключей для класса keys'),
('01234567-8901-4abc-def0-123456789010', 'currency_type', 'Подтипы валют для класса currencies'),

-- Специализированные классификаторы качества
('01234567-8901-4abc-def0-123456789011', 'tool_quality_levels', 'Доступные уровни качества для инструментов'),
('01234567-8901-4abc-def0-123456789012', 'key_quality_levels', 'Доступные уровни качества для ключей'),

-- Дополнительные классификаторы
('01234567-8901-4abc-def0-123456789013', 'blueprint_type', 'Подтипы чертежей для класса blueprints'),
('01234567-8901-4abc-def0-123456789014', 'standard_collection', 'Стандартная коллекция для базовых предметов'),
('01234567-8901-4abc-def0-123456789015', 'standard_quality', 'Стандартный уровень качества для базовых предметов')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- Элементы классификаторов (~90 элементов)
-- =====================================================

-- Классы предметов (item_class)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('11111111-1111-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789000', 'resources', 'Базовые ресурсы для строительства и производства'),
('11111111-1111-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789000', 'reagents', 'Реагенты для производства инструментов'),
('11111111-1111-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789000', 'boosters', 'Ускорители различных процессов'),
('11111111-1111-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789000', 'blueprints', 'Чертежи для производства инструментов'),
('11111111-1111-4abc-def0-123456789004', '01234567-8901-4abc-def0-123456789000', 'tools', 'Инструменты для добычи ресурсов'),
('11111111-1111-4abc-def0-123456789005', '01234567-8901-4abc-def0-123456789000', 'keys', 'Ключи для открытия сундуков'),
('11111111-1111-4abc-def0-123456789006', '01234567-8901-4abc-def0-123456789000', 'currencies', 'Игровые валюты')
ON CONFLICT (id) DO NOTHING;

-- Уровни качества (quality_level)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('22222222-2222-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789001', 'wooden', 'Деревянный уровень'),
('22222222-2222-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789001', 'stone', 'Каменный уровень'),
('22222222-2222-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789001', 'metal', 'Металлический уровень'),
('22222222-2222-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789001', 'diamond', 'Бриллиантовый уровень'),
('22222222-2222-4abc-def0-123456789004', '01234567-8901-4abc-def0-123456789001', 'small', 'Малый размер (для ключей)'),
('22222222-2222-4abc-def0-123456789005', '01234567-8901-4abc-def0-123456789001', 'medium', 'Средний размер (для ключей)'),
('22222222-2222-4abc-def0-123456789006', '01234567-8901-4abc-def0-123456789001', 'large', 'Большой размер (для ключей)')
ON CONFLICT (id) DO NOTHING;

-- Коллекции (collection)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('33333333-3333-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789002', 'winter_2025', 'Зимняя коллекция 2025'),
('33333333-3333-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789002', 'spring_2025', 'Весенняя коллекция 2025'),
('33333333-3333-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789002', 'summer_2025', 'Летняя коллекция 2025'),
('33333333-3333-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789002', 'autumn_2025', 'Осенняя коллекция 2025')
ON CONFLICT (id) DO NOTHING;

-- Разделы инвентаря (inventory_section)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('44444444-4444-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789003', 'main', 'Основной инвентарь пользователя'),
('44444444-4444-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789003', 'factory', 'Фабричный инвентарь (зарезервированные материалы)'),
('44444444-4444-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789003', 'trade', 'Торговый инвентарь (предметы на продаже)')
ON CONFLICT (id) DO NOTHING;

-- Типы операций (operation_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('55555555-5555-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789004', 'chest_reward', 'Получение из сундука'),
('55555555-5555-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789004', 'craft_result', 'Результат производства'),
('55555555-5555-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789004', 'trade_sale', 'Продажа в торговле'),
('55555555-5555-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789004', 'trade_purchase', 'Покупка в торговле'),
('55555555-5555-4abc-def0-123456789004', '01234567-8901-4abc-def0-123456789004', 'admin_adjustment', 'Административная корректировка'),
('55555555-5555-4abc-def0-123456789005', '01234567-8901-4abc-def0-123456789004', 'system_reward', 'Системная награда'),
('55555555-5555-4abc-def0-123456789006', '01234567-8901-4abc-def0-123456789004', 'system_penalty', 'Системное списание'),
('55555555-5555-4abc-def0-123456789007', '01234567-8901-4abc-def0-123456789004', 'daily_quest_reward', 'Награда за ежедневное задание'),
('55555555-5555-4abc-def0-123456789008', '01234567-8901-4abc-def0-123456789004', 'factory_reservation', 'Резервирование материалов для производства'),
('55555555-5555-4abc-def0-123456789009', '01234567-8901-4abc-def0-123456789004', 'factory_return', 'Возврат материалов при отмене'),
('55555555-5555-4abc-def0-123456789010', '01234567-8901-4abc-def0-123456789004', 'factory_consumption', 'Уничтожение материалов при производстве'),
('55555555-5555-4abc-def0-123456789011', '01234567-8901-4abc-def0-123456789004', 'trade_listing', 'Размещение предмета на продажу'),
('55555555-5555-4abc-def0-123456789012', '01234567-8901-4abc-def0-123456789004', 'trade_delisting', 'Снятие предмета с продажи')
ON CONFLICT (id) DO NOTHING;

-- Типы ресурсов (resource_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('66666666-6666-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789005', 'stone', 'Камень'),
('66666666-6666-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789005', 'wood', 'Дерево'),
('66666666-6666-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789005', 'ore', 'Руда'),
('66666666-6666-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789005', 'diamond', 'Алмаз')
ON CONFLICT (id) DO NOTHING;

-- Типы реагентов (reagent_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('77777777-7777-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789006', 'abrasive', 'Абразив'),
('77777777-7777-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789006', 'disc', 'Диск'),
('77777777-7777-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789006', 'inductor', 'Индуктор'),
('77777777-7777-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789006', 'paste', 'Паста')
ON CONFLICT (id) DO NOTHING;

-- Типы ускорителей (booster_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('88888888-8888-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789007', 'repair_tool', 'Починка инструмента'),
('88888888-8888-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789007', 'speed_processing', 'Ускорение переработки'),
('88888888-8888-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789007', 'speed_crafting', 'Ускорение создания')
ON CONFLICT (id) DO NOTHING;

-- Типы инструментов (tool_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('99999999-9999-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789008', 'shovel', 'Лопата'),
('99999999-9999-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789008', 'sickle', 'Серп'),
('99999999-9999-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789008', 'axe', 'Топор'),
('99999999-9999-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789008', 'pickaxe', 'Кирка')
ON CONFLICT (id) DO NOTHING;

-- Типы ключей (key_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('aaaaaaaa-aaaa-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789009', 'key', 'Обычный ключ'),
('aaaaaaaa-aaaa-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789009', 'blueprint_key', 'Ключ для чертежей')
ON CONFLICT (id) DO NOTHING;

-- Типы валют (currency_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('bbbbbbbb-bbbb-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789010', 'diamonds', 'Диаманты')
ON CONFLICT (id) DO NOTHING;

-- Уровни качества инструментов (tool_quality_levels)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('cccccccc-cccc-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789011', 'wooden', 'Деревянный инструмент'),
('cccccccc-cccc-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789011', 'stone', 'Каменный инструмент'),
('cccccccc-cccc-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789011', 'metal', 'Металлический инструмент'),
('cccccccc-cccc-4abc-def0-123456789003', '01234567-8901-4abc-def0-123456789011', 'diamond', 'Бриллиантовый инструмент')
ON CONFLICT (id) DO NOTHING;

-- Уровни качества ключей (key_quality_levels)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('dddddddd-dddd-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789012', 'small', 'Малый ключ (Key-S)'),
('dddddddd-dddd-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789012', 'medium', 'Средний ключ (Key-M)'),
('dddddddd-dddd-4abc-def0-123456789002', '01234567-8901-4abc-def0-123456789012', 'large', 'Большой ключ (Key-L)')
ON CONFLICT (id) DO NOTHING;

-- Типы чертежей (blueprint_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('eeeeeeee-eeee-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789013', 'tool_blueprint', 'Чертеж инструмента'),
('eeeeeeee-eeee-4abc-def0-123456789001', '01234567-8901-4abc-def0-123456789013', 'booster_blueprint', 'Чертеж ускорителя')
ON CONFLICT (id) DO NOTHING;

-- Стандартная коллекция (standard_collection)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('ffffffff-ffff-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789014', 'standard', 'Стандартная коллекция (базовая)')
ON CONFLICT (id) DO NOTHING;

-- Стандартный уровень качества (standard_quality)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
('00000000-0000-4abc-def0-123456789000', '01234567-8901-4abc-def0-123456789015', 'standard', 'Стандартный уровень качества (базовый)')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- Завершение миграции
-- =====================================================
-- Проверка количества загруженных записей
DO $$
DECLARE
    classifiers_count INTEGER;
    classifier_items_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO classifiers_count FROM inventory.classifiers;
    SELECT COUNT(*) INTO classifier_items_count FROM inventory.classifier_items;
    
    RAISE NOTICE 'Загружено классификаторов: %', classifiers_count;
    RAISE NOTICE 'Загружено элементов классификаторов: %', classifier_items_count;
    
    -- Проверяем, что загружено ожидаемое количество
    IF classifiers_count < 16 THEN
        RAISE EXCEPTION 'Ошибка: загружено классификаторов %, ожидалось минимум 16', classifiers_count;
    END IF;
    
    IF classifier_items_count < 63 THEN
        RAISE EXCEPTION 'Ошибка: загружено элементов классификаторов %, ожидалось минимум 63', classifier_items_count;
    END IF;
END $$;
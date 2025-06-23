-- =====================================================
-- Dev Test Data: 001_test_items_and_operations.sql
-- Description: Тестовые данные для разработки inventory-service
-- Environment: ТОЛЬКО DEV - НЕ ПРИМЕНЯТЬ В PRODUCTION!
-- Author: AI Assistant
-- Created: 2025-06-23
-- =====================================================

-- ВНИМАНИЕ: Этот файл содержит тестовые данные ТОЛЬКО для dev среды!
-- НЕ ПРИМЕНЯЙТЕ в staging или production!

-- =====================================================
-- Тестовые предметы (items)
-- =====================================================

-- Получаем ID классификаторов для создания предметов
INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id) VALUES
-- Ресурсы
('aaaaaaaa-test-4abc-def0-000000000001', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'resources'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'stone'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

('aaaaaaaa-test-4abc-def0-000000000002', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'resources'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'wood'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

('aaaaaaaa-test-4abc-def0-000000000003', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'resources'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'ore'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

-- Инструменты
('aaaaaaaa-test-4abc-def0-000000000011', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'tools'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'axe'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

('aaaaaaaa-test-4abc-def0-000000000012', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'tools'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'pickaxe'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

-- Ключи
('aaaaaaaa-test-4abc-def0-000000000021', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'keys'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'key'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'key_quality_levels'),
 (SELECT id FROM inventory.classifiers WHERE code = 'collection')),

-- Валюты
('aaaaaaaa-test-4abc-def0-000000000031', 
 (SELECT id FROM inventory.classifier_items WHERE code = 'currencies'), 
 (SELECT id FROM inventory.classifier_items WHERE code = 'diamonds'), 
 (SELECT id FROM inventory.classifiers WHERE code = 'standard_quality'),
 (SELECT id FROM inventory.classifiers WHERE code = 'standard_collection'))
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- Тестовые изображения предметов
-- =====================================================

INSERT INTO inventory.item_images (item_id, collection_id, quality_level_id, image_url) VALUES
-- Камень в разных коллекциях и качествах
('aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 'https://cdn.shardlegends.com/items/stone/winter_2025/wooden.png'),

('aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'stone'),
 'https://cdn.shardlegends.com/items/stone/winter_2025/stone.png'),

-- Дерево
('aaaaaaaa-test-4abc-def0-000000000002',
 (SELECT id FROM inventory.classifier_items WHERE code = 'summer_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 'https://cdn.shardlegends.com/items/wood/summer_2025/wooden.png'),

-- Топор
('aaaaaaaa-test-4abc-def0-000000000011',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 'https://cdn.shardlegends.com/items/axe/winter_2025/wooden.png'),

('aaaaaaaa-test-4abc-def0-000000000011',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'metal'),
 'https://cdn.shardlegends.com/items/axe/winter_2025/metal.png'),

-- Ключ
('aaaaaaaa-test-4abc-def0-000000000021',
 (SELECT id FROM inventory.classifier_items WHERE code = 'autumn_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'small'),
 'https://cdn.shardlegends.com/items/key/autumn_2025/small.png'),

-- Диаманты
('aaaaaaaa-test-4abc-def0-000000000031',
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 'https://cdn.shardlegends.com/items/diamonds/standard/standard.png')
ON CONFLICT (item_id, collection_id, quality_level_id) DO NOTHING;

-- =====================================================
-- Тестовые дневные остатки
-- =====================================================
-- Создаем тестовые остатки для пользователя с ID test-user-uuid-here
-- В реальной системе user_id будет браться из auth.users

INSERT INTO inventory.daily_balances 
(user_id, section_id, item_id, collection_id, quality_level_id, balance_date, quantity) VALUES

-- Камень в основном инвентаре
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 CURRENT_DATE - INTERVAL '1 day',
 100),

-- Дерево в основном инвентаре
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000002',
 (SELECT id FROM inventory.classifier_items WHERE code = 'summer_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 CURRENT_DATE - INTERVAL '1 day',
 50),

-- Топор в основном инвентаре
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000011',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 CURRENT_DATE - INTERVAL '1 day',
 1),

-- Диаманты в основном инвентаре
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000031',
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 CURRENT_DATE - INTERVAL '1 day',
 500)
ON CONFLICT (user_id, section_id, item_id, collection_id, quality_level_id, balance_date) DO NOTHING;

-- =====================================================
-- Тестовые операции для демонстрации
-- =====================================================

INSERT INTO inventory.operations 
(user_id, section_id, item_id, collection_id, quality_level_id, quantity_change, operation_type_id, operation_id, comment, created_at) VALUES

-- Получение камня из сундука сегодня
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 25,
 (SELECT id FROM inventory.classifier_items WHERE code = 'chest_reward'),
 'chest-reward-' || extract(epoch from now())::text,
 'Награда из зимнего сундука',
 NOW() - INTERVAL '2 hours'),

-- Трата дерева на производство
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000002',
 (SELECT id FROM inventory.classifier_items WHERE code = 'summer_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 -10,
 (SELECT id FROM inventory.classifier_items WHERE code = 'craft_result'),
 'craft-operation-' || extract(epoch from now())::text,
 'Создание деревянного инструмента',
 NOW() - INTERVAL '1 hour'),

-- Покупка диамантов
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000031',
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'standard'),
 100,
 (SELECT id FROM inventory.classifier_items WHERE code = 'trade_purchase'),
 'trade-purchase-' || extract(epoch from now())::text,
 'Покупка диамантов в магазине',
 NOW() - INTERVAL '30 minutes'),

-- Резервирование камня для производства (основной -> фабричный)
('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'main'),
 'aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 -20,
 (SELECT id FROM inventory.classifier_items WHERE code = 'factory_reservation'),
 'factory-reservation-001',
 'Резервирование камня для производства',
 NOW() - INTERVAL '15 minutes'),

('bbbbbbbb-test-4abc-def0-testuser0001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'factory'),
 'aaaaaaaa-test-4abc-def0-000000000001',
 (SELECT id FROM inventory.classifier_items WHERE code = 'winter_2025'),
 (SELECT id FROM inventory.classifier_items WHERE code = 'wooden'),
 20,
 (SELECT id FROM inventory.classifier_items WHERE code = 'factory_reservation'),
 'factory-reservation-001',
 'Резервирование камня для производства',
 NOW() - INTERVAL '15 minutes')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- Проверка и отчет о тестовых данных
-- =====================================================
DO $$
DECLARE
    items_count INTEGER;
    images_count INTEGER;
    balances_count INTEGER;
    operations_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO items_count FROM inventory.items WHERE id LIKE 'aaaaaaaa-test-%';
    SELECT COUNT(*) INTO images_count FROM inventory.item_images WHERE item_id LIKE 'aaaaaaaa-test-%';
    SELECT COUNT(*) INTO balances_count FROM inventory.daily_balances WHERE user_id LIKE 'bbbbbbbb-test-%';
    SELECT COUNT(*) INTO operations_count FROM inventory.operations WHERE user_id LIKE 'bbbbbbbb-test-%';
    
    RAISE NOTICE '=== ТЕСТОВЫЕ ДАННЫЕ ЗАГРУЖЕНЫ ===';
    RAISE NOTICE 'Тестовых предметов: %', items_count;
    RAISE NOTICE 'Тестовых изображений: %', images_count;
    RAISE NOTICE 'Тестовых остатков: %', balances_count;
    RAISE NOTICE 'Тестовых операций: %', operations_count;
    RAISE NOTICE 'Тестовый пользователь: bbbbbbbb-test-4abc-def0-testuser0001';
    RAISE NOTICE '=== ГОТОВ К ТЕСТИРОВАНИЮ ===';
END $$;
-- Migration: Add missing classifiers and classifier items
-- Purpose: Add tool_quality_levels, key_quality_levels classifiers and missing classifier items

-- Add missing classifiers
INSERT INTO classifiers (id, code, description, created_at, updated_at) VALUES 
    ('33333333-3333-3333-3333-333333333333', 'tool_quality_levels', 'Уровни качества инструментов', NOW(), NOW()),
    ('44444444-4444-4444-4444-444444444444', 'key_quality_levels', 'Уровни качества ключей', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Add missing classifier items for tool_quality_levels
INSERT INTO classifier_items (id, classifier_id, code, description, is_active, created_at, updated_at) VALUES
    ('dd000001-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'wooden', 'Деревянный инструмент', TRUE, NOW(), NOW()),
    ('dd000002-0000-0000-0000-000000000002', '33333333-3333-3333-3333-333333333333', 'stone', 'Каменный инструмент', TRUE, NOW(), NOW()),
    ('dd000003-0000-0000-0000-000000000003', '33333333-3333-3333-3333-333333333333', 'metal', 'Металлический инструмент', TRUE, NOW(), NOW()),
    ('dd000004-0000-0000-0000-000000000004', '33333333-3333-3333-3333-333333333333', 'diamond', 'Бриллиантовый инструмент', TRUE, NOW(), NOW())
ON CONFLICT (classifier_id, code) DO NOTHING;

-- Add missing classifier items for key_quality_levels  
INSERT INTO classifier_items (id, classifier_id, code, description, is_active, created_at, updated_at) VALUES
    ('ee000001-0000-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'small', 'Малый ключ (Key-S)', TRUE, NOW(), NOW()),
    ('ee000002-0000-0000-0000-000000000002', '44444444-4444-4444-4444-444444444444', 'medium', 'Средний ключ (Key-M)', TRUE, NOW(), NOW()),
    ('ee000003-0000-0000-0000-000000000003', '44444444-4444-4444-4444-444444444444', 'large', 'Большой ключ (Key-L)', TRUE, NOW(), NOW())
ON CONFLICT (classifier_id, code) DO NOTHING;

-- Add missing classifier item for item_class (chests)
INSERT INTO classifier_items (id, classifier_id, code, description, is_active, created_at, updated_at) VALUES
    ('ff000001-0000-0000-0000-000000000001', 
     (SELECT id FROM classifiers WHERE code = 'item_class'), 
     'chests', 'Сундуки с различными наградами', TRUE, NOW(), NOW())
ON CONFLICT (classifier_id, code) DO NOTHING;

-- Add missing classifier item for quality_level (base)
INSERT INTO classifier_items (id, classifier_id, code, description, is_active, created_at, updated_at) VALUES
    ('gg000001-0000-0000-0000-000000000001', 
     (SELECT id FROM classifiers WHERE code = 'quality_level'), 
     'base', 'Базовый уровень (для ресурсов, реагентов, ускорителей)', TRUE, NOW(), NOW())
ON CONFLICT (classifier_id, code) DO NOTHING;

-- =============================================================================
-- DEV DATA Script: 002_reset_classifiers.sql
-- Description: Полная перезагрузка всех классификаторов системы inventory
-- Service: inventory-service
-- Created: 2025-06-28
-- 
-- ВНИМАНИЕ: Этот скрипт предназначен только для dev окружения!
-- Он полностью очищает и перезагружает все классификаторы.
-- =============================================================================

BEGIN;

-- =============================================================================
-- Шаг 1: Очистка существующих данных
-- =============================================================================
TRUNCATE TABLE inventory.classifier_items CASCADE;
TRUNCATE TABLE inventory.classifiers CASCADE;

-- =============================================================================
-- Шаг 2: Загрузка классификаторов
-- =============================================================================
INSERT INTO inventory.classifiers (id, code, description) VALUES
    (gen_random_uuid(), 'item_class', 'Классы предметов - высший уровень группировки игровых предметов'),
    (gen_random_uuid(), 'resource_type', 'Типы ресурсов для класса resources'),
    (gen_random_uuid(), 'reagent_type', 'Типы реагентов для класса reagents'),
    (gen_random_uuid(), 'booster_type', 'Типы ускорителей для класса boosters'),
    (gen_random_uuid(), 'tool_type', 'Типы инструментов для класса tools'),
    (gen_random_uuid(), 'key_type', 'Типы ключей для класса keys'),
    (gen_random_uuid(), 'currency_type', 'Типы валют для класса currencies'),
    (gen_random_uuid(), 'quality_level', 'Уровни качества предметов'),
    (gen_random_uuid(), 'collection', 'Коллекции предметов (сезонные и базовые)'),
    (gen_random_uuid(), 'inventory_section', 'Разделы инвентаря пользователя'),
    (gen_random_uuid(), 'operation_type', 'Типы операций с инвентарем'),
    (gen_random_uuid(), 'production_operation_class', 'Классы операций для производственных процессов'),
    (gen_random_uuid(), 'recipe_limit_type', 'Типы временных ограничений для производственных рецептов'),
    (gen_random_uuid(), 'recipe_limit_object', 'Объекты ограничений для производственных рецептов'),
    (gen_random_uuid(), 'production_task_status', 'Возможные статусы производственных заданий'),
    (gen_random_uuid(), 'production_slot_type', 'Типы производственных слотов пользователя'),
    (gen_random_uuid(), 'modifier_type', 'Типы модификаторов для производственных процессов'),
    (gen_random_uuid(), 'modifier_source', 'Источники происхождения модификаторов'),
    (gen_random_uuid(), 'user_status', 'Статусы активности пользователей'),
    (gen_random_uuid(), 'tool_quality_levels', 'Доступные уровни качества для инструментов'),
    (gen_random_uuid(), 'key_quality_levels', 'Доступные уровни качества для ключей')
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- Шаг 3: Загрузка элементов классификаторов
-- =============================================================================

-- CL001: Классы предметов (item_class)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'resources', 'Базовые ресурсы для строительства и производства'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'reagents', 'Реагенты для производства инструментов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'boosters', 'Ускорители различных процессов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'blueprints', 'Чертежи для производства инструментов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'tools', 'Инструменты для добычи ресурсов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'keys', 'Ключи для открытия сундуков'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'currencies', 'Игровые валюты'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'item_class'), 'chests', 'Сундуки с различными наградами')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL002: Типы ресурсов (resource_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'resource_type'), 'stone', 'Камень'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'resource_type'), 'wood', 'Дерево'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'resource_type'), 'ore', 'Руда'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'resource_type'), 'diamond', 'Алмаз')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL003: Типы реагентов (reagent_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type'), 'abrasive', 'Абразив'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type'), 'disc', 'Диск'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type'), 'inductor', 'Индуктор'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'reagent_type'), 'paste', 'Паста')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL004: Типы ускорителей (booster_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'booster_type'), 'repair_tool', 'Починка инструмента'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'booster_type'), 'speed_processing', 'Ускорение переработки'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'booster_type'), 'speed_crafting', 'Ускорение создания')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL005: Типы инструментов (tool_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_type'), 'shovel', 'Лопата'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_type'), 'sickle', 'Серп'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_type'), 'axe', 'Топор'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_type'), 'pickaxe', 'Кирка')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL006: Типы ключей (key_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'key_type'), 'key', 'Обычный ключ'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'key_type'), 'blueprint_key', 'Ключ для чертежей')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL007: Типы валют (currency_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'currency_type'), 'diamonds', 'Диаманты')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL009: Уровни качества (quality_level)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'base', 'Базовый уровень (для ресурсов, реагентов, ускорителей)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'wooden', 'Деревянный уровень (для инструментов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'stone', 'Каменный уровень (для инструментов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'metal', 'Металлический уровень (для инструментов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'diamond', 'Бриллиантовый уровень (для инструментов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'small', 'Малый размер (для сундуков и ключей Key-S)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'medium', 'Средний размер (для сундуков и ключей Key-M)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'quality_level'), 'large', 'Большой размер (для сундуков и ключей Key-L)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL010: Коллекции (collection)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'collection'), 'base', 'Базовая коллекция (для всех типов предметов)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL011: Разделы инвентаря (inventory_section)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'inventory_section'), 'main', 'Основной инвентарь пользователя'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'inventory_section'), 'factory', 'Фабричный инвентарь (зарезервированные материалы)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL012: Типы операций инвентаря (operation_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'chest_reward', 'Получение из сундука'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'craft_result', 'Результат производства'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'admin_adjustment', 'Административная корректировка'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'system_reward', 'Системная награда'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'system_penalty', 'Системное списание'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'daily_quest_reward', 'Награда за ежедневное задание'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'factory_reservation', 'Резервирование материалов для производства'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'factory_return', 'Возврат материалов при отмене'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'operation_type'), 'factory_consumption', 'Уничтожение материалов при производстве')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL013: Классы производственных операций (production_operation_class)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_operation_class'), 'crafting', 'Крафт инструментов, создание предметов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_operation_class'), 'smelting', 'Переплавка, обработка сырья'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_operation_class'), 'chest_opening', 'Открытие сундуков')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL014: Типы лимитов рецептов (recipe_limit_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'total', 'Общий лимит на все время жизни профиля'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_day', 'Лимит на календарные сутки (00:00-23:59 UTC)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_week', 'Лимит на календарную неделю (понедельник-воскресенье)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_month', 'Лимит на календарный месяц'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_quarter', 'Лимит на квартал (3 месяца)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_year', 'Лимит на календарный год'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_season', 'Лимит на сезон (4 недели для чертежей)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_type'), 'per_event', 'Лимит на все время проведения события')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL015: Объекты лимитов рецептов (recipe_limit_object)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_object'), 'recipe_execution', 'Ограничение на количество запусков рецепта'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'recipe_limit_object'), 'item_receipt', 'Ограничение на количество получений конкретного предмета')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL016: Статусы производственных заданий (production_task_status)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'pending', 'Ожидает свободного слота для запуска'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'in_progress', 'Выполняется в производственном слоте'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'completed', 'Завершено, результат готов к Claim'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'claimed', 'Результат получен пользователем'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'cancelled', 'Отменено до начала выполнения'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_task_status'), 'failed', 'Не удалось выполнить по техническим причинам')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL017: Типы производственных слотов (production_slot_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'universal', 'Универсальный слот (поддерживает все операции)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'smithy', 'Кузнечный слот (crafting + smelting)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'crafting', 'Слот для операции крафта (crafting)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'smelting', 'Слот для операции переплавки (smelting)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'chest_opening', 'Слот для операции открытия сундуков (chest_opening)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'mining', 'Шахтерский слот (resource_gathering)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'production_slot_type'), 'special', 'Специальный слот (special)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL018: Типы модификаторов (modifier_type)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_type'), 'speed_bonus', 'Ускорение производства (сокращение времени)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_type'), 'quantity_bonus', 'Увеличение количества (модификация диапазонов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_type'), 'probability_bonus', 'Повышение шансов (улучшение вероятностей)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_type'), 'cost_reduction', 'Снижение затрат (уменьшение материалов)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_type'), 'quality_bonus', 'Повышение качества результата')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL019: Источники модификаторов (modifier_source)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'booster_item', 'Модификатор от предмета-ускорителя'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'vip_status', 'Модификатор от VIP статуса'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'character_level', 'Модификатор от уровня персонажа'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'achievement', 'Модификатор от достижения'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'clan_bonus', 'Модификатор от клановых бонусов'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'seasonal_event', 'Модификатор от сезонного события'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'lunar_cycle', 'Модификатор от лунного цикла'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'server_buff', 'Модификатор от серверного бафа'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'modifier_source'), 'special_event', 'Модификатор от специального события')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL021: Статусы пользователей (user_status)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'active', 'Активный пользователь'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'inactive', 'Неактивный пользователь'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'suspended', 'Заблокированный пользователь'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'banned', 'Забаненный пользователь'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'under_investigation', 'Подозрение в мошенничестве (ограниченные действия до расследования)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'user_status'), 'shadow_restricted', 'Ограниченный статус для подтверждённых мошенников (игра без полноценной выгоды)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL025: Уровни качества инструментов (tool_quality_levels)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), 'wooden', 'Деревянный инструмент'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), 'stone', 'Каменный инструмент'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), 'metal', 'Металлический инструмент'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'tool_quality_levels'), 'diamond', 'Бриллиантовый инструмент')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- CL026: Уровни качества ключей (key_quality_levels)
INSERT INTO inventory.classifier_items (id, classifier_id, code, description) VALUES
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'key_quality_levels'), 'small', 'Малый ключ (Key-S)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'key_quality_levels'), 'medium', 'Средний ключ (Key-M)'),
    (gen_random_uuid(), (SELECT id FROM inventory.classifiers WHERE code = 'key_quality_levels'), 'large', 'Большой ключ (Key-L)')
ON CONFLICT (classifier_id, code) DO NOTHING;

-- =============================================================================
-- Шаг 4: Проверка загруженных данных
-- =============================================================================
DO $$
DECLARE
    v_classifiers_count INTEGER;
    v_items_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_classifiers_count FROM inventory.classifiers;
    SELECT COUNT(*) INTO v_items_count FROM inventory.classifier_items;
    
    RAISE NOTICE 'Загружено классификаторов: %', v_classifiers_count;
    RAISE NOTICE 'Загружено элементов классификаторов: %', v_items_count;
    
    -- Проверка ожидаемого количества
    IF v_classifiers_count < 20 THEN
        RAISE WARNING 'Загружено меньше классификаторов, чем ожидалось (ожидается минимум 20)';
    END IF;
    
    IF v_items_count < 80 THEN
        RAISE WARNING 'Загружено меньше элементов, чем ожидалось (ожидается минимум 80)';
    END IF;
END $$;

COMMIT;

-- =============================================================================
-- Скрипт успешно завершен
-- =============================================================================
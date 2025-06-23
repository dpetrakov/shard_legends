-- Migration DOWN: 003_populate_classifiers.down.sql
-- Description: Откат дистрибутивных данных классификаторов
-- Service: inventory-service
-- Created: 2025-06-23

-- ВНИМАНИЕ: Этот скрипт удаляет все данные классификаторов!
-- Используйте только для отката миграции в dev/test средах.

-- =====================================================
-- Проверка и отчет перед удалением
-- =====================================================
DO $$
DECLARE
    classifiers_count INTEGER;
    classifier_items_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO classifiers_count FROM inventory.classifiers;
    SELECT COUNT(*) INTO classifier_items_count FROM inventory.classifier_items;
    
    RAISE NOTICE 'Будет удалено классификаторов: %', classifiers_count;
    RAISE NOTICE 'Будет удалено элементов классификаторов: %', classifier_items_count;
END $$;

-- =====================================================
-- Удаление данных в правильном порядке (учет FK)
-- =====================================================

-- Сначала удаляем элементы классификаторов (они ссылаются на классификаторы)
DELETE FROM inventory.classifier_items;

-- Затем удаляем сами классификаторы
DELETE FROM inventory.classifiers;

-- =====================================================
-- Сброс sequences (если используются)
-- =====================================================
-- PostgreSQL с UUID gen_random_uuid() не использует sequences,
-- но на всякий случай оставляем этот блок для будущих изменений

-- =====================================================
-- Подтверждение отката
-- =====================================================
DO $$
DECLARE
    classifiers_count INTEGER;
    classifier_items_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO classifiers_count FROM inventory.classifiers;
    SELECT COUNT(*) INTO classifier_items_count FROM inventory.classifier_items;
    
    RAISE NOTICE 'После отката осталось классификаторов: %', classifiers_count;
    RAISE NOTICE 'После отката осталось элементов классификаторов: %', classifier_items_count;
    
    -- Проверяем, что все удалено
    IF classifiers_count > 0 THEN
        RAISE WARNING 'Внимание: не все классификаторы были удалены!';
    END IF;
    
    IF classifier_items_count > 0 THEN
        RAISE WARNING 'Внимание: не все элементы классификаторов были удалены!';
    END IF;
    
    IF classifiers_count = 0 AND classifier_items_count = 0 THEN
        RAISE NOTICE 'Откат данных классификаторов завершен успешно';
    END IF;
END $$;
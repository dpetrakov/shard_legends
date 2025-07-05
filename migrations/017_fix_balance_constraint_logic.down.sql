-- Migration DOWN: 017_fix_balance_constraint_logic.down.sql
-- Description: Откат исправления логики проверки баланса в constraint
-- Service: inventory-service
-- Created: 2025-07-05

-- =====================================================
-- Восстановление оригинальной функции проверки баланса
-- =====================================================

-- Возвращаем оригинальную функцию из миграции 006
CREATE OR REPLACE FUNCTION inventory.check_sufficient_balance()
RETURNS TRIGGER AS $$
DECLARE
    current_balance BIGINT := 0;
    main_section_id UUID;
    daily_balance BIGINT := 0;
    operations_sum BIGINT := 0;
BEGIN
    -- Получаем ID секции main
    SELECT ci.id INTO main_section_id
    FROM inventory.classifier_items ci
    JOIN inventory.classifiers c ON ci.classifier_id = c.id
    WHERE c.code = 'inventory_section' AND ci.code = 'main';
    
    -- Проверяем только для main секции и отрицательных изменений (расходы)
    IF NEW.section_id = main_section_id AND NEW.quantity_change < 0 THEN
        
        -- Получаем последний дневной остаток
        SELECT COALESCE(quantity, 0) INTO daily_balance
        FROM inventory.daily_balances
        WHERE user_id = NEW.user_id
          AND section_id = NEW.section_id
          AND item_id = NEW.item_id
          AND collection_id = NEW.collection_id
          AND quality_level_id = NEW.quality_level_id
          AND balance_date <= CURRENT_DATE
        ORDER BY balance_date DESC
        LIMIT 1;
        
        -- Суммируем операции за текущий день (оригинальная логика)
        SELECT COALESCE(SUM(quantity_change), 0) INTO operations_sum
        FROM inventory.operations
        WHERE user_id = NEW.user_id
          AND section_id = NEW.section_id
          AND item_id = NEW.item_id
          AND collection_id = NEW.collection_id
          AND quality_level_id = NEW.quality_level_id
          AND DATE(created_at) = CURRENT_DATE;
        
        -- Рассчитываем текущий баланс
        current_balance := daily_balance + operations_sum;
        
        -- Проверяем достаточность баланса
        IF current_balance + NEW.quantity_change < 0 THEN
            RAISE EXCEPTION 'Insufficient balance for user % item %: requested %, available %', 
                NEW.user_id, NEW.item_id, ABS(NEW.quantity_change), current_balance
                USING ERRCODE = '23514'; -- check_violation
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Восстановление оригинального комментария
-- =====================================================

COMMENT ON FUNCTION inventory.check_sufficient_balance() IS 'Функция-триггер для проверки достаточности баланса при операциях расхода в main секции. Предотвращает oversell.';

-- =====================================================
-- Завершение отката миграции
-- =====================================================

-- Откат успешно завершен
-- Восстановлена оригинальная логика из миграции 006 
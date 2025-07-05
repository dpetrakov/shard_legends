-- Migration UP: 017_fix_balance_constraint_logic.up.sql
-- Description: Исправление логики проверки баланса в constraint - использовать операции с даты последнего daily_balance
-- Service: inventory-service
-- Depends: 006_add_balance_constraints.up.sql
-- Created: 2025-07-05
-- Issue: Database constraint использует неправильную логику подсчета операций

-- =====================================================
-- Обновление функции для правильной проверки баланса
-- =====================================================

-- Заменяем функцию для проверки достаточности баланса перед операциями
CREATE OR REPLACE FUNCTION inventory.check_sufficient_balance()
RETURNS TRIGGER AS $$
DECLARE
    current_balance BIGINT := 0;
    main_section_id UUID;
    daily_balance BIGINT := 0;
    operations_sum BIGINT := 0;
    last_balance_date DATE;
BEGIN
    -- Получаем ID секции main
    SELECT ci.id INTO main_section_id
    FROM inventory.classifier_items ci
    JOIN inventory.classifiers c ON ci.classifier_id = c.id
    WHERE c.code = 'inventory_section' AND ci.code = 'main';
    
    -- Проверяем только для main секции и отрицательных изменений (расходы)
    IF NEW.section_id = main_section_id AND NEW.quantity_change < 0 THEN
        
        -- Получаем последний дневной остаток и дату
        SELECT COALESCE(quantity, 0), COALESCE(balance_date, '1970-01-01'::date) 
        INTO daily_balance, last_balance_date
        FROM inventory.daily_balances
        WHERE user_id = NEW.user_id
          AND section_id = NEW.section_id
          AND item_id = NEW.item_id
          AND collection_id = NEW.collection_id
          AND quality_level_id = NEW.quality_level_id
          AND balance_date <= CURRENT_DATE
        ORDER BY balance_date DESC
        LIMIT 1;
        
        -- Если нет daily_balance, устанавливаем значения по умолчанию
        IF last_balance_date IS NULL THEN
            daily_balance := 0;
            last_balance_date := '1970-01-01'::date;
        END IF;
        
        -- Суммируем операции С ДАТЫ ПОСЛЕДНЕГО DAILY_BALANCE (исправленная логика)
        -- Это соответствует логике в GetUserInventoryOptimized и CheckAndLockBalances
        SELECT COALESCE(SUM(quantity_change), 0) INTO operations_sum
        FROM inventory.operations
        WHERE user_id = NEW.user_id
          AND section_id = NEW.section_id
          AND item_id = NEW.item_id
          AND collection_id = NEW.collection_id
          AND quality_level_id = NEW.quality_level_id
          AND created_at > (last_balance_date + INTERVAL '1 day');
        
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
-- Комментарии к обновленной функции
-- =====================================================

COMMENT ON FUNCTION inventory.check_sufficient_balance() IS 'Функция-триггер для проверки достаточности баланса при операциях расхода в main секции. Использует операции с даты последнего daily_balance (соответствует логике CheckAndLockBalances).';

-- =====================================================
-- Завершение миграции
-- =====================================================

-- Миграция успешно завершена
-- Исправлено:
-- 1. Логика подсчета операций в database constraint теперь соответствует логике в CheckAndLockBalances
-- 2. Операции суммируются с даты последнего daily_balance, а не только за текущий день
-- 3. Устранено расхождение между application-level и database-level проверками баланса 
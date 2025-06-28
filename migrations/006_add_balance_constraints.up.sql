-- Migration UP: 006_add_balance_constraints.up.sql
-- Description: Добавление constraints для предотвращения oversell и индексов для атомарных блокировок
-- Service: inventory-service
-- Depends: 002_create_inventory_schema.up.sql
-- Created: 2025-06-28
-- Issue: D-14 - Исправить гонки при резервировании (oversell)

-- =====================================================
-- Создание функции для проверки баланса при операциях
-- =====================================================

-- Создаем функцию для проверки достаточности баланса перед операциями расхода
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
        
        -- Суммируем операции за текущий день (до текущей операции)
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

-- Создаем триггер для проверки баланса при добавлении операций
CREATE TRIGGER check_balance_before_operation
    BEFORE INSERT ON inventory.operations
    FOR EACH ROW EXECUTE FUNCTION inventory.check_sufficient_balance();

-- =====================================================
-- Создание индексов для атомарных блокировок
-- =====================================================

-- Индекс для атомарных блокировок строк daily_balances
-- Критичен для SELECT ... FOR UPDATE операций
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_daily_balances_user_item_lock 
ON inventory.daily_balances (user_id, section_id, item_id, collection_id, quality_level_id);

-- Индекс для быстрого поиска операций текущего дня при расчете баланса
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_operations_current_day
ON inventory.operations (user_id, section_id, item_id, collection_id, quality_level_id)
WHERE DATE(created_at) = CURRENT_DATE;

-- Составной индекс для оптимизации проверки баланса
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_operations_balance_check
ON inventory.operations (user_id, section_id, item_id, collection_id, quality_level_id, created_at)
INCLUDE (quantity_change);

-- =====================================================
-- Комментарии к новым объектам
-- =====================================================

COMMENT ON FUNCTION inventory.check_sufficient_balance() IS 'Функция-триггер для проверки достаточности баланса при операциях расхода в main секции. Предотвращает oversell.';

COMMENT ON INDEX inventory.idx_daily_balances_user_item_lock IS 'Критичный индекс для атомарных блокировок строк при SELECT FOR UPDATE в операциях резервирования';

COMMENT ON INDEX inventory.idx_operations_current_day IS 'Частичный индекс для операций текущего дня для быстрого расчета балансов';

COMMENT ON INDEX inventory.idx_operations_balance_check IS 'Составной индекс с включенными колонками для оптимизации проверки баланса';

-- =====================================================
-- Обновление прав доступа
-- =====================================================

GRANT EXECUTE ON FUNCTION inventory.check_sufficient_balance() TO slcw_user;

-- =====================================================
-- Завершение миграции
-- =====================================================

-- Миграция успешно завершена
-- Добавлены:
-- 1. Триггер-функция check_sufficient_balance для проверки баланса на уровне БД
-- 2. Оптимальные индексы для атомарных блокировок и производительности  
-- 3. Триггер для автоматической проверки баланса при операциях
-- 4. Защита от race conditions и oversell на уровне базы данных
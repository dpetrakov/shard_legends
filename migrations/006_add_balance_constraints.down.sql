-- Migration DOWN: 006_add_balance_constraints.down.sql
-- Description: Откат добавления constraints и индексов для предотвращения oversell
-- Service: inventory-service
-- Created: 2025-06-28
-- Issue: D-14 - Откат исправления гонок при резервировании

-- =====================================================
-- Удаление триггера и функции проверки баланса
-- =====================================================

-- Удаляем триггер
DROP TRIGGER IF EXISTS check_balance_before_operation ON inventory.operations;

-- Удаляем функцию проверки баланса
DROP FUNCTION IF EXISTS inventory.check_sufficient_balance();

-- =====================================================
-- Удаление индексов
-- =====================================================

-- Удаляем индексы для атомарных блокировок
DROP INDEX CONCURRENTLY IF EXISTS inventory.idx_daily_balances_user_item_lock;
DROP INDEX CONCURRENTLY IF EXISTS inventory.idx_operations_current_day;
DROP INDEX CONCURRENTLY IF EXISTS inventory.idx_operations_balance_check;

-- =====================================================
-- Откат завершен
-- =====================================================

-- Откат миграции завершен
-- Удалены:
-- 1. Триггер check_balance_before_operation
-- 2. Функция check_sufficient_balance
-- 3. Индексы для атомарных блокировок
-- 4. Защита от race conditions (система вернулась к исходному состоянию с уязвимостью)
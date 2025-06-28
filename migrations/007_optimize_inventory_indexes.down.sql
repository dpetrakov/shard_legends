-- Migration DOWN: 007_optimize_inventory_indexes.down.sql
-- Description: Откат оптимизации индексов для inventory-service
-- Service: inventory-service
-- Related: 007_optimize_inventory_indexes.up.sql
-- Created: 2025-06-28

-- =====================================================
-- Удаление функций и view
-- =====================================================

DROP FUNCTION IF EXISTS inventory.analyze_inventory_performance(uuid, uuid);
DROP VIEW IF EXISTS inventory.index_usage_monitoring;

-- =====================================================
-- Удаление оптимизированных индексов
-- =====================================================

-- Индексы для daily_balances
DROP INDEX IF EXISTS inventory.idx_daily_balances_user_section_positive;
DROP INDEX IF EXISTS inventory.idx_daily_balances_full_lookup;
DROP INDEX IF EXISTS inventory.idx_daily_balances_latest_lookup;
DROP INDEX IF EXISTS inventory.idx_daily_balances_locking;

-- Индексы для operations
DROP INDEX IF EXISTS inventory.idx_operations_balance_calculation;
DROP INDEX IF EXISTS inventory.idx_operations_today_sum;
DROP INDEX IF EXISTS inventory.idx_operations_today_items;

-- Индексы для items и classifier_items
DROP INDEX IF EXISTS inventory.idx_items_details_lookup;
DROP INDEX IF EXISTS inventory.idx_classifier_items_lookup;
DROP INDEX IF EXISTS inventory.idx_classifier_items_classifier_id_lookup;

-- =====================================================
-- Восстановление оригинальных индексов
-- =====================================================

-- Восстанавливаем оригинальный индекс daily_balances_user_date
CREATE INDEX IF NOT EXISTS idx_daily_balances_user_date 
ON inventory.daily_balances (user_id, balance_date);

-- =====================================================
-- Завершение отката
-- =====================================================
-- Откат миграции 007 завершен
-- Индексы оптимизации производительности удалены
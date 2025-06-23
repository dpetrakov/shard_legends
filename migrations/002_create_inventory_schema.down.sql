-- Migration DOWN: 002_create_inventory_schema.down.sql
-- Description: Откат создания схемы inventory и всех таблиц
-- Service: inventory-service
-- Created: 2025-06-23

-- ВНИМАНИЕ: Этот скрипт полностью удаляет схему inventory и все данные!
-- Используйте только для отката миграции в dev/test средах.

-- =====================================================
-- Удаление таблиц в правильном порядке (учет FK)
-- =====================================================

-- Удаляем таблицы с foreign keys сначала
DROP TABLE IF EXISTS inventory.operations CASCADE;
DROP TABLE IF EXISTS inventory.daily_balances CASCADE;
DROP TABLE IF EXISTS inventory.item_images CASCADE;
DROP TABLE IF EXISTS inventory.items CASCADE;
DROP TABLE IF EXISTS inventory.classifier_items CASCADE;
DROP TABLE IF EXISTS inventory.classifiers CASCADE;

-- =====================================================
-- Удаление функций и триггеров
-- =====================================================
DROP FUNCTION IF EXISTS inventory.update_updated_at_column() CASCADE;

-- =====================================================
-- Удаление схемы inventory полностью
-- =====================================================
DROP SCHEMA IF EXISTS inventory CASCADE;

-- =====================================================
-- Откат завершен
-- =====================================================
package storage

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

type itemStorage struct {
	pool    *pgxpool.Pool
	logger  *slog.Logger
	metrics *metrics.Metrics
}

func (s *itemStorage) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error) {
	s.logger.Info("GetItemByID called", "item_id", itemID)
	
	query := `
		SELECT 
			id, 
			item_class_id, 
			item_type_id, 
			quality_levels_classifier_id, 
			collections_classifier_id, 
			created_at, 
			updated_at
		FROM inventory.items 
		WHERE id = $1
	`
	
	var item models.Item
	err := s.pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID,
		&item.ItemClassID,
		&item.ItemTypeID,
		&item.QualityLevelsClassifierID,
		&item.CollectionsClassifierID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("Item not found", "item_id", itemID)
			return nil, nil
		}
		s.logger.Error("Failed to get item by ID", "error", err)
		return nil, err
	}
	
	return &item, nil
}

func (s *itemStorage) GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error) {
	s.logger.Info("GetItemsByClass called", "class_code", classCode)
	
	query := `
		SELECT 
			i.id, 
			i.item_class_id, 
			i.item_type_id, 
			i.quality_levels_classifier_id, 
			i.collections_classifier_id, 
			i.created_at, 
			i.updated_at
		FROM inventory.items i
		JOIN inventory.classifier_items ci ON i.item_class_id = ci.id
		WHERE ci.code = $1
		ORDER BY i.created_at ASC
	`
	
	rows, err := s.pool.Query(ctx, query, classCode)
	if err != nil {
		s.logger.Error("Failed to get items by class", "error", err)
		return nil, err
	}
	defer rows.Close()
	
	var items []*models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ID,
			&item.ItemClassID,
			&item.ItemTypeID,
			&item.QualityLevelsClassifierID,
			&item.CollectionsClassifierID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			s.logger.Error("Failed to scan item", "error", err)
			return nil, err
		}
		items = append(items, &item)
	}
	
	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}
	
	s.logger.Info("Found items by class", "class_code", classCode, "count", len(items))
	return items, nil
}

func (s *itemStorage) GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error) {
	s.logger.Info("GetItemWithDetails called", "item_id", itemID)
	
	query := `
		SELECT 
			i.id, 
			i.item_class_id, 
			i.item_type_id, 
			i.quality_levels_classifier_id, 
			i.collections_classifier_id, 
			i.created_at, 
			i.updated_at,
			ci_class.code as item_class,
			ci_type.code as item_type
		FROM inventory.items i
		JOIN inventory.classifier_items ci_class ON i.item_class_id = ci_class.id
		JOIN inventory.classifier_items ci_type ON i.item_type_id = ci_type.id
		WHERE i.id = $1
	`
	
	var itemWithDetails models.ItemWithDetails
	err := s.pool.QueryRow(ctx, query, itemID).Scan(
		&itemWithDetails.ID,
		&itemWithDetails.ItemClassID,
		&itemWithDetails.ItemTypeID,
		&itemWithDetails.QualityLevelsClassifierID,
		&itemWithDetails.CollectionsClassifierID,
		&itemWithDetails.CreatedAt,
		&itemWithDetails.UpdatedAt,
		&itemWithDetails.ItemClass,
		&itemWithDetails.ItemType,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("Item with details not found", "item_id", itemID)
			return nil, nil
		}
		s.logger.Error("Failed to get item with details", "error", err)
		return nil, err
	}
	
	return &itemWithDetails, nil
}
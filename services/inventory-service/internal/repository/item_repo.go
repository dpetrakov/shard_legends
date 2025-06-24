package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/shard-legends/inventory-service/internal/models"
)

// itemRepo implements ItemRepository
type itemRepo struct {
	db *sqlx.DB
}

// NewItemRepository creates a new ItemRepository
func NewItemRepository(db *sqlx.DB) ItemRepository {
	return &itemRepo{
		db: db,
	}
}

// GetItemByID retrieves an item by its ID
func (r *itemRepo) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.Item, error) {
	query := `
		SELECT id, item_class_id, item_type_id, quality_levels_classifier_id, 
		       collections_classifier_id, created_at, updated_at
		FROM item
		WHERE id = $1
	`
	
	var item models.Item
	err := r.db.GetContext(ctx, &item, query, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get item by ID")
	}
	
	return &item, nil
}

// GetItemsByClass retrieves all items for a specific class
func (r *itemRepo) GetItemsByClass(ctx context.Context, classCode string) ([]*models.Item, error) {
	query := `
		SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id,
		       i.collections_classifier_id, i.created_at, i.updated_at
		FROM item i
		JOIN classifier_item ci ON ci.id = i.item_class_id
		WHERE ci.code = $1 AND ci.is_active = true
		ORDER BY i.created_at
	`
	
	var items []*models.Item
	err := r.db.SelectContext(ctx, &items, query, classCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get items by class")
	}
	
	return items, nil
}

// GetItemsByClassAndType retrieves items by class and type codes
func (r *itemRepo) GetItemsByClassAndType(ctx context.Context, classCode, typeCode string) ([]*models.Item, error) {
	query := `
		SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id,
		       i.collections_classifier_id, i.created_at, i.updated_at
		FROM item i
		JOIN classifier_item ci_class ON ci_class.id = i.item_class_id
		JOIN classifier_item ci_type ON ci_type.id = i.item_type_id
		WHERE ci_class.code = $1 AND ci_type.code = $2
		  AND ci_class.is_active = true AND ci_type.is_active = true
		ORDER BY i.created_at
	`
	
	var items []*models.Item
	err := r.db.SelectContext(ctx, &items, query, classCode, typeCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get items by class and type")
	}
	
	return items, nil
}

// GetItemImage retrieves an image for a specific item variant
func (r *itemRepo) GetItemImage(ctx context.Context, itemID, collectionID, qualityLevelID uuid.UUID) (*models.ItemImage, error) {
	query := `
		SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at
		FROM item_image
		WHERE item_id = $1 AND collection_id = $2 AND quality_level_id = $3 AND is_active = true
	`
	
	var image models.ItemImage
	err := r.db.GetContext(ctx, &image, query, itemID, collectionID, qualityLevelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get item image")
	}
	
	return &image, nil
}

// GetItemImages retrieves all active images for an item
func (r *itemRepo) GetItemImages(ctx context.Context, itemID uuid.UUID) ([]*models.ItemImage, error) {
	query := `
		SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at
		FROM item_image
		WHERE item_id = $1 AND is_active = true
		ORDER BY collection_id, quality_level_id
	`
	
	var images []*models.ItemImage
	err := r.db.SelectContext(ctx, &images, query, itemID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item images")
	}
	
	return images, nil
}

// GetItemWithDetails retrieves an item with classifier details loaded
func (r *itemRepo) GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*models.ItemWithDetails, error) {
	query := `
		SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id,
		       i.collections_classifier_id, i.created_at, i.updated_at,
		       ci_class.code as item_class, ci_type.code as item_type
		FROM item i
		JOIN classifier_item ci_class ON ci_class.id = i.item_class_id
		JOIN classifier_item ci_type ON ci_type.id = i.item_type_id
		WHERE i.id = $1
	`
	
	var item models.ItemWithDetails
	err := r.db.GetContext(ctx, &item, query, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get item with details")
	}
	
	return &item, nil
}
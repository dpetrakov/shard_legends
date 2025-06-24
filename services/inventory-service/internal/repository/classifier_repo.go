package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/shard-legends/inventory-service/internal/models"
)

const (
	// Cache TTL for classifiers (24 hours)
	classifierCacheTTL = 24 * time.Hour
	
	// Cache key prefixes
	cacheKeyClassifier      = "inventory:classifier:%s"
	cacheKeyClassifierItems = "inventory:classifier_items:%s"
	cacheKeyCodeToUUID      = "inventory:code_to_uuid:%s"
	cacheKeyUUIDToCode      = "inventory:uuid_to_code:%s"
	cacheKeyAllClassifiers  = "inventory:all_classifiers"
)

// classifierRepo implements ClassifierRepository
type classifierRepo struct {
	db    *sqlx.DB
	cache Cache
}

// NewClassifierRepository creates a new ClassifierRepository
func NewClassifierRepository(db *sqlx.DB, cache Cache) ClassifierRepository {
	return &classifierRepo{
		db:    db,
		cache: cache,
	}
}

// GetClassifierByCode retrieves a classifier by its code
func (r *classifierRepo) GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error) {
	cacheKey := fmt.Sprintf(cacheKeyClassifier, code)
	
	// Try to get from cache
	var classifier models.Classifier
	if err := r.cache.Get(ctx, cacheKey, &classifier); err == nil {
		return &classifier, nil
	}
	
	// Query from database
	query := `
		SELECT id, code, description, created_at, updated_at
		FROM classifier
		WHERE code = $1
	`
	
	err := r.db.GetContext(ctx, &classifier, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get classifier by code")
	}
	
	// Cache the result
	_ = r.cache.Set(ctx, cacheKey, &classifier, classifierCacheTTL)
	
	return &classifier, nil
}

// GetClassifierItems retrieves all items for a classifier
func (r *classifierRepo) GetClassifierItems(ctx context.Context, classifierID uuid.UUID) ([]*models.ClassifierItem, error) {
	cacheKey := fmt.Sprintf(cacheKeyClassifierItems, classifierID.String())
	
	// Try to get from cache
	var cachedData []byte
	if err := r.cache.Get(ctx, cacheKey, &cachedData); err == nil {
		var items []*models.ClassifierItem
		if err := json.Unmarshal(cachedData, &items); err == nil {
			return items, nil
		}
	}
	
	// Query from database
	query := `
		SELECT id, classifier_id, code, description, is_active, created_at, updated_at
		FROM classifier_item
		WHERE classifier_id = $1 AND is_active = true
		ORDER BY code
	`
	
	var items []*models.ClassifierItem
	err := r.db.SelectContext(ctx, &items, query, classifierID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get classifier items")
	}
	
	// Cache the result
	if data, err := json.Marshal(items); err == nil {
		_ = r.cache.Set(ctx, cacheKey, data, classifierCacheTTL)
	}
	
	return items, nil
}

// GetClassifierItemByCode retrieves a specific classifier item by code
func (r *classifierRepo) GetClassifierItemByCode(ctx context.Context, classifierID uuid.UUID, code string) (*models.ClassifierItem, error) {
	// Get all items from cache or DB
	items, err := r.GetClassifierItems(ctx, classifierID)
	if err != nil {
		return nil, err
	}
	
	// Find the specific item
	for _, item := range items {
		if item.Code == code {
			return item, nil
		}
	}
	
	return nil, ErrNotFound
}

// GetAllClassifiersWithItems retrieves all classifiers with their items
func (r *classifierRepo) GetAllClassifiersWithItems(ctx context.Context) (map[string][]*models.ClassifierItem, error) {
	// Try to get from cache
	var cachedData []byte
	if err := r.cache.Get(ctx, cacheKeyAllClassifiers, &cachedData); err == nil {
		result := make(map[string][]*models.ClassifierItem)
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return result, nil
		}
	}
	
	// Query classifiers
	classifierQuery := `
		SELECT id, code, description, created_at, updated_at
		FROM classifier
		ORDER BY code
	`
	
	var classifiers []models.Classifier
	err := r.db.SelectContext(ctx, &classifiers, classifierQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get classifiers")
	}
	
	// Query all classifier items
	itemsQuery := `
		SELECT ci.id, ci.classifier_id, ci.code, ci.description, ci.is_active, ci.created_at, ci.updated_at,
		       c.code as classifier_code
		FROM classifier_item ci
		JOIN classifier c ON c.id = ci.classifier_id
		WHERE ci.is_active = true
		ORDER BY c.code, ci.code
	`
	
	rows, err := r.db.QueryxContext(ctx, itemsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query classifier items")
	}
	defer rows.Close()
	
	result := make(map[string][]*models.ClassifierItem)
	for rows.Next() {
		var item models.ClassifierItem
		var classifierCode string
		
		err := rows.Scan(
			&item.ID, &item.ClassifierID, &item.Code, &item.Description,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt, &classifierCode,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan classifier item")
		}
		
		result[classifierCode] = append(result[classifierCode], &item)
	}
	
	// Cache the result
	if data, err := json.Marshal(result); err == nil {
		_ = r.cache.Set(ctx, cacheKeyAllClassifiers, data, classifierCacheTTL)
	}
	
	return result, nil
}

// GetCodeToUUIDMapping returns a mapping of codes to UUIDs for a classifier
func (r *classifierRepo) GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	cacheKey := fmt.Sprintf(cacheKeyCodeToUUID, classifierCode)
	
	// Try to get from cache
	var cachedData []byte
	if err := r.cache.Get(ctx, cacheKey, &cachedData); err == nil {
		result := make(map[string]uuid.UUID)
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return result, nil
		}
	}
	
	// Get classifier
	classifier, err := r.GetClassifierByCode(ctx, classifierCode)
	if err != nil {
		return nil, err
	}
	
	// Get items
	items, err := r.GetClassifierItems(ctx, classifier.ID)
	if err != nil {
		return nil, err
	}
	
	// Build mapping
	result := make(map[string]uuid.UUID)
	for _, item := range items {
		result[item.Code] = item.ID
	}
	
	// Cache the result
	if data, err := json.Marshal(result); err == nil {
		_ = r.cache.Set(ctx, cacheKey, data, classifierCacheTTL)
	}
	
	return result, nil
}

// GetUUIDToCodeMapping returns a mapping of UUIDs to codes for a classifier
func (r *classifierRepo) GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	cacheKey := fmt.Sprintf(cacheKeyUUIDToCode, classifierCode)
	
	// Try to get from cache
	var cachedData []byte
	if err := r.cache.Get(ctx, cacheKey, &cachedData); err == nil {
		result := make(map[uuid.UUID]string)
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return result, nil
		}
	}
	
	// Get classifier
	classifier, err := r.GetClassifierByCode(ctx, classifierCode)
	if err != nil {
		return nil, err
	}
	
	// Get items
	items, err := r.GetClassifierItems(ctx, classifier.ID)
	if err != nil {
		return nil, err
	}
	
	// Build mapping
	result := make(map[uuid.UUID]string)
	for _, item := range items {
		result[item.ID] = item.Code
	}
	
	// Cache the result
	if data, err := json.Marshal(result); err == nil {
		_ = r.cache.Set(ctx, cacheKey, data, classifierCacheTTL)
	}
	
	return result, nil
}

// InvalidateCache invalidates the cache for a specific classifier
func (r *classifierRepo) InvalidateCache(ctx context.Context, classifierCode string) error {
	// Delete specific classifier cache
	if err := r.cache.Delete(ctx, fmt.Sprintf(cacheKeyClassifier, classifierCode)); err != nil {
		return errors.Wrap(err, "failed to delete classifier cache")
	}
	
	// Delete mappings cache
	if err := r.cache.Delete(ctx, fmt.Sprintf(cacheKeyCodeToUUID, classifierCode)); err != nil {
		return errors.Wrap(err, "failed to delete code to UUID cache")
	}
	
	if err := r.cache.Delete(ctx, fmt.Sprintf(cacheKeyUUIDToCode, classifierCode)); err != nil {
		return errors.Wrap(err, "failed to delete UUID to code cache")
	}
	
	// Delete all classifiers cache
	if err := r.cache.Delete(ctx, cacheKeyAllClassifiers); err != nil {
		return errors.Wrap(err, "failed to delete all classifiers cache")
	}
	
	// Delete classifier items cache by pattern
	pattern := fmt.Sprintf("inventory:classifier_items:*")
	if err := r.cache.DeletePattern(ctx, pattern); err != nil {
		return errors.Wrap(err, "failed to delete classifier items cache")
	}
	
	return nil
}
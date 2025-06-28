package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/internal/service"
)

type itemStorage struct {
	pool    *pgxpool.Pool
	logger  *slog.Logger
	metrics service.MetricsInterface
	cache   service.CacheInterface
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

// GetTranslationsBatch gets translations for multiple entities
func (s *itemStorage) GetTranslationsBatch(ctx context.Context, entityType string, entityIDs []uuid.UUID, languageCode string) (map[uuid.UUID]map[string]string, error) {
	s.logger.Info("GetTranslationsBatch called", "entity_type", entityType, "entity_count", len(entityIDs), "language", languageCode)

	if len(entityIDs) == 0 {
		return make(map[uuid.UUID]map[string]string), nil
	}

	result := make(map[uuid.UUID]map[string]string)
	uncachedIDs := make([]uuid.UUID, 0)

	// Try to get translations from cache first
	for _, entityID := range entityIDs {
		cacheKey := fmt.Sprintf("i18n:translations:%s:%s:%s", entityType, entityID.String(), languageCode)
		var cachedTranslations map[string]string

		if err := s.cache.Get(ctx, cacheKey, &cachedTranslations); err == nil && cachedTranslations != nil {
			result[entityID] = cachedTranslations
			s.metrics.RecordCacheHit("translations")
		} else {
			uncachedIDs = append(uncachedIDs, entityID)
			s.metrics.RecordCacheMiss("translations")
		}
	}

	// Query uncached translations from database
	if len(uncachedIDs) > 0 {
		query := `
			SELECT entity_id, field_name, content
			FROM i18n.translations
			WHERE entity_type = $1 
			  AND entity_id = ANY($2) 
			  AND language_code = $3
		`

		rows, err := s.pool.Query(ctx, query, entityType, uncachedIDs, languageCode)
		if err != nil {
			s.logger.Error("Failed to get translations batch", "error", err)
			return nil, err
		}
		defer rows.Close()

		dbResults := make(map[uuid.UUID]map[string]string)

		for rows.Next() {
			var entityID uuid.UUID
			var fieldName, content string

			if err := rows.Scan(&entityID, &fieldName, &content); err != nil {
				s.logger.Error("Failed to scan translation", "error", err)
				return nil, err
			}

			if dbResults[entityID] == nil {
				dbResults[entityID] = make(map[string]string)
			}
			dbResults[entityID][fieldName] = content
		}

		if err := rows.Err(); err != nil {
			s.logger.Error("Rows iteration error", "error", err)
			return nil, err
		}

		// Cache the results and add to final result
		for _, entityID := range uncachedIDs {
			translations := dbResults[entityID]
			if translations == nil {
				translations = make(map[string]string) // Cache empty result to prevent repeated queries
			}

			result[entityID] = translations

			// Cache with 24 hour TTL
			cacheKey := fmt.Sprintf("i18n:translations:%s:%s:%s", entityType, entityID.String(), languageCode)
			if err := s.cache.Set(ctx, cacheKey, translations, 24*time.Hour); err != nil {
				s.logger.Warn("Failed to cache translations", "error", err, "entity_id", entityID)
			}
		}
	}

	s.logger.Info("Found translations", "entity_type", entityType, "language", languageCode, "total_entities", len(result), "cached", len(entityIDs)-len(uncachedIDs), "from_db", len(uncachedIDs))
	return result, nil
}

// GetDefaultLanguage gets the default language for fallback
func (s *itemStorage) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	s.logger.Debug("GetDefaultLanguage called")

	query := `
		SELECT code, name, is_default, is_active, created_at, updated_at
		FROM i18n.languages
		WHERE is_default = true AND is_active = true
		LIMIT 1
	`

	var lang models.Language
	err := s.pool.QueryRow(ctx, query).Scan(
		&lang.Code,
		&lang.Name,
		&lang.IsDefault,
		&lang.IsActive,
		&lang.CreatedAt,
		&lang.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Warn("No default language found")
			return nil, nil
		}
		s.logger.Error("Failed to get default language", "error", err)
		return nil, err
	}

	return &lang, nil
}

// GetItemImagesBatch gets image URLs for multiple item variants
func (s *itemStorage) GetItemImagesBatch(ctx context.Context, requests []models.ItemDetailRequestItem) (map[string]string, error) {
	s.logger.Info("GetItemImagesBatch called", "item_count", len(requests))

	if len(requests) == 0 {
		return make(map[string]string), nil
	}

	// Build maps to resolve codes to UUIDs for collections and quality levels
	collectionCodeToID := make(map[string]uuid.UUID)
	qualityCodeToID := make(map[string]uuid.UUID)

	// Collect unique codes for batch resolution
	collectionCodes := make(map[string]bool)
	qualityCodes := make(map[string]bool)

	for _, req := range requests {
		if req.Collection != nil {
			collectionCodes[*req.Collection] = true
		}
		if req.QualityLevel != nil {
			qualityCodes[*req.QualityLevel] = true
		}
	}

	// Resolve collection codes to UUIDs if any
	if len(collectionCodes) > 0 {
		var codes []string
		for code := range collectionCodes {
			codes = append(codes, code)
		}

		query := `
			SELECT ci.code, ci.id
			FROM inventory.classifier_items ci
			JOIN inventory.classifiers c ON ci.classifier_id = c.id
			WHERE c.code = 'collection' AND ci.code = ANY($1)
		`

		rows, err := s.pool.Query(ctx, query, codes)
		if err != nil {
			s.logger.Error("Failed to resolve collection codes", "error", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var code string
			var id uuid.UUID
			if err := rows.Scan(&code, &id); err != nil {
				s.logger.Error("Failed to scan collection code", "error", err)
				return nil, err
			}
			collectionCodeToID[code] = id
		}

		if err := rows.Err(); err != nil {
			s.logger.Error("Collection codes query error", "error", err)
			return nil, err
		}
	}

	// Resolve quality level codes to UUIDs if any
	if len(qualityCodes) > 0 {
		var codes []string
		for code := range qualityCodes {
			codes = append(codes, code)
		}

		query := `
			SELECT ci.code, ci.id
			FROM inventory.classifier_items ci
			JOIN inventory.classifiers c ON ci.classifier_id = c.id
			WHERE c.code = 'quality_level' AND ci.code = ANY($1)
		`

		rows, err := s.pool.Query(ctx, query, codes)
		if err != nil {
			s.logger.Error("Failed to resolve quality level codes", "error", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var code string
			var id uuid.UUID
			if err := rows.Scan(&code, &id); err != nil {
				s.logger.Error("Failed to scan quality level code", "error", err)
				return nil, err
			}
			qualityCodeToID[code] = id
		}

		if err := rows.Err(); err != nil {
			s.logger.Error("Quality level codes query error", "error", err)
			return nil, err
		}
	}

	// Build the main query for item images
	type imageKey struct {
		itemID         uuid.UUID
		collectionID   uuid.UUID
		qualityLevelID uuid.UUID
	}

	var imageKeys []imageKey
	requestKeyMap := make(map[imageKey]string) // maps imageKey to request identifier

	for _, req := range requests {
		// Create identifier for this request
		reqID := req.ItemID.String()
		if req.Collection != nil {
			reqID += "_" + *req.Collection
		}
		if req.QualityLevel != nil {
			reqID += "_" + *req.QualityLevel
		}

		// Get collection and quality level UUIDs (use zero UUID if not specified)
		var collectionID, qualityLevelID uuid.UUID

		if req.Collection != nil {
			if id, exists := collectionCodeToID[*req.Collection]; exists {
				collectionID = id
			} else {
				s.logger.Warn("Collection code not found", "code", *req.Collection)
				continue // Skip this request if collection not found
			}
		}

		if req.QualityLevel != nil {
			if id, exists := qualityCodeToID[*req.QualityLevel]; exists {
				qualityLevelID = id
			} else {
				s.logger.Warn("Quality level code not found", "code", *req.QualityLevel)
				continue // Skip this request if quality level not found
			}
		}

		key := imageKey{
			itemID:         req.ItemID,
			collectionID:   collectionID,
			qualityLevelID: qualityLevelID,
		}
		imageKeys = append(imageKeys, key)
		requestKeyMap[key] = reqID
	}

	if len(imageKeys) == 0 {
		s.logger.Warn("No valid image keys to query")
		return make(map[string]string), nil
	}

	// Query item images with cache
	result := make(map[string]string)
	uncachedKeys := make([]imageKey, 0)

	// Try to get images from cache first
	for _, key := range imageKeys {
		cacheKey := fmt.Sprintf("i18n:item_images:%s:%s:%s", key.itemID.String(), key.collectionID.String(), key.qualityLevelID.String())
		var cachedImageURL string

		if err := s.cache.Get(ctx, cacheKey, &cachedImageURL); err == nil && cachedImageURL != "" {
			if reqID, exists := requestKeyMap[key]; exists {
				result[reqID] = cachedImageURL
				s.metrics.RecordCacheHit("item_images")
			}
		} else {
			uncachedKeys = append(uncachedKeys, key)
			s.metrics.RecordCacheMiss("item_images")
		}
	}

	// Query uncached images from database
	for _, key := range uncachedKeys {
		query := `
			SELECT image_url
			FROM inventory.item_images
			WHERE item_id = $1 AND collection_id = $2 AND quality_level_id = $3
		`

		var imageURL string
		err := s.pool.QueryRow(ctx, query, key.itemID, key.collectionID, key.qualityLevelID).Scan(&imageURL)

		if err != nil {
			if err == pgx.ErrNoRows {
				s.logger.Debug("No image found for item variant",
					"item_id", key.itemID,
					"collection_id", key.collectionID,
					"quality_level_id", key.qualityLevelID)
				// Cache empty result to prevent repeated queries
				cacheKey := fmt.Sprintf("i18n:item_images:%s:%s:%s", key.itemID.String(), key.collectionID.String(), key.qualityLevelID.String())
				if err := s.cache.Set(ctx, cacheKey, "", 24*time.Hour); err != nil {
					s.logger.Warn("Failed to cache empty image result", "error", err)
				}
				continue
			}
			s.logger.Error("Failed to get item image", "error", err)
			return nil, err
		}

		// Cache the found image URL
		cacheKey := fmt.Sprintf("i18n:item_images:%s:%s:%s", key.itemID.String(), key.collectionID.String(), key.qualityLevelID.String())
		if err := s.cache.Set(ctx, cacheKey, imageURL, 24*time.Hour); err != nil {
			s.logger.Warn("Failed to cache image URL", "error", err)
		}

		if reqID, exists := requestKeyMap[key]; exists {
			result[reqID] = imageURL
		}
	}

	s.logger.Info("Retrieved item images", "requested", len(requests), "found", len(result), "cached", len(imageKeys)-len(uncachedKeys), "from_db", len(uncachedKeys))
	return result, nil
}

// GetItemsBatch gets multiple items by their IDs
func (s *itemStorage) GetItemsBatch(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID]*models.ItemWithDetails, error) {
	s.logger.Info("GetItemsBatch called", "item_count", len(itemIDs))

	if len(itemIDs) == 0 {
		return make(map[uuid.UUID]*models.ItemWithDetails), nil
	}

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
		WHERE i.id = ANY($1)
	`

	rows, err := s.pool.Query(ctx, query, itemIDs)
	if err != nil {
		s.logger.Error("Failed to get items batch", "error", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.ItemWithDetails)

	for rows.Next() {
		var item models.ItemWithDetails
		if err := rows.Scan(
			&item.ID,
			&item.ItemClassID,
			&item.ItemTypeID,
			&item.QualityLevelsClassifierID,
			&item.CollectionsClassifierID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ItemClass,
			&item.ItemType,
		); err != nil {
			s.logger.Error("Failed to scan item", "error", err)
			return nil, err
		}
		result[item.ID] = &item
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}

	s.logger.Info("Found items", "requested", len(itemIDs), "found", len(result))
	return result, nil
}

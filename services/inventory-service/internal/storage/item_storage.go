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
	// Add classifier repo to use its cached methods
	classifierRepo service.ClassifierRepositoryInterface
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

// GetItemImagesBatch gets image URLs for multiple item variants with a fallback strategy.
// 1. Exact match (item, collection, quality)
// 2. Base collection match (item, 'base' collection, quality)
// 3. Base fallback (item, 'base' collection, 'base' quality)
func (s *itemStorage) GetItemImagesBatch(ctx context.Context, requests []models.ItemDetailRequestItem) (map[string]string, error) {
	s.logger.Info("GetItemImagesBatch called with fallback logic", "item_count", len(requests))
	if len(requests) == 0 {
		return make(map[string]string), nil
	}

	// Get item details to determine which quality classifier to use for each item
	itemIDs := make([]uuid.UUID, len(requests))
	for i, req := range requests {
		itemIDs[i] = req.ItemID
	}

	itemsMap, err := s.GetItemsBatch(ctx, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get items batch: %w", err)
	}

	// Get all classifier mappings once using the cached repository method.
	collectionCodeToID, err := s.classifierRepo.GetCodeToUUIDMapping(ctx, "collection")
	if err != nil {
		return nil, fmt.Errorf("failed to get collection mapping: %w", err)
	}

	// Get quality level mappings for all required classifiers
	qualityClassifierCodes := make(map[string]bool)
	qualityClassifierCodes["quality_level"] = true // Default classifier

	// Determine which quality classifiers are needed
	for _, req := range requests {
		if item, exists := itemsMap[req.ItemID]; exists {
			// Get the quality classifier code from the item
			qualityClassifierID := item.QualityLevelsClassifierID
			// Query the classifier code directly from the database
			var classifierCode string
			query := `SELECT code FROM inventory.classifiers WHERE id = $1`
			err := s.pool.QueryRow(ctx, query, qualityClassifierID).Scan(&classifierCode)
			if err == nil {
				qualityClassifierCodes[classifierCode] = true
			}
		}
	}

	// Load all required quality level mappings
	qualityMappings := make(map[string]map[string]uuid.UUID)
	for classifierCode := range qualityClassifierCodes {
		mapping, err := s.classifierRepo.GetCodeToUUIDMapping(ctx, classifierCode)
		if err != nil {
			s.logger.Warn("Failed to get quality level mapping", "classifier", classifierCode, "error", err)
			continue
		}
		qualityMappings[classifierCode] = mapping
	}

	// Ensure default quality_level mapping exists
	if _, exists := qualityMappings["quality_level"]; !exists {
		defaultMapping, err := s.classifierRepo.GetCodeToUUIDMapping(ctx, "quality_level")
		if err != nil {
			return nil, fmt.Errorf("failed to get default quality level mapping: %w", err)
		}
		qualityMappings["quality_level"] = defaultMapping
	}

	// Final map for results, keyed by the composite key the service layer expects.
	finalImageMap := make(map[string]string)
	// Create a copy of requests to keep track of items that still need an image.
	pendingRequests := make([]models.ItemDetailRequestItem, len(requests))
	copy(pendingRequests, requests)

	// --- Fallback Attempt 1: Exact Match ---
	s.logger.Debug("Image fallback: Attempt 1 (Exact Match)", "pending_count", len(pendingRequests))
	pendingRequests, err = s.findImagesForRequests(ctx, pendingRequests, finalImageMap, collectionCodeToID, qualityMappings, itemsMap, false, false)
	if err != nil {
		return nil, fmt.Errorf("error in fallback attempt 1 (exact match): %w", err)
	}

	// --- Fallback Attempt 2: Base Collection, Original Quality ---
	if len(pendingRequests) > 0 {
		s.logger.Debug("Image fallback: Attempt 2 (Base Collection)", "pending_count", len(pendingRequests))
		pendingRequests, err = s.findImagesForRequests(ctx, pendingRequests, finalImageMap, collectionCodeToID, qualityMappings, itemsMap, true, false)
		if err != nil {
			return nil, fmt.Errorf("error in fallback attempt 2 (base collection): %w", err)
		}
	}

	// --- Fallback Attempt 3: Base Collection & Base Quality ---
	if len(pendingRequests) > 0 {
		s.logger.Debug("Image fallback: Attempt 3 (Base Collection & Quality)", "pending_count", len(pendingRequests))
		// This is the last attempt, so we don't need to update the pending list anymore.
		_, err = s.findImagesForRequests(ctx, pendingRequests, finalImageMap, collectionCodeToID, qualityMappings, itemsMap, true, true)
		if err != nil {
			return nil, fmt.Errorf("error in fallback attempt 3 (base fallback): %w", err)
		}
	}

	s.logger.Info("Finished image search", "total_requested", len(requests), "found", len(finalImageMap))
	return finalImageMap, nil
}

// findImagesForRequests is a helper function that attempts to find images for a slice of requests using a specific strategy.
// It returns a slice of requests for which an image was still not found.
func (s *itemStorage) findImagesForRequests(
	ctx context.Context,
	requests []models.ItemDetailRequestItem,
	finalImageMap map[string]string,
	collectionCodeToID map[string]uuid.UUID,
	qualityMappings map[string]map[string]uuid.UUID,
	itemsMap map[uuid.UUID]*models.ItemWithDetails,
	useBaseCollection, useBaseQuality bool,
) ([]models.ItemDetailRequestItem, error) {

	if len(requests) == 0 {
		return nil, nil
	}

	baseCollectionID, baseCollOK := collectionCodeToID["base"]
	if !baseCollOK {
		return requests, fmt.Errorf("base collection not found in classifiers")
	}

	// Get default quality level ID from the default quality_level classifier
	defaultQualityMapping, hasDefault := qualityMappings["quality_level"]
	if !hasDefault {
		return requests, fmt.Errorf("default quality_level classifier not found")
	}
	baseQualityLevelID, baseQualOK := defaultQualityMapping["base"]
	if !baseQualOK {
		return requests, fmt.Errorf("base quality level not found in default classifier")
	}

	batch := &pgx.Batch{}
	for _, req := range requests {
		query := `SELECT image_url FROM inventory.item_images WHERE item_id = $1 AND collection_id = $2 AND quality_level_id = $3 LIMIT 1`

		var collectionID, qualityLevelID uuid.UUID

		// Determine Collection ID for the query
		if useBaseCollection {
			collectionID = baseCollectionID
		} else if req.Collection != nil {
			id, ok := collectionCodeToID[*req.Collection]
			if ok {
				collectionID = id
			} else {
				collectionID = baseCollectionID // Fallback to base if code is unknown
			}
		} else {
			collectionID = baseCollectionID // Fallback to base if not provided
		}

		// Determine Quality Level ID for the query
		if useBaseQuality {
			qualityLevelID = baseQualityLevelID
		} else if req.QualityLevel != nil {
			// Find the correct quality classifier for this item
			item, itemExists := itemsMap[req.ItemID]
			if itemExists {
				// Get the quality classifier code from the item
				qualityClassifierID := item.QualityLevelsClassifierID
				var classifierCode string
				classifierQuery := `SELECT code FROM inventory.classifiers WHERE id = $1`
				err := s.pool.QueryRow(ctx, classifierQuery, qualityClassifierID).Scan(&classifierCode)
				if err == nil {
					if qualityMapping, exists := qualityMappings[classifierCode]; exists {
						if id, ok := qualityMapping[*req.QualityLevel]; ok {
							qualityLevelID = id
						} else {
							qualityLevelID = baseQualityLevelID // Fallback to base if code is unknown
						}
					} else {
						qualityLevelID = baseQualityLevelID // Fallback to base if classifier not found
					}
				} else {
					qualityLevelID = baseQualityLevelID // Fallback to base if classifier lookup fails
				}
			} else {
				qualityLevelID = baseQualityLevelID // Fallback to base if item not found
			}
		} else {
			qualityLevelID = baseQualityLevelID // Fallback to base if not provided
		}

		batch.Queue(query, req.ItemID, collectionID, qualityLevelID)
	}

	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	var stillPending []models.ItemDetailRequestItem
	for _, req := range requests {
		var imageURL string
		err := results.QueryRow().Scan(&imageURL)

		if err == nil && imageURL != "" {
			// Found an image. Build the key that the service layer expects and add it to the final map.
			key := s.buildImageMapKey(req)
			finalImageMap[key] = imageURL
		} else {
			// Image not found with this strategy, keep it pending for the next attempt.
			stillPending = append(stillPending, req)
			if err != pgx.ErrNoRows {
				s.logger.Warn("Failed to scan image_url during batch", "item_id", req.ItemID, "error", err)
			}
		}
	}

	return stillPending, nil
}

// buildImageMapKey generates the composite key used by the service layer to look up an image URL.
// This logic must be identical to the key generation in the service layer.
func (s *itemStorage) buildImageMapKey(req models.ItemDetailRequestItem) string {
	key := req.ItemID.String()
	var collection, qualityLevel string

	if req.Collection != nil && *req.Collection != "" {
		collection = *req.Collection
	} else {
		collection = "base"
	}

	if req.QualityLevel != nil && *req.QualityLevel != "" {
		qualityLevel = *req.QualityLevel
	} else {
		qualityLevel = "base"
	}

	return key + "_" + collection + "_" + qualityLevel
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

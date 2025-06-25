package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/pkg/metrics"
)

type classifierStorage struct {
	pool    *pgxpool.Pool
	redis   *redis.Client
	logger  *slog.Logger
	metrics *metrics.Metrics
}

func (s *classifierStorage) GetClassifierByCode(ctx context.Context, code string) (*models.Classifier, error) {
	s.logger.Info("GetClassifierByCode called", "code", code)
	
	query := `
		SELECT 
			id, 
			code, 
			description, 
			created_at, 
			updated_at
		FROM inventory.classifiers 
		WHERE code = $1
	`
	
	var classifier models.Classifier
	err := s.pool.QueryRow(ctx, query, code).Scan(
		&classifier.ID,
		&classifier.Code,
		&classifier.Description,
		&classifier.CreatedAt,
		&classifier.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("Classifier not found", "code", code)
			return nil, nil
		}
		s.logger.Error("Failed to get classifier by code", "error", err)
		return nil, err
	}
	
	return &classifier, nil
}

func (s *classifierStorage) GetCodeToUUIDMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	s.logger.Info("GetCodeToUUIDMapping called", "classifier_code", classifierCode)
	
	// Try to get from cache first
	cacheKey := "classifier:code_to_uuid:" + classifierCode
	cachedData, err := s.redis.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(cachedData) > 0 {
		s.logger.Debug("Found mapping in cache", "classifier_code", classifierCode)
		mapping := make(map[string]uuid.UUID)
		for code, uuidStr := range cachedData {
			if parsedUUID, err := uuid.Parse(uuidStr); err == nil {
				mapping[code] = parsedUUID
			}
		}
		return mapping, nil
	}
	
	// If not in cache, query database
	query := `
		SELECT ci.code, ci.id
		FROM inventory.classifier_items ci
		JOIN inventory.classifiers c ON ci.classifier_id = c.id
		WHERE c.code = $1
	`
	
	rows, err := s.pool.Query(ctx, query, classifierCode)
	if err != nil {
		s.logger.Error("Failed to get code to UUID mapping", "error", err)
		return nil, err
	}
	defer rows.Close()
	
	mapping := make(map[string]uuid.UUID)
	cacheData := make(map[string]interface{})
	
	for rows.Next() {
		var code string
		var id uuid.UUID
		if err := rows.Scan(&code, &id); err != nil {
			s.logger.Error("Failed to scan classifier item", "error", err)
			return nil, err
		}
		mapping[code] = id
		cacheData[code] = id.String()
	}
	
	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}
	
	// Cache the result for 1 hour
	if len(cacheData) > 0 {
		pipe := s.redis.Pipeline()
		pipe.HMSet(ctx, cacheKey, cacheData)
		pipe.Expire(ctx, cacheKey, time.Hour)
		_, err = pipe.Exec(ctx)
		if err != nil {
			s.logger.Warn("Failed to cache mapping", "error", err)
		}
	}
	
	s.logger.Info("Found classifier mapping", "classifier_code", classifierCode, "count", len(mapping))
	return mapping, nil
}

func (s *classifierStorage) GetUUIDToCodeMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	s.logger.Info("GetUUIDToCodeMapping called", "classifier_code", classifierCode)
	
	// Try to get from cache first
	cacheKey := "classifier:uuid_to_code:" + classifierCode
	cachedData, err := s.redis.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(cachedData) > 0 {
		s.logger.Debug("Found reverse mapping in cache", "classifier_code", classifierCode)
		mapping := make(map[uuid.UUID]string)
		for uuidStr, code := range cachedData {
			if parsedUUID, err := uuid.Parse(uuidStr); err == nil {
				mapping[parsedUUID] = code
			}
		}
		return mapping, nil
	}
	
	// If not in cache, query database
	query := `
		SELECT ci.id, ci.code
		FROM inventory.classifier_items ci
		JOIN inventory.classifiers c ON ci.classifier_id = c.id
		WHERE c.code = $1
	`
	
	rows, err := s.pool.Query(ctx, query, classifierCode)
	if err != nil {
		s.logger.Error("Failed to get UUID to code mapping", "error", err)
		return nil, err
	}
	defer rows.Close()
	
	mapping := make(map[uuid.UUID]string)
	cacheData := make(map[string]interface{})
	
	for rows.Next() {
		var id uuid.UUID
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			s.logger.Error("Failed to scan classifier item", "error", err)
			return nil, err
		}
		mapping[id] = code
		cacheData[id.String()] = code
	}
	
	if err := rows.Err(); err != nil {
		s.logger.Error("Rows iteration error", "error", err)
		return nil, err
	}
	
	// Cache the result for 1 hour
	if len(cacheData) > 0 {
		pipe := s.redis.Pipeline()
		pipe.HMSet(ctx, cacheKey, cacheData)
		pipe.Expire(ctx, cacheKey, time.Hour)
		_, err = pipe.Exec(ctx)
		if err != nil {
			s.logger.Warn("Failed to cache reverse mapping", "error", err)
		}
	}
	
	s.logger.Info("Found reverse classifier mapping", "classifier_code", classifierCode, "count", len(mapping))
	return mapping, nil
}

func (s *classifierStorage) InvalidateCache(ctx context.Context, classifierCode string) error {
	s.logger.Info("InvalidateCache called", "classifier_code", classifierCode)
	
	// Delete both mapping caches
	cacheKeys := []string{
		"classifier:code_to_uuid:" + classifierCode,
		"classifier:uuid_to_code:" + classifierCode,
	}
	
	pipe := s.redis.Pipeline()
	for _, key := range cacheKeys {
		pipe.Del(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("Failed to invalidate cache", "error", err)
		return err
	}
	
	s.logger.Info("Cache invalidated successfully", "classifier_code", classifierCode)
	return nil
}
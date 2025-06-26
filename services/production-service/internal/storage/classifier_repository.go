package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// classifierRepository реализует ClassifierRepository
type classifierRepository struct {
	db      DatabaseInterface
	cache   CacheInterface
	metrics MetricsInterface
}

// NewClassifierRepository создает новый экземпляр репозитория классификаторов
func NewClassifierRepository(deps *RepositoryDependencies) ClassifierRepository {
	return &classifierRepository{
		db:      deps.DB,
		cache:   deps.Cache,
		metrics: deps.MetricsCollector,
	}
}

// GetClassifierMapping возвращает маппинг код -> UUID для классификатора
func (r *classifierRepository) GetClassifierMapping(ctx context.Context, classifierName string) (map[string]uuid.UUID, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_classifier_mapping", time.Since(start))
	}()

	// Проверяем кеш
	cacheKey := fmt.Sprintf("classifier_mapping:%s", classifierName)
	if cachedMapping, err := r.getCachedMapping(ctx, cacheKey); err == nil && cachedMapping != nil {
		r.metrics.IncCacheHit("classifier_mapping")
		return cachedMapping, nil
	}
	r.metrics.IncCacheMiss("classifier_mapping")

	r.metrics.IncDBQuery("get_classifier_mapping")

	// Запрашиваем из базы данных
	query := `
		SELECT ci.code, ci.id
		FROM inventory.classifiers c
		JOIN inventory.classifier_items ci ON c.id = ci.classifier_id
		WHERE c.name = $1 AND ci.is_active = true
		ORDER BY ci.code
	`

	rows, err := r.db.Query(ctx, query, classifierName)
	if err != nil {
		return nil, fmt.Errorf("failed to query classifier mapping for %s: %w", classifierName, err)
	}
	defer rows.Close()

	mapping := make(map[string]uuid.UUID)
	for rows.Next() {
		var code string
		var id uuid.UUID
		if err := rows.Scan(&code, &id); err != nil {
			return nil, fmt.Errorf("failed to scan classifier item: %w", err)
		}
		mapping[code] = id
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Кешируем результат на 1 час
	if err := r.setCachedMapping(ctx, cacheKey, mapping, time.Hour); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		// TODO: добавить логирование
	}

	return mapping, nil
}

// GetReverseClassifierMapping возвращает маппинг UUID -> код для классификатора
func (r *classifierRepository) GetReverseClassifierMapping(ctx context.Context, classifierName string) (map[uuid.UUID]string, error) {
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("get_reverse_classifier_mapping", time.Since(start))
	}()

	// Проверяем кеш
	cacheKey := fmt.Sprintf("reverse_classifier_mapping:%s", classifierName)
	if cachedMapping, err := r.getCachedReverseMapping(ctx, cacheKey); err == nil && cachedMapping != nil {
		r.metrics.IncCacheHit("reverse_classifier_mapping")
		return cachedMapping, nil
	}
	r.metrics.IncCacheMiss("reverse_classifier_mapping")

	r.metrics.IncDBQuery("get_reverse_classifier_mapping")

	// Запрашиваем из базы данных
	query := `
		SELECT ci.id, ci.code
		FROM inventory.classifiers c
		JOIN inventory.classifier_items ci ON c.id = ci.classifier_id
		WHERE c.name = $1 AND ci.is_active = true
		ORDER BY ci.code
	`

	rows, err := r.db.Query(ctx, query, classifierName)
	if err != nil {
		return nil, fmt.Errorf("failed to query reverse classifier mapping for %s: %w", classifierName, err)
	}
	defer rows.Close()

	mapping := make(map[uuid.UUID]string)
	for rows.Next() {
		var id uuid.UUID
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, fmt.Errorf("failed to scan classifier item: %w", err)
		}
		mapping[id] = code
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Кешируем результат на 1 час
	if err := r.setCachedReverseMapping(ctx, cacheKey, mapping, time.Hour); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		// TODO: добавить логирование
	}

	return mapping, nil
}

// ConvertCodeToUUID преобразует код в UUID через классификатор
func (r *classifierRepository) ConvertCodeToUUID(ctx context.Context, classifierName, code string) (*uuid.UUID, error) {
	mapping, err := r.GetClassifierMapping(ctx, classifierName)
	if err != nil {
		return nil, fmt.Errorf("failed to get classifier mapping: %w", err)
	}

	if id, exists := mapping[code]; exists {
		return &id, nil
	}

	return nil, fmt.Errorf("code '%s' not found in classifier '%s'", code, classifierName)
}

// ConvertUUIDToCode преобразует UUID в код через классификатор
func (r *classifierRepository) ConvertUUIDToCode(ctx context.Context, classifierName string, id uuid.UUID) (*string, error) {
	mapping, err := r.GetReverseClassifierMapping(ctx, classifierName)
	if err != nil {
		return nil, fmt.Errorf("failed to get reverse classifier mapping: %w", err)
	}

	if code, exists := mapping[id]; exists {
		return &code, nil
	}

	return nil, fmt.Errorf("UUID '%s' not found in classifier '%s'", id, classifierName)
}

// getCachedMapping получает кешированный маппинг код -> UUID
func (r *classifierRepository) getCachedMapping(ctx context.Context, cacheKey string) (map[string]uuid.UUID, error) {
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	// Простая сериализация: "code1:uuid1,code2:uuid2,..."
	mapping := make(map[string]uuid.UUID)
	if data == "" {
		return nil, fmt.Errorf("empty cache data")
	}

	pairs := splitString(data, ",")
	for _, pair := range pairs {
		parts := splitString(pair, ":")
		if len(parts) != 2 {
			continue
		}
		
		code := parts[0]
		uuidStr := parts[1]
		
		if id, err := uuid.Parse(uuidStr); err == nil {
			mapping[code] = id
		}
	}

	return mapping, nil
}

// setCachedMapping сохраняет маппинг код -> UUID в кеш
func (r *classifierRepository) setCachedMapping(ctx context.Context, cacheKey string, mapping map[string]uuid.UUID, ttl time.Duration) error {
	if len(mapping) == 0 {
		return nil
	}

	// Простая сериализация: "code1:uuid1,code2:uuid2,..."
	var pairs []string
	for code, id := range mapping {
		pairs = append(pairs, fmt.Sprintf("%s:%s", code, id.String()))
	}
	
	data := joinStrings(pairs, ",")
	return r.cache.Set(ctx, cacheKey, data, ttl)
}

// getCachedReverseMapping получает кешированный маппинг UUID -> код
func (r *classifierRepository) getCachedReverseMapping(ctx context.Context, cacheKey string) (map[uuid.UUID]string, error) {
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	// Простая сериализация: "uuid1:code1,uuid2:code2,..."
	mapping := make(map[uuid.UUID]string)
	if data == "" {
		return nil, fmt.Errorf("empty cache data")
	}

	pairs := splitString(data, ",")
	for _, pair := range pairs {
		parts := splitString(pair, ":")
		if len(parts) != 2 {
			continue
		}
		
		uuidStr := parts[0]
		code := parts[1]
		
		if id, err := uuid.Parse(uuidStr); err == nil {
			mapping[id] = code
		}
	}

	return mapping, nil
}

// setCachedReverseMapping сохраняет маппинг UUID -> код в кеш
func (r *classifierRepository) setCachedReverseMapping(ctx context.Context, cacheKey string, mapping map[uuid.UUID]string, ttl time.Duration) error {
	if len(mapping) == 0 {
		return nil
	}

	// Простая сериализация: "uuid1:code1,uuid2:code2,..."
	var pairs []string
	for id, code := range mapping {
		pairs = append(pairs, fmt.Sprintf("%s:%s", id.String(), code))
	}
	
	data := joinStrings(pairs, ",")
	return r.cache.Set(ctx, cacheKey, data, ttl)
}

// Вспомогательные функции для работы со строками (аналогично strings.Split и strings.Join)
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	start := 0
	for i := 0; i < len(s); {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start
		} else {
			i++
		}
	}
	result = append(result, s[start:])
	return result
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	totalLen := len(sep) * (len(strs) - 1)
	for _, s := range strs {
		totalLen += len(s)
	}
	
	result := make([]byte, 0, totalLen)
	result = append(result, strs[0]...)
	for _, s := range strs[1:] {
		result = append(result, sep...)
		result = append(result, s...)
	}
	return string(result)
}
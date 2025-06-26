package adapters

import (
	"time"

	"github.com/shard-legends/production-service/internal/storage"
	"github.com/shard-legends/production-service/pkg/metrics"
)

// MetricsAdapter адаптирует metrics для storage.MetricsInterface
type MetricsAdapter struct{}

// NewMetricsAdapter создает новый адаптер для метрик
func NewMetricsAdapter() storage.MetricsInterface {
	return &MetricsAdapter{}
}

// IncDBQuery увеличивает счетчик запросов к БД
func (a *MetricsAdapter) IncDBQuery(operation string) {
	metrics.DBQueriesTotal.WithLabelValues(operation, "recipes").Inc()
}

// IncCacheHit увеличивает счетчик попаданий в кеш
func (a *MetricsAdapter) IncCacheHit(cacheType string) {
	metrics.RedisOperationsTotal.WithLabelValues("get", "hit").Inc()
}

// IncCacheMiss увеличивает счетчик промахов кеша
func (a *MetricsAdapter) IncCacheMiss(cacheType string) {
	metrics.RedisOperationsTotal.WithLabelValues("get", "miss").Inc()
}

// ObserveDBQueryDuration записывает время выполнения запроса к БД
func (a *MetricsAdapter) ObserveDBQueryDuration(operation string, duration time.Duration) {
	metrics.DBQueryDuration.WithLabelValues(operation, "recipes").Observe(duration.Seconds())
}

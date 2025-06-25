package metrics

import (
	"time"
)

// ServiceMetrics implements the service.MetricsInterface for the service layer.
type ServiceMetrics struct {
	metrics *Metrics
}

// NewServiceMetrics creates a new ServiceMetrics instance.
// It acts as an adapter to provide metrics to the service layer.
func NewServiceMetrics(m *Metrics) *ServiceMetrics {
	return &ServiceMetrics{
		metrics: m,
	}
}

// RecordInventoryOperation records an inventory operation metric.
func (sm *ServiceMetrics) RecordInventoryOperation(operationType, section, status string) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.InventoryOperationsTotal.WithLabelValues(operationType, section, status).Inc()
}

// RecordBalanceCalculation records balance calculation metrics.
func (sm *ServiceMetrics) RecordBalanceCalculation(section string, duration time.Duration, status string) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.BalanceCalculations.WithLabelValues(section, status).Inc()
	if status == "success" {
		sm.metrics.BalanceCalculationDuration.WithLabelValues(section).Observe(duration.Seconds())
	}
}

// RecordCacheHit records a cache hit.
func (sm *ServiceMetrics) RecordCacheHit(cacheType string) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss.
func (sm *ServiceMetrics) RecordCacheMiss(cacheType string) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.CacheMisses.WithLabelValues(cacheType).Inc()
}

// RecordTransactionMetrics records transaction-related metrics.
func (sm *ServiceMetrics) RecordTransactionMetrics(operationType string, operationCount int, duration time.Duration) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.TransactionOperations.WithLabelValues(operationType).Observe(float64(operationCount))
	sm.metrics.TransactionDuration.WithLabelValues(operationType).Observe(duration.Seconds())
}

// RecordItemsPerInventory records the number of items in a user's inventory.
func (sm *ServiceMetrics) RecordItemsPerInventory(section string, count int) {
	if sm.metrics == nil {
		return
	}
	sm.metrics.ItemsPerInventory.WithLabelValues(section).Observe(float64(count))
}

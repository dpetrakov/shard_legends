package metrics

import (
	"testing"
	"time"
)

var globalMetrics *Metrics

func TestMetricsCreation(t *testing.T) {
	if globalMetrics == nil {
		globalMetrics = New()
	}
	metrics := globalMetrics

	if metrics == nil {
		t.Fatal("Expected metrics to be created, got nil")
	}

	// Test that all metrics are initialized
	if metrics.HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal not initialized")
	}
	if metrics.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration not initialized")
	}
	if metrics.HTTPRequestsInFlight == nil {
		t.Error("HTTPRequestsInFlight not initialized")
	}
	if metrics.DatabaseConnections == nil {
		t.Error("DatabaseConnections not initialized")
	}
	if metrics.RedisConnections == nil {
		t.Error("RedisConnections not initialized")
	}
	if metrics.InventoryOperationsTotal == nil {
		t.Error("InventoryOperationsTotal not initialized")
	}
	if metrics.DependencyHealth == nil {
		t.Error("DependencyHealth not initialized")
	}
}

func TestMetricsInitialize(t *testing.T) {
	if globalMetrics == nil {
		globalMetrics = New()
	}
	metrics := globalMetrics
	metrics.Initialize()

	// Test should not panic and complete successfully
}

func TestServiceMetrics(t *testing.T) {
	if globalMetrics == nil {
		globalMetrics = New()
	}
	serviceMetrics := NewServiceMetrics(globalMetrics)

	if serviceMetrics == nil {
		t.Fatal("Expected service metrics to be created, got nil")
	}

	// Test that methods can be called without panicking
	serviceMetrics.RecordInventoryOperation("test", "section1", "success")
	serviceMetrics.RecordBalanceCalculation("section1", time.Millisecond*100, "success")
	serviceMetrics.RecordCacheHit("balance")
	serviceMetrics.RecordCacheMiss("balance")
	serviceMetrics.RecordTransactionMetrics("reserve", 5, time.Millisecond*50)
	serviceMetrics.RecordItemsPerInventory("section1", 10)
}

func TestServiceMetricsWithNilMetrics(t *testing.T) {
	serviceMetrics := NewServiceMetrics(nil)

	if serviceMetrics == nil {
		t.Fatal("Expected service metrics to be created even with nil metrics, got nil")
	}

	// Test that methods can be called without panicking even with nil metrics
	serviceMetrics.RecordInventoryOperation("test", "section1", "success")
	serviceMetrics.RecordBalanceCalculation("section1", time.Millisecond*100, "success")
	serviceMetrics.RecordCacheHit("balance")
	serviceMetrics.RecordCacheMiss("balance")
	serviceMetrics.RecordTransactionMetrics("reserve", 5, time.Millisecond*50)
	serviceMetrics.RecordItemsPerInventory("section1", 10)
}

func TestUpdateDependencyHealth(t *testing.T) {
	if globalMetrics == nil {
		globalMetrics = New()
	}
	metrics := globalMetrics

	// Test updating dependency health
	metrics.UpdateDependencyHealth("postgres", true)
	metrics.UpdateDependencyHealth("redis", false)

	// Should not panic
}

func TestShutdown(t *testing.T) {
	if globalMetrics == nil {
		globalMetrics = New()
	}
	metrics := globalMetrics

	// Test shutdown
	metrics.Shutdown()

	// Should not panic
}

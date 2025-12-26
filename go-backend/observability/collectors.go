package observability

import (
	"context"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// StartSystemMetricsCollection starts a background goroutine to collect system metrics
// Returns a stop function to cleanup the goroutine
func StartSystemMetricsCollection(metrics *Metrics, interval time.Duration) func() {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				collectSystemMetrics(metrics)
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// Return stop function
	return cancel
}

func collectSystemMetrics(metrics *Metrics) {
	ctx := context.Background()
	
	// Collect goroutine count
	numGoroutines := int64(runtime.NumGoroutine())
	metrics.GoroutinesActive.Record(ctx, numGoroutines)
	
	// Collect memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	metrics.MemoryAllocated.Record(ctx, int64(m.Alloc))
	metrics.MemoryHeap.Record(ctx, int64(m.HeapAlloc))
}

// RecordDBMetrics records database connection pool metrics as absolute values
// Note: These should be called with current connection counts, not deltas
func RecordDBMetrics(metrics *Metrics, activeChange, idleChange int64) {
	ctx := context.Background()
	// Use Add for UpDownCounter with delta values
	metrics.DBConnectionsActive.Add(ctx, activeChange)
	metrics.DBConnectionsIdle.Add(ctx, idleChange)
}

// RecordKafkaPartitionMetrics records Kafka partition information
func RecordKafkaPartitionMetrics(metrics *Metrics, topic string, partition int32, offset, lag int64) {
	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("topic", topic),
		attribute.Int("partition", int(partition)),
	}
	
	metrics.KafkaConsumerOffset.Record(ctx, offset, metric.WithAttributes(attrs...))
	metrics.KafkaConsumerLag.Record(ctx, lag, metric.WithAttributes(attrs...))
}

// RecordKafkaPartitionCount records the number of partitions for a topic
func RecordKafkaPartitionCount(metrics *Metrics, topic string, count int64) {
	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("topic", topic),
	}
	metrics.KafkaPartitionCount.Record(ctx, count, metric.WithAttributes(attrs...))
}

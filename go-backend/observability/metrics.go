package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	// Global meter for creating metrics
	meter metric.Meter
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	HTTPRequestSize     metric.Int64Histogram
	HTTPResponseSize    metric.Int64Histogram
	HTTPActiveRequests  metric.Int64UpDownCounter
	
	// Business metrics
	IndexingJobsTotal      metric.Int64Counter
	IndexingJobsActive     metric.Int64UpDownCounter
	IndexingJobsDuration   metric.Float64Histogram
	IndexingPagesTotal     metric.Int64Counter
	IndexingErrors         metric.Int64Counter
	
	// Database metrics
	DBQueriesTotal        metric.Int64Counter
	DBQueryDuration       metric.Float64Histogram
	DBConnectionsActive   metric.Int64UpDownCounter
	DBConnectionsIdle     metric.Int64UpDownCounter
	DBErrors              metric.Int64Counter
	
	// Kafka metrics - Producer
	KafkaMessagesProduced       metric.Int64Counter
	KafkaProducerErrors         metric.Int64Counter
	KafkaProducerLatency        metric.Float64Histogram
	KafkaProducerBatchSize      metric.Int64Histogram
	
	// Kafka metrics - Consumer
	KafkaMessagesConsumed       metric.Int64Counter
	KafkaConsumerErrors         metric.Int64Counter
	KafkaConsumerLag            metric.Int64Gauge
	KafkaConsumerOffset         metric.Int64Gauge
	KafkaPartitionCount         metric.Int64Gauge
	KafkaConsumerLatency        metric.Float64Histogram
	
	// System metrics
	GoroutinesActive      metric.Int64Gauge
	MemoryAllocated       metric.Int64Gauge
	MemoryHeap            metric.Int64Gauge
}

// InitMetrics initializes OpenTelemetry metrics with Prometheus exporter
func InitMetrics(serviceName string, serviceVersion string) (*Metrics, *http.ServeMux, error) {
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	// Create meter provider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	// Set global meter provider
	otel.SetMeterProvider(provider)

	// Create meter
	meter = provider.Meter(serviceName)

	// Initialize metrics
	metrics := &Metrics{}
	
	// HTTP metrics
	metrics.HTTPRequestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_requests_total counter: %w", err)
	}

	metrics.HTTPRequestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_request_duration_seconds histogram: %w", err)
	}

	metrics.HTTPRequestSize, err = meter.Int64Histogram(
		"http_request_size_bytes",
		metric.WithDescription("HTTP request size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_request_size_bytes histogram: %w", err)
	}

	metrics.HTTPResponseSize, err = meter.Int64Histogram(
		"http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_response_size_bytes histogram: %w", err)
	}

	metrics.HTTPActiveRequests, err = meter.Int64UpDownCounter(
		"http_requests_active",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_requests_active gauge: %w", err)
	}

	// Indexing job metrics
	metrics.IndexingJobsTotal, err = meter.Int64Counter(
		"indexing_jobs_total",
		metric.WithDescription("Total number of indexing jobs"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create indexing_jobs_total counter: %w", err)
	}

	metrics.IndexingJobsActive, err = meter.Int64UpDownCounter(
		"indexing_jobs_active",
		metric.WithDescription("Number of currently active indexing jobs"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create indexing_jobs_active gauge: %w", err)
	}

	metrics.IndexingJobsDuration, err = meter.Float64Histogram(
		"indexing_job_duration_seconds",
		metric.WithDescription("Indexing job duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create indexing_job_duration_seconds histogram: %w", err)
	}

	metrics.IndexingPagesTotal, err = meter.Int64Counter(
		"indexing_pages_total",
		metric.WithDescription("Total number of pages indexed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create indexing_pages_total counter: %w", err)
	}

	metrics.IndexingErrors, err = meter.Int64Counter(
		"indexing_errors_total",
		metric.WithDescription("Total number of indexing errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create indexing_errors_total counter: %w", err)
	}

	// Database metrics
	metrics.DBQueriesTotal, err = meter.Int64Counter(
		"db_queries_total",
		metric.WithDescription("Total number of database queries"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db_queries_total counter: %w", err)
	}

	metrics.DBQueryDuration, err = meter.Float64Histogram(
		"db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db_query_duration_seconds histogram: %w", err)
	}

	metrics.DBConnectionsActive, err = meter.Int64UpDownCounter(
		"db_connections_active",
		metric.WithDescription("Number of active database connections"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db_connections_active gauge: %w", err)
	}

	metrics.DBConnectionsIdle, err = meter.Int64UpDownCounter(
		"db_connections_idle",
		metric.WithDescription("Number of idle database connections"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db_connections_idle gauge: %w", err)
	}

	metrics.DBErrors, err = meter.Int64Counter(
		"db_errors_total",
		metric.WithDescription("Total number of database errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db_errors_total counter: %w", err)
	}

	// Kafka metrics - Producer
	metrics.KafkaMessagesProduced, err = meter.Int64Counter(
		"kafka_messages_produced_total",
		metric.WithDescription("Total number of Kafka messages produced"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_messages_produced_total counter: %w", err)
	}

	metrics.KafkaProducerErrors, err = meter.Int64Counter(
		"kafka_producer_errors_total",
		metric.WithDescription("Total number of Kafka producer errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_producer_errors_total counter: %w", err)
	}

	metrics.KafkaProducerLatency, err = meter.Float64Histogram(
		"kafka_producer_latency_seconds",
		metric.WithDescription("Kafka producer message latency in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_producer_latency_seconds histogram: %w", err)
	}

	metrics.KafkaProducerBatchSize, err = meter.Int64Histogram(
		"kafka_producer_batch_size",
		metric.WithDescription("Kafka producer batch size"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_producer_batch_size histogram: %w", err)
	}

	// Kafka metrics - Consumer
	metrics.KafkaMessagesConsumed, err = meter.Int64Counter(
		"kafka_messages_consumed_total",
		metric.WithDescription("Total number of Kafka messages consumed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_messages_consumed_total counter: %w", err)
	}

	metrics.KafkaConsumerErrors, err = meter.Int64Counter(
		"kafka_consumer_errors_total",
		metric.WithDescription("Total number of Kafka consumer errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_consumer_errors_total counter: %w", err)
	}

	metrics.KafkaConsumerLag, err = meter.Int64Gauge(
		"kafka_consumer_lag",
		metric.WithDescription("Kafka consumer lag (messages behind)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_consumer_lag gauge: %w", err)
	}

	metrics.KafkaConsumerOffset, err = meter.Int64Gauge(
		"kafka_consumer_offset",
		metric.WithDescription("Current Kafka consumer offset"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_consumer_offset gauge: %w", err)
	}

	metrics.KafkaPartitionCount, err = meter.Int64Gauge(
		"kafka_partition_count",
		metric.WithDescription("Number of Kafka partitions"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_partition_count gauge: %w", err)
	}

	metrics.KafkaConsumerLatency, err = meter.Float64Histogram(
		"kafka_consumer_latency_seconds",
		metric.WithDescription("Kafka consumer message processing latency in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka_consumer_latency_seconds histogram: %w", err)
	}

	// System metrics
	metrics.GoroutinesActive, err = meter.Int64Gauge(
		"goroutines_active",
		metric.WithDescription("Number of active goroutines"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create goroutines_active gauge: %w", err)
	}

	metrics.MemoryAllocated, err = meter.Int64Gauge(
		"memory_allocated_bytes",
		metric.WithDescription("Memory allocated in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create memory_allocated_bytes gauge: %w", err)
	}

	metrics.MemoryHeap, err = meter.Int64Gauge(
		"memory_heap_bytes",
		metric.WithDescription("Memory heap in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create memory_heap_bytes gauge: %w", err)
	}

	// Create HTTP handler for Prometheus metrics endpoint
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("OpenTelemetry metrics initialized with Prometheus exporter")

	return metrics, mux, nil
}

// GetMeter returns the global meter instance
func GetMeter() metric.Meter {
	return meter
}

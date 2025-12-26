# Observability Setup - MCP-Docs Platform

This document describes the complete observability stack integrated into the MCP-Docs platform using OpenTelemetry, Prometheus, and Grafana.

## Overview

The MCP-Docs platform implements comprehensive observability without requiring an OpenTelemetry Collector. Metrics are collected directly from services using OpenTelemetry SDKs and exposed via Prometheus endpoints.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Services Layer                            │
│  ┌──────────────┐         ┌──────────────────┐             │
│  │  Go Backend  │         │  Python RAG      │             │
│  │  :9090       │         │  Agent :9091     │             │
│  │  (metrics)   │         │  (metrics)       │             │
│  └──────────────┘         └──────────────────┘             │
└─────────────────────────────────────────────────────────────┘
                    │                    │
                    │  Scrape metrics    │
                    ▼                    ▼
            ┌────────────────────────────────┐
            │      Prometheus :9090          │
            │  (Metrics Collection & Storage)│
            └────────────────────────────────┘
                            │
                            │  Query metrics
                            ▼
            ┌────────────────────────────────┐
            │       Grafana :3001            │
            │   (Metrics Visualization)      │
            └────────────────────────────────┘
```

## Components

### 1. OpenTelemetry Integration

Both Go backend and Python RAG agent export OpenTelemetry metrics directly to Prometheus format without requiring an OTEL Collector.

#### Go Backend Metrics (Port 9090)

**HTTP Metrics:**
- `http_requests_total` - Total HTTP requests by method, path, status
- `http_request_duration_seconds` - Request latency percentiles
- `http_request_size_bytes` - Request payload sizes
- `http_response_size_bytes` - Response payload sizes
- `http_requests_active` - Currently active requests

**Business Metrics:**
- `indexing_jobs_total` - Total indexing jobs by status
- `indexing_jobs_active` - Currently active indexing jobs
- `indexing_job_duration_seconds` - Job processing duration
- `indexing_pages_total` - Total pages indexed
- `indexing_errors_total` - Total indexing errors

**Database Metrics:**
- `db_queries_total` - Total database queries
- `db_query_duration_seconds` - Query execution time
- `db_connections_active` - Active DB connections
- `db_connections_idle` - Idle DB connections
- `db_errors_total` - Database errors

**Kafka Metrics (Producer):**
- `kafka_messages_produced_total` - Messages produced by topic
- `kafka_producer_errors_total` - Producer errors
- `kafka_producer_latency_seconds` - Message send latency
- `kafka_producer_batch_size` - Batch sizes

**Kafka Metrics (Consumer):**
- `kafka_messages_consumed_total` - Messages consumed by topic
- `kafka_consumer_errors_total` - Consumer errors
- `kafka_consumer_lag` - Consumer lag by topic/partition
- `kafka_consumer_offset` - Current offset by topic/partition
- `kafka_partition_count` - Number of partitions per topic
- `kafka_consumer_latency_seconds` - Message processing latency

**System Metrics:**
- `goroutines_active` - Active goroutines
- `memory_allocated_bytes` - Memory allocated
- `memory_heap_bytes` - Heap memory usage

#### Python RAG Agent Metrics (Port 9091)

**Kafka Metrics:**
- `kafka_messages_consumed_total` - Messages consumed
- `kafka_messages_processed_total` - Successfully processed messages
- `kafka_processing_errors_total` - Processing errors
- `kafka_consumer_lag` - Consumer lag
- `kafka_processing_duration_seconds` - Processing duration

**RAG Processing Metrics:**
- `rag_instances_active` - Active RAG instances
- `rag_instances_total` - Total RAG instances created
- `rag_documents_processed_total` - Documents processed
- `rag_processing_duration_seconds` - RAG processing duration

**Document Processing:**
- `html_documents_cleaned_total` - HTML docs cleaned
- `html_cleaning_duration_seconds` - Cleaning duration
- `document_chunks_created_total` - Document chunks created

**Indexing Jobs:**
- `indexing_jobs_started_total` - Jobs started
- `indexing_jobs_completed_total` - Jobs completed
- `indexing_jobs_failed_total` - Jobs failed
- `indexing_pages_total` - Pages indexed

**API Calls:**
- `api_calls_total` - API calls to Go backend
- `api_call_duration_seconds` - API call duration
- `api_call_errors_total` - API call errors

### 2. Prometheus

Prometheus scrapes metrics from both services every 15 seconds.

**Configuration:** `/prometheus/prometheus.yml`

**Access:** http://localhost:9090

**Key Features:**
- 15-second scrape interval
- Service discovery via static configs
- Labels for service identification
- Data retention (default 15 days)

### 3. Grafana

Grafana provides visualization dashboards for all metrics.

**Access:** http://localhost:3001
- Username: `admin`
- Password: `admin`

**Pre-configured Dashboards:**
1. **MCP-Docs Platform Overview** - Comprehensive system overview
   - HTTP request rate and latency percentiles (p50, p95, p99)
   - Kafka message throughput and consumer lag
   - Indexing job metrics
   - Memory usage
   - Database query latency
   - HTTP status code distribution

## Getting Started

### 1. Start the Observability Stack

```bash
# Ensure .env file is configured
cp .env.example .env

# Start all services including observability stack
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 2. Access Metrics Endpoints

**Go Backend Metrics:**
```bash
curl http://localhost:9090/metrics
```

**Python RAG Agent Metrics:**
```bash
curl http://localhost:9091/metrics
```

### 3. View Prometheus

Open http://localhost:9090 in your browser to access the Prometheus UI.

**Example Queries:**
```promql
# Average API latency (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Kafka consumer lag
kafka_consumer_lag

# Active indexing jobs
indexing_jobs_active

# Memory usage trend
memory_allocated_bytes
```

### 4. View Grafana Dashboards

1. Open http://localhost:3001
2. Login with admin/admin
3. Navigate to Dashboards → MCP-Docs Platform Overview

## Monitoring Best Practices

### 1. Key Metrics to Watch

**Service Health:**
- HTTP error rate (5xx responses)
- API latency (p95, p99)
- Active connections
- Memory usage

**Kafka Health:**
- Consumer lag (should be < 1000)
- Message throughput
- Processing errors
- Partition distribution

**Business Metrics:**
- Indexing job success rate
- Pages indexed per hour
- RAG instance pool utilization
- Document processing latency

### 2. Alerting (Future Enhancement)

Consider setting up alerts for:
- Consumer lag > 1000 messages
- Error rate > 5%
- API latency p95 > 1 second
- Memory usage > 80%
- Failed indexing jobs

### 3. Dashboard Customization

You can customize the Grafana dashboard by:
1. Logging into Grafana (http://localhost:3001)
2. Navigate to the dashboard
3. Click "Edit" on any panel
4. Modify queries, thresholds, or visualizations
5. Save the dashboard

## Troubleshooting

### Metrics Not Appearing

**Go Backend:**
```bash
# Check if metrics endpoint is accessible
docker exec mcp-docs-go-backend wget -O- http://localhost:9090/metrics

# Check logs
docker logs mcp-docs-go-backend
```

**Python RAG Agent:**
```bash
# Check if metrics endpoint is accessible
docker exec mcp-docs-rag-agent wget -O- http://localhost:9091/metrics

# Check logs
docker logs mcp-docs-rag-agent
```

**Prometheus:**
```bash
# Check if Prometheus is scraping targets
# Go to http://localhost:9090/targets

# Check Prometheus logs
docker logs mcp-docs-prometheus
```

**Grafana:**
```bash
# Check Grafana logs
docker logs mcp-docs-grafana

# Verify datasource connection
# Go to http://localhost:3001/datasources
```

### Common Issues

1. **Prometheus can't reach services**
   - Ensure all services are on the same Docker network
   - Check service names in prometheus.yml match container names

2. **Grafana shows "No data"**
   - Verify Prometheus datasource is configured correctly
   - Check time range in dashboard
   - Ensure Prometheus is collecting metrics

3. **Metrics show as 0 or empty**
   - Check if application is generating traffic
   - Verify metric names in Prometheus (http://localhost:9090/graph)
   - Ensure services have initialized metrics

## Environment Variables

Add these to your `.env` file:

```bash
# Metrics ports (optional - defaults shown)
METRICS_PORT=9091  # Python RAG Agent metrics port
```

## Performance Impact

The observability stack has minimal performance impact:
- **Go Backend:** < 1% CPU overhead
- **Python Agent:** < 2% CPU overhead
- **Prometheus:** ~50-100MB RAM for storage
- **Grafana:** ~100MB RAM

## Advanced Configuration

### Custom Metrics

**Go Backend:**
```go
// Access global metrics from observability package
import "mcpdocs/observability"

// Record custom metric
ctx := context.Background()
container.Metrics.IndexingJobsTotal.Add(ctx, 1, 
    metric.WithAttributes(
        attribute.String("status", "started"),
    ),
)
```

**Python RAG Agent:**
```python
# Access metrics from RAGAgent instance
self.metrics.indexing_jobs_started.add(1, {
    "project_id": str(project_id)
})
```

### Metric Retention

Prometheus default retention is 15 days. To change:

Edit `docker-compose.yml`:
```yaml
prometheus:
  command:
    - '--storage.tsdb.retention.time=30d'  # 30 days
```

## End-to-End Observability

The platform provides complete end-to-end visibility:

1. **API Request** → HTTP metrics in Go backend
2. **Kafka Message** → Producer metrics in Go, Consumer metrics in Python
3. **Document Processing** → RAG processing metrics in Python
4. **Database Queries** → DB metrics in Go backend
5. **System Health** → Memory, goroutines, connections

This allows you to:
- Track a request from API → Kafka → Processing → Storage
- Identify bottlenecks in the pipeline
- Monitor system resource utilization
- Understand business metrics (pages indexed, jobs completed)

## Next Steps

1. Set up alerting with Prometheus Alertmanager
2. Add custom business dashboards
3. Integrate with log aggregation (ELK, Loki)
4. Add distributed tracing with Jaeger
5. Export metrics to external monitoring (DataDog, NewRelic)

## Support

For issues or questions about observability:
1. Check Prometheus targets: http://localhost:9090/targets
2. Review service logs: `docker logs <service-name>`
3. Verify metric endpoints are accessible
4. Check Grafana datasource configuration

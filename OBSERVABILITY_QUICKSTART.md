# Quick Start Guide - Observability Stack

This guide will help you quickly set up and verify the observability stack for the MCP-Docs platform.

## Prerequisites

- Docker and Docker Compose installed
- `.env` file configured (copy from `.env.example`)
- At least 4GB of available RAM

## Step 1: Start the Stack

```bash
# Clone the repository (if you haven't already)
git clone <repository-url>
cd mcp-docs

# Copy environment file
cp .env.example .env

# Edit .env and set your passwords
# At minimum, set POSTGRES_PASSWORD

# Start all services
docker-compose up -d

# Check that all services are running
docker-compose ps
```

You should see all services in "Up" state:
- postgres
- zookeeper
- kafka
- go-backend
- frontend
- nginx
- rag-agent
- prometheus
- grafana

## Step 2: Verify Metrics Endpoints

### Go Backend Metrics
```bash
# Check if metrics are being collected
curl http://localhost:9090/metrics | grep http_requests_total
```

You should see output like:
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/",status="200",status_class="2xx"} 5
```

### Python RAG Agent Metrics
```bash
# Check if Python agent metrics are available
curl http://localhost:9091/metrics | grep kafka_messages_consumed
```

You should see output like:
```
# HELP kafka_messages_consumed_total Total number of Kafka messages consumed
# TYPE kafka_messages_consumed_total counter
kafka_messages_consumed_total 0
```

## Step 3: Access Prometheus

1. Open your browser and navigate to: http://localhost:9090
2. Click on "Status" → "Targets" in the top menu
3. Verify that both `go-backend` and `python-rag-agent` targets are "UP"

### Run a Test Query

In the Prometheus UI:

1. Go to the "Graph" tab
2. Enter this query: `rate(http_requests_total[5m])`
3. Click "Execute"
4. Switch to the "Graph" tab to see a visualization

## Step 4: Access Grafana

1. Open your browser and navigate to: http://localhost:3001
2. Login with:
   - Username: `admin`
   - Password: `admin`
3. Skip the password change prompt (or set a new password)
4. Click on "Dashboards" in the left menu
5. Select "MCP-Docs Platform Overview"

You should see a comprehensive dashboard with:
- HTTP request rate and latency
- Kafka metrics
- Indexing jobs
- Memory usage
- And more!

## Step 5: Generate Some Traffic

To see metrics in action, generate some API traffic:

```bash
# Make some API calls to the health check endpoint
for i in {1..50}; do
  curl -s http://localhost:8080/ > /dev/null
  echo "Request $i completed"
  sleep 0.1
done
```

Now refresh your Grafana dashboard to see the metrics update!

## Step 6: Explore Metrics

### In Prometheus

Try these queries:

```promql
# Average API latency (95th percentile)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Total requests per second
rate(http_requests_total[1m])

# Active goroutines
goroutines_active

# Memory usage
memory_allocated_bytes
```

### In Grafana

1. Click on any panel to expand it
2. Click "Edit" to see the query
3. Modify the query to explore different metrics
4. Add new panels with the "+" icon

## Troubleshooting

### Services Not Starting

```bash
# Check logs for a specific service
docker logs mcp-docs-go-backend
docker logs mcp-docs-prometheus
docker logs mcp-docs-grafana

# Restart a specific service
docker-compose restart go-backend
```

### Metrics Not Showing

```bash
# Verify Go backend metrics endpoint
docker exec mcp-docs-go-backend wget -qO- http://localhost:9090/metrics

# Verify Python agent metrics endpoint
docker exec mcp-docs-rag-agent wget -qO- http://localhost:9091/metrics

# Check Prometheus targets
# Go to http://localhost:9090/targets
```

### Grafana Dashboard Empty

1. Check that Prometheus is collecting data: http://localhost:9090/graph
2. Run a query: `up{job="go-backend"}`
3. If it returns results, check Grafana datasource configuration
4. Go to Configuration → Data Sources → Prometheus
5. Click "Test" to verify connection

### Port Already in Use

If ports are already in use, you can change them in `docker-compose.yml`:

```yaml
prometheus:
  ports:
    - "9095:9090"  # Change host port to 9095

grafana:
  ports:
    - "3002:3000"  # Change host port to 3002
```

## What's Next?

1. **Explore the Dashboard**: Familiarize yourself with the pre-built dashboard
2. **Create Custom Dashboards**: Add panels for specific metrics you want to track
3. **Set Up Alerts**: Configure Prometheus alerts for critical thresholds
4. **Read the Full Documentation**: See [OBSERVABILITY.md](./OBSERVABILITY.md) for comprehensive details
5. **Customize Metrics**: Add custom business metrics to your application

## Key Metrics to Monitor

### For Development:
- HTTP request latency (should be < 100ms for most endpoints)
- Active requests (to spot traffic patterns)
- Memory usage (to detect leaks)
- Kafka consumer lag (should be near 0)

### For Production:
- Error rates (5xx responses)
- Consumer lag (alerts if > 1000)
- Database query latency
- Indexing job success rate

## Quick Commands Reference

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart observability stack
docker-compose restart prometheus grafana

# Check service health
docker-compose ps
curl http://localhost:9090/-/healthy  # Prometheus
curl http://localhost:3001/api/health  # Grafana
```

## Support

For detailed information, see:
- [OBSERVABILITY.md](./OBSERVABILITY.md) - Complete observability documentation
- [ENV_SETUP.md](./ENV_SETUP.md) - Environment setup guide
- Grafana Documentation: https://grafana.com/docs/
- Prometheus Documentation: https://prometheus.io/docs/

Happy monitoring! 🚀📊

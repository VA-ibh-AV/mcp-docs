# Observability Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          MCP-Docs Platform                                   │
│                         Observability Stack                                  │
└─────────────────────────────────────────────────────────────────────────────┘

                         ┌──────────────────────────┐
                         │   External Access        │
                         │   (localhost)            │
                         └──────────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
         ┌─────────────────┐ ┌─────────────┐ ┌──────────────┐
         │  Grafana :3001  │ │ Prom :9090  │ │ Go :9090     │
         │  (Dashboards)   │ │ (Metrics DB)│ │ (API Metrics)│
         └─────────────────┘ └─────────────┘ └──────────────┘
                │                    │               │
                │                    │               │
                │    ┌───────────────┘               │
                │    │                               │
                ▼    ▼                               ▼
         ┌──────────────────────────────────────────────────┐
         │           Docker Network: mcp-docs-network       │
         │                                                  │
         │  ┌──────────────────────────────────────────┐  │
         │  │  Application Services                     │  │
         │  │                                           │  │
         │  │  ┌────────────────┐  ┌────────────────┐ │  │
         │  │  │  Go Backend    │  │ Python RAG     │ │  │
         │  │  │  Port: 8005    │  │ Agent          │ │  │
         │  │  │  Metrics: 9090 │  │ Metrics: 9091  │ │  │
         │  │  └────────────────┘  └────────────────┘ │  │
         │  │         │                    │           │  │
         │  │         └─────────┬──────────┘           │  │
         │  │                   │                      │  │
         │  │                   ▼                      │  │
         │  │         ┌──────────────────┐            │  │
         │  │         │  Kafka + ZK      │            │  │
         │  │         │  Port: 9092      │            │  │
         │  │         └──────────────────┘            │  │
         │  │                   │                      │  │
         │  │                   ▼                      │  │
         │  │         ┌──────────────────┐            │  │
         │  │         │  PostgreSQL      │            │  │
         │  │         │  Port: 5432      │            │  │
         │  │         └──────────────────┘            │  │
         │  └──────────────────────────────────────────┘  │
         │                                                  │
         │  ┌──────────────────────────────────────────┐  │
         │  │  Observability Services                   │  │
         │  │                                           │  │
         │  │  ┌────────────────┐                      │  │
         │  │  │  Prometheus    │◄─────────────────┐   │  │
         │  │  │  Port: 9090    │                  │   │  │
         │  │  │  (Time Series  │                  │   │  │
         │  │  │   Database)    │                  │   │  │
         │  │  └────────────────┘                  │   │  │
         │  │         │                            │   │  │
         │  │         │ Scrapes every 15s          │   │  │
         │  │         │                            │   │  │
         │  │         ▼                            │   │  │
         │  │  ┌─────────────────┐                │   │  │
         │  │  │ Metrics Sources │                │   │  │
         │  │  ├─────────────────┤                │   │  │
         │  │  │ go-backend:9090 │────────────────┘   │  │
         │  │  │ rag-agent:9091  │────────────────────┘  │
         │  │  └─────────────────┘                       │
         │  │                                            │  │
         │  │  ┌────────────────┐                       │  │
         │  │  │  Grafana       │                       │  │
         │  │  │  Port: 3000    │                       │  │
         │  │  │  (exposed:3001)│                       │  │
         │  │  │  User:pass     │                       │  │
         │  │  │  admin:admin   │                       │  │
         │  │  └────────────────┘                       │  │
         │  │         │                                 │  │
         │  │         │ Queries Prometheus              │  │
         │  │         └──────────────┐                  │  │
         │  │                        │                  │  │
         │  └────────────────────────┼──────────────────┘  │
         │                            ▼                     │
         └──────────────────────────────────────────────────┘
                                     │
                         ┌───────────┴──────────┐
                         │   Persistent Storage │
                         │   Docker Volumes     │
                         │   - prometheus_data  │
                         │   - grafana_data     │
                         └──────────────────────┘


Metrics Flow:
═════════════

1. Applications → OpenTelemetry SDK → Prometheus Format
2. Prometheus → Scrapes metrics endpoints every 15s → Stores time-series data
3. Grafana → Queries Prometheus → Displays dashboards
4. Users → Access Grafana → View metrics and create alerts


Key Metrics Collected:
══════════════════════

┌─────────────────────────────────────────────────────────────────┐
│ HTTP Metrics (Go Backend)                                       │
├─────────────────────────────────────────────────────────────────┤
│ • Request rate (req/sec)                                        │
│ • Latency percentiles (p50, p95, p99)                          │
│ • Request/response sizes                                        │
│ • Status codes distribution                                     │
│ • Active concurrent requests                                    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Kafka Metrics (Both Services)                                   │
├─────────────────────────────────────────────────────────────────┤
│ • Messages produced/consumed (by topic)                         │
│ • Consumer lag (by topic/partition)                            │
│ • Processing latency                                            │
│ • Partition counts                                              │
│ • Producer/consumer errors                                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Business Metrics                                                │
├─────────────────────────────────────────────────────────────────┤
│ • Indexing jobs (started, completed, failed)                   │
│ • Pages indexed                                                 │
│ • RAG instances (active, total)                                │
│ • Documents processed                                           │
│ • HTML documents cleaned                                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ System Metrics                                                  │
├─────────────────────────────────────────────────────────────────┤
│ • Memory usage (allocated, heap)                               │
│ • Goroutines count (Go)                                        │
│ • Database connections (active, idle)                          │
│ • Database query latency                                        │
└─────────────────────────────────────────────────────────────────┘


Access Points:
═════════════

Service              | URL                          | Credentials
─────────────────────┼──────────────────────────────┼──────────────
Grafana Dashboards   | http://localhost:3001        | admin/admin
Prometheus UI        | http://localhost:9090        | None
Go Backend Metrics   | http://localhost:9090/metrics| None  
Python Agent Metrics | http://localhost:9091/metrics| None
Frontend             | http://localhost:3000        | Registration
API (via Nginx)      | http://localhost:80          | Auth required


Technology Stack:
════════════════

┌──────────────────┬────────────────────────────────────────────┐
│ Component        │ Technology                                 │
├──────────────────┼────────────────────────────────────────────┤
│ Instrumentation  │ OpenTelemetry SDK (Go + Python)           │
│ Metrics Format   │ Prometheus (native format)                │
│ Collection       │ Prometheus (pull-based scraping)          │
│ Visualization    │ Grafana (dashboards)                      │
│ Storage          │ Prometheus TSDB (time-series database)    │
└──────────────────┴────────────────────────────────────────────┘


Benefits:
════════

✅ No OTEL Collector required (direct Prometheus export)
✅ Complete end-to-end visibility across all services
✅ Real-time metrics (15-second refresh)
✅ Pre-built dashboards for immediate insights
✅ Low overhead (< 2% CPU, minimal memory)
✅ Production-ready configuration
✅ Easy to extend with custom metrics
✅ Kafka topic/partition monitoring
✅ API latency tracking (p50, p95, p99)
✅ System resource monitoring
```

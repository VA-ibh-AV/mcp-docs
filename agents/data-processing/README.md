# Python RAG Agent

A scalable document processing agent that consumes indexing jobs from Kafka and stores them in LightRAG with PostgreSQL backend.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Python RAG Agent                            │
│                                                                 │
│  ┌─────────────────┐   ┌─────────────────┐   ┌───────────────┐ │
│  │  Kafka Consumer │──▶│ Document        │──▶│ LightRAG Pool │ │
│  │  (aiokafka)     │   │ Processor       │   │ (per collection)│
│  └─────────────────┘   └─────────────────┘   └───────────────┘ │
│          │                     │                     │          │
│          │                     │                     │          │
│          ▼                     ▼                     ▼          │
│  ┌─────────────────┐   ┌─────────────────┐   ┌───────────────┐ │
│  │ Consumer Group  │   │ Status Client   │   │  PostgreSQL   │ │
│  │ (horizontal     │   │ (Go Backend     │   │  (pgvector)   │ │
│  │  scaling)       │   │  updates)       │   │               │ │
│  └─────────────────┘   └─────────────────┘   └───────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Features

- **Scalable**: Horizontal scaling via Kafka consumer groups
- **Concurrent Processing**: Configurable worker pool with semaphore-based concurrency
- **LightRAG Integration**: Automatic workspace isolation per collection_id
- **LRU Instance Pool**: Memory-efficient management of LightRAG instances
- **Status Reporting**: HTTP callbacks to Go backend for job status updates
- **Graceful Shutdown**: Clean resource cleanup on termination

## Project Structure

```
agents/data-processing/
├── app/
│   ├── __init__.py
│   ├── main.py           # Entry point
│   └── config.py         # Settings management
│
├── consumer/
│   ├── __init__.py
│   └── kafka_consumer.py # Kafka consumer implementation
│
├── processor/
│   ├── __init__.py
│   ├── pipeline.py       # Processing pipeline
│   └── html_cleaner.py   # HTML → text conversion
│
├── rag/
│   ├── __init__.py
│   └── pool.py           # LightRAG instance pool
│
├── api/
│   ├── __init__.py
│   └── status_client.py  # Go backend API client
│
├── models/
│   ├── __init__.py
│   └── messages.py       # Pydantic models
│
├── Dockerfile
├── Dockerfile.dev
├── requirements.txt
└── README.md
```

## Configuration

Environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka broker addresses | `localhost:9092` |
| `KAFKA_GROUP_ID` | Consumer group ID | `rag-processor` |
| `KAFKA_TOPIC` | Topic to consume | `indexing_jobs` |
| `MAX_WORKERS` | Concurrent processors | `5` |
| `MAX_RAG_INSTANCES` | Max LightRAG pools | `10` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `GO_BACKEND_URL` | Go backend URL | `http://localhost:8005` |

## Running

### With Docker Compose (Recommended)

From the project root:

```bash
docker compose up rag-agent
```

### Standalone

```bash
# Install dependencies
pip install -r requirements.txt

# Copy and configure environment
cp .env.example .env
# Edit .env with your settings

# Run
python -m app.main
```

## Kafka Message Format

The agent consumes messages from `indexing_jobs` topic:

```json
{
    "job_id": 12345,
    "request_id": 100,
    "project_id": 50,
    "user_id": "user_abc123",
    "collection_id": "550e8400-e29b-41d4-a716-446655440000",
    "url": "https://docs.example.com/getting-started",
    "depth": 2,
    "parent_url": "https://docs.example.com",
    "content": {
        "html": "<gzip+base64 encoded>",
        "text": "Pre-extracted text content...",
        "title": "Getting Started",
        "content_type": "text/html",
        "encoding": "gzip+base64"
    },
    "discovered_at": "2025-12-24T10:30:00Z",
    "metadata": {
        "base_url": "https://docs.example.com"
    }
}
```

## Scaling

To scale horizontally, simply run multiple instances:

```bash
# Docker Compose
docker compose up --scale rag-agent=3

# Or with different group members
KAFKA_GROUP_ID=rag-processor docker compose up rag-agent
```

Kafka will automatically balance partitions across consumers in the same group.

## Monitoring

The agent logs processing metrics:

- Jobs processed per second
- Processing time per job
- Active RAG instances
- Consumer lag

## Development

### Hot Reloading

The dev Dockerfile includes hot reloading via `watchfiles`:

```bash
docker compose up rag-agent
# Changes to Python files will auto-restart the agent
```

### Testing

```bash
# Run tests (TODO: Add test suite)
pytest tests/
```

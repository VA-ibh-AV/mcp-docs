# Python RAG Agent Design Document
## LightRAG-based Document Processor with Kafka Integration

**Version:** 1.0  
**Date:** December 25, 2025  
**Status:** ✅ Implemented

---

## Table of Contents

1. [Overview](#overview)
2. [Current State Analysis](#current-state-analysis)
3. [Go Backend Changes](#go-backend-changes)
4. [Python Agent Architecture](#python-agent-architecture)
5. [Kafka Message Contracts](#kafka-message-contracts)
6. [Scalability Design](#scalability-design)
7. [LightRAG Integration](#lightrag-integration)
8. [Error Handling & Retry Strategy](#error-handling--retry-strategy)
9. [Configuration Management](#configuration-management)
10. [Open Questions for Discussion](#open-questions-for-discussion)

---

## Overview

### Goal
Build a **scalable Python agent** that:
1. Consumes `indexing_jobs` messages from Kafka (produced by Go indexer)
2. Processes HTML content (clean, chunk)
3. Uses **LightRAG** to generate embeddings and store in PostgreSQL vector DB
4. Reports job completion status back to Go backend
5. Scales horizontally with multiple consumer instances

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Go Backend (Indexer)                               │
│                                                                             │
│   URL Discovery → Link Extraction → HTML Extraction → Kafka Producer        │
└───────────────────────────────────────┬─────────────────────────────────────┘
                                        │
                                        ▼
                              ┌─────────────────┐
                              │     Kafka       │
                              │  indexing_jobs  │
                              │  (Partitioned)  │
                              └────────┬────────┘
                                       │
              ┌────────────────────────┼────────────────────────┐
              │                        │                        │
              ▼                        ▼                        ▼
       ┌──────────────┐        ┌──────────────┐        ┌──────────────┐
       │ Python Agent │        │ Python Agent │        │ Python Agent │
       │  Instance 1  │        │  Instance 2  │        │  Instance N  │
       │              │        │              │        │              │
       │ ┌──────────┐ │        │ ┌──────────┐ │        │ ┌──────────┐ │
       │ │Consumer  │ │        │ │Consumer  │ │        │ │Consumer  │ │
       │ │Group     │ │        │ │Group     │ │        │ │Group     │ │
       │ └────┬─────┘ │        │ └────┬─────┘ │        │ └────┬─────┘ │
       │      │       │        │      │       │        │      │       │
       │ ┌────▼─────┐ │        │ ┌────▼─────┐ │        │ ┌────▼─────┐ │
       │ │Processor │ │        │ │Processor │ │        │ │Processor │ │
       │ │ Workers  │ │        │ │ Workers  │ │        │ │ Workers  │ │
       │ └────┬─────┘ │        │ └────┬─────┘ │        │ └────┬─────┘ │
       │      │       │        │      │       │        │      │       │
       │ ┌────▼─────┐ │        │ ┌────▼─────┐ │        │ ┌────▼─────┐ │
       │ │ LightRAG │ │        │ │ LightRAG │ │        │ │ LightRAG │ │
       │ └──────────┘ │        │ └──────────┘ │        │ └──────────┘ │
       └──────────────┘        └──────────────┘        └──────────────┘
              │                        │                        │
              └────────────────────────┼────────────────────────┘
                                       │
                                       ▼
                              ┌─────────────────┐
                              │   PostgreSQL    │
                              │  (pgvector +    │
                              │   LightRAG)     │
                              └─────────────────┘
```

---

## Current State Analysis

### Go Backend (Completed)
- ✅ Website crawler with Playwright
- ✅ Concurrent worker pool
- ✅ URL filtering and prioritization
- ✅ HTML content extraction
- ✅ Kafka producer for `indexing_jobs` topic
- ✅ Job tracking in PostgreSQL

### Python PoC (lite-rag-poc.py)
```python
rag = LightRAG(
    embedding_func=openai_embed,
    llm_model_func=gpt_4o_mini_complete,
    kv_storage="PGKVStorage",
    vector_storage="PGVectorStorage",
    graph_storage="PGGraphStorage",
    doc_status_storage="PGDocStatusStorage",
    workspace="unique_workspace_name_2",  # ← This needs to be dynamic!
)
```

**Key Observations:**
1. LightRAG uses PostgreSQL for all storage types
2. `workspace` parameter isolates data per collection
3. Each insert/query uses async methods

---

## Go Backend Changes

### 1. Add `collection_id` to IndexingRequest

**Rationale:** The `collection_id` (workspace in LightRAG terms) must be:
- Generated at request creation time
- Unique per indexing request (can be UUID)
- Passed to Python agent for LightRAG isolation
- Consistent across all jobs in a request

**Option A: UUID-based Collection ID (Recommended)**
```go
// models/indexing_request.go
type IndexingRequest struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    UserID        string    `gorm:"not null" json:"user_id"`
    CollectionID  string    `gorm:"type:varchar(36);not null;uniqueIndex" json:"collection_id"` // NEW
    Endpoint      string    `gorm:"type:varchar(255);not null" json:"endpoint"`
    TotalJobs     int       `gorm:"not null" json:"total_jobs"`
    CompletedJobs int       `gorm:"not null" json:"completed_jobs"`
    ProjectID     uint      `gorm:"not null" json:"project_id"`
    Status        string    `gorm:"type:varchar(50);not null" json:"status"`
    ErrorMsg      string    `gorm:"type:text" json:"error_msg,omitempty"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

**Option B: Project-based Collection ID (Alternative)**
```go
// Collection ID = project_id + request_id combination
CollectionID: fmt.Sprintf("project_%d_request_%d", req.ProjectID, request.ID)
```

**Discussion Points:**
- Q1: Should collection_id be tied to project or request?
  - Per-request: Each crawl creates new collection (version-like)
  - Per-project: Multiple crawls update same collection (might cause conflicts)
  - **Recommendation:** Per-request with UUID for isolation

- Q2: Should users be able to specify custom collection names?
  - Pro: User-friendly naming ("langchain-docs-v2")
  - Con: Name collision handling required
  - **Recommendation:** Auto-generate UUID, allow optional user alias

### 2. Update Kafka Message Contract

```go
// indexer/types.go - Add CollectionID to message
type IndexingJobMessage struct {
    JobID        uint   `json:"job_id"`
    RequestID    uint   `json:"request_id"`
    ProjectID    uint   `json:"project_id"`
    UserID       string `json:"user_id"`
    CollectionID string `json:"collection_id"` // NEW - LightRAG workspace
    URL          string `json:"url"`
    // ... rest unchanged
}
```

### 3. Update CrawlRequest

```go
// indexer/types.go
type CrawlRequest struct {
    RequestID    uint   `json:"request_id"`
    ProjectID    uint   `json:"project_id"`
    UserID       string `json:"user_id"`
    CollectionID string `json:"collection_id"` // NEW
    BaseURL      string `json:"base_url"`
    MaxPages     int    `json:"max_pages"`
    MaxDepth     int    `json:"max_depth"`
}
```

### 4. Update Indexing Consumer

```go
// consumers/indexing_consumer.go
type IndexingRequestPayload struct {
    ID           uint   `json:"id"`
    UserID       string `json:"user_id"`
    CollectionID string `json:"collection_id"` // NEW
    Endpoint     string `json:"endpoint"`
    ProjectID    uint   `json:"project_id"`
    Status       string `json:"status"`
}
```

### 5. Service Layer Changes

```go
// services/indexing_service.go
func (s *IndexingService) CreateIndexingRequest(...) (*models.IndexingRequest, error) {
    // Generate unique collection ID
    collectionID := uuid.New().String()
    
    request := &models.IndexingRequest{
        UserID:       userID,
        ProjectID:    req.ProjectID,
        CollectionID: collectionID, // NEW
        Endpoint:     req.Endpoint,
        Status:       "pending",
        // ...
    }
    // ...
}
```

---

## Python Agent Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Python Agent Process                          │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    Kafka Consumer Manager                      │  │
│  │                                                               │  │
│  │  - Consumer Group: "rag-processor"                            │  │
│  │  - Topic: "indexing_jobs"                                     │  │
│  │  - Partitions: Auto-assigned by Kafka                         │  │
│  │  - Commit Strategy: Manual commit after processing            │  │
│  └───────────────────────────┬───────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    Message Processor                           │  │
│  │                                                               │  │
│  │  - Async task queue (asyncio)                                 │  │
│  │  - Concurrent workers: configurable (default: 5)              │  │
│  │  - Backpressure: max in-flight messages                       │  │
│  └───────────────────────────┬───────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    Processing Pipeline                         │  │
│  │                                                               │  │
│  │  1. Decompress HTML (gzip+base64 → raw HTML)                  │  │
│  │  2. Clean HTML → Plain Text (BeautifulSoup)                   │  │
│  │  3. LightRAG Insert (embedding + graph + storage)             │  │
│  │  4. Update job status via API callback                        │  │
│  └───────────────────────────┬───────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    LightRAG Instance Pool                      │  │
│  │                                                               │  │
│  │  - Pool per collection_id (workspace)                         │  │
│  │  - Connection pooling to PostgreSQL                           │  │
│  │  - LRU eviction for inactive workspaces                       │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. Kafka Consumer (`consumer.py`)

```python
# Conceptual structure
class IndexingJobConsumer:
    def __init__(
        self,
        bootstrap_servers: list[str],
        group_id: str = "rag-processor",
        topic: str = "indexing_jobs",
        max_workers: int = 5,
    ):
        self.consumer = AIOKafkaConsumer(
            topic,
            bootstrap_servers=bootstrap_servers,
            group_id=group_id,
            enable_auto_commit=False,  # Manual commit after processing
            auto_offset_reset="earliest",
            value_deserializer=lambda m: json.loads(m.decode("utf-8")),
        )
        self.processor = MessageProcessor(max_workers)
        
    async def start(self):
        await self.consumer.start()
        await self.processor.start()
        
    async def consume(self):
        async for message in self.consumer:
            await self.processor.submit(message)
            await self.consumer.commit()  # Commit after successful submit
```

#### 2. Message Processor (`processor.py`)

```python
class MessageProcessor:
    def __init__(self, max_workers: int = 5):
        self.semaphore = asyncio.Semaphore(max_workers)
        self.rag_pool = LightRAGPool()
        self.status_client = StatusUpdateClient()
        
    async def submit(self, message: KafkaMessage):
        async with self.semaphore:  # Limit concurrency
            await self.process(message.value)
            
    async def process(self, job: dict):
        collection_id = job["collection_id"]
        job_id = job["job_id"]
        
        try:
            # 1. Decompress HTML
            html = decompress_html(job["content"]["html"], job["content"]["encoding"])
            
            # 2. Clean HTML to text (or use provided text)
            text = job["content"]["text"] or clean_html(html)
            
            # 3. Get or create LightRAG instance for this collection
            rag = await self.rag_pool.get(collection_id)
            
            # 4. Insert into LightRAG
            doc_id = f"{job['url']}"  # Use URL as document ID
            await rag.ainsert(doc_id, text)
            
            # 5. Report success
            await self.status_client.update_job_status(
                job_id=job_id,
                status="completed"
            )
            
        except Exception as e:
            await self.status_client.update_job_status(
                job_id=job_id,
                status="failed",
                error_msg=str(e)
            )
```

#### 3. LightRAG Instance Pool (`rag_pool.py`)

```python
class LightRAGPool:
    """
    Manages LightRAG instances per collection_id (workspace).
    Uses LRU eviction for inactive workspaces.
    """
    
    def __init__(self, max_instances: int = 10, ttl_seconds: int = 300):
        self.instances: OrderedDict[str, LightRAGWrapper] = OrderedDict()
        self.max_instances = max_instances
        self.ttl_seconds = ttl_seconds
        self.lock = asyncio.Lock()
        
    async def get(self, collection_id: str) -> LightRAG:
        async with self.lock:
            if collection_id in self.instances:
                # Move to end (most recently used)
                self.instances.move_to_end(collection_id)
                wrapper = self.instances[collection_id]
                wrapper.last_used = time.time()
                return wrapper.rag
            
            # Evict oldest if at capacity
            while len(self.instances) >= self.max_instances:
                oldest_id, oldest = self.instances.popitem(last=False)
                await oldest.rag.finalize_storages()
                
            # Create new instance
            rag = await self._create_instance(collection_id)
            self.instances[collection_id] = LightRAGWrapper(rag)
            return rag
            
    async def _create_instance(self, collection_id: str) -> LightRAG:
        rag = LightRAG(
            embedding_func=openai_embed,
            llm_model_func=gpt_4o_mini_complete,
            kv_storage="PGKVStorage",
            vector_storage="PGVectorStorage",
            graph_storage="PGGraphStorage",
            doc_status_storage="PGDocStatusStorage",
            workspace=collection_id,  # Use collection_id as workspace
        )
        await rag.initialize_storages()
        return rag
```

#### 4. Status Update Client (`status_client.py`)

```python
class StatusUpdateClient:
    """
    Reports job completion status back to Go backend.
    """
    
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session = aiohttp.ClientSession()
        
    async def update_job_status(
        self,
        job_id: int,
        status: str,
        error_msg: str = None
    ):
        """
        PATCH /api/v1/indexing/jobs/{job_id}/status
        """
        url = f"{self.base_url}/api/v1/indexing/jobs/{job_id}/status"
        payload = {"status": status}
        if error_msg:
            payload["error_msg"] = error_msg
            
        async with self.session.patch(url, json=payload) as resp:
            if resp.status != 200:
                logger.error(f"Failed to update job status: {await resp.text()}")
```

---

## Kafka Message Contracts

### Topic: `indexing_jobs`

**Message consumed by Python Agent:**

```json
{
    "job_id": 12345,
    "request_id": 100,
    "project_id": 50,
    "user_id": "user_abc123",
    "collection_id": "550e8400-e29b-41d4-a716-446655440000",  // NEW
    "url": "https://docs.example.com/guides/getting-started",
    "depth": 2,
    "parent_url": "https://docs.example.com/guides",
    "discovered_at": "2025-12-24T10:30:00Z",
    "content": {
        "html": "<gzip+base64 encoded>",
        "text": "Plain text extracted from body...",
        "title": "Getting Started Guide",
        "content_type": "text/html",
        "encoding": "gzip+base64"
    },
    "metadata": {
        "base_url": "https://docs.example.com",
        "crawl_session_id": "crawl_xyz789"
    }
}
```

### Topic: `indexing_job_results` (Optional - Alternative to API callback)

**If we want to use Kafka for status updates instead of HTTP:**

```json
{
    "job_id": 12345,
    "request_id": 100,
    "status": "completed",  // or "failed"
    "error_msg": null,
    "processed_at": "2025-12-24T10:31:00Z",
    "metrics": {
        "text_length": 5420,
        "chunks_created": 12,
        "embedding_time_ms": 350,
        "storage_time_ms": 120
    }
}
```

**Discussion Point:**
- Q: Use HTTP callback or Kafka topic for status updates?
  - HTTP: Simpler, direct update to DB
  - Kafka: Decoupled, better for high-volume status updates
  - **Recommendation:** Start with HTTP callback, add Kafka topic later if needed

---

## Scalability Design

### Horizontal Scaling Strategy

```
                              ┌─────────────────┐
                              │     Kafka       │
                              │  indexing_jobs  │
                              │                 │
                              │  Partition 0    │────────┐
                              │  Partition 1    │────────┼─────┐
                              │  Partition 2    │────────┼─────┼─────┐
                              │  Partition N    │────────┼─────┼─────┼─────┐
                              └─────────────────┘        │     │     │     │
                                                         │     │     │     │
Consumer Group: "rag-processor"                          │     │     │     │
                              ┌───────────────────────────┘     │     │     │
                              │        ┌───────────────────────┘     │     │
                              │        │         ┌──────────────────┘     │
                              │        │         │         ┌──────────────┘
                              ▼        ▼         ▼         ▼
                         ┌─────────────────────────────────────────┐
                         │        Python Agent Instances           │
                         │                                         │
                         │   Agent 1    Agent 2    Agent 3   ...   │
                         │   (P0, P1)   (P2)       (P3, PN)        │
                         │                                         │
                         │   * Kafka auto-balances partitions      │
                         │   * Each message processed exactly once │
                         └─────────────────────────────────────────┘
```

### Concurrency Model

**Per-Instance Concurrency:**
```python
# Each Python agent instance
MAX_WORKERS = 5  # Concurrent message processors per instance
MAX_RAG_INSTANCES = 10  # Max LightRAG pools in memory

# Example: 3 agent instances × 5 workers = 15 concurrent document processors
```

**Scaling Considerations:**

| Factor | Constraint | Solution |
|--------|------------|----------|
| OpenAI API Rate Limits | ~10k tokens/min per org | Batch requests, use pool |
| PostgreSQL Connections | Connection pool limit | Use pgBouncer or limit instances |
| Memory per LightRAG | ~100-200MB | LRU eviction, limit pool size |
| Kafka Partitions | Max parallelism = partitions | Create enough partitions upfront |

### Recommended Kafka Configuration

```yaml
# Topic: indexing_jobs
num.partitions: 12          # Allows up to 12 parallel consumers
replication.factor: 1       # For dev (use 3 for production)
retention.ms: 604800000     # 7 days retention
max.message.bytes: 10485760 # 10MB for large HTML
```

---

## LightRAG Integration

### Workspace Isolation

Each `collection_id` maps to a LightRAG `workspace`, which creates isolated:
- Vector storage tables
- KV storage tables  
- Graph storage
- Document status tracking

```python
# Workspace naming convention
workspace = collection_id  # e.g., "550e8400-e29b-41d4-a716-446655440000"

# LightRAG creates tables like:
# - lightrag_{workspace}_vectors
# - lightrag_{workspace}_kv
# - lightrag_{workspace}_graph
# - lightrag_{workspace}_doc_status
```

### Insert Strategy

**Option A: Insert per Document (Current)**
```python
await rag.ainsert(doc_id=url, content=text)
```
- Pro: Simple, immediate
- Con: Many small DB writes

**Option B: Batch Insert (Recommended for Performance)**
```python
# Accumulate documents, batch insert
batch = []
async for message in consumer:
    batch.append((message.url, message.text))
    if len(batch) >= BATCH_SIZE or time_since_last > BATCH_TIMEOUT:
        await rag.ainsert_batch(batch)
        batch = []
```
- Pro: Fewer DB round-trips, better embedding batching
- Con: Slightly delayed processing

**Discussion Point:**
- Q: Should we batch inserts or process individually?
  - Individual: Simpler error handling, immediate status updates
  - Batch: Better performance, but error handling is complex
  - **Recommendation:** Start with individual, optimize to batch later

### Document Metadata

```python
# Include metadata in LightRAG insert
await rag.ainsert(
    doc_id=url,
    content=text,
    metadata={
        "url": url,
        "title": title,
        "depth": depth,
        "parent_url": parent_url,
        "request_id": request_id,
        "indexed_at": datetime.utcnow().isoformat(),
    }
)
```

---

## Error Handling & Retry Strategy

### Error Categories

| Category | Example | Action |
|----------|---------|--------|
| Transient | OpenAI timeout, DB connection | Retry with backoff |
| Permanent | Invalid content, parsing error | Mark failed, skip |
| Rate Limit | OpenAI 429 | Exponential backoff |
| Infrastructure | Kafka disconnect | Auto-reconnect |

### Retry Configuration

```python
@dataclass
class RetryConfig:
    max_retries: int = 3
    initial_backoff: float = 1.0
    max_backoff: float = 30.0
    backoff_multiplier: float = 2.0
    retryable_exceptions: tuple = (
        asyncio.TimeoutError,
        ConnectionError,
        RateLimitError,
    )
```

### Dead Letter Queue (DLQ)

```python
# Messages that fail after max retries go to DLQ
async def handle_permanent_failure(message, error):
    await producer.send(
        "indexing_jobs_dlq",
        key=message.key,
        value={
            "original_message": message.value,
            "error": str(error),
            "failed_at": datetime.utcnow().isoformat(),
            "retry_count": message.retry_count,
        }
    )
```

---

## Configuration Management

### Environment Variables

```bash
# Kafka
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_GROUP_ID=rag-processor
KAFKA_TOPIC=indexing_jobs
KAFKA_AUTO_OFFSET_RESET=earliest

# Processing
MAX_WORKERS=5
MAX_RAG_INSTANCES=10
RAG_INSTANCE_TTL_SECONDS=300

# PostgreSQL (for LightRAG)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=mcpdocs

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_LLM_MODEL=gpt-4o-mini

# Go Backend (for status updates)
GO_BACKEND_URL=http://localhost:8005
```

### Configuration File (`config.py`)

```python
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # Kafka
    kafka_bootstrap_servers: str = "localhost:9092"
    kafka_group_id: str = "rag-processor"
    kafka_topic: str = "indexing_jobs"
    
    # Processing
    max_workers: int = 5
    max_rag_instances: int = 10
    rag_instance_ttl: int = 300
    
    # Retry
    max_retries: int = 3
    retry_backoff: float = 1.0
    
    class Config:
        env_file = ".env"
```

---

## Open Questions for Discussion

### 1. **Collection ID Strategy**
   - Q: UUID vs Project+Request composite key?
   - Q: Should users be able to name their collections?
   - **Recommendation:** UUID with optional user alias

### 2. **Status Update Mechanism**
   - Q: HTTP callback to Go backend or Kafka topic?
   - Q: Should we increment `completed_jobs` per job or batch?
   - **Recommendation:** HTTP callback, increment per job

### 3. **LightRAG Instance Lifecycle**
   - Q: Keep instances alive or create/destroy per batch?
   - Q: How many concurrent workspaces to support?
   - **Recommendation:** Pool with LRU eviction (max 10 instances)

### 4. **HTML Processing**
   - Q: Use text from Go (already extracted) or re-extract from HTML?
   - Q: Should Python do additional cleaning beyond Go?
   - **Recommendation:** Use pre-extracted text, add option for re-processing

### 5. **Batching Strategy**
   - Q: Process individually or batch by collection_id?
   - Q: What batch size and timeout?
   - **Recommendation:** Start individual, add batching as optimization

### 6. **Consumer Offset Commit Strategy**
   - Q: Commit after each message or batch commit?
   - Q: What happens if processing fails after commit?
   - **Recommendation:** Commit after successful processing, use DLQ for failures

### 7. **Monitoring & Observability**
   - Q: What metrics to expose? (Prometheus, OpenTelemetry?)
   - Q: How to track processing lag?
   - **Recommendation:** Prometheus metrics + structured logging

### 8. **Docker & Deployment**
   - Q: Separate container or same as Go backend?
   - Q: How to scale in Kubernetes/Docker Swarm?
   - **Recommendation:** Separate container, scale via replicas

### 9. **Testing Strategy**
   - Q: How to test without hitting real OpenAI?
   - Q: Integration tests with Kafka?
   - **Recommendation:** Mock embedding functions, use testcontainers for Kafka

### 10. **Rate Limiting for OpenAI**
   - Q: Global rate limit or per-collection?
   - Q: How to handle embedding bursts?
   - **Recommendation:** Global semaphore with configurable concurrency

---

## Proposed File Structure

```
agents/data-processing/
├── app/
│   ├── __init__.py
│   ├── main.py                 # Entry point
│   ├── config.py               # Settings management
│   └── dependencies.py         # Dependency injection
│
├── consumer/
│   ├── __init__.py
│   ├── kafka_consumer.py       # Kafka consumer implementation
│   └── message_handler.py      # Message routing
│
├── processor/
│   ├── __init__.py
│   ├── pipeline.py             # Processing pipeline
│   ├── html_cleaner.py         # HTML → text conversion
│   └── decompressor.py         # gzip+base64 decompression
│
├── rag/
│   ├── __init__.py
│   ├── pool.py                 # LightRAG instance pool
│   └── wrapper.py              # LightRAG wrapper utilities
│
├── api/
│   ├── __init__.py
│   └── status_client.py        # Go backend API client
│
├── models/
│   ├── __init__.py
│   └── messages.py             # Pydantic models for Kafka messages
│
├── utils/
│   ├── __init__.py
│   ├── retry.py                # Retry decorators
│   └── metrics.py              # Prometheus metrics
│
├── tests/
│   ├── __init__.py
│   ├── test_consumer.py
│   ├── test_processor.py
│   └── test_rag_pool.py
│
├── Dockerfile
├── docker-compose.yml          # For local development
├── requirements.txt
├── pyproject.toml
└── README.md
```

---

## Next Steps (After Discussion)

1. **Finalize Go backend changes** (collection_id field)
2. **Define final Kafka message schema**
3. **Implement MVP Python agent:**
   - Kafka consumer
   - Basic processor (no batching)
   - LightRAG integration
   - HTTP status callback
4. **Add Docker support**
5. **Integration testing**
6. **Monitoring & metrics**
7. **Performance optimization** (batching, pooling)

---

## Summary of Changes Needed

### Go Backend
1. Add `collection_id` field to `IndexingRequest` model
2. Generate UUID for `collection_id` in service layer
3. Include `collection_id` in Kafka message (`IndexingJobMessage`)
4. Update `CrawlRequest` to include `collection_id`
5. Update consumer payload struct

### Python Agent (New)
1. Create new structured Python project
2. Implement Kafka consumer with aiokafka
3. Create processing pipeline (decompress → clean → LightRAG insert)
4. Implement LightRAG instance pool
5. Add status update client for Go backend
6. Docker containerization
7. Configuration management

---

*Document ready for review and discussion.*

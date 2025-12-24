# Website Indexer Design Document
## Go Backend - Playwright-based Concurrent Web Scraper

**Version:** 1.0  
**Date:** December 24, 2025  
**Status:** Design Phase - Discussion

---

## Table of Contents

1. [Overview](#overview)
2. [Current Python Approach Analysis](#current-python-approach-analysis)
3. [Proposed Go Architecture](#proposed-go-architecture)
4. [Concurrency Model](#concurrency-model)
5. [Component Design](#component-design)
6. [Data Flow](#data-flow)
7. [Kafka Message Contracts](#kafka-message-contracts)
8. [Scalability Considerations](#scalability-considerations)
9. [Error Handling & Retry Strategy](#error-handling--retry-strategy)
10. [Open Questions for Discussion](#open-questions-for-discussion)

---

## Overview

### Goal
Build a **website indexer** in Go that:
1. Receives indexing requests via API
2. Crawls websites using Playwright (headless browser)
3. Discovers and filters relevant documentation URLs
4. Pushes individual page URLs to Kafka for Python agent processing
5. Tracks crawl progress and job status

### Scope Clarification
| Go Backend (Indexer) | Python Agent (Processor) |
|---------------------|--------------------------|
| URL discovery & crawling | Text cleaning |
| Link extraction | Chunking |
| **HTML content extraction** | Embedding generation |
| URL filtering | Vector DB storage |
| Depth management | |
| Kafka publishing (URL + HTML) | |
| Progress tracking | |

### Why Extract HTML in Go?

**Key Insight:** To extract links, we must load the page in Playwright and have DOM access. Discarding the HTML after link extraction and having Python re-visit the same page is wasteful.

```
❌ BAD: Double-visit (URL-only approach)
   Go visits page → extracts links → discards HTML
   Python visits SAME page → gets HTML → processes
   Result: 2x network requests per page!

✅ GOOD: Single-visit (URL + HTML approach)  
   Go visits page → extracts links + HTML → sends both to Kafka
   Python receives HTML → cleans → chunks → embeds
   Result: 1x network request per page!
```

**Benefits:**
- 50% reduction in crawl time
- Lower load on target websites (politeness)
- Reduced rate limiting risk
- More efficient resource usage

---

## Current Python Approach Analysis

### Lib Indexer Pipeline (`lib/indexer/`)

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  scrapper   │ → │   cleaner   │ → │   chunker   │ → │  embedder   │ → │  db_writer  │
│   .py       │   │    .py      │   │    .py      │   │    .py      │   │    .py      │
└─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘
```

#### Key Components:

**1. Scraper (`scrapper.py`)**
- Uses Playwright with Chromium (headless)
- BFS-style crawling with queue: `[(url, depth)]`
- Tracks visited URLs in a `set()`
- Configurable: `max_pages`, `max_depth`
- Returns: `[{url, html, text}]`
- Single browser context, sequential page visits

**2. URL Filter (`url_filter.py`)**
- Filters by domain match
- Blocks: assets, media, non-doc pages
- Forbidden keywords: pricing, blog, auth, etc.
- Called before adding to queue

**3. Cleaner (`cleaner.py`)**
- Simple whitespace normalization
- `re.sub(r"\s+", " ", text).strip()`

**4. Chunker (`chunker.py`)**
- Token-based chunking (500 tokens default)
- Simple word split approach

### Limitations of Current Approach
- **Single-threaded**: One page at a time
- **No persistence**: Queue lost on crash
- **Blocking**: Entire pipeline runs synchronously
- **Resource intensive**: Browser stays open for entire crawl

---

## Proposed Go Architecture

### High-Level Design

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           Go Backend                                        │
│                                                                            │
│   ┌──────────────────┐      ┌──────────────────┐      ┌────────────────┐  │
│   │   API Handler    │      │  Indexer Service │      │  Kafka Producer│  │
│   │                  │─────▶│                  │─────▶│                │  │
│   │ /indexing/start  │      │  - Coordinator   │      │  indexing_jobs │  │
│   └──────────────────┘      │  - Worker Pool   │      └───────┬────────┘  │
│                             │  - URL Frontier  │              │           │
│                             └──────────────────┘              │           │
│                                      │                        │           │
│                             ┌────────▼────────┐               │           │
│                             │ Browser Manager │               │           │
│                             │   (Playwright)  │               │           │
│                             │                 │               │           │
│                             │ ┌─────┐ ┌─────┐ │               │           │
│                             │ │Tab 1│ │Tab 2│ │               │           │
│                             │ └─────┘ └─────┘ │               │           │
│                             └─────────────────┘               │           │
└────────────────────────────────────────────────────────────────┼───────────┘
                                                                 │
                                                                 ▼
                                                        ┌────────────────┐
                                                        │     Kafka      │
                                                        │                │
                                                        │ indexing_jobs  │
                                                        └───────┬────────┘
                                                                │
                                                                ▼
                                                        ┌────────────────┐
                                                        │  Python Agent  │
                                                        │                │
                                                        │ - Scrape HTML  │
                                                        │ - Clean        │
                                                        │ - Chunk        │
                                                        │ - Embed        │
                                                        │ - Store        │
                                                        └────────────────┘
```

---

## Concurrency Model

### Option A: Worker Pool with Shared Browser (Recommended)

```go
// Conceptual structure
type CrawlCoordinator struct {
    browser      *playwright.Browser
    urlFrontier  *URLFrontier       // Thread-safe URL queue
    visited      *sync.Map          // Concurrent visited set
    workers      int                // Number of concurrent workers
    maxPages     int
    maxDepth     int
    results      chan CrawlResult
    kafkaProducer *kafka.Producer
}
```

**Flow:**
```
                    ┌─────────────────────────────────────┐
                    │         URL Frontier (Queue)        │
                    │  Thread-safe priority queue by      │
                    │  depth, with deduplication          │
                    └──────────────┬──────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
       ┌──────────┐         ┌──────────┐        ┌──────────┐
       │ Worker 1 │         │ Worker 2 │        │ Worker N │
       │          │         │          │        │          │
       │ Browser  │         │ Browser  │        │ Browser  │
       │ Tab/Page │         │ Tab/Page │        │ Tab/Page │
       └────┬─────┘         └────┬─────┘        └────┬─────┘
            │                    │                    │
            └────────────────────┼────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────────────────┐
                    │         Results Channel             │
                    │    {url, discovered_links, depth}   │
                    └──────────────┬──────────────────────┘
                                   │
                                   ▼
                    ┌─────────────────────────────────────┐
                    │         Kafka Publisher             │
                    │    Batch publish discovered URLs    │
                    └─────────────────────────────────────┘
```

### Option B: Multiple Browser Instances (Higher Resource Usage)

```
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│   Browser 1      │   │   Browser 2      │   │   Browser N      │
│   (Chromium)     │   │   (Chromium)     │   │   (Chromium)     │
│                  │   │                  │   │                  │
│  Worker Pool 1   │   │  Worker Pool 2   │   │  Worker Pool N   │
└──────────────────┘   └──────────────────┘   └──────────────────┘
```

### Comparison

| Aspect | Option A (Shared Browser) | Option B (Multi Browser) |
|--------|--------------------------|--------------------------|
| Memory | Lower (~200-400MB) | Higher (N × 200MB) |
| Startup | Faster (one browser) | Slower |
| Isolation | Less (shared context) | More (separate browsers) |
| Complexity | Simpler | More complex |
| Recommended for | Most cases | Very large crawls |

---

## Component Design

### 1. URL Frontier (Thread-Safe Queue)

```go
// URLFrontier manages the crawl queue with concurrency safety
type URLFrontier struct {
    queue       *PriorityQueue  // Ordered by depth (BFS)
    visited     *sync.Map       // URL -> bool
    inQueue     *sync.Map       // Prevent duplicates in queue
    mu          sync.RWMutex    // Protect queue operations
    domain      string          // Base domain for filtering
    baseURL     string          // Original starting URL
    
    // Limits
    maxDepth    int
    maxPages    int
    
    // Counters
    pagesQueued   atomic.Int64
    pagesVisited  atomic.Int64
}

// URLItem represents a URL in the frontier
type URLItem struct {
    URL       string
    Depth     int
    ParentURL string
    Priority  int  // Lower = higher priority (BFS: depth)
}

// Key Methods:
// - Push(url, depth, parentURL) error  // Add URL if valid & not visited
// - Pop() (*URLItem, bool)             // Get next URL to crawl
// - MarkVisited(url)                   // Mark URL as processed
// - Size() int                         // Current queue size
// - IsFinished() bool                  // Check if crawl complete
```

### 2. URL Filter

```go
// URLFilter determines if a URL should be crawled
type URLFilter struct {
    allowedDomain    string
    baseURL          string
    forbiddenExts    []string  // png, jpg, css, js, etc.
    forbiddenKeywords []string // pricing, blog, auth, etc.
}

// Key Methods:
// - IsRelevant(url string) bool
// - NormalizeURL(url string) string  // Remove fragments, normalize
// - IsSameDomain(url string) bool
```

**Forbidden Extensions:**
```go
var ForbiddenExtensions = []string{
    "png", "jpg", "jpeg", "gif", "svg", "webp",
    "css", "js", "ico", "map",
    "pdf", "zip", "tar", "gz", "7z", "rar",
    "mp4", "mp3", "wav", "avi", "mov",
    "woff", "woff2", "ttf", "eot",
}
```

**Forbidden Keywords:**
```go
var ForbiddenKeywords = []string{
    // Commercial
    "pricing", "plans", "enterprise", "customers", "case-study",
    // Company
    "about", "company", "team", "careers", "jobs", "hiring",
    // Legal
    "security", "legal", "privacy", "terms", "gdpr",
    // Content
    "blog", "newsletter", "community", "event", "news",
    "press", "changelog", "roadmap",
    // Commerce
    "store", "shop", "buy", "subscribe",
    // Support
    "contact", "support", "help-center", "faq",
    // External
    "github.com", "gitlab.com", "bitbucket.org",
    // Auth
    "login", "logout", "signin", "signup", "register", "auth",
    // Tracking
    "utm_", "ref=", "source=",
}
```

### 3. Browser Manager

```go
// BrowserManager handles Playwright browser lifecycle
type BrowserManager struct {
    pw        *playwright.Playwright
    browser   playwright.Browser
    contexts  []*BrowserContext  // Pool of contexts
    
    // Configuration
    headless    bool
    userAgent   string
    viewport    Viewport
    timeout     time.Duration
}

// BrowserContext wraps a browser context with a page pool
type BrowserContext struct {
    ctx       playwright.BrowserContext
    pages     chan playwright.Page  // Page pool
    maxPages  int
}

// Key Methods:
// - NewPage() (playwright.Page, error)      // Get a page from pool
// - ReleasePage(page)                       // Return page to pool
// - ExtractLinks(page) ([]string, error)   // Get all href links
// - Close()                                 // Cleanup
```

### 4. Crawl Worker

```go
// CrawlWorker processes URLs from the frontier
type CrawlWorker struct {
    id            int
    frontier      *URLFrontier
    browserMgr    *BrowserManager
    filter        *URLFilter
    results       chan<- CrawlResult
    done          <-chan struct{}
    
    // Stats
    pagesProcessed atomic.Int64
    errors         atomic.Int64
}

// CrawlResult represents the outcome of crawling a URL
type CrawlResult struct {
    URL             string
    DiscoveredLinks []string
    Depth           int
    
    // Content (extracted in same visit)
    HTML            string    // Raw HTML content
    Text            string    // Extracted text from body
    Title           string    // Page title
    
    // Metadata
    Success         bool
    Error           error
    ProcessedAt     time.Time
    ResponseTimeMs  int64
    StatusCode      int
    ContentType     string
}

// Key Methods:
// - Run()                              // Main worker loop
// - processURL(item) (*CrawlResult, error)  // Crawl single URL
// - extractContent(page) (html, text, title, error)
// - extractAndFilterLinks(page) []string
```

### 5. Crawl Coordinator

```go
// CrawlCoordinator orchestrates the entire crawl operation
type CrawlCoordinator struct {
    // Dependencies
    frontier      *URLFrontier
    browserMgr    *BrowserManager
    filter        *URLFilter
    kafkaProducer *kafka.Producer
    jobRepo       repository.IndexingJobRepository
    requestRepo   repository.IndexingRequestRepository
    
    // Configuration
    workerCount   int
    batchSize     int           // Kafka batch size
    flushInterval time.Duration // Kafka flush interval
    
    // State
    requestID     uint
    projectID     uint
    results       chan CrawlResult
    workers       []*CrawlWorker
    wg            sync.WaitGroup
    ctx           context.Context
    cancel        context.CancelFunc
}

// Key Methods:
// - Start(baseURL string) error
// - Stop()
// - Wait() error
// - GetProgress() CrawlProgress
```

---

## Data Flow

### Complete Flow Sequence

```
┌─────────┐     ┌───────────┐     ┌──────────────┐     ┌─────────────┐
│   API   │     │  Indexing │     │    Crawl     │     │   Kafka     │
│ Handler │────▶│  Service  │────▶│ Coordinator  │────▶│  Producer   │
└─────────┘     └───────────┘     └──────────────┘     └─────────────┘
     │                                   │                    │
     │                                   │                    │
     ▼                                   ▼                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Detailed Flow                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. API receives POST /indexing/request                             │
│     └─▶ Creates IndexingRequest (status: "pending")                 │
│                                                                     │
│  2. IndexingService.StartCrawl(requestID, baseURL)                  │
│     └─▶ Update status to "crawling"                                 │
│     └─▶ Initialize CrawlCoordinator                                 │
│                                                                     │
│  3. CrawlCoordinator.Start()                                        │
│     └─▶ Add baseURL to URLFrontier (depth: 0)                       │
│     └─▶ Start N CrawlWorkers                                        │
│     └─▶ Start ResultProcessor goroutine                             │
│                                                                     │
│  4. CrawlWorker.Run() [N concurrent]                                │
│     └─▶ Pop URL from Frontier                                       │
│     └─▶ Navigate with Playwright                                    │
│     └─▶ Extract links from page                                     │
│     └─▶ Filter links (domain, keywords, extensions)                 │
│     └─▶ Send CrawlResult to results channel                         │
│                                                                     │
│  5. ResultProcessor (single goroutine)                              │
│     └─▶ Receive CrawlResult                                         │
│     └─▶ Add discovered links to Frontier                            │
│     └─▶ Create IndexingJob in DB (status: "pending")                │
│     └─▶ Batch publish to Kafka (topic: indexing_jobs)               │
│                                                                     │
│  6. On completion (Frontier empty + workers idle)                   │
│     └─▶ Update IndexingRequest (status: "crawl_complete")           │
│     └─▶ Set total_jobs count                                        │
│     └─▶ Cleanup browser resources                                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Kafka Message Contracts

### Topic: `indexing_jobs`

**Message Key:** Domain (for partitioning)
```
"docs.example.com"
```

**Message Value (with HTML - Recommended):**
```json
{
    "job_id": 12345,
    "request_id": 100,
    "project_id": 50,
    "user_id": "user_abc123",
    "url": "https://docs.example.com/guides/getting-started",
    "depth": 2,
    "parent_url": "https://docs.example.com/guides",
    "discovered_at": "2025-12-24T10:30:00Z",
    "content": {
        "html": "<compressed base64 gzip content>",
        "text": "Page text content extracted from body...",
        "title": "Getting Started Guide",
        "content_type": "text/html",
        "encoding": "gzip+base64"
    },
    "metadata": {
        "base_url": "https://docs.example.com",
        "max_depth": 3,
        "crawl_session_id": "crawl_xyz789",
        "response_time_ms": 1250,
        "status_code": 200
    }
}
```

**Handling Large HTML:**
```go
// Option 1: Compress HTML (recommended for <1MB)
compressedHTML := gzipCompress(rawHTML)  // 70-90% reduction
base64HTML := base64.StdEncoding.EncodeToString(compressedHTML)

// Option 2: Store in object storage (for very large pages)
// Store HTML in S3/MinIO, send reference
{
    "content": {
        "storage_type": "s3",
        "storage_key": "crawls/request_100/job_12345.html.gz",
        "html_size_bytes": 524288
    }
}
```

**Message Size Considerations:**
| Content Size | Strategy |
|--------------|----------|
| < 500KB HTML | Gzip compress, embed in message |
| 500KB - 2MB | Gzip compress, increase Kafka max.message.bytes |
| > 2MB | Store in S3/MinIO, send reference |
```

### Topic: `indexing_status` (Status Updates)

**Message Value:**
```json
{
    "request_id": 100,
    "event_type": "url_discovered|crawl_progress|crawl_complete|crawl_error",
    "timestamp": "2025-12-24T10:30:00Z",
    "data": {
        "urls_discovered": 150,
        "urls_queued": 120,
        "urls_processed": 50,
        "current_depth": 2,
        "error_message": null
    }
}
```

---

## Scalability Considerations

### 1. Horizontal Scaling

```
                              Load Balancer
                                    │
              ┌─────────────────────┼─────────────────────┐
              │                     │                     │
              ▼                     ▼                     ▼
       ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
       │ Go Backend 1 │     │ Go Backend 2 │     │ Go Backend 3 │
       │              │     │              │     │              │
       │ Crawl Coord  │     │ Crawl Coord  │     │ Crawl Coord  │
       │ Worker Pool  │     │ Worker Pool  │     │ Worker Pool  │
       └──────────────┘     └──────────────┘     └──────────────┘
              │                     │                     │
              └─────────────────────┼─────────────────────┘
                                    │
                                    ▼
                           ┌────────────────┐
                           │     Kafka      │
                           │  (Partitioned) │
                           └────────────────┘
```

**Key Strategy:** Each Go instance handles independent crawl requests. No shared state between instances.

### 2. Rate Limiting & Politeness

```go
type RateLimiter struct {
    domainLimits  map[string]*rate.Limiter
    defaultRate   rate.Limit  // e.g., 2 requests/second
    mu            sync.RWMutex
}

// Configurable per-domain delays
// Respect robots.txt (future enhancement)
// Exponential backoff on errors
```

### 3. Resource Management

| Resource | Limit | Strategy |
|----------|-------|----------|
| Browser instances | 1-3 per Go instance | Pool |
| Pages per browser | 5-10 concurrent | Channel pool |
| Memory per browser | ~200-400MB | Monitor & restart |
| Goroutines | 10-50 workers | Configurable |
| Kafka batch size | 100-500 messages | Time + size based flush |

### 4. Backpressure Handling

```go
// If URL Frontier grows too large
if frontier.Size() > MaxQueueSize {
    // Option 1: Reduce max_depth dynamically
    // Option 2: Pause adding new URLs
    // Option 3: Persist queue to Redis/DB
}

// If Kafka is slow
// Use buffered channels with timeout
// Implement circuit breaker pattern
```

---

## Error Handling & Retry Strategy

### Error Categories

| Category | Example | Action |
|----------|---------|--------|
| Transient | Timeout, 503 | Retry with backoff |
| Permanent | 404, 401 | Skip, log error |
| Browser | Crash, OOM | Restart browser |
| Network | DNS failure | Retry with backoff |
| Filter | Invalid URL | Skip silently |

### Retry Configuration

```go
type RetryConfig struct {
    MaxRetries     int           // 3
    InitialBackoff time.Duration // 1s
    MaxBackoff     time.Duration // 30s
    BackoffFactor  float64       // 2.0
    RetryableStatus []int        // [429, 500, 502, 503, 504]
}
```

### Circuit Breaker (for external sites)

```go
type CircuitBreaker struct {
    domain         string
    failureCount   atomic.Int64
    threshold      int           // 5 failures
    resetTimeout   time.Duration // 60s
    state          atomic.Int32  // 0=closed, 1=open, 2=half-open
}
```

---

## Open Questions for Discussion

### 1. **Browser Management Strategy**
   - Q: Use single shared browser or multiple isolated instances?
   - Q: How many concurrent pages per browser context?
   - Recommendation: Start with single browser, 5 concurrent pages

### 2. **Playwright Go Library**
   - Q: Use `playwright-go` or shell out to Node.js Playwright?
   - Pro `playwright-go`: Native Go, simpler deployment
   - Pro Node.js: More mature, better documentation
   - Recommendation: `playwright-go` for simplicity

### 3. **URL Queue Persistence**
   - Q: In-memory only or persist to Redis/DB?
   - For small crawls (<10k URLs): In-memory OK
   - For large crawls: Consider Redis sorted set
   - Recommendation: Start in-memory, add Redis later

### 4. **Depth vs. Breadth Priority**
   - Q: Pure BFS or prioritize certain URL patterns?
   - Option: Boost priority for /docs, /api, /reference paths
   - Recommendation: BFS with optional priority boosting

### 5. **Content Extraction in Go vs Python** ✅ DECIDED
   - **Decision:** Go extracts BOTH links AND HTML content
   - **Rationale:** Page is already loaded for link extraction - discarding HTML is wasteful
   - **Python only needs to:** Clean → Chunk → Embed → Store
   - **Trade-off:** Larger Kafka messages (mitigated with gzip compression)

### 6. **Kafka Batching Strategy**
   - Q: Batch by count, time, or both?
   - Recommendation: Both - flush every 100 URLs or 5 seconds

### 7. **Progress Reporting**
   - Q: Real-time WebSocket updates or polling?
   - Current: Polling via API
   - Future: WebSocket for real-time
   - Recommendation: Start with polling, add WebSocket later

### 8. **robots.txt Compliance**
   - Q: Respect robots.txt or ignore (with rate limiting)?
   - For documentation sites: Usually ignore (docs are meant to be read)
   - Recommendation: Configurable, default to respecting

### 9. **Handling SPAs**
   - Q: How long to wait for JS rendering?
   - Current Python: `wait_until="networkidle"` + 2s delay
   - Recommendation: Same approach, configurable timeout

### 10. **Failure Recovery**
   - Q: Resume interrupted crawls?
   - Requires: Persistent URL frontier
   - Recommendation: Phase 2 feature, start without resume

---

## Next Steps (After Discussion)

1. **Finalize architecture decisions** from open questions
2. **Define package structure** for Go indexer
3. **Implement MVP** with:
   - Single browser, 3 workers
   - In-memory URL frontier
   - Basic URL filtering
   - Kafka publishing
4. **Add observability** (metrics, logging)
5. **Integration testing** with Python agent
6. **Performance tuning** based on real usage

---

## Appendix: Package Structure (Proposed)

```
go-backend/
├── indexer/
│   ├── coordinator.go      # CrawlCoordinator
│   ├── worker.go           # CrawlWorker
│   ├── frontier.go         # URLFrontier
│   ├── filter.go           # URLFilter
│   ├── browser.go          # BrowserManager
│   ├── config.go           # IndexerConfig
│   ├── types.go            # Shared types
│   └── indexer_test.go     # Tests
├── kafka/
│   ├── producer.go         # (existing)
│   ├── consumer.go         # (existing)
│   └── messages.go         # Kafka message types
├── services/
│   └── indexing_service.go # (extend existing)
└── handlers/
    └── indexing.go         # (extend existing)
```

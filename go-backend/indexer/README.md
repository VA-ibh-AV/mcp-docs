# Website Indexer

A high-performance, concurrent website crawler built in Go using Playwright for JavaScript-rendered pages. Designed to discover and extract content from documentation websites, pushing results to Kafka for downstream processing.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Programmatic Usage](#programmatic-usage)
- [Configuration](#configuration)
- [URL Filtering](#url-filtering)
- [Kafka Integration](#kafka-integration)
- [Sample Consumer](#sample-consumer)
- [Troubleshooting](#troubleshooting)

---

## Features

- **Concurrent Crawling**: Worker pool with configurable concurrency
- **JavaScript Support**: Uses Playwright/Chromium for JS-rendered pages
- **Rate Limiting**: Per-domain token bucket rate limiter
- **Smart URL Filtering**: 70+ forbidden keywords, 30+ blocked extensions
- **BFS Crawling**: Breadth-first search with depth limits
- **HTML Compression**: Gzip + Base64 encoding (70-90% size reduction)
- **Kafka Integration**: Publishes discovered URLs with content to Kafka
- **Graceful Shutdown**: Proper cleanup on SIGINT/SIGTERM
- **Progress Tracking**: Real-time stats on crawl progress

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Coordinator                               │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ URL Frontier│  │Rate Limiter │  │    Browser Manager      │  │
│  │ (BFS Queue) │  │ (Per Domain)│  │    (Page Pool)          │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         │                │                     │                │
│         └────────────────┼─────────────────────┘                │
│                          ▼                                      │
│         ┌────────────────────────────────────────────┐          │
│         │           Worker Pool (N workers)           │          │
│         │    ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  │          │
│         │    │ W1   │  │ W2   │  │ W3   │  │ WN   │  │          │
│         │    └──────┘  └──────┘  └──────┘  └──────┘  │          │
│         └────────────────────────────────────────────┘          │
│                          │                                      │
│                          ▼                                      │
│         ┌────────────────────────────────────────────┐          │
│         │            Results Channel                  │          │
│         │      {URL, HTML, Text, Links}              │          │
│         └────────────────────────────────────────────┘          │
│                          │                                      │
│                          ▼                                      │
│         ┌────────────────────────────────────────────┐          │
│         │           Kafka Publisher                   │          │
│         │      Compress HTML → Send Message          │          │
│         └────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Components

| Component | File | Description |
|-----------|------|-------------|
| **Coordinator** | `coordinator.go` | Orchestrates the crawl, manages workers, publishes to Kafka |
| **Worker** | `worker.go` | Concurrent crawl workers that process URLs |
| **URL Frontier** | `frontier.go` | Thread-safe priority queue (BFS) for URLs |
| **Browser Manager** | `browser.go` | Playwright browser lifecycle and page pooling |
| **URL Filter** | `filter.go` | Filters irrelevant URLs (assets, non-docs) |
| **Rate Limiter** | `rate_limiter.go` | Per-domain token bucket rate limiter |
| **Config** | `config.go` | Configuration with sensible defaults |
| **Types** | `types.go` | Core data structures |
| **Utils** | `utils.go` | Compression, serialization utilities |

---

## Installation

### Prerequisites

- Go 1.21+
- Chromium browser (installed via Playwright)

### Install Dependencies

```bash
cd go-backend

# Add playwright-go dependency
go get github.com/playwright-community/playwright-go@latest

# Update vendor (if using vendor mode)
go mod vendor

# Install Chromium browser for Playwright
go run -mod=mod github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

### Build CLI Tools

```bash
# Build the indexer CLI
go build -o bin/indexer-cli cmd/indexer_cli/main.go

# Build the sample consumer
go build -o bin/sample-consumer cmd/sample_consumer/main.go
```

---

## Quick Start

### Dry Run (Without Kafka)

Test the indexer without Kafka - useful for development and testing:

```bash
# Basic crawl (10 pages, 3 depth)
go run cmd/indexer_cli/main.go -url "https://docs.example.com" -no-kafka

# Custom limits
go run cmd/indexer_cli/main.go \
  -url "https://docs.example.com" \
  -max-pages 50 \
  -max-depth 5 \
  -concurrency 5 \
  -no-kafka
```

### With Kafka

```bash
# Start Kafka (if using docker-compose)
docker compose up -d kafka zookeeper

# Run indexer with Kafka
go run cmd/indexer_cli/main.go \
  -url "https://docs.example.com" \
  -max-pages 100 \
  -brokers "localhost:9092" \
  -topic "indexing_jobs"

# In another terminal, run the sample consumer
go run cmd/sample_consumer/main.go -brokers "localhost:9092" -topic "indexing_jobs"
```

---

## CLI Usage

### Indexer CLI

```
Usage: indexer_cli [options]

Options:
  -url string
        URL to crawl (required)
  -max-pages int
        Maximum pages to crawl (default 10)
  -max-depth int
        Maximum crawl depth (default 3)
  -concurrency int
        Number of concurrent workers (default 3)
  -rps float
        Requests per second rate limit (default 2.0)
  -brokers string
        Kafka broker addresses (default "localhost:9092")
  -topic string
        Kafka topic for publishing jobs (default "indexing_jobs")
  -no-kafka
        Run without Kafka (dry run mode)
```

### Examples

```bash
# Minimal crawl
go run cmd/indexer_cli/main.go -url "https://docs.python.org" -max-pages 5 -no-kafka

# Large crawl with high concurrency
go run cmd/indexer_cli/main.go \
  -url "https://docs.langchain.com" \
  -max-pages 500 \
  -max-depth 10 \
  -concurrency 10 \
  -rps 5.0 \
  -no-kafka

# Production crawl with Kafka
go run cmd/indexer_cli/main.go \
  -url "https://docs.langchain.com" \
  -max-pages 1000 \
  -max-depth 10 \
  -concurrency 5 \
  -rps 2.0 \
  -brokers "kafka:9092" \
  -topic "indexing_jobs"
```

### Sample Consumer CLI

```
Usage: sample_consumer [options]

Options:
  -brokers string
        Kafka broker addresses (default "localhost:9092")
  -topic string
        Kafka topic to consume (default "indexing_jobs")
  -group string
        Consumer group ID (default "indexer-test-consumer")
```

---

## Programmatic Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "mcpdocs/indexer"
    "mcpdocs/kafka"
)

func main() {
    // Create config with defaults
    config := indexer.DefaultConfig()
    
    // Or customize with options
    config := indexer.NewConfig(
        indexer.WithMaxPages(100),
        indexer.WithMaxDepth(5),
        indexer.WithConcurrency(5),
        indexer.WithRateLimit(2.0),
    )
    
    // Create Kafka producer (optional)
    producer, err := kafka.NewProducer([]string{"localhost:9092"})
    if err != nil {
        log.Fatal(err)
    }
    defer producer.Close()
    
    // Create coordinator
    coordinator := indexer.NewCoordinator(config, producer)
    
    // Define crawl request
    request := &indexer.CrawlRequest{
        RequestID: 1,
        ProjectID: 1,
        UserID:    "user-123",
        BaseURL:   "https://docs.example.com",
        MaxPages:  100,
        MaxDepth:  5,
    }
    
    // Start crawling
    ctx := context.Background()
    if err := coordinator.Start(ctx, request); err != nil {
        log.Fatal(err)
    }
    
    // Wait for completion
    coordinator.Wait()
    
    // Get final stats
    stats := coordinator.GetStats()
    log.Printf("Crawled %d URLs in %v", stats.TotalURLsCrawled, stats.CrawlDuration)
}
```

### Without Kafka (Dry Run)

```go
// Pass nil as producer for dry run
coordinator := indexer.NewCoordinator(config, nil)
```

### With Job Callback

```go
// Set callback to be notified when jobs are created
coordinator.SetJobCreatedCallback(func(jobID uint, url string) {
    log.Printf("Job %d created for URL: %s", jobID, url)
    // Save to database, etc.
})
```

### Graceful Shutdown

```go
// Handle Ctrl+C
ctx, cancel := context.WithCancel(context.Background())
signals := make(chan os.Signal, 1)
signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-signals
    cancel()
    coordinator.Stop()
}()

coordinator.Start(ctx, request)
coordinator.Wait()
```

---

## Configuration

### Default Configuration

```go
config := indexer.DefaultConfig()

// Defaults:
// - Headless: true
// - UserAgent: "Mozilla/5.0 ... MCPDocsBot/1.0"
// - MaxPages: 100
// - MaxDepth: 5
// - MaxConcurrency: 5
// - PageTimeout: 30s
// - NetworkTimeout: 30s
// - IdleTimeout: 2s (wait for JS rendering)
// - RequestsPerSecond: 2.0
// - KafkaTopic: "indexing_jobs"
// - KafkaBatchSize: 50
// - CompressHTML: true
// - MaxHTMLSize: 5MB
// - MaxTextLength: 1MB
```

### Configuration Options

```go
config := indexer.NewConfig(
    indexer.WithMaxPages(500),        // Max pages to crawl
    indexer.WithMaxDepth(10),         // Max depth (BFS levels)
    indexer.WithConcurrency(10),      // Concurrent workers
    indexer.WithRateLimit(5.0),       // Requests per second
    indexer.WithKafkaTopic("my-topic"), // Kafka topic
)
```

### Manual Configuration

```go
config := &indexer.Config{
    Headless:          true,
    UserAgent:         "MyBot/1.0",
    MaxPages:          1000,
    MaxDepth:          10,
    MaxConcurrency:    10,
    PageTimeout:       60 * time.Second,
    RequestsPerSecond: 5.0,
    CompressHTML:      true,
}
```

---

## URL Filtering

The indexer automatically filters out irrelevant URLs.

### Blocked Extensions (30+)

```
Images:     .png, .jpg, .jpeg, .gif, .svg, .webp, .ico
Styles:     .css, .scss, .sass, .less
Scripts:    .js, .mjs, .map
Documents:  .pdf, .doc, .docx, .xls, .xlsx, .ppt
Archives:   .zip, .tar, .gz, .7z, .rar
Media:      .mp4, .mp3, .wav, .avi, .mov, .webm
Fonts:      .woff, .woff2, .ttf, .eot, .otf
Data:       .json, .xml, .csv, .yaml, .yml
```

### Blocked Keywords (70+)

```
Commercial:  pricing, plans, enterprise, customers, demo
Company:     about, team, careers, jobs, hiring
Legal:       privacy, terms, gdpr, legal, security
Blog:        blog, news, newsletter, changelog, roadmap
Commerce:    store, shop, buy, subscribe, cart
Support:     contact, support, faq, help-center
Auth:        login, logout, signin, signup, register
External:    github.com, gitlab.com, bitbucket.org
Tracking:    utm_source, utm_medium, ref=, source=
```

### Custom Filtering

```go
// The filter is created automatically from base URL
filter, _ := indexer.NewURLFilter("https://docs.example.com")

// Check if URL should be crawled
if filter.IsRelevant("https://docs.example.com/api/guide") {
    // Will be crawled
}

if filter.IsRelevant("https://docs.example.com/pricing") {
    // Won't be crawled (blocked keyword)
}
```

---

## Kafka Integration

### Message Format

Each crawled URL produces a Kafka message:

```json
{
  "job_id": 123,
  "request_id": 1,
  "project_id": 1,
  "user_id": "user-123",
  "url": "https://docs.example.com/guides/intro",
  "depth": 2,
  "parent_url": "https://docs.example.com/guides",
  "discovered_at": "2025-12-24T15:30:00Z",
  "content": {
    "html": "<base64 gzip compressed HTML>",
    "text": "Plain text content extracted from body...",
    "title": "Introduction Guide",
    "content_type": "text/html",
    "encoding": "gzip+base64",
    "html_size": 125000
  },
  "metadata": {
    "base_url": "https://docs.example.com",
    "crawl_session_id": "crawl_1"
  }
}
```

### Message Key

Messages are keyed by domain for partitioning:
```
Key: "docs.example.com"
```

### Decompressing HTML

```go
// In Python consumer
import base64
import gzip

def decompress_html(compressed: str) -> str:
    data = base64.b64decode(compressed)
    return gzip.decompress(data).decode('utf-8')
```

```go
// In Go
html, err := indexer.DecompressHTML(message.Content.HTML)
```

---

## Sample Consumer

The sample consumer shows how to process Kafka messages:

```bash
# Run consumer
go run cmd/sample_consumer/main.go -brokers "localhost:9092" -topic "indexing_jobs"
```

Output:
```
============================================================
📥 Received message:
   Topic: indexing_jobs
   Partition: 0
   Offset: 42
   Key: docs.example.com

📄 Job Details:
   Job ID: 5
   Request ID: 1
   Project ID: 1
   URL: https://docs.example.com/guides/intro
   Depth: 2

📝 Content:
   Title: Introduction Guide
   Encoding: gzip+base64
   HTML Size: 125000 bytes
   Text Length: 45000 chars
   Text Preview: Welcome to the documentation...

📊 Metadata:
   base_url: https://docs.example.com
   crawl_session_id: crawl_1
============================================================
```

---

## Troubleshooting

### Common Issues

#### 1. "playwright: timeout" errors

Pages taking too long to load (>30s). Solutions:
- Increase timeout: Modify `PageTimeout` in config
- These are usually heavy SPAs - the crawler handles gracefully and continues

#### 2. "failed to acquire page" 

All browser pages are busy. Solutions:
- Increase concurrency for more pages
- Wait for current pages to complete

#### 3. Only 1 URL crawled

The first page is still loading when idle check triggers. This was fixed in the latest version with in-flight tracking.

#### 4. Rate limiting errors

Too many requests. Solutions:
- Lower `RequestsPerSecond` (default: 2.0)
- Increase `RequestDelay`

#### 5. Memory issues

Browser consuming too much RAM. Solutions:
- Reduce `MaxConcurrency`
- Reduce `MaxPages`
- Restart between large crawls

### Debug Mode

Add logging to see what's happening:

```go
// Check frontier stats
stats := coordinator.GetProgress()
log.Printf("Queued: %d, In-Flight: %d, Processed: %d", 
    stats.URLsQueued, stats.URLsInFlight, stats.URLsProcessed)

// Check filter decisions
filter, _ := indexer.NewURLFilter(baseURL)
for _, url := range urls {
    if !filter.IsRelevant(url) {
        log.Printf("Filtered: %s", url)
    }
}
```

### Performance Tuning

| Scenario | Concurrency | RPS | Max Pages |
|----------|-------------|-----|-----------|
| Testing | 2-3 | 1.0 | 10-20 |
| Normal | 5 | 2.0 | 100-500 |
| Aggressive | 10 | 5.0 | 1000+ |
| Polite | 2 | 0.5 | Any |

---

## License

MIT License - See LICENSE file for details.

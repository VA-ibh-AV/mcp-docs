package indexer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"mcpdocs/kafka"
)

// Coordinator orchestrates the entire crawl operation
type Coordinator struct {
	// Configuration
	config  *Config
	request *CrawlRequest

	// Components
	frontier    *URLFrontier
	browserMgr  *BrowserManager
	filter      *URLFilter
	rateLimiter *RateLimiter
	producer    *kafka.Producer

	// Workers
	workers []*CrawlWorker
	results chan CrawlResult

	// State
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	done      chan struct{}
	startTime time.Time

	// Stats
	jobsCreated atomic.Int64
	jobsSent    atomic.Int64

	// Callbacks
	onJobCreated func(jobID uint, url string) // Called when a job is created in DB
}

// NewCoordinator creates a new crawl coordinator
func NewCoordinator(config *Config, producer *kafka.Producer) *Coordinator {
	return &Coordinator{
		config:   config,
		producer: producer,
		results:  make(chan CrawlResult, config.MaxConcurrency*2),
		done:     make(chan struct{}),
	}
}

// SetJobCreatedCallback sets a callback for when jobs are created
func (c *Coordinator) SetJobCreatedCallback(callback func(jobID uint, url string)) {
	c.onJobCreated = callback
}

// Start begins the crawl operation
func (c *Coordinator) Start(ctx context.Context, request *CrawlRequest) error {
	c.request = request
	c.startTime = time.Now()

	// Create context with cancellation
	c.ctx, c.cancel = context.WithCancel(ctx)

	log.Printf("Starting crawl for %s (max_pages: %d, max_depth: %d, concurrency: %d)",
		request.BaseURL, c.config.MaxPages, c.config.MaxDepth, c.config.MaxConcurrency)

	// Initialize URL filter
	filter, err := NewURLFilter(request.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to create URL filter: %w", err)
	}
	c.filter = filter

	// Initialize URL frontier
	c.frontier = NewURLFrontier(filter, c.config.MaxDepth, c.config.MaxPages)

	// Add base URL to frontier (depth 0 - always crawl)
	c.frontier.Push(URLItem{
		URL:       request.BaseURL,
		Depth:     0,
		ParentURL: "",
	})

	// Initialize rate limiter
	c.rateLimiter = NewRateLimiter(c.config.RequestsPerSecond, 5)

	// Initialize browser manager
	c.browserMgr, err = NewBrowserManager(c.config)
	if err != nil {
		return fmt.Errorf("failed to create browser manager: %w", err)
	}

	// Start result processor
	c.wg.Add(1)
	go c.processResults()

	// Start workers
	c.workers = make([]*CrawlWorker, c.config.MaxConcurrency)
	for i := 0; i < c.config.MaxConcurrency; i++ {
		worker := NewCrawlWorker(
			i,
			c.frontier,
			c.browserMgr,
			c.filter,
			c.rateLimiter,
			c.results,
			c.done,
		)
		c.workers[i] = worker

		c.wg.Add(1)
		go func(w *CrawlWorker) {
			defer c.wg.Done()
			w.Run(c.ctx)
		}(worker)
	}

	// Start completion monitor
	go c.monitorCompletion()

	return nil
}

// processResults handles crawl results and publishes to Kafka
func (c *Coordinator) processResults() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.done:
			return
		case result, ok := <-c.results:
			if !ok {
				return
			}
			c.handleResult(result)
		}
	}
}

// handleResult processes a single crawl result
func (c *Coordinator) handleResult(result CrawlResult) {
	if !result.Success {
		log.Printf("Crawl failed for %s: %s", result.URL, result.Error)
		return
	}

	// Add discovered links to frontier with priority based on source
	for _, link := range result.DiscoveredLinks {
		priority := LinkPriorityContent // Default
		switch link.Source {
		case LinkSourceSidebar:
			priority = LinkPrioritySidebar
		case LinkSourceFooter:
			priority = LinkPriorityFooter
		}
		
		c.frontier.Push(URLItem{
			URL:       link.URL,
			Depth:     result.Depth + 1,
			ParentURL: result.URL,
			Priority:  priority,
			Source:    link.Source,
		})
	}

	// Create job ID (in real implementation, this would come from DB)
	jobID := uint(c.jobsCreated.Add(1))

	// Call job created callback if set
	if c.onJobCreated != nil {
		c.onJobCreated(jobID, result.URL)
	}

	// Create Kafka message
	metadata := map[string]string{
		"base_url":         c.request.BaseURL,
		"crawl_session_id": fmt.Sprintf("crawl_%d", c.request.RequestID),
	}

	msg, err := CreateKafkaMessage(
		jobID,
		c.request.RequestID,
		c.request.ProjectID,
		c.request.UserID,
		&result,
		metadata,
		c.config.CompressHTML,
	)
	if err != nil {
		log.Printf("Failed to create Kafka message for %s: %v", result.URL, err)
		return
	}

	// Serialize and send to Kafka
	msgJSON, err := SerializeMessage(msg)
	if err != nil {
		log.Printf("Failed to serialize message for %s: %v", result.URL, err)
		return
	}

	if c.producer != nil {
		domain := c.filter.GetDomain()
		if err := c.producer.SendMessage(c.config.KafkaTopic, domain, msgJSON); err != nil {
			log.Printf("Failed to send message to Kafka for %s: %v", result.URL, err)
			return
		}
		c.jobsSent.Add(1)
		log.Printf("Sent job %d to Kafka: %s", jobID, result.URL)
	} else {
		// No Kafka producer - just log
		log.Printf("Would send job %d to Kafka: %s (no producer configured)", jobID, result.URL)
		c.jobsSent.Add(1)
	}
}

// monitorCompletion checks if the crawl is complete
func (c *Coordinator) monitorCompletion() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	idleCount := 0
	stuckCount := 0
	const idleThreshold = 3  // 15 seconds of no work
	const stuckThreshold = 6 // 30 seconds of no progress while in_flight > 0
	
	var lastProcessed, lastJobsSent int64
	
	// Calculate max duration (default to 10 minutes if not set)
	maxDuration := c.config.MaxCrawlDuration
	if maxDuration == 0 {
		maxDuration = 10 * time.Minute
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			stats := c.frontier.Stats()
			currentProcessed := stats["urls_popped"]
			currentJobsSent := c.jobsSent.Load()
			inFlight := stats["urls_in_flight"]
			queueSize := stats["queue_size"]
			
			// Check max crawl duration
			elapsed := time.Since(c.startTime)
			if elapsed > maxDuration {
				log.Printf("Crawl timeout: max duration %v exceeded (elapsed: %v)", maxDuration, elapsed)
				c.Stop()
				return
			}
			
			// Check if there's any work left (queue OR in-flight URLs)
			if !c.frontier.HasWork() {
				idleCount++
				stuckCount = 0
				if idleCount >= idleThreshold {
					log.Printf("Crawl complete: no work remaining for %d checks", idleThreshold)
					c.Stop()
					return
				}
			} else {
				idleCount = 0
				
				// Check if we're stuck (in_flight > 0 but no progress)
				if inFlight > 0 && queueSize == 0 && currentJobsSent == lastJobsSent {
					stuckCount++
					if stuckCount >= stuckThreshold {
						log.Printf("Crawl stuck: %d URLs in-flight but no progress for %d seconds. Force stopping.",
							inFlight, stuckCount*5)
						c.Stop()
						return
					}
				} else {
					stuckCount = 0
				}
			}

			// Only log if there's progress
			if currentProcessed != lastProcessed || currentJobsSent != lastJobsSent {
				log.Printf("Progress: queued=%d, in_flight=%d, processed=%d, jobs_sent=%d, elapsed=%v",
					queueSize, inFlight, currentProcessed, currentJobsSent, elapsed.Round(time.Second))
				lastProcessed = currentProcessed
				lastJobsSent = currentJobsSent
			}
		}
	}
}

// Stop gracefully stops the crawl operation
func (c *Coordinator) Stop() {
	log.Printf("Stopping crawl coordinator...")

	// Signal workers to stop
	close(c.done)

	// Cancel context
	if c.cancel != nil {
		c.cancel()
	}

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Printf("Timeout waiting for workers to stop")
	}

	// Close resources
	if c.browserMgr != nil {
		c.browserMgr.Close()
	}
	if c.rateLimiter != nil {
		c.rateLimiter.Close()
	}

	// Close results channel
	close(c.results)

	log.Printf("Crawl stopped. Duration: %v", time.Since(c.startTime))
}

// Wait blocks until the crawl is complete
func (c *Coordinator) Wait() {
	c.wg.Wait()
}

// GetProgress returns current crawl progress
func (c *Coordinator) GetProgress() *CrawlProgress {
	stats := c.frontier.Stats()

	status := "crawling"
	select {
	case <-c.done:
		status = "completed"
	default:
	}

	return &CrawlProgress{
		RequestID:      c.request.RequestID,
		URLsDiscovered: int(stats["urls_added"]),
		URLsQueued:     int(stats["queue_size"]),
		URLsProcessed:  int(stats["urls_popped"]),
		URLsFailed:     int(stats["urls_filtered"]),
		Status:         status,
		StartedAt:      c.startTime,
		UpdatedAt:      time.Now(),
	}
}

// GetStats returns crawl statistics
func (c *Coordinator) GetStats() *CrawlStats {
	frontierStats := c.frontier.Stats()

	return &CrawlStats{
		TotalURLsFound:    frontierStats["urls_added"],
		TotalURLsFiltered: frontierStats["urls_filtered"],
		TotalURLsCrawled:  frontierStats["urls_popped"],
		CrawlDuration:     time.Since(c.startTime),
	}
}

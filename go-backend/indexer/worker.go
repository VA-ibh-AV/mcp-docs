package indexer

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

// CrawlWorker processes URLs from the frontier
type CrawlWorker struct {
	id          int
	frontier    *URLFrontier
	browserMgr  *BrowserManager
	filter      *URLFilter
	rateLimiter *RateLimiter
	results     chan<- CrawlResult
	done        <-chan struct{}

	// Stats
	pagesProcessed atomic.Int64
	pagesFailed    atomic.Int64
}

// NewCrawlWorker creates a new crawl worker
func NewCrawlWorker(
	id int,
	frontier *URLFrontier,
	browserMgr *BrowserManager,
	filter *URLFilter,
	rateLimiter *RateLimiter,
	results chan<- CrawlResult,
	done <-chan struct{},
) *CrawlWorker {
	return &CrawlWorker{
		id:          id,
		frontier:    frontier,
		browserMgr:  browserMgr,
		filter:      filter,
		rateLimiter: rateLimiter,
		results:     results,
		done:        done,
	}
}

// Run starts the worker's main loop
func (w *CrawlWorker) Run(ctx context.Context) {
	log.Printf("[Worker %d] Starting", w.id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Context cancelled, stopping", w.id)
			return
		case <-w.done:
			log.Printf("[Worker %d] Done signal received, stopping", w.id)
			return
		default:
			// Try to get a URL from the frontier
			item := w.frontier.Pop()
			if item == nil {
				// Queue is empty, wait a bit and try again
				select {
				case <-ctx.Done():
					return
				case <-w.done:
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			// Process the URL
			result := w.processURL(ctx, item)
			
			// Send result
			select {
			case w.results <- result:
			case <-ctx.Done():
				return
			case <-w.done:
				return
			}
		}
	}
}

// processURL crawls a single URL and returns the result
func (w *CrawlWorker) processURL(ctx context.Context, item *URLItem) CrawlResult {
	result := CrawlResult{
		URL:         item.URL,
		Depth:       item.Depth,
		ParentURL:   item.ParentURL,
		ProcessedAt: time.Now(),
	}

	// Rate limit
	domain := w.filter.GetDomain()
	if err := w.rateLimiter.Wait(ctx, domain); err != nil {
		result.Success = false
		result.Error = "rate limiter cancelled: " + err.Error()
		w.pagesFailed.Add(1)
		w.frontier.MarkComplete(item.URL) // Mark as complete (not in-flight anymore)
		return result
	}

	// Acquire a browser page
	page, err := w.browserMgr.AcquirePage()
	if err != nil {
		result.Success = false
		result.Error = "failed to acquire page: " + err.Error()
		w.pagesFailed.Add(1)
		w.frontier.MarkComplete(item.URL) // Mark as complete (not in-flight anymore)
		return result
	}
	defer w.browserMgr.ReleasePage(page)

	// Fetch the page
	log.Printf("[Worker %d] Crawling: %s (depth: %d)", w.id, item.URL, item.Depth)
	
	fetchResult, err := w.browserMgr.FetchPage(page, item.URL)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		w.pagesFailed.Add(1)
		w.frontier.MarkComplete(item.URL) // Mark as complete (not in-flight anymore)
		log.Printf("[Worker %d] Failed to crawl %s: %v", w.id, item.URL, err)
		return result
	}

	// Mark as visited (this also removes from in-flight)
	w.frontier.MarkVisited(item.URL)

	// Populate result
	result.Success = true
	result.HTML = fetchResult.HTML
	result.Text = fetchResult.Text
	result.Title = fetchResult.Title
	result.StatusCode = fetchResult.StatusCode
	result.ResponseTimeMs = fetchResult.ResponseTimeMs
	result.ContentType = "text/html"

	// Filter and add discovered links
	for _, link := range fetchResult.Links {
		// Resolve relative URLs
		resolvedLink := w.filter.ResolveURL(item.URL, link)
		// Normalize
		normalizedLink := w.filter.NormalizeURL(resolvedLink)
		
		// Only add links that pass the filter
		if w.filter.IsRelevant(normalizedLink) {
			result.DiscoveredLinks = append(result.DiscoveredLinks, normalizedLink)
		}
	}

	w.pagesProcessed.Add(1)
	log.Printf("[Worker %d] Crawled %s: %d chars, %d links found", 
		w.id, item.URL, len(fetchResult.HTML), len(result.DiscoveredLinks))

	return result
}

// Stats returns worker statistics
func (w *CrawlWorker) Stats() map[string]int64 {
	return map[string]int64{
		"pages_processed": w.pagesProcessed.Load(),
		"pages_failed":    w.pagesFailed.Load(),
	}
}

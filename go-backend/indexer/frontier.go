package indexer

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// URLFrontier is a thread-safe priority queue for URLs to crawl
// It uses BFS ordering (lower depth = higher priority)
type URLFrontier struct {
	mu        sync.Mutex
	queue     *priorityQueue
	visited   sync.Map // URL -> bool
	inQueue   sync.Map // URL -> bool (prevent duplicates in queue)
	inFlight  sync.Map // URL -> bool (currently being processed)
	filter    *URLFilter
	maxDepth  int
	maxPages  int

	// Stats
	urlsAdded    atomic.Int64
	urlsPopped   atomic.Int64
	urlsFiltered atomic.Int64
	urlsInFlight atomic.Int64
}

// NewURLFrontier creates a new URL frontier
func NewURLFrontier(filter *URLFilter, maxDepth, maxPages int) *URLFrontier {
	pq := &priorityQueue{}
	heap.Init(pq)

	return &URLFrontier{
		queue:    pq,
		filter:   filter,
		maxDepth: maxDepth,
		maxPages: maxPages,
	}
}

// Push adds a URL to the frontier if it passes all checks
// Returns true if the URL was added, false otherwise
func (f *URLFrontier) Push(item URLItem) bool {
	// Normalize URL
	normalizedURL := f.filter.NormalizeURL(item.URL)
	item.URL = normalizedURL

	// Check depth limit
	if item.Depth > f.maxDepth {
		f.urlsFiltered.Add(1)
		return false
	}

	// Check if we've hit max pages
	if f.urlsAdded.Load() >= int64(f.maxPages) {
		return false
	}

	// Check if already visited
	if _, visited := f.visited.Load(normalizedURL); visited {
		return false
	}

	// Check if already in queue
	if _, inQueue := f.inQueue.Load(normalizedURL); inQueue {
		return false
	}

	// Apply URL filter (skip for depth 0 - always crawl the base URL)
	if item.Depth > 0 && !f.filter.IsRelevant(normalizedURL) {
		f.urlsFiltered.Add(1)
		return false
	}

	// Add to queue
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring lock
	if _, inQueue := f.inQueue.Load(normalizedURL); inQueue {
		return false
	}

	heap.Push(f.queue, &pqItem{
		url:      normalizedURL,
		depth:    item.Depth,
		parent:   item.ParentURL,
		priority: item.Depth, // Lower depth = higher priority (BFS)
	})
	f.inQueue.Store(normalizedURL, true)
	f.urlsAdded.Add(1)

	return true
}

// Pop removes and returns the next URL to crawl
// Returns nil if the queue is empty
func (f *URLFrontier) Pop() *URLItem {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.queue.Len() == 0 {
		return nil
	}

	item := heap.Pop(f.queue).(*pqItem)
	f.inQueue.Delete(item.url)
	f.urlsPopped.Add(1)
	
	// Mark as in-flight
	f.inFlight.Store(item.url, true)
	f.urlsInFlight.Add(1)

	return &URLItem{
		URL:       item.url,
		Depth:     item.depth,
		ParentURL: item.parent,
	}
}

// MarkVisited marks a URL as visited and removes from in-flight
func (f *URLFrontier) MarkVisited(url string) {
	normalizedURL := f.filter.NormalizeURL(url)
	f.visited.Store(normalizedURL, true)
	
	// Remove from in-flight
	if _, ok := f.inFlight.LoadAndDelete(normalizedURL); ok {
		f.urlsInFlight.Add(-1)
	}
}

// MarkComplete marks a URL as no longer in-flight (for failed URLs)
func (f *URLFrontier) MarkComplete(url string) {
	normalizedURL := f.filter.NormalizeURL(url)
	if _, ok := f.inFlight.LoadAndDelete(normalizedURL); ok {
		f.urlsInFlight.Add(-1)
	}
}

// IsVisited checks if a URL has been visited
func (f *URLFrontier) IsVisited(url string) bool {
	normalizedURL := f.filter.NormalizeURL(url)
	_, visited := f.visited.Load(normalizedURL)
	return visited
}

// Size returns the current queue size
func (f *URLFrontier) Size() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.queue.Len()
}

// Stats returns frontier statistics
func (f *URLFrontier) Stats() map[string]int64 {
	return map[string]int64{
		"urls_added":     f.urlsAdded.Load(),
		"urls_popped":    f.urlsPopped.Load(),
		"urls_filtered":  f.urlsFiltered.Load(),
		"urls_in_flight": f.urlsInFlight.Load(),
		"queue_size":     int64(f.Size()),
	}
}

// IsEmpty returns true if the queue is empty
func (f *URLFrontier) IsEmpty() bool {
	return f.Size() == 0
}

// HasWork returns true if there's work to do (queue not empty OR urls in flight)
func (f *URLFrontier) HasWork() bool {
	return f.Size() > 0 || f.urlsInFlight.Load() > 0
}

// InFlightCount returns the number of URLs currently being processed
func (f *URLFrontier) InFlightCount() int64 {
	return f.urlsInFlight.Load()
}

// VisitedCount returns the number of visited URLs
func (f *URLFrontier) VisitedCount() int {
	count := 0
	f.visited.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// --- Priority Queue Implementation ---

// pqItem is an item in the priority queue
type pqItem struct {
	url      string
	depth    int
	parent   string
	priority int // Lower = higher priority
	index    int // Index in the heap
}

// priorityQueue implements heap.Interface
type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Lower priority number = higher priority (BFS: prefer lower depth)
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // Avoid memory leak
	item.index = -1 // Mark as removed
	*pq = old[0 : n-1]
	return item
}

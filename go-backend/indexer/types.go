package indexer

import (
	"time"
)

// CrawlRequest represents a request to crawl a website
type CrawlRequest struct {
	RequestID uint   `json:"request_id"`
	ProjectID uint   `json:"project_id"`
	UserID    string `json:"user_id"`
	BaseURL   string `json:"base_url"`
	MaxPages  int    `json:"max_pages"`
	MaxDepth  int    `json:"max_depth"`
}

// URLItem represents a URL in the crawl queue
type URLItem struct {
	URL       string `json:"url"`
	Depth     int    `json:"depth"`
	ParentURL string `json:"parent_url"`
}

// CrawlResult represents the outcome of crawling a URL
type CrawlResult struct {
	URL             string    `json:"url"`
	Depth           int       `json:"depth"`
	ParentURL       string    `json:"parent_url"`
	DiscoveredLinks []string  `json:"discovered_links"`

	// Content extracted from the page
	HTML  string `json:"html"`
	Text  string `json:"text"`
	Title string `json:"title"`

	// Metadata
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
	ProcessedAt    time.Time `json:"processed_at"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"`
	ContentType    string    `json:"content_type"`
}

// CrawlProgress represents the current state of a crawl
type CrawlProgress struct {
	RequestID      uint      `json:"request_id"`
	URLsDiscovered int       `json:"urls_discovered"`
	URLsQueued     int       `json:"urls_queued"`
	URLsProcessed  int       `json:"urls_processed"`
	URLsFailed     int       `json:"urls_failed"`
	CurrentDepth   int       `json:"current_depth"`
	Status         string    `json:"status"` // crawling, completed, failed
	StartedAt      time.Time `json:"started_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IndexingJobMessage is the Kafka message sent for each discovered URL
type IndexingJobMessage struct {
	JobID      uint   `json:"job_id"`
	RequestID  uint   `json:"request_id"`
	ProjectID  uint   `json:"project_id"`
	UserID     string `json:"user_id"`
	URL        string `json:"url"`
	Depth      int    `json:"depth"`
	ParentURL  string `json:"parent_url"`

	// Content (compressed)
	Content *PageContent `json:"content,omitempty"`

	// Metadata
	DiscoveredAt time.Time         `json:"discovered_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PageContent holds the extracted page content
type PageContent struct {
	HTML        string `json:"html"`         // Gzip + Base64 encoded
	Text        string `json:"text"`         // Plain text from body
	Title       string `json:"title"`        // Page title
	ContentType string `json:"content_type"` // MIME type
	Encoding    string `json:"encoding"`     // e.g., "gzip+base64"
	HTMLSize    int    `json:"html_size"`    // Original HTML size in bytes
}

// CrawlStats holds statistics about the crawl
type CrawlStats struct {
	TotalURLsFound     int64         `json:"total_urls_found"`
	TotalURLsFiltered  int64         `json:"total_urls_filtered"`
	TotalURLsCrawled   int64         `json:"total_urls_crawled"`
	TotalURLsFailed    int64         `json:"total_urls_failed"`
	TotalBytesReceived int64         `json:"total_bytes_received"`
	AverageResponseMs  float64       `json:"average_response_ms"`
	CrawlDuration      time.Duration `json:"crawl_duration"`
}

package indexer

import "time"

// Config holds all configuration for the indexer
type Config struct {
	// Browser settings
	Headless  bool   `json:"headless"`
	UserAgent string `json:"user_agent"`

	// Crawl limits
	MaxPages         int           `json:"max_pages"`
	MaxDepth         int           `json:"max_depth"`
	MaxConcurrency   int           `json:"max_concurrency"`    // Number of concurrent workers
	MaxCrawlDuration time.Duration `json:"max_crawl_duration"` // Maximum time for entire crawl

	// Timeouts
	PageTimeout    time.Duration `json:"page_timeout"`
	NetworkTimeout time.Duration `json:"network_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"` // Wait for JS rendering

	// Rate limiting
	RequestsPerSecond float64       `json:"requests_per_second"` // Per domain
	RequestDelay      time.Duration `json:"request_delay"`       // Min delay between requests

	// Kafka settings
	KafkaTopic     string        `json:"kafka_topic"`
	KafkaBatchSize int           `json:"kafka_batch_size"`
	KafkaFlushTime time.Duration `json:"kafka_flush_time"`

	// Content settings
	CompressHTML  bool `json:"compress_html"`
	MaxHTMLSize   int  `json:"max_html_size"`   // Max HTML size to store (bytes)
	ExtractText   bool `json:"extract_text"`    // Extract text content
	MaxTextLength int  `json:"max_text_length"` // Max text length to store
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		// Browser
		Headless:  true,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 MCPDocsBot/1.0",

		// Crawl limits
		MaxPages:         20,
		MaxDepth:         5,
		MaxConcurrency:   5,
		MaxCrawlDuration: 10 * time.Minute, // Max 10 minutes per crawl

		// Timeouts
		PageTimeout:    30 * time.Second,
		NetworkTimeout: 30 * time.Second,
		IdleTimeout:    2 * time.Second,

		// Rate limiting - be polite
		RequestsPerSecond: 2.0, // 2 requests per second per domain
		RequestDelay:      500 * time.Millisecond,

		// Kafka
		KafkaTopic:     "indexing_jobs",
		KafkaBatchSize: 50,
		KafkaFlushTime: 5 * time.Second,

		// Content
		CompressHTML:  true,
		MaxHTMLSize:   5 * 1024 * 1024, // 5MB
		ExtractText:   true,
		MaxTextLength: 1024 * 1024, // 1MB
	}
}

// ConfigOption is a function that modifies config
type ConfigOption func(*Config)

// WithMaxPages sets the maximum pages to crawl
func WithMaxPages(n int) ConfigOption {
	return func(c *Config) {
		c.MaxPages = n
	}
}

// WithMaxDepth sets the maximum crawl depth
func WithMaxDepth(n int) ConfigOption {
	return func(c *Config) {
		c.MaxDepth = n
	}
}

// WithConcurrency sets the number of concurrent workers
func WithConcurrency(n int) ConfigOption {
	return func(c *Config) {
		c.MaxConcurrency = n
	}
}

// WithRateLimit sets requests per second
func WithRateLimit(rps float64) ConfigOption {
	return func(c *Config) {
		c.RequestsPerSecond = rps
	}
}

// WithKafkaTopic sets the Kafka topic
func WithKafkaTopic(topic string) ConfigOption {
	return func(c *Config) {
		c.KafkaTopic = topic
	}
}

// NewConfig creates a new config with the given options
func NewConfig(opts ...ConfigOption) *Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

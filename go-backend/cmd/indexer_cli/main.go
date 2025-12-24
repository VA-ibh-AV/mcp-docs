package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mcpdocs/indexer"
	"mcpdocs/kafka"
)

func main() {
	// Parse flags
	url := flag.String("url", "", "URL to crawl (required)")
	maxPages := flag.Int("max-pages", 10, "Maximum pages to crawl")
	maxDepth := flag.Int("max-depth", 3, "Maximum crawl depth")
	concurrency := flag.Int("concurrency", 3, "Number of concurrent workers")
	rps := flag.Float64("rps", 2.0, "Requests per second (rate limit)")
	brokers := flag.String("brokers", "localhost:9092", "Kafka broker addresses")
	topic := flag.String("topic", "indexing_jobs", "Kafka topic")
	noKafka := flag.Bool("no-kafka", false, "Run without Kafka (dry run)")
	
	flag.Parse()

	if *url == "" {
		log.Fatal("URL is required. Use -url flag.")
	}

	log.Printf("=== Website Indexer CLI ===")
	log.Printf("URL: %s", *url)
	log.Printf("Max Pages: %d", *maxPages)
	log.Printf("Max Depth: %d", *maxDepth)
	log.Printf("Concurrency: %d", *concurrency)
	log.Printf("Rate Limit: %.2f req/s", *rps)
	log.Printf("Kafka: %s (topic: %s)", *brokers, *topic)
	log.Printf("Dry Run: %v", *noKafka)

	// Create config
	config := indexer.NewConfig(
		indexer.WithMaxPages(*maxPages),
		indexer.WithMaxDepth(*maxDepth),
		indexer.WithConcurrency(*concurrency),
		indexer.WithRateLimit(*rps),
		indexer.WithKafkaTopic(*topic),
	)

	// Create Kafka producer (if not dry run)
	var producer *kafka.Producer
	var err error
	if !*noKafka {
		producer, err = kafka.NewProducer([]string{*brokers})
		if err != nil {
			log.Printf("Warning: Failed to create Kafka producer: %v", err)
			log.Printf("Continuing in dry-run mode...")
			producer = nil
		} else {
			defer producer.Close()
		}
	}

	// Create coordinator
	coordinator := indexer.NewCoordinator(config, producer)

	// Handle signals for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		log.Println("\nReceived shutdown signal, stopping crawl...")
		cancel()
		coordinator.Stop()
	}()

	// Create crawl request
	request := &indexer.CrawlRequest{
		RequestID: 1,
		ProjectID: 1,
		UserID:    "test-user",
		BaseURL:   *url,
		MaxPages:  *maxPages,
		MaxDepth:  *maxDepth,
	}

	// Start crawling
	log.Println("\nStarting crawl...")
	if err := coordinator.Start(ctx, request); err != nil {
		log.Fatalf("Failed to start crawl: %v", err)
	}

	// Wait for completion
	coordinator.Wait()

	// Print final stats
	stats := coordinator.GetStats()
	log.Println("\n=== Crawl Complete ===")
	log.Printf("Total URLs Found: %d", stats.TotalURLsFound)
	log.Printf("Total URLs Filtered: %d", stats.TotalURLsFiltered)
	log.Printf("Total URLs Crawled: %d", stats.TotalURLsCrawled)
	log.Printf("Duration: %v", stats.CrawlDuration)

	log.Println("\nDone!")
}

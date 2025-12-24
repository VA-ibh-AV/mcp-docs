package consumers

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"mcpdocs/indexer"
	"mcpdocs/kafka"
	"mcpdocs/models"

	"github.com/IBM/sarama"
)

// IndexingRequestPayload represents the message from indexing_requests topic
type IndexingRequestPayload struct {
	ID        uint   `json:"id"`
	UserID    string `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	ProjectID uint   `json:"project_id"`
	Status    string `json:"status"`
}

// IndexingConsumerCallbacks contains callbacks for status updates
type IndexingConsumerCallbacks struct {
	// Called to update IndexingRequest status
	OnStatusUpdate func(ctx context.Context, requestID uint, status string, errorMsg string) error
	// Called to update total jobs count after crawl completes
	OnTotalJobsUpdate func(ctx context.Context, requestID uint, totalJobs int) error
}

// IndexingConsumer consumes indexing requests and triggers the crawler
type IndexingConsumer struct {
	consumer      sarama.Consumer
	producer      *kafka.Producer
	indexerConfig *indexer.Config
	callbacks     IndexingConsumerCallbacks
	wg            sync.WaitGroup
	stopCh        chan struct{}
}

// NewIndexingConsumer creates a new indexing consumer
func NewIndexingConsumer(
	brokers []string,
	producer *kafka.Producer,
	indexerConfig *indexer.Config,
	callbacks IndexingConsumerCallbacks,
) (*IndexingConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &IndexingConsumer{
		consumer:      consumer,
		producer:      producer,
		indexerConfig: indexerConfig,
		callbacks:     callbacks,
		stopCh:        make(chan struct{}),
	}, nil
}

// Start begins consuming from the indexing_requests topic
func (c *IndexingConsumer) Start(topic string) error {
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			slog.Error("Failed to consume partition", "partition", partition, "error", err)
			continue
		}

		c.wg.Add(1)
		go c.consumePartition(pc)
	}

	slog.Info("IndexingConsumer started", "topic", topic, "partitions", len(partitions))
	return nil
}

// consumePartition processes messages from a single partition
func (c *IndexingConsumer) consumePartition(pc sarama.PartitionConsumer) {
	defer c.wg.Done()
	defer pc.Close()

	for {
		select {
		case <-c.stopCh:
			return
		case err := <-pc.Errors():
			slog.Error("Consumer error", "error", err)
		case msg := <-pc.Messages():
			c.handleMessage(msg)
		}
	}
}

// handleMessage processes a single indexing request message
func (c *IndexingConsumer) handleMessage(msg *sarama.ConsumerMessage) {
	slog.Info("Received indexing request",
		"key", string(msg.Key),
		"partition", msg.Partition,
		"offset", msg.Offset,
	)

	// Parse the indexing request
	var payload IndexingRequestPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		slog.Error("Failed to unmarshal indexing request", "error", err)
		return
	}

	slog.Info("Processing indexing request",
		"request_id", payload.ID,
		"endpoint", payload.Endpoint,
		"project_id", payload.ProjectID,
	)

	// Update status to in_progress
	ctx := context.Background()
	if c.callbacks.OnStatusUpdate != nil {
		if err := c.callbacks.OnStatusUpdate(ctx, payload.ID, models.IndexingStatusInProgress, ""); err != nil {
			slog.Error("Failed to update status to in_progress", "error", err)
		}
	}

	// Start the crawl
	totalJobs, err := c.runCrawl(ctx, &payload)
	if err != nil {
		slog.Error("Crawl failed", "request_id", payload.ID, "error", err)
		if c.callbacks.OnStatusUpdate != nil {
			c.callbacks.OnStatusUpdate(ctx, payload.ID, models.IndexingStatusFailed, err.Error())
		}
		return
	}

	// Update total jobs count
	if c.callbacks.OnTotalJobsUpdate != nil {
		if err := c.callbacks.OnTotalJobsUpdate(ctx, payload.ID, totalJobs); err != nil {
			slog.Error("Failed to update total jobs", "error", err)
		}
	}

	// Update status to completed (crawl phase done, Python agent will process)
	if c.callbacks.OnStatusUpdate != nil {
		if err := c.callbacks.OnStatusUpdate(ctx, payload.ID, models.IndexingStatusCrawlComplete, ""); err != nil {
			slog.Error("Failed to update status to crawl_complete", "error", err)
		}
	}

	slog.Info("Crawl completed",
		"request_id", payload.ID,
		"total_jobs", totalJobs,
	)
}

// runCrawl executes the crawler and returns the total number of pages crawled
func (c *IndexingConsumer) runCrawl(ctx context.Context, payload *IndexingRequestPayload) (int, error) {
	// Create crawl request
	crawlRequest := &indexer.CrawlRequest{
		RequestID: payload.ID,
		ProjectID: payload.ProjectID,
		UserID:    payload.UserID,
		BaseURL:   payload.Endpoint,
		MaxPages:  c.indexerConfig.MaxPages,
		MaxDepth:  c.indexerConfig.MaxDepth,
	}

	// Create coordinator
	coordinator := indexer.NewCoordinator(c.indexerConfig, c.producer)

	// Start crawl
	if err := coordinator.Start(ctx, crawlRequest); err != nil {
		return 0, err
	}

	// Wait for completion
	coordinator.Wait()

	// Get stats
	stats := coordinator.GetStats()
	return int(stats.TotalURLsCrawled), nil
}

// Stop gracefully stops the consumer
func (c *IndexingConsumer) Stop() {
	close(c.stopCh)
	c.wg.Wait()
	c.consumer.Close()
	slog.Info("IndexingConsumer stopped")
}

// RunBlocking starts the consumer and blocks until Stop is called
func (c *IndexingConsumer) RunBlocking(topic string) error {
	if err := c.Start(topic); err != nil {
		return err
	}
	c.wg.Wait()
	return nil
}

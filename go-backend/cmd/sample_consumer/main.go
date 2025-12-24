package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"mcpdocs/indexer"

	"github.com/IBM/sarama"
)

func main() {
	// Parse flags
	brokers := flag.String("brokers", "localhost:9092", "Kafka broker addresses (comma-separated)")
	topic := flag.String("topic", "indexing_jobs", "Kafka topic to consume")
	groupID := flag.String("group", "indexer-test-consumer", "Consumer group ID")
	flag.Parse()

	log.Printf("Starting sample consumer...")
	log.Printf("Brokers: %s", *brokers)
	log.Printf("Topic: %s", *topic)
	log.Printf("Group: %s", *groupID)

	// Create Kafka consumer config
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	// Create consumer group
	consumer, err := sarama.NewConsumerGroup([]string{*brokers}, *groupID, config)
	if err != nil {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	defer consumer.Close()

	// Handle signals
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Create handler
	handler := &ConsumerHandler{}

	// Start consuming
	log.Println("Waiting for messages... (press Ctrl+C to exit)")
	for {
		if err := consumer.Consume(ctx, []string{*topic}, handler); err != nil {
			log.Printf("Consumer error: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
	}

	log.Println("Consumer stopped")
}

// ConsumerHandler implements sarama.ConsumerGroupHandler
type ConsumerHandler struct{}

func (h *ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	log.Println("Consumer setup complete")
	return nil
}

func (h *ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Println("Consumer cleanup complete")
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		separator := strings.Repeat("=", 60)
		log.Printf("\n%s", separator)
		log.Printf("📥 Received message:")
		log.Printf("   Topic: %s", message.Topic)
		log.Printf("   Partition: %d", message.Partition)
		log.Printf("   Offset: %d", message.Offset)
		log.Printf("   Key: %s", string(message.Key))

		// Parse the message
		var job indexer.IndexingJobMessage
		if err := json.Unmarshal(message.Value, &job); err != nil {
			log.Printf("   ⚠ Failed to parse message: %v", err)
			log.Printf("   Raw value: %s", string(message.Value))
		} else {
			log.Printf("\n📄 Job Details:")
			log.Printf("   Job ID: %d", job.JobID)
			log.Printf("   Request ID: %d", job.RequestID)
			log.Printf("   Project ID: %d", job.ProjectID)
			log.Printf("   User ID: %s", job.UserID)
			log.Printf("   URL: %s", job.URL)
			log.Printf("   Depth: %d", job.Depth)
			log.Printf("   Parent URL: %s", job.ParentURL)
			log.Printf("   Discovered At: %s", job.DiscoveredAt)

			if job.Content != nil {
				log.Printf("\n📝 Content:")
				log.Printf("   Title: %s", job.Content.Title)
				log.Printf("   Encoding: %s", job.Content.Encoding)
				log.Printf("   HTML Size: %d bytes", job.Content.HTMLSize)
				log.Printf("   Text Length: %d chars", len(job.Content.Text))

				// Show first 200 chars of text
				text := job.Content.Text
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				log.Printf("   Text Preview: %s", text)

				// If HTML is compressed, decompress and show size
				if job.Content.Encoding == "gzip+base64" {
					html, err := indexer.DecompressHTML(job.Content.HTML)
					if err == nil {
						log.Printf("   Decompressed HTML: %d bytes", len(html))
					}
				}
			}

			if job.Metadata != nil {
				log.Printf("\n📊 Metadata:")
				for k, v := range job.Metadata {
					log.Printf("   %s: %s", k, v)
				}
			}
		}

		// Mark message as processed
		session.MarkMessage(message, "")
		log.Printf("%s\n", separator)
	}
	return nil
}

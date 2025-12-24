package kafka

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: consumer}, nil
}

func (c *Consumer) Consume(topic string) {
	partitionList, err := c.consumer.Partitions(topic)
	if err != nil {
		log.Printf("Error retrieving partition list for topic %s: %v", topic, err)
		return
	}

	var wg sync.WaitGroup

	for _, partition := range partitionList {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error consuming partition %d: %v", partition, err)
			continue
		}

		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for message := range pc.Messages() {
				// Parse message to extract key fields without the large base64 content
				var msg struct {
					JobID     uint   `json:"job_id"`
					RequestID uint   `json:"request_id"`
					ProjectID uint   `json:"project_id"`
					URL       string `json:"url"`
				}
				if err := json.Unmarshal(message.Value, &msg); err == nil {
					log.Printf("Message received: job_id=%d request_id=%d project_id=%d url=%s topic=%s partition=%d offset=%d",
						msg.JobID, msg.RequestID, msg.ProjectID, msg.URL, message.Topic, message.Partition, message.Offset)
				} else {
					// Fallback: just log metadata without value
					log.Printf("Message received: key=%s topic=%s partition=%d offset=%d (payload parsing failed)",
						string(message.Key), message.Topic, message.Partition, message.Offset)
				}
			}
		}(pc)
	}

	wg.Wait()
}

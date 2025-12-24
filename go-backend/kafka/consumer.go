package kafka

import (
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
				log.Printf("Message received: key=%s value=%s topic=%s partition=%d offset=%d",
					string(message.Key), string(message.Value), message.Topic, message.Partition, message.Offset)
			}
		}(pc)
	}

	wg.Wait()
}

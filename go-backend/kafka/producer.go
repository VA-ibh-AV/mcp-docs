package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) SendMessage(topic, key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

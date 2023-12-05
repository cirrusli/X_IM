package kafka

import (
	"X_IM/pkg/logger"
	"github.com/Shopify/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

var log = logger.WithField("pkg", "kafka")

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: producer, topic: topic}, nil
}

func (p *Producer) SendMessage(msg []byte) error {
	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(msg),
	}
	_, _, err := p.producer.SendMessage(message)
	if err != nil {
		log.Println("Error while sending message to Kafka:", err)
		return err
	}
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

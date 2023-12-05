package kafka

import (
	"github.com/Shopify/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
	topic    string
}

func NewConsumer(brokers []string, topic string) (*Consumer, error) {
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, err
	}
	return &Consumer{consumer: consumer, topic: topic}, nil
}

func (c *Consumer) ConsumeMessages(handler func([]byte)) error {
	partitionConsumer, err := c.consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Errorln("creating consumer partition:", err)
		return err
	}
	defer func(partitionConsumer sarama.PartitionConsumer) {
		err = partitionConsumer.Close()
		if err != nil {
			log.Errorln("closing consumer partition:", err)
		}
	}(partitionConsumer)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			handler(msg.Value)
		case err = <-partitionConsumer.Errors():
			log.Errorln("consuming message error:", err)
			return err
			//case <-time.After(2 * time.Second):
			//	//NOTE: THE CASE JUST FOR TEST
			//	log.Infoln("No message received in 2 seconds, exiting...")
			//	return nil
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"testing"
)

func TestConsumeMessages(t *testing.T) {
	topic := "test-topic"

	mockConsumer := mocks.NewConsumer(t, nil)
	mockPartitionConsumer := mockConsumer.ExpectConsumePartition(topic, 0, sarama.OffsetNewest)
	mockPartitionConsumer.ExpectMessagesDrainedOnClose()

	consumer := &Consumer{
		consumer: mockConsumer,
		topic:    topic,
	}

	handler := func(msg []byte) {
		t.Logf("Received message: %v", string(msg))
	}
	// Generate mock messages
	messages := []string{"test message1", "test message2", "test message3"}
	for _, msg := range messages {
		mockPartitionConsumer.YieldMessage(&sarama.ConsumerMessage{Value: []byte(msg)})
	}
	err := consumer.ConsumeMessages(handler)
	if err != nil {
		t.Errorf("Expected ConsumeMessages to succeed, got err: %v", err)
	}
}

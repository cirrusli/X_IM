package kafka

import (
	"github.com/Shopify/sarama/mocks"
	"testing"
)

func TestSendMessage(t *testing.T) {
	topic := "test-topic"

	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	producer := &Producer{
		producer: mockProducer,
		topic:    topic,
	}

	err := producer.SendMessage([]byte("test message"))
	if err != nil {
		t.Errorf("Expected SendMessage to succeed, got err: %v", err)
	}
}

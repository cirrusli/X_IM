package main

import (
	"github.com/Shopify/sarama"
	"testing"
	"time"
)

func TestKafkaConsumer(t *testing.T) {
	kafkaURL := "8.146.198.70:9092" // 设置有效的 Kafka 服务器地址
	testTopic := "mytopic"          // 设置你要测试的主题名称

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// 创建消费者
	consumer, err := sarama.NewConsumer([]string{kafkaURL}, config)
	if err != nil {
		t.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer func(consumer sarama.Consumer) {
		_ = consumer.Close()
	}(consumer)

	// 创建分区消费者，消费最新生产的消息
	partitionConsumer, err := consumer.ConsumePartition(testTopic, 0, sarama.OffsetNewest)
	if err != nil {
		t.Fatalf("Failed to create Kafka partition consumer: %v", err)
	}
	defer func(partitionConsumer sarama.PartitionConsumer) {
		_ = partitionConsumer.Close()
	}(partitionConsumer)

	// 设置一个超时时间，防止测试函数无限等待
	timeout := time.After(20 * time.Second)
	received := false

ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			t.Logf("Received message: %v", string(msg.Value))
			received = true
			break ConsumerLoop
		case <-timeout:
			t.Fatal("Test timed out before message was received")
			return
		case err := <-partitionConsumer.Errors():
			t.Fatalf("Received error while consuming: %v", err)
			return
		}
	}

	if !received {
		t.Fatal("Expected to receive a message, but didn't")
	}
}

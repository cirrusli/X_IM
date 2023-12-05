package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
	"time"
)

const kafkaURL = "8.146.198.70:9092"

func TestKafka(t *testing.T) {
	// Initialize Kafka producer
	producer, err := NewProducer([]string{kafkaURL}, "test")
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer func(producer *Producer) {
		_ = producer.Close()
	}(producer)

	// Initialize Kafka consumer
	consumer, err := NewConsumer([]string{kafkaURL}, "test")
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer func(consumer *Consumer) {
		_ = consumer.Close()
	}(consumer)

	// Start consuming messages in a new goroutine
	go func() {
		err := consumer.ConsumeMessages(func(msg []byte) {
			// Handle the message here, e.g., write it to the database
			log.Printf("Received message: %s", string(msg))
		})
		if err != nil {
			log.Fatalf("Failed to consume messages: %v", err)
		}
	}()

	// In your message handling function, send the message to Kafka
	err = producer.SendMessage([]byte("test message"))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
}
func TestConsumer(t *testing.T) {
	testTopic := "test" // 设置你要测试的主题名称

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
	exitLoop := false

	for !exitLoop {
		select {
		case msg := <-partitionConsumer.Messages():
			t.Logf("Received message: %v", string(msg.Value))
			received = true
			exitLoop = true
		case <-timeout:
			t.Fatal("Test timed out before message was received")
			return
		case err = <-partitionConsumer.Errors():
			t.Fatalf("Received error while consuming: %v", err)
			return
		}
	}

	if !received {
		t.Fatal("Expected to receive a message, but didn't")
	}
}

func TestProducerWithOffset(t *testing.T) {
	testTopic := "test" // 设置你要测试的主题名称

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	// 创建一个AdminClient来获取偏移量
	adminClient, err := sarama.NewClusterAdmin([]string{kafkaURL}, config)
	if err != nil {
		t.Fatalf("Failed to create Kafka admin client: %v", err)
	}
	defer func(adminClient sarama.ClusterAdmin) {
		_ = adminClient.Close()
	}(adminClient)

	// 获取主题的元数据
	metadata, err := adminClient.DescribeTopics([]string{testTopic})
	if err != nil {
		t.Fatalf("Failed to get metadata for topic: %v", err)
	}

	// 创建一个消费者来获取偏移量
	consumer, err := sarama.NewConsumer([]string{kafkaURL}, config)
	if err != nil {
		t.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer func(consumer sarama.Consumer) {
		_ = consumer.Close()
	}(consumer)

	// 获取最新的偏移量
	var latestOffset int64
	for _, topicMetadata := range metadata {
		for _, partitionMetadata := range topicMetadata.Partitions {
			offset, err := consumer.ConsumePartition(testTopic, partitionMetadata.ID, sarama.OffsetNewest)
			if err != nil {
				t.Fatalf("Failed to get offset for partition %d: %v", partitionMetadata.ID, err)
			}
			if offset.HighWaterMarkOffset() > latestOffset {
				latestOffset = offset.HighWaterMarkOffset()
			}
		}
	}

	// 创建生产者
	producer, err := sarama.NewSyncProducer([]string{kafkaURL}, config)
	if err != nil {
		t.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer func(producer sarama.SyncProducer) {
		_ = producer.Close()
	}(producer)

	// 创建消息，将偏移量添加到消息中
	msg := &sarama.ProducerMessage{}
	msg.Topic = testTopic
	msg.Value = sarama.StringEncoder(fmt.Sprintf("\nMessage4 with offset: %d\n", latestOffset))

	// 发送消息
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
}

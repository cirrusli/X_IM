package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
)

func TestKafkaProducerWithOffset(t *testing.T) {
	kafkaURL := "8.146.198.70:9092" // 设置有效的 Kafka 服务器地址
	testTopic := "mytopic"          // 设置你要测试的主题名称

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

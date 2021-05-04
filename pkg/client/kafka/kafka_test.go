package kafka

import (
	"context"
	"fmt"
	"testing"

	"github.com/douyu/jupiter/pkg/client/kafka/config"
)

func runTransaction(p *config.Producer, topics ...string) {
	var ctx = context.Background()
	p.InitTransactions(ctx)

	var err = p.BeginTransaction()
	if err != nil {
		fmt.Printf("Begin transaction failed: %v\n", err)
		return
	}

	for _, topic := range topics {
		for _, word := range []string{"Send", "Kafka", "Transaction", "Message"} {
			var err = p.ProduceTo(topic, 0, 0, []byte(word))
			if err != nil {
				fmt.Printf("Transaction produce failed: %v\n", err)
				err = p.AbortTransaction(ctx)
				if err != nil {
					fmt.Printf("Abort transaction failed: %v\n", err)
				}
				return
			}

			fmt.Printf("Transaction produce %s to %s Success\n", word, topic)
		}
	}

	err = p.CommitTransaction(ctx)
	if err != nil {
		fmt.Printf("Commit transaction failed: %v\n", err)
	}
}

func runProducer(p *config.Producer, topics ...string) {
	for _, topic := range topics {
		for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
			var err = p.ProduceTo(topic, 1, 0, []byte(word))
			if err != nil {
				fmt.Printf("Produce failed: %v\n", err)
			} else {
				fmt.Printf("Produce %s to %s Success\n", word, topic)
			}
		}
	}
}

func runConsumer(c *config.Consumer) {
	for {
		var msg, err = c.ReadMessage(-1)
		if err != nil {
			fmt.Printf("Comsume failed: %v\n", err)
			continue
		}

		fmt.Printf("Receive: %s: %s\n", msg.TopicPartition, string(msg.Value))
	}
}

func TestKafka(t *testing.T) {
	var topics = []string{"testTopicA", "testTopicB"}
	var producerConfig = config.DefaultProducerConfig()
	var txProducerConfig = config.DefaultProducerConfig()
	var consumerConfig = config.DefaultConsumerConfig()
	consumerConfig.KafkaConfig.GroupID = "TestConsumerGroup"

	// 测试机 kafka
	producerConfig.KafkaConfig.MetadataBrokerList = "192.168.1.3:39092"
	txProducerConfig.KafkaConfig.MetadataBrokerList = "192.168.1.3:39092"
	consumerConfig.KafkaConfig.MetadataBrokerList = "192.168.1.3:39092"

	// 开启事务需要的配置
	txProducerConfig.KafkaConfig.TransactionalID = "TestTransactionID"
	txProducerConfig.KafkaConfig.EnableIdempotence = true
	txProducerConfig.KafkaConfig.MaxInFlightRequestsPerConnection = 5
	txProducerConfig.KafkaConfig.MessageTimeoutMs = 6e4

	var p = producerConfig.BuildProducer()
	defer func() {
		p.Flush(15 * 1000)
		p.Close()
	}()

	var txp = txProducerConfig.BuildProducer()
	defer func() {
		txp.Flush(15 * 1000)
		txp.Close()
	}()

	// 	var newTopic = "test_topic_3"
	// 	var a = p.NewAdminClient()
	// 	var md, err = a.GetMetadata(&newTopic, false, 1000)
	// 	if err != nil {
	// 		fmt.Printf("get md failed: %v\n", err)
	// 		return
	// 	}
	// 	fmt.Printf("get md: %+v\n", md)
	//
	// 	result, err := a.DeleteTopics(context.Background(), topics)
	// 	if err != nil {
	// 		fmt.Printf("rm topics failed: %v\n", err)
	// 		return
	// 	}
	//
	// 	fmt.Printf("result: %+v\n", result)

	p.RunMonitor()
	txp.RunMonitor()

	var c = consumerConfig.BuildConsumer()
	c.SubscribeTopics(topics, nil)

	defer c.Close()
	go runConsumer(c)

	runProducer(p, topics...)
	runTransaction(txp, topics...)

	<-make(chan struct{})
}

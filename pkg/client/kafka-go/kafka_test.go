package kafka

import (
	"context"
	"fmt"
	"testing"
)

func runProducer(p *Producer) {
	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		var err = p.ProduceTo(context.Background(), 1, []byte(word))

		if err != nil {
			fmt.Printf("Produce failed: %v\n", err)
		} else {
			fmt.Printf("Produce %s Success\n", word)
		}
	}
}

func runConsumer(c *Consumer) {
	for {
		var msg, err = c.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Comsume failed: %v\n", err)
			continue
		}

		if err == nil {
			fmt.Printf("Receive: %s[%d]: %s\n", msg.Topic, msg.Partition, string(msg.Value))
		}
	}
}

func TestKafka(t *testing.T) {
	var producerConfig = DefaultProducerConfig()
	producerConfig.Topic = "test_consumer_result"

	var consumerConfig = DefaultConsumerConfig()
	consumerConfig.GroupID = "UserBalanceServiceGroup"
	consumerConfig.Topic = "test_consumer_result"

	var p = producerConfig.BuildProducer()

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

	go runProducer(p)

	var c = consumerConfig.BuildConsumer()

	runConsumer(c)
}

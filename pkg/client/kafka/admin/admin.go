package admin

import "github.com/confluentinc/confluent-kafka-go/kafka"

type Admin struct {
	*kafka.AdminClient
}

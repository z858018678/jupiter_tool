package kafka

import (
	"context"
	"encoding/json"

	"github.com/douyu/jupiter/pkg/conf"
	file_datasource "github.com/douyu/jupiter/pkg/datasource/file"
	"github.com/douyu/jupiter/pkg/xlog"
	kf "github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers  []string `json:"brokers"`
	Topic    string   `json:"topic"`
	GroupID  string   `json:"group_id"`
	MinBytes int      `json:"max_bytes"`
	MaxBytes int      `json:"max_bytes"`

	logger *xlog.Logger
}

func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:  []string{"localhost:9092"},
		MinBytes: 1e4,
		MaxBytes: 1e6,
		logger:   xlog.JupiterLogger,
	}
}

// StdConsumerConfig ...
func StdConsumerConfig(path string) *ConsumerConfig {
	var cf = DefaultConsumerConfig()
	provider := file_datasource.NewDataSource(path, true)
	var c = conf.New()
	if err := c.LoadFromDataSource(provider, json.Unmarshal); err != nil {
		xlog.Panic("unmarshal kafka config",
			xlog.String("path", path),
			xlog.Any("kafka config", path),
			xlog.String("error", err.Error()))
	}

	if err := c.UnmarshalKey("", &cf, conf.TagName("json")); err != nil {
		xlog.Panic("unmarshal kafka config",
			xlog.String("path", path),
			xlog.Any("kafka config", path),
			xlog.String("error", err.Error()))
	}

	return cf
}

type Consumer struct {
	*ConsumerConfig
	rd *kf.Reader
}

// Build ...
func (config *ConsumerConfig) BuildConsumer() *Consumer {
	if config == nil {
		return nil
	}

	var consumer Consumer
	consumer.ConsumerConfig = config

	var conf = kf.ReaderConfig{
		GroupID:  config.GroupID,
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
	}

	consumer.rd = kf.NewReader(conf)
	return &consumer
}

func (c *Consumer) ReadMessage(ctx context.Context) (kf.Message, error) {
	return c.rd.ReadMessage(ctx)
}

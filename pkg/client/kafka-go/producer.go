package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/douyu/jupiter/pkg/conf"
	file_datasource "github.com/douyu/jupiter/pkg/datasource/file"
	"github.com/douyu/jupiter/pkg/xlog"
	kf "github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`

	logger *xlog.Logger
}

func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers: []string{"localhost:9092"},
		logger:  xlog.JupiterLogger,
	}
}

// StdProducerConfig...
func StdProducerConfig(path string) *ProducerConfig {
	var cf = DefaultProducerConfig()
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

type Producer struct {
	*ProducerConfig
	// map[partitionID]*kafka.Conn
	conns sync.Map
}

// Build ...
func (config *ProducerConfig) BuildProducer() *Producer {
	if config == nil {
		return nil
	}

	var producer Producer
	producer.ProducerConfig = config

	return &producer
}

func (p *Producer) ProduceTo(ctx context.Context, partition int, data []byte) error {
	var conn, ok = p.conns.Load(partition)
	// 没有拿到 partition 连接
	if !ok {
		c, err := kf.DialLeader(ctx, "tcp", strings.Join(p.ProducerConfig.Brokers, ","), p.ProducerConfig.Topic, partition)
		if err != nil {
			p.ProducerConfig.logger.Errorf("new producer conn failed: %v", err)
			return err
		}

		p.conns.Store(partition, c)
		conn = c
	}

	_, err := conn.(*kf.Conn).Write(data)
	return err
}

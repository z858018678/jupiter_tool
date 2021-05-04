package config

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/douyu/jupiter/pkg/client/kafka/admin"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/douyu/jupiter/pkg/conf"
	file_datasource "github.com/douyu/jupiter/pkg/datasource/file"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/fatih/structs"
)

type ProducerTopicConfigHighLevel struct {
	// This field indicates the number of acknowledgements the leader broker must receive from ISR brokers before responding to the request: 0=Broker does not send any response/ack to client, -1 or all=Broker will block until message is committed by all in sync replicas (ISRs). If there are less than min.insync.replicas (broker configuration) in the ISR set the produce request will fail.
	// range: -1 ~ 1e3
	// Type: integer
	// alias: acks
	RequestRequiredAcks int `json:"request.required.acks"`

	// Local message timeout. This value is only enforced locally and limits the time a produced message waits for successful delivery. A time of 0 is infinite. This is the maximum time librdkafka may use to deliver a message (including retries). Delivery error occurs when either the retry count or the message timeout are exceeded. The message timeout is automatically adjusted to transaction.timeout.ms if transactional.id is configured.
	// range: 0 ~ 2147483647
	// Type: integer
	// alias: delivery.timeout.ms
	// 如果配置了事务，此参数值应 <= transaction.timeout.ms
	MessageTimeoutMs int `json:"message.timeout.ms"`

	// Partitioner: random - random distribution, consistent - CRC32 hash of key (Empty and NULL keys are mapped to single partition), consistent_random - CRC32 hash of key (Empty and NULL keys are randomly partitioned), murmur2 - Java Producer compatible Murmur2 hash of key (NULL keys are mapped to single partition), murmur2_random - Java Producer compatible Murmur2 hash of key (NULL keys are randomly partitioned. This is functionally equivalent to the default partitioner in the Java Producer.), fnv1a - FNV-1a hash of key (NULL keys are mapped to single partition), fnv1a_random - FNV-1a hash of key (NULL keys are randomly partitioned).
	// Type: string
	Partitioner string `json:"partitioner"`

	// Compression codec to use for compressing message sets. inherit = inherit global compression.codec configuration.
	// range: none, gzip, snappy, lz4, zstd, inherit
	// Type: enum value
	// alias: compression.type
	CompressionCodec string `json:"compression.codec"`
}

func DefaultProducerTopicConfigHigh() ProducerTopicConfigHighLevel {
	return ProducerTopicConfigHighLevel{
		RequestRequiredAcks: -1,
		MessageTimeoutMs:    3e5,
		Partitioner:         "consistent_random",
		CompressionCodec:    "none",
	}
}

type ProducerConfigHighLevel struct {
	ConfigHighLevel `json:"config_high_level,flatten"`

	// Enables the transactional producer. The transactional.id is used to identify the same transactional producer instance across process restarts. It allows the producer to guarantee that transactions corresponding to earlier instances of the same producer have been finalized prior to starting any new transactions, and that any zombie instances are fenced off. If no transactional.id is provided, then the producer is limited to idempotent delivery (if enable.idempotence is set). Requires broker version >= 0.11.0.
	// Type: string
	TransactionalID string `json:"transactional.id"`

	// When set to true, the producer will ensure that messages are successfully produced exactly once and in the original produce order. The following configuration properties are adjusted automatically (if not modified by the user) when idempotence is enabled: max.in.flight.requests.per.connection=5 (must be less than or equal to 5), retries=INT32_MAX (must be greater than 0), acks=all, queuing.strategy=fifo. Producer instantation will fail if user-supplied configuration is incompatible.
	// Type: boolean
	EnableIdempotence bool `json:"enable.idempotence"`

	// Maximum number of messages allowed on the producer queue. This queue is shared by all topics and partitions.
	// range: 1 ~ 1e7
	// Type: integer
	QueueBufferingMaxMessages int `json:"queue.buffering.max.messages"`

	// Maximum total message size sum allowed on the producer queue. This queue is shared by all topics and partitions. This property has higher priority than queue.buffering.max.messages.
	// range 1 ~ 2147483647
	// Type: integer
	QueueBufferingMaxKbytes int `json:"queue.buffering.max.kbytes"`

	// Delay in milliseconds to wait for messages in the producer queue to accumulate before constructing message batches (MessageSets) to transmit to brokers. A higher value allows larger and more effective (less overhead, improved compression) batches of messages to accumulate at the expense of increased message delivery latency.
	// range: 0 ~ 9e5
	// Type: float; 实际上应该是 integer
	// alias: linger.ms
	QueueBufferingMaxMs int `json:"queue.buffering.max.ms"`

	// How many times to retry sending a failing Message. Note: retrying may cause reordering unless enable.idempotence is set to true.
	// range: 0 ~ 10000000
	// Type: integer
	// alias: retries
	MessageSendMaxRetries int `json:"message.send.max.retries"`
}

func DefaultProducerConfigHigh() ProducerConfigHighLevel {
	return ProducerConfigHighLevel{
		ConfigHighLevel:           DefaultConfigHigh(),
		TransactionalID:           "",
		EnableIdempotence:         false,
		QueueBufferingMaxMessages: 1e5,
		QueueBufferingMaxKbytes:   1048576,
		QueueBufferingMaxMs:       5,
		MessageSendMaxRetries:     10000000,
	}
}

type ProducerTopicConfigMediumLevel struct {
	// The ack timeout of the producer request in milliseconds. This value is only enforced by the broker and relies on request.required.acks being != 0.
	// range 1 ~ 9e5
	// Type: integer
	RequestTimeoutMs int `json:"request.timeout.ms"`

	// Compression level parameter for algorithm selected by configuration property compression.codec. Higher values will result in better compression at the cost of more CPU usage. Usable range is algorithm-dependent: [0-9] for gzip; [0-12] for lz4; only 0 for snappy; -1 = codec-dependent default compression level.
	// range: -1 ~ 12
	// Type: integer
	CompressionLevel int `json:"compression.level"`
}

func DefaultProducerTopicConfigMedium() ProducerTopicConfigMediumLevel {
	return ProducerTopicConfigMediumLevel{
		RequestTimeoutMs: 3e4,
		CompressionLevel: -1,
	}
}

type ProducerConfigMediumLevel struct {
	ConfigMediumLevel `json:"config_medium_level,flatten"`

	// The maximum amount of time in milliseconds that the transaction coordinator will wait for a transaction status update from the producer before proactively aborting the ongoing transaction. If this value is larger than the transaction.max.timeout.ms setting in the broker, the init_transactions() call will fail with ERR_INVALID_TRANSACTION_TIMEOUT. The transaction timeout automatically adjusts message.timeout.ms and socket.timeout.ms, unless explicitly configured in which case they must not exceed the transaction timeout (socket.timeout.ms must be at least 100ms lower than transaction.timeout.ms). This is also the default timeout value if no timeout (-1) is supplied to the transactional API methods.
	// range: 1e3 ~ 2147483647
	// Type: integer
	TransactionTimeoutMs int `json:"transaction.timeout.ms"`

	// The backoff time in milliseconds before retrying a protocol request.
	// range: 1 ~ 3e5
	// Type: integer
	RetryBackoffMs int `json:"retry.backoff.ms"`

	// compression codec to use for compressing message sets. This is the default value for all topics, may be overridden by the topic configuration property compression.codec.
	// range: none, gzip, snappy, lz4, zstd
	// Type: enum value
	// alias: compression.type
	CompressionCodec string `json:"compression.codec"`

	// Maximum number of messages batched in one MessageSet. The total MessageSet size is also limited by batch.size and message.max.bytes.
	// range: 1 ~ 1e6
	// Type: integer
	BatchNumMessages int `json:"batch.num.messages"`

	// Maximum size (in bytes) of all messages batched in one MessageSet, including protocol framing overhead. This limit is applied after the first message has been added to the batch, regardless of the first message's size, this is to ensure that messages that exceed batch.size are produced. The total MessageSet size is also limited by batch.num.messages and message.max.bytes.
	// range: 1 ~ 2147483647
	// Type: integer
	// NOTE: 此项配置在 https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	// 中有介绍可以使用，但在实际使用中，被被提示 'No such configuration property'
	// BatchSize int `json:"batch.size"`
}

func DefaultProducerConfigMedium() ProducerConfigMediumLevel {
	return ProducerConfigMediumLevel{
		ConfigMediumLevel:    DefaultConfigMedium(),
		TransactionTimeoutMs: 6e4,
		RetryBackoffMs:       100,
		CompressionCodec:     "none",
		BatchNumMessages:     1e4,
		// BatchSize: 1e6,
	}
}

// type TopicConfigLowLevel struct {
// 	// Custom partitioner callback (set with rd_kafka_topic_conf_set_partitioner_cb())
// 	// Type: see dedicated API
// 	PartitionerCb interface{} `json:"-"`
// }
//
// func DefaultTopicConfigLow() TopicConfigLowLevel {
// 	return TopicConfigLowLevel{}
// }

type ProducerConfigLowLevel struct {
	ConfigLowLevel `json:"config_low_level,flatten"`

	// EXPERIMENTAL: subject to change or removal. When set to true, any error that could result in a gap in the produced message series when a batch of messages fails, will raise a fatal error (ERR__GAPLESS_GUARANTEE) and stop the producer. Messages failing due to message.timeout.ms are not covered by this guarantee. Requires enable.idempotence=true.
	// Type: boolean
	EnableGaplessGuarantee bool `json:"enable.gapless.guarantee"`

	// The threshold of outstanding not yet transmitted broker requests needed to backpressure the producer's message accumulator. If the number of not yet transmitted requests equals or exceeds this number, produce request creation that would have otherwise been triggered (for example, in accordance with linger.ms) will be delayed. A lower number yields larger and more effective batches. A higher value can improve latency when using compression on slow machines.
	// range: 1 ~ 1e6
	// Type: integer
	QueueBufferingBackpressureThreshold int `json:"queue.buffering.backpressure.threshold"`

	// 	Only provide delivery reports for failed messages.
	// Type: boolean
	DeliveryReportOnlyError bool `json:"delivery.report.only.error"`

	// Delivery report callback (set with rd_kafka_conf_set_dr_cb())
	// Type: see dedicated API
	DrCb interface{} `json:"-"`

	// Delivery report callback (set with rd_kafka_conf_set_dr_msg_cb())
	// Type: see dedicated API
	DrMsgCb interface{} `json:"-"`
}

func DefaultProducerConfigLow() ProducerConfigLowLevel {
	return ProducerConfigLowLevel{
		ConfigLowLevel:                      DefaultConfigLow(),
		EnableGaplessGuarantee:              false,
		QueueBufferingBackpressureThreshold: 1,
		DeliveryReportOnlyError:             false,
	}
}

type kafkaProducerConfig struct {
	ProducerConfigHighLevel        `json:"producer_config_high_level,flatten"`
	ProducerConfigMediumLevel      `json:"producer_config_medium_level,flatten"`
	ProducerConfigLowLevel         `json:"producer_config_low_level,flatten"`
	ProducerTopicConfigHighLevel   `json:"producer_topic_config_high_level,flatten"`
	ProducerTopicConfigMediumLevel `json:"producer_topic_config_medium_level,flatten"`
	// TopicConfigLowLevel    `json:",flatten"`
}

type ProducerConfig struct {
	KafkaConfig kafkaProducerConfig `json:"kafka_config"`
	logger      *xlog.Logger
}

func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		KafkaConfig: kafkaProducerConfig{
			ProducerConfigHighLevel:        DefaultProducerConfigHigh(),
			ProducerConfigMediumLevel:      DefaultProducerConfigMedium(),
			ProducerConfigLowLevel:         DefaultProducerConfigLow(),
			ProducerTopicConfigHighLevel:   DefaultProducerTopicConfigHigh(),
			ProducerTopicConfigMediumLevel: DefaultProducerTopicConfigMedium(),
		},
		// TopicConfigLowLevel:    DefaultTopicConfigLow(),
		logger: xlog.JupiterLogger,
	}
}

// StdKafkaConfig ...
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

	if err := c.UnmarshalKey("", &cf.KafkaConfig, conf.TagName("json")); err != nil {
		xlog.Panic("unmarshal kafka config",
			xlog.String("path", path),
			xlog.Any("kafka config", path),
			xlog.String("error", err.Error()))
	}

	return cf
}

// Build ...
func (config *ProducerConfig) BuildProducer() *Producer {
	if config == nil {
		return nil
	}

	var producer Producer
	producer.ProducerConfig = config

	structs.DefaultTagName = "json"
	var m = structs.Map(config.KafkaConfig)

	var kafkaConf = make(kafka.ConfigMap)
	for k, v := range m {
		kafkaConf.SetKey(k, v)
	}

	var p, err = kafka.NewProducer(&kafkaConf)
	if err != nil {
		config.logger.Panic("new kafka producer failed", xlog.String("error", err.Error()))
		return nil
	}

	producer.Producer = p
	return &producer
}

type Producer struct {
	*ProducerConfig
	*kafka.Producer
}

func (p *Producer) ProduceTo(topic string, partition int32, offset kafka.Offset, data []byte) error {
	var err = p.Producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: partition,
				Offset:    offset,
			},
			Value: data,
		},
		nil,
	)

	if err != nil {
		return err
	}

	p.Flush(1)
	return nil
}

func (p *Producer) ReadProducedEvent() {
	e := <-p.Events()
	switch ev := e.(type) {
	case *kafka.Message:
		if ev.TopicPartition.Error != nil {
			p.ProducerConfig.logger.Error("kafka event", xlog.Any("topic partition", ev.TopicPartition), xlog.Any("data", string(ev.Value)))
			return
		}

		p.ProducerConfig.logger.Info("kafka event", xlog.Any("topic partition", ev.TopicPartition), xlog.Any("data", string(ev.Value)))
	}
}

func (p *Producer) RunMonitor() {
	go func() {
		for {
			p.ReadProducedEvent()
		}
	}()
}

func (p *Producer) NewAdminClient() *admin.Admin {
	var a, err = kafka.NewAdminClientFromProducer(p.Producer)
	if err != nil {
		p.logger.Panic("new kafka admin failed", xlog.String("error", err.Error()))
		return nil
	}

	var ac admin.Admin
	ac.AdminClient = a
	return &ac
}

func CreatePartitions(host string, topic string) (partition int32, err error) {
	var producerConfig = DefaultProducerConfig()
	producerConfig.KafkaConfig.MetadataBrokerList = host
	var p = producerConfig.BuildProducer()

	var a = p.NewAdminClient()
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	meta, err := a.GetMetadata(&topic, false, 1000)
	if err != nil {
		return
	}
	currentPartitionNum := len(meta.Topics[topic].Partitions)
	res, err := a.CreatePartitions(
		ctx,
		[]kafka.PartitionsSpecification{
			{
				Topic:      topic,
				IncreaseTo: currentPartitionNum + 1,
			},
		})
	if err != nil {
		p.logger.Errorf("Expected CreatePartitions err:%s\n", err.Error())
		return
	}
	if res == nil {
		err = errors.New("CreatePartitions res is null")
		p.logger.Errorf("Expected CreatePartitions to fail, but got result:%v\n", res)
		return
	}

	if ctx.Err() == context.DeadlineExceeded {
		err = errors.New("CreatePartitions DeadlineExceeded")
		p.logger.Errorf("Expected DeadlineExceeded, not %v", ctx.Err())
		return
	}
	return int32(currentPartitionNum + 1), nil
}

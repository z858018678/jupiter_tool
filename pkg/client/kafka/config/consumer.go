package config

import (
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/douyu/jupiter/pkg/conf"
	file_datasource "github.com/douyu/jupiter/pkg/datasource/file"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gqcn/structs"
)

type ConsumerTopicConfigHighLevel struct {
	// Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error which is retrieved by consuming messages and checking 'message->err'.
	// smallest, earliest, beginning, largest, latest, end, error
	// Type: enum value
	AutoOffsetReset string `json:"auto.offset.reset"`
}

func DefaultConsumerTopicConfigHigh() ConsumerTopicConfigHighLevel {
	return ConsumerTopicConfigHighLevel{
		AutoOffsetReset: "earliest",
	}
}

type ConsumerConfigHighLevel struct {
	ConfigHighLevel `json:"config_high_level,flatten"`

	// Client group id string. All clients sharing the same group.id belong to the same group.
	// Type: string
	GroupID string `json:"group.id"`

	// Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance. The allowed range is configured with the broker configuration properties group.min.session.timeout.ms and group.max.session.timeout.ms. Also see max.poll.interval.ms.
	// range: 1 ~ 3.6e6
	// Type: integer
	SessionTimeoutMs int `json:"session.timeout.ms"`

	// Maximum allowed time between calls to consume messages (e.g., rd_kafka_consumer_poll()) for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member. Warning: Offset commits may be not possible at this point. Note: It is recommended to set enable.auto.offset.store=false for long-time processing applications and then explicitly store offsets (using offsets_store()) after message processing, to make sure offsets are not auto-committed prior to processing has finished. The interval is checked two times per second. See KIP-62 for more information.
	// range: 1 ~ 86400e3
	// Type: integer
	MaxPollIntervalMs int `json:"max.poll.interval.ms"`

	// Automatically and periodically commit offsets in the background. Note: setting this to false does not prevent the consumer from fetching previously committed start offsets. To circumvent this behaviour set specific start offsets per partition in the call to assign().
	// Type: boolean
	EnableAutoCommit bool `json:"enable.auto.commit"`

	// Automatically store offset of last message provided to application. The offset store is an in-memory store of the next offset to (auto-)commit for each partition.
	// Type: boolean
	EnableAutoOffsetStore bool `json:"enable.auto.offset.store"`

	// Controls how to read messages written transactionally: read_committed - only return transactional messages which have been committed. read_uncommitted - return all messages, even transactional messages which have been aborted.
	// range: read_uncommitted, read_committed
	// Type: enum value
	IsolationLevel string `json:"isolation.level"`
}

func DefaultConsumerConfigHigh() ConsumerConfigHighLevel {
	return ConsumerConfigHighLevel{
		ConfigHighLevel:       DefaultConfigHigh(),
		GroupID:               "",
		SessionTimeoutMs:      1e4,
		MaxPollIntervalMs:     3e5,
		EnableAutoCommit:      true,
		EnableAutoOffsetStore: true,
		IsolationLevel:        "read_committed",
	}
}

// type TopicConfigMediumLevel struct{}
//
// func DefaultTopicConfigMedium() TopicConfigMediumLevel {
// 	return TopicConfigMediumLevel{}
// }

type ConsumerConfigMediumLevel struct {
	ConfigMediumLevel `json:"config_medium_level,flatten"`

	// Enable static group membership. Static group members are able to leave and rejoin a group within the configured session.timeout.ms without prompting a group rebalance. This should be used in combination with a larger session.timeout.ms to avoid group rebalances caused by transient unavailability (e.g. process restarts). Requires broker version >= 2.3.0.
	// Type: string
	GroupInstanceId string `json:"group.instance.id"`

	// Name of partition assignment strategy to use when elected group leader assigns partitions to group members.
	// Type: string
	PartitionAssignmentStrategy string `json:"partition.assignment.strategy"`

	// The frequency in milliseconds that the consumer offsets are committed (written) to offset storage. (0 = disable). This setting is used by the high-level consumer.
	// range: 0 ~ 86400e3
	// Type: integer
	AutoCommitIntervalMs int `json:"auto.commit.interval.ms"`

	// Minimum number of messages per topic+partition librdkafka tries to maintain in the local consumer queue.
	// range 1 ~ 1e7
	// Type: integer
	QueuedMinMessages int `json:"queued.min.messages"`

	// Maximum number of kilobytes of queued pre-fetched messages in the local consumer queue. If using the high-level consumer this setting applies to the single consumer queue, regardless of the number of partitions. When using the legacy simple consumer or when separate partition queues are used this setting applies per partition. This value may be overshot by fetch.message.max.bytes. This property has higher priority than queued.min.messages.
	// range: 1 ~ 2097151
	// Type: integer
	QueuedMaxMessagesKbytes int `json:"queued.max.messages.kbytes"`

	// Maximum time the broker may wait to fill the Fetch response with fetch.min.bytes of messages.
	// range: 1 ~ 1e9
	// Type: integer
	// alias: max.partition.fetch.bytes
	FetchMessageMaxBytes int `json:"fetch.message.max.bytes"`

	// Maximum amount of data the broker shall return for a Fetch request. Messages are fetched in batches by the consumer and if the first message batch in the first non-empty partition of the Fetch request is larger than this value, then the message batch will still be returned to ensure the consumer can make progress. The maximum message batch size accepted by the broker is defined via message.max.bytes (broker config) or max.message.bytes (broker topic config). fetch.max.bytes is automatically adjusted upwards to be at least message.max.bytes (consumer config).
	// range: 0 ~ 2147483135
	// Type: integer
	FetchMaxBytes int `json:"fetch.max.bytes"`

	// How long to postpone the next fetch request for a topic+partition in case of a fetch error.
	// range: 0 ~ 3e6
	// Type: integer
	FetchErrorBackoffMs int `json:"fetch.error.backoff.ms"`

	// Verify CRC32 of consumed messages, ensuring no on-the-wire or on-disk corruption to the messages occurred. This check comes at slightly increased CPU usage.
	// Type: boolean
	CheckCrcs bool `json:"check.crcs"`
}

func DefaultConsumerConfigMedium() ConsumerConfigMediumLevel {
	return ConsumerConfigMediumLevel{
		ConfigMediumLevel:           DefaultConfigMedium(),
		GroupInstanceId:             "",
		PartitionAssignmentStrategy: "range,roundrobin",
		AutoCommitIntervalMs:        5e3,
		QueuedMinMessages:           1e5,
		QueuedMaxMessagesKbytes:     65536,
		FetchMessageMaxBytes:        1048576,
		FetchMaxBytes:               52428800,
		FetchErrorBackoffMs:         500,
	}
}

type ConsumerTopicConfigLowLevel struct {
	// Maximum number of messages to dispatch in one rd_kafka_consume_callback*() call (0 = unlimited)
	// range: 0 ~ 1e6
	// Type: integer
	ConsumeCallbackMaxMessages int `json:"consume.callback.max.messages"`
}

func DefaultConsumerTopicConfigLow() ConsumerTopicConfigLowLevel {
	return ConsumerTopicConfigLowLevel{
		ConsumeCallbackMaxMessages: 0,
	}
}

type ConsumerConfigLowLevel struct {
	ConfigLowLevel `json:"config_low_level,flatten"`

	// Group session keepalive heartbeat interval.
	// range: 1 ~ 3.6e6
	// Type: integer
	HeartbeatIntervalMs int `json:"heartbeat.interval.ms"`

	// Group protocol type
	// Type: string
	GroupProtocolType string `json:"group.protocol.type"`

	// How often to query for the current client group coordinator. If the currently assigned coordinator is down the configured query interval will be divided by ten to more quickly recover in case of coordinator reassignment.
	// range: 1 ~ 3.6e6
	// Type: integer
	CoordinatorQueryIntervalMs int `json:"coordinator.query.interval.ms"`

	// Maximum time the broker may wait to fill the Fetch response with fetch.min.bytes of messages.
	// range: 0 ~ 3e6
	// Type: integer
	FetchWaitMaxMs int `json:"fetch.wait.max.ms"`

	// Minimum number of bytes the broker responds with. If fetch.wait.max.ms expires the accumulated data will be sent to the client regardless of this setting.
	// range: 1 ~ 1e8
	// Type: integer
	FetchMinBytes int `json:"fetch.min.bytes"`

	// Message consume callback (set with rd_kafka_conf_set_consume_cb())
	// Type: see dedicated API
	ConsumeCb interface{} `json:"-"`

	// Called after consumer group has been rebalanced (set with rd_kafka_conf_set_rebalance_cb())
	// Type: see dedicated API
	RebalanceCb interface{} `json:"-"`

	// Offset commit result propagation callback. (set with rd_kafka_conf_set_offset_commit_cb())
	// Type: see dedicated API
	OffsetCommitCb interface{} `json:"-"`

	// Emit RD_KAFKA_RESP_ERR__PARTITION_EOF event whenever the consumer reaches the end of a partition.
	// Type: boolean
	EnablePartitionEof bool `json:"enable.partition.eof"`

	// Allow automatic topic creation on the broker when subscribing to or assigning non-existent topics. The broker must also be configured with auto.create.topics.enable=true for this configuraiton to take effect. Note: The default value (false) is different from the Java consumer (true). Requires broker version >= 0.11.0.0, for older broker versions only the broker configuration applies.
	// Type: boolean
	// NOTE: 此项配置在 https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	// 中有介绍可以使用，但在实际使用中，被被提示 'No such configuration property'
	// AllowAutoCreateTopics bool `json:"allow.auto.create.topics"`

	// A rack identifier for this client. This can be any string value which indicates where this client is physically located. It corresponds with the broker config broker.rack.
	// Type: string
	ClientRack string `json:"client.rack"`

	// Enable the Events() channel. Messages and events will be pushed on the Events() channel and the Poll() interface will be disabled.
	EnableEventsChannel bool `json:"go.events.channel.enable"`
}

func DefaultConsumerConfigLow() ConsumerConfigLowLevel {
	return ConsumerConfigLowLevel{
		ConfigLowLevel:             DefaultConfigLow(),
		HeartbeatIntervalMs:        3e3,
		GroupProtocolType:          "consumer",
		CoordinatorQueryIntervalMs: 6e5,
		FetchWaitMaxMs:             500,
		FetchMinBytes:              1,
		EnablePartitionEof:         false,
		ClientRack:                 "",
		EnableEventsChannel:		false,
		// AllowAutoCreateTopics: false,
	}
}

type kafkaConsumerConfig struct {
	ConsumerConfigHighLevel      `json:"consumer_config_high_level,flatten"`
	ConsumerConfigMediumLevel    `json:"consumer_config_medium_level,flatten"`
	ConsumerConfigLowLevel       `json:"consumer_config_low_level,flatten"`
	ConsumerTopicConfigHighLevel `json:"consumer_topic_config_high_level,flatten"`
	// 	TopicConfigMediumLevel `json:",flatten"`
	ConsumerTopicConfigLowLevel `json:"consumer_topic_config_low_level,flatten"`
}

type ConsumerConfig struct {
	KafkaConfig kafkaConsumerConfig `json:"kafka_config"`
	logger      *xlog.Logger
}

func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		KafkaConfig: kafkaConsumerConfig{
			ConsumerConfigHighLevel:      DefaultConsumerConfigHigh(),
			ConsumerConfigMediumLevel:    DefaultConsumerConfigMedium(),
			ConsumerConfigLowLevel:       DefaultConsumerConfigLow(),
			ConsumerTopicConfigHighLevel: DefaultConsumerTopicConfigHigh(),
			// ConsumerTopicConfigMediumLevel: DefaultTopicConfigMedium(),
			ConsumerTopicConfigLowLevel: DefaultConsumerTopicConfigLow(),
		},

		logger: xlog.JupiterLogger,
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

	if err := c.UnmarshalKey("", &cf.KafkaConfig, conf.TagName("json")); err != nil {
		xlog.Panic("unmarshal kafka config",
			xlog.String("path", path),
			xlog.Any("kafka config", path),
			xlog.String("error", err.Error()))
	}

	return cf
}

// Build ...
func (config *ConsumerConfig) BuildConsumer() *Consumer {
	if config == nil {
		return nil
	}

	var consumer Consumer
	consumer.ConsumerConfig = config

	structs.DefaultTagName = "json"
	var m = structs.Map(config.KafkaConfig)

	var kafkaConf = make(kafka.ConfigMap)
	for k, v := range m {
		kafkaConf.SetKey(k, v)
	}

	var c, err = kafka.NewConsumer(&kafkaConf)
	if err != nil {
		config.logger.Panic("new kafka consumer failed", xlog.String("error", err.Error()))
		return nil
	}

	consumer.Consumer = c
	return &consumer
}

type Consumer struct {
	*ConsumerConfig
	*kafka.Consumer
}

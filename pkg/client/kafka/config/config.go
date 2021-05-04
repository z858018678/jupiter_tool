package config

type ConfigHighLevel struct {
	// Initial list of brokers as a CSV list of broker host or host:port. The application may also use rd_kafka_brokers_add() to add brokers during runtime.
	// kafka broker 地址
	// 对应为 metadata.broker.list 或 bootstrap.servers
	// alias: bootstrap.servers
	MetadataBrokerList string `json:"metadata.broker.list"`

	// librdkafka statistics emit interval. The application also needs to register a stats callback using rd_kafka_conf_set_stats_cb(). The granularity is 1000ms. A value of 0 disables statistics.
	// range: 0 ~ 86400e3
	// Type: integer
	StatisticsIntervalMs int `json:"statistics.interval.ms"`

	// high	Request broker's supported API versions to adjust functionality to available protocol features. If set to false, or the ApiVersionRequest fails, the fallback version broker.version.fallback will be used. NOTE: Depends on broker version >=0.10.0. If the request is not supported by (an older) broker the broker.version.fallback fallback is used.
	// Type: boolean
	ApiVersionRequest bool `json:"api.version.request"`

	// Protocol used to communicate with brokers.
	// range plaintext, ssl, sasl_plaintext, sasl_ssl
	// Type: enum value
	SecurityProtocol string `json:"security.protocol"`

	// SASL mechanism to use for authentication. Supported: GSSAPI, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, OAUTHBEARER. NOTE: Despite the name only one mechanism must be configured.
	// Type: string
	// alias: sasl.mechanism
	SaslMechanisms string `json:"sasl.mechanisms"`

	// SASL username for use with the PLAIN and SASL-SCRAM-.. mechanisms
	// Type: string
	SaslUsername string `json:"sasl.username"`

	// SASL password for use with the PLAIN and SASL-SCRAM-.. mechanism
	// Type: string
	SaslPassword string `json:"sasl.password"`
}

func DefaultConfigHigh() ConfigHighLevel {
	return ConfigHighLevel{
		MetadataBrokerList:   "localhost",
		StatisticsIntervalMs: 0,
		ApiVersionRequest:    true,
		SecurityProtocol:     "PLAINTEXT",
		SaslMechanisms:       "PLAIN",
		SaslUsername:         "",
		SaslPassword:         "",
	}
}

type ConfigMediumLevel struct {
	// Maximum Kafka protocol request message size. Due to differing framing overhead between protocol versions the producer is unable to reliably enforce a strict max message limit at produce time and may exceed the maximum size by one message in protocol ProduceRequests, the broker will enforce the the topic's max.message.bytes limit (see Apache Kafka documentation).
	// range: 1e3 ~ 1e9
	// Type: integer
	MessageMaxBytes int `json:"message.max.bytes"`

	// Maximum Kafka protocol response message size. This serves as a safety precaution to avoid memory exhaustion in case of protocol hickups. This value must be at least fetch.max.bytes + 512 to allow for protocol overhead; the value is adjusted automatically unless the configuration property is explicitly set.
	// range: 1e3 ~ 2^31-1
	// Type: integer
	ReceiveMessageMaxBytes int `json:"receive.message.max.bytes"`

	// A comma-separated list of debug contexts to enable. Detailed Producer debugging: broker,topic,msg. Consumer: consumer,cgrp,topic,fetch
	// range: generic, broker, topic, metadata, feature, queue, msg, protocol, cgrp, security, fetch, interceptor, plugin, consumer, admin, eos, mock, all
	// Type: CSV flags
	Debug string `json:"debug"`

	// The initial time to wait before reconnecting to a broker after the connection has been closed. The time is increased exponentially until reconnect.backoff.max.ms is reached. -25% to +50% jitter is applied to each reconnect backoff. A value of 0 disables the backoff and reconnects immediately.
	// range 0 ~ 3.6e6
	// Type: integer
	ReconnectBackoffMs int `json:"reconnect.backoff.ms"`

	// The maximum time to wait before reconnecting to a broker after the connection has been closed.
	// range 0 ~ 3.6e6
	// Type: integer
	ReconnectBackoffMaxMs int `json:"reconnect.backoff.max.ms"`

	// Dictates how long the broker.version.fallback fallback is used in the case the ApiVersionRequest fails. NOTE: The ApiVersionRequest is only issued when a new connection to the broker is made (such as after an upgrade).
	// range 0 ~ 6048e5
	// Type: integer
	ApiVersionFallbackMs int `json:"api.version.fallback.ms"`

	// Older broker versions (before 0.10.0) provide no way for a client to query for supported protocol features (ApiVersionRequest, see api.version.request) making it impossible for the client to know what features it may use. As a workaround a user may set this property to the expected broker version and the client will automatically adjust its feature set accordingly if the ApiVersionRequest fails (or is disabled). The fallback broker version will be used for api.version.fallback.ms. Valid values are: 0.9.0, 0.8.2, 0.8.1, 0.8.0. Any other value >= 0.10, such as 0.10.2.1, enables ApiVersionRequests.
	// Type: string
	BrockerVersionFallback string `json:"broker.version.fallback"`
}

func DefaultConfigMedium() ConfigMediumLevel {
	return ConfigMediumLevel{
		MessageMaxBytes:        1e6,
		ReceiveMessageMaxBytes: 1e8,
		Debug:                  "admin",
		ReconnectBackoffMs:     1e2,
		ReconnectBackoffMaxMs:  1e4,
		ApiVersionFallbackMs:   0,
		BrockerVersionFallback: "0.10.0",
	}
}

type ConfigLowLevel struct {
	// Indicates the builtin features for this build of librdkafka. An application can either query this value or attempt to set it with its list of required features to check for library support.
	// Type: CSV flags
	BuiltInFeatures string `json:"builtin.features"`

	// client identifier
	ClientID string `json:"client.id"`

	// Maximum size for message to be copied to buffer. Messages larger than this will be passed by reference (zero-copy) at the expense of larger iovecs.
	// range: 0 ~ 1e9
	// Type: integer
	MessageCopyMaxBytes int `json:"message.copy.max.bytes"`

	// Maximum number of in-flight requests per broker connection. This is a generic property applied to all broker communication, however it is primarily relevant to produce requests. In particular, note that other mechanisms limit the number of outstanding consumer fetch request per broker to one.
	// range: 1 ~ 1e6
	// Type: integer
	// alias: max.in.flight
	// 限制客户端在单个连接上能够发送的未响应请求的个数。设置此值是1表示kafka broker在响应请求之前client不能再向同一个broker发送请求。注意：设置此参数是为了避免消息乱序
	// 如果要开启事务，此值必须 <= 5
	MaxInFlightRequestsPerConnection int `json:"max.in.flight.requests.per.connection"`

	// Non-topic request timeout in milliseconds. This is for metadata requests, etc.
	// range: 10 ~ 9e5
	// Type: integer
	MetadataRequestTimeoutMs int `json:"metadata.request.timeout.ms"`

	// Period of time in milliseconds at which topic and broker metadata is refreshed in order to proactively discover any new brokers, topics, partitions or partition leader changes. Use -1 to disable the intervalled refresh (not recommended). If there are no locally referenced topics (no topic objects created, no messages produced, no subscription or no assignment) then only the broker list will be refreshed every interval but no more often than every 10s.
	// range: -1 ~ 3.6e6
	// Type: integer
	TopicMetadataRefreshIntervalMs int `json:"topic.metadata.refresh.interval.ms"`

	// Metadata cache max age. Defaults to topic.metadata.refresh.interval.ms * 3
	// range: 1 ~ 86400e3
	// Type: integer
	MetadataMaxAgeMs int `json:"metadata.max.age.ms"`

	// When a topic loses its leader a new metadata request will be enqueued with this initial interval, exponentially increasing until the topic metadata has been refreshed. This is used to recover quickly from transitioning leader brokers.
	// range 1 ~ 6e5
	// Type: integer
	TopicMetadataRefreshFastIntervalMs int `json:"topic.metadata.refresh.fast.interval.ms"`

	// Sparse metadata requests (consumes less network bandwidth)
	// Type: boolean
	TopicMetadataRefreshSparse bool `json:"topic.metadata.refresh.sparse"`

	// Apache Kafka topic creation is asynchronous and it takes some time for a new topic to propagate throughout the cluster to all brokers. If a client requests topic metadata after manual topic creation but before the topic has been fully propagated to the broker the client is requesting metadata from, the topic will seem to be non-existent and the client will mark the topic as such, failing queued produced messages with ERR__UNKNOWN_TOPIC. This setting delays marking a topic as non-existent until the configured propagation max time has passed. The maximum propagation time is calculated from the time the topic is first referenced in the client, e.g., on produce().
	// range 0 ~ 3.6e6
	// Type: integer
	// NOTE: 此项配置在 https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	// 中有介绍可以使用，但在实际使用中，被被提示 'No such configuration property'
	// TopicMetadataPropagationMaxMs int `json:"topic.metadata.propagation.max.ms"`

	// Topic blacklist, a comma-separated list of regular expressions for matching topic names that should be ignored in broker metadata information as if the topics did not exist.
	// Type: pattern list
	TopicBlacklist string `json:"topic.blacklist"`
	// Default timeout for network requests. Producer: ProduceRequests will use the lesser value of socket.timeout.ms and remaining message.timeout.ms for the first message in the batch. Consumer: FetchRequests will use fetch.wait.max.ms + socket.timeout.ms. Admin: Admin requests will use socket.timeout.ms or explicitly set rd_kafka_AdminOptions_set_operation_timeout() value.
	// range 10 ~ 3e5
	// Type: integer
	SocketTimeoutMs int `json:"socket.timeout.ms"`

	// Broker socket send buffer size. System default is used if 0.
	// range: 0 ~ 1e8
	// Type: integer
	SocketSendBufferBytes int `json:"socket.send.buffer.bytes"`

	// Broker socket receive buffer size. System default is used if 0.
	// range: 0 ~ 1e8
	// Type: integer
	SocketReceiveBufferBytes int `json:"socket.receive.buffer.bytes"`

	// Enable TCP keep-alives (SO_KEEPALIVE) on broker sockets
	// Type: boolean
	SocketKeepaliveEnable bool `json:"socket.keepalive.enable"`

	// Disable the Nagle algorithm (TCP_NODELAY) on broker sockets.
	// Type: boolean
	SocketNagleDisable bool `json:"socket.nagle.disable"`

	// Disconnect from broker when this number of send failures (e.g., timed out requests) is reached. Disable with 0. WARNING: It is highly recommended to leave this setting at its default value of 1 to avoid the client and broker to become desynchronized in case of request timeouts. NOTE: The connection is automatically re-established.
	// range 0 ~ 1e6
	// Type: integer
	SocketMaxFails int `json:"socket.max.fails"`

	// How long to cache the broker address resolving results (milliseconds).
	// range 0 ~ 86400e3
	// Type: integer
	BrockerAddressTTL int `json:"broker.address.ttl"`

	// Allowed broker IP address families: any, v4, v6
	// range: any, v4, v6
	// Type: enum value
	BrockerAddressFamily string `json:"broker.address.family"`
	// See rd_kafka_conf_set_events()
	// range: 0 ~ 2^31-1
	// Type: integer
	EnabledEvents int `json:"enabled_events"`

	// Error callback (set with rd_kafka_conf_set_error_cb())
	// Type: see dedicated API
	ErrorCb interface{} `json:"-"`

	// Throttle callback (set with rd_kafka_conf_set_throttle_cb())
	// Type: see dedicated API
	ThrottleCb interface{} `json:"-"`

	// Statistics callback (set with rd_kafka_conf_set_stats_cb())
	// Type: see dedicated API
	StatsCb interface{} `json:"-"`

	// Log callback (set with rd_kafka_conf_set_log_cb())
	// Type: see dedicated API
	LogCb interface{} `json:"-"`

	// Logging level (syslog(3) levels)
	// range: 0 ~ 7
	// Type: integer
	LogLevel int `json:"log_level"`

	// Disable spontaneous log_cb from internal librdkafka threads, instead enqueue log messages on queue set with rd_kafka_set_log_queue() and serve log callbacks or events through the standard poll APIs. NOTE: Log messages will linger in a temporary queue until the log queue has been set.
	// Type: boolean
	LogQueue bool `json:"log.queue"`

	// Print internal thread name in log messages (useful for debugging librdkafka internals)
	// Type: boolean
	LogThreadName bool `json:"log.thread.name"`

	// If enabled librdkafka will initialize the POSIX PRNG with srand(current_time.milliseconds) on the first invocation of rd_kafka_new(). If disabled the application must call srand() prior to calling rd_kafka_new().
	// Type: boolean
	EnableRandomSeed bool `json:"enable.random.seed"`

	// Log broker disconnects. It might be useful to turn this off when interacting with 0.9 brokers with an aggressive connection.max.idle.ms value.
	// Type: boolean
	LogConnectionClose bool `json:"log.connection.close"`

	// Background queue event callback (set with rd_kafka_conf_set_background_event_cb())
	// Type: see dedicated API
	BackgroundEventCb interface{} `json:"-"`

	// Socket creation callback to provide race-free CLOEXEC
	// Type: see dedicated API
	SocketCb interface{} `json:"-"`

	// 	Socket connect callback
	// Type: see dedicated API
	ConnectCb interface{} `json:"-"`

	// Socket close callback
	// Type: see dedicated API
	ClosesocketCb interface{} `json:"-"`

	// 	File open callback to provide race-free CLOEXEC
	// Type: see dedicated API
	OpenCb interface{} `json:"-"`

	// Application opaque (set with rd_kafka_conf_set_opaque())
	// Type: see dedicated API
	Opaque interface{} `json:"-"`

	// Default topic configuration for automatically subscribed topics
	// Type: see dedicated API
	DefaultTopicConf interface{} `json:"-"`

	// Signal that librdkafka will use to quickly terminate on rd_kafka_destroy(). If this signal is not set then there will be a delay before rd_kafka_wait_destroyed() returns true as internal threads are timing out their system calls. If this signal is set however the delay will be minimal. The application should mask this signal as an internal signal handler is installed.
	// range 0 ~ 128
	// Type: integer
	InternalTerminaltionSignal int `json:"internal.termination.signal"`

	// Timeout for broker API version requests.
	// range 1 ~ 3e5
	// Type: integer
	ApiVersionRequestTimeoutMs int `json:"api.version.request.timeout.ms"`

	// A cipher suite is a named combination of authentication, encryption, MAC and key exchange algorithm used to negotiate the security settings for a network connection using TLS or SSL network protocol. See manual page for ciphers(1) and `SSL_CTX_set_cipher_list(3).
	// Type: string
	SslCipherSuites string `json:"ssl.cipher.suites"`

	// The supported-curves extension in the TLS ClientHello message specifies the curves (standard/named, or 'explicit' GF(2^k) or GF(p)) the client is willing to have the server use. See manual page for SSL_CTX_set1_curves_list(3). OpenSSL >= 1.0.2 required.
	// Type: string
	SslCurvesList string `json:"ssl.curves.list"`

	// The client uses the TLS ClientHello signature_algorithms extension to indicate to the server which signature/hash algorithm pairs may be used in digital signatures. See manual page for SSL_CTX_set1_sigalgs_list(3). OpenSSL >= 1.0.2 required.
	// Type: string
	SslSigalgsList string `json:"ssl.sigalgs.list"`

	// Path to client's private key (PEM) used for authentication.
	// Type: string
	SslKeyLocation string `json:"ssl.key.location"`

	// Private key passphrase (for use with ssl.key.location and set_ssl_cert())
	// Type: string
	SslKeyPassword string `json:"ssl.key.password"`

	// Client's private key string (PEM format) used for authentication.
	// Type: string
	SslKeyPem string `json:"ssl.key.pem"`

	// 	Client's private key as set by rd_kafka_conf_set_ssl_cert()
	// Type: see dedicated API
	SslKey interface{} `json:"-"`

	// Path to client's public key (PEM) used for authentication.
	// Type: string
	SslCertificateLocaltion string `json:"ssl.certificate.location"`

	// Client's public key string (PEM format) used for authentication.
	// Type: string
	SslCertificatePem string `json:"ssl.certificate.pem"`

	// Client's public key as set by rd_kafka_conf_set_ssl_cert()
	// Type: see dedicated API
	SslCertificate interface{} `json:"-"`

	// File or directory path to CA certificate(s) for verifying the broker's key. Defaults: On Windows the system's CA certificates are automatically looked up in the Windows Root certificate store. On Mac OSX it is recommended to install openssl using Homebrew, to provide CA certificates. On Linux install the distribution's ca-certificates package. If OpenSSL is statically linked or ssl.ca.location is set to probe a list of standard paths will be probed and the first one found will be used as the default CA certificate location path. If OpenSSL is dynamically linked the OpenSSL library's default path will be used (see OPENSSLDIR in openssl version -a).
	// Type: string
	SslCaLocation string `json:"ssl.ca.location"`

	// 	CA certificate as set by rd_kafka_conf_set_ssl_cert()
	// Type: see dedicated API
	SslCa interface{} `json:"-"`

	// Path to CRL for verifying broker's certificate validity.
	// Type: string
	SslCrlLocation string `json:"ssl.crl.location"`

	// Path to client's keystore (PKCS#12) used for authentication.
	// Type: string
	SslKeystoreLocation string `json:"ssl.keystore.location"`

	// Client's keystore (PKCS#12) password.
	// Type: string
	SslKeystorePassword string `json:"ssl.keystore.password"`

	// Enable OpenSSL's builtin broker (server) certificate verification. This verification can be extended by the application by implementing a certificate_verify_cb.
	// Type: boolean
	EnableSslCertificateVerification bool `json:"enable.ssl.certificate.verification"`

	// Endpoint identification algorithm to validate broker hostname using broker certificate. https - Server (broker) hostname verification as specified in RFC2818. none - No endpoint verification. OpenSSL >= 1.0.2 required.
	// range: none, https
	// Type: enum value
	SslEndpointIdentificationAlgorithm string `json:"ssl.endpoint.identification.algorithm"`

	// 	Callback to verify the broker certificate chain.
	// Type: see dedicated API
	SslCertificateVerifyCb interface{} `json:"-"`

	// Kerberos principal name that Kafka runs as, not including /hostname@REALM
	// Type: string
	SaslKerberosServiceName string `json:"sasl.kerberos.service.name"`

	// This client's Kerberos principal name. (Not supported on Windows, will use the logon user's principal).
	// Type: string
	SaslKerberosPrincipal string `json:"sasl.kerberos.principal"`

	// Shell command to refresh or acquire the client's Kerberos ticket. This command is executed on client creation and every sasl.kerberos.min.time.before.relogin (0=disable). %{config.prop.name} is replaced by corresponding config object value.
	// Type: string
	SaslKerberosKinitCmd string `json:"sasl.kerberos.kinit.cmd"`

	// Path to Kerberos keytab file. This configuration property is only used as a variable in sasl.kerberos.kinit.cmd as ... -t "%{sasl.kerberos.keytab}".
	// Type: string
	SaslKerberosKeytab string `json:"sasl.kerberos.keytab"`

	// Minimum time in milliseconds between key refresh attempts. Disable automatic key refresh by setting this property to 0.
	// range: 0 ~ 86400e3
	// Type: integer
	SaslKerberosMinTimeBeforeRelogin int `json:"sasl.kerberos.min.time.before.relogin"`

	// SASL/OAUTHBEARER configuration. The format is implementation-dependent and must be parsed accordingly. The default unsecured token implementation (see https://tools.ietf.org/html/rfc7515#appendix-A.5) recognizes space-separated name=value pairs with valid names including principalClaimName, principal, scopeClaimName, scope, and lifeSeconds. The default value for principalClaimName is "sub", the default value for scopeClaimName is "scope", and the default value for lifeSeconds is 3600. The scope value is CSV format with the default value being no/empty scope. For example: principalClaimName=azp principal=admin scopeClaimName=roles scope=role1,role2 lifeSeconds=600. In addition, SASL extensions can be communicated to the broker via extension_NAME=value. For example: principal=admin extension_traceId=123
	// Type: string
	SaslOauthbearerConfig string `json:"sasl.oauthbearer.config"`

	// Enable the builtin unsecure JWT OAUTHBEARER token handler if no oauthbearer_refresh_cb has been set. This builtin handler should only be used for development or testing, and not in production.
	// Type: boolean
	EnableSaslOauthbearerUnsecureJwt bool `json:"enable.sasl.oauthbearer.unsecure.jwt"`

	// SASL/OAUTHBEARER token refresh callback (set with rd_kafka_conf_set_oauthbearer_token_refresh_cb(), triggered by rd_kafka_poll(), et.al. This callback will be triggered when it is time to refresh the client's OAUTHBEARER token.
	// Type: see dedicated API
	OauthbearerTokenRefreshCb interface{} `json:"-"`

	// List of plugin libraries to load (; separated). The library search path is platform dependent (see dlopen(3) for Unix and LoadLibrary() for Windows). If no filename extension is specified the platform-specific extension (such as .dll or .so) will be appended automatically.
	// Type: string
	PluginLibraryPaths string `json:"plugin.library.paths"`

	// Interceptors added through rd_kafka_conf_interceptor_add_..() and any configuration handled by interceptors.
	// Type: see dedicated API
	Interceptors interface{} `json:"-"`
}

func DefaultConfigLow() ConfigLowLevel {
	return ConfigLowLevel{
		// BuiltInFeatures:                    "gzip, snappy, ssl, sasl, regex, lz4, sasl_gssapi, sasl_plain, sasl_scram, plugins, zstd, sasl_oauthbearer",
		BuiltInFeatures:                    "gzip, snappy, ssl, sasl, regex, lz4, sasl_plain, sasl_scram, plugins, zstd, sasl_oauthbearer",
		ClientID:                           "rdkafka",
		MessageCopyMaxBytes:                65535,
		MaxInFlightRequestsPerConnection:   1e6,
		MetadataRequestTimeoutMs:           6e4,
		TopicMetadataRefreshIntervalMs:     3e5,
		MetadataMaxAgeMs:                   9e5,
		TopicMetadataRefreshFastIntervalMs: 250,
		TopicMetadataRefreshSparse:         true,
		SocketTimeoutMs:                    6e4,
		SocketSendBufferBytes:              0,
		SocketReceiveBufferBytes:           0,
		SocketKeepaliveEnable:              false,
		SocketNagleDisable:                 false,
		SocketMaxFails:                     1,
		BrockerAddressTTL:                  1e3,
		BrockerAddressFamily:               "any",
		EnabledEvents:                      0,
		LogLevel:                           6,
		LogQueue:                           false,
		LogThreadName:                      true,
		EnableRandomSeed:                   true,
		LogConnectionClose:                 true,
		InternalTerminaltionSignal:         0,
		ApiVersionRequestTimeoutMs:         1e4,
		SslCipherSuites:                    "",
		SslCurvesList:                      "",
		SslSigalgsList:                     "",
		SslKeyLocation:                     "",
		SslKeyPassword:                     "",
		SslKeyPem:                          "",
		SslCaLocation:                      "",
		SslCrlLocation:                     "",
		SslKeystoreLocation:                "",
		EnableSslCertificateVerification:   true,
		SslEndpointIdentificationAlgorithm: "none",
		// 	SaslKerberosServiceName:            "kafka",
		// 	SaslKerberosPrincipal:              "kafkaclient",
		// 	SaslKerberosKinitCmd:               "kinit -R -t \"%{sasl.kerberos.keytab}\" -k %{sasl.kerberos.principal} || kinit -t \"%{sasl.kerberos.keytab}\" -k %{sasl.kerberos.principal}",
		// 	SaslKerberosKeytab:                 "",
		// 	SaslKerberosMinTimeBeforeRelogin:   6e4,
		EnableSaslOauthbearerUnsecureJwt: false,
		PluginLibraryPaths:               "",

		// TopicMetadataPropagationMaxMs: 3e5,
	}
}

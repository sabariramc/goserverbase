package env

const (
	// ServiceName is the environment variable for the name of the service.
	ServiceName = "SERVICE_NAME"

	// LogLevel is the environment variable for the log level setting.
	LogLevel = "LOG__LEVEL"
	// LogFileTrace is the environment variable for enabling file trace logging.
	LogFileTrace = "LOG__FILE_TRACE"
	// LogWriter is the environment variable for specifying the log writer type.
	LogWriter = "LOG__WRITER"

	// NotifierTopic is the environment variable for the notifier topic.
	NotifierTopic = "NOTIFIER__TOPIC"

	// KafkaBroker is the environment variable for the Kafka broker address.
	KafkaBroker = "KAFAK__BROKER"
	// KafkaProducerAcknowledge is the environment variable for Kafka producer acknowledgment setting.
	KafkaProducerAcknowledge = "KAFKA__PRODUCER__ACKNOWLEDGE"
	// KafkaProducerMaxBuffer is the environment variable for the maximum buffer size for Kafka producer.
	KafkaProducerMaxBuffer = "KAFKA__PRODUCER__MAX_BUFFER"
	// KafkaProducerAutoFlushInterval is the environment variable for the auto flush interval for Kafka producer.
	KafkaProducerAutoFlushInterval = "KAFKA__PRODUCER__AUTO_FLUSH_INTERVAL"
	// KafkaProducerAsync is the environment variable for enabling asynchronous Kafka producer.
	KafkaProducerAsync = "KAFKA__PRODUCER__ASYNC"
	// KafkaProducerBatch is the environment variable for enabling batch processing in Kafka producer.
	KafkaProducerBatch = "KAFKA__PRODUCER__BATCH"
	// KafkaConsumerGroupID is the environment variable for the Kafka consumer group ID.
	KafkaConsumerGroupID = "KAFKA__CONSUMER__GROUP_ID"
	// KafkaConsumerTopics is the environment variable for the Kafka consumer topics.
	KafkaConsumerTopics = "KAFKA__CONSUMER__TOPICS"
	// KafkaConsumerAutoCommit is the environment variable for enabling auto commit in Kafka consumer.
	KafkaConsumerAutoCommit = "KAFKA__CONSUMER__AUTO_COMMIT"
	// KafkaConsumerMaxBuffer is the environment variable for the maximum buffer size for Kafka consumer.
	KafkaConsumerMaxBuffer = "KAFKA__CONSUMER__MAX_BUFFER"
	// KafkaConsumerAutoCommitInterval is the environment variable for the auto commit interval for Kafka consumer.
	KafkaConsumerAutoCommitInterval = "KAFKA__CONSUMER__AUTO_COMMIT_INTERVAL"

	// MongoConnectionString is the environment variable for the MongoDB connection string.
	MongoConnectionString = "MONGO__CONNECTION_STRING"

	// KafkaClientHealthCheckInterval is the environment variable for the Kafka client health check interval.
	KafkaClientHealthCheckInterval = "KAFKA_CLIENT__HEALTH_CHECK_INTERVAL"
	// KafkaClientHealthCheckResultPath is the environment variable for the Kafka client health check result path.
	KafkaClientHealthCheckResultPath = "KAFKA_CLIENT__HEALTH_CHECK_RESULT_PATH"

	// HTTPServerHost is the environment variable for the HTTP server host.
	HTTPServerHost = "HTTP_SERVER__HOST"
	// HTTPServerPort is the environment variable for the HTTP server port.
	HTTPServerPort = "HTTP_SERVER__PORT"
	// HTTPServerMaskHeaderKeyList is the environment variable for the list of HTTP headers to mask.
	HTTPServerMaskHeaderKeyList = "HTTP_SERVER__MASK__HEADER_KEY_LIST"
	// HTTPServerDocHost is the environment variable for the documentation server host.
	HTTPServerDocHost = "HTTP_SERVER__DOC_HOST"
	// HTTPServerDocRootFolder is the environment variable for the root folder for documentation.
	HTTPServerDocRootFolder = "HTTP_SERVER__DOC_ROOT_FOLDER"
	// HTTPServerTLSPublicKey is the environment variable for the path to the TLS public key.
	HTTPServerTLSPublicKey = "HTTP_SERVER__TLS_PUBLIC_KEY"
	// HTTPServerTLSPrivateKey is the environment variable for the path to the TLS private key.
	HTTPServerTLSPrivateKey = "HTTP_SERVER__TLS_PRIVATE_KEY"
)

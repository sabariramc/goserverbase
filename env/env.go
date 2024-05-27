package env

const (
	ServiceName = "SERVICE_NAME"

	LogLevel     = "LOG__LEVEL"
	LogFileTrace = "LOG__FILE_TRACE"
	LogWriter    = "LOG__WRITER"

	NotifierTopic = "NOTIFIER__TOPIC"

	KafkaBroker                     = "KAFAK__BROKER"
	KafkaProducerAcknowledge        = "KAFKA__PRODUCER__ACKNOWLEDGE"
	KafkaProducerMaxBuffer          = "KAFKA__PRODUCER__MAX_BUFFER"
	KafkaProducerAutoFlushInterval  = "KAFKA__PRODUCER__AUTO_FLUSH_INTERVAL"
	KafkaProducerAsync              = "KAFKA__PRODUCER__ASYNC"
	KafkaProducerBatch              = "KAFKA__PRODUCER__BATCH"
	KafkaConsumerGroupID            = "KAFKA__CONSUMER__GROUP_ID"
	KafkaConsumerTopics             = "KAFKA__CONSUMER__TOPICS"
	KafkaConsumerAutoCommit         = "KAFKA__CONSUMER__AUTO_COMMIT"
	KafkaConsumerMaxBuffer          = "KAFKA__CONSUMER__MAX_BUFFER"
	KafkaConsumerAutoCommitInterval = "KAFKA__CONSUMER__AUTO_COMMIT_INTERVAL"

	MongoConnectionString = "MONGO__CONNECTION_STRING"

	KafkaClientHealthCheckInterval   = "KAFKA_CLIENT__HEALTH_CHECK_INTERVAL"
	KafkaClientHealthCheckResultPath = "KAFKA_CLIENT__HEALTH_CHECK_RESULT_PATH"

	HTTPServerHost              = "HTTP_SERVER__HOST"
	HTTPServerPort              = "HTTP_SERVER__PORT"
	HTTPServerMaskHeaderKeyList = "HTTP_SERVER__MASK__HEADER_KEY_LIST"
	HTTPServerDocHost           = "HTTP_SERVER__DOC_HOST"
	HTTPServerDocRootFolder     = "HTTP_SERVER__DOC_ROOT_FOLDER"
	HTTPServerTLSPublicKey      = "HTTP_SERVER__TLS_PUBLIC_KEY"
	HTTPServerTLSPrivateKey     = "HTTP_SERVER__TLS_PRIVATE_KEY"
)

package envvariables

const (
	ServiceName = "SERVICE_NAME"

	LogLevel = "LOG__LEVEL"

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
)

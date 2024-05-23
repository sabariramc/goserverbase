package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/envvariables"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
)

const (
	ModuleProducer = "KafkaProducer"
	ModuleConsumer = "KafkaConsumer"
)

// CredConfig holds the configuration for Kafka credentials and connection details.
/*
	Environment Variables
	- KAFAK__BROKER: Sets [Brokers]
	- KAFAK__SASL__TYPE: Sets [SASLType]
*/
type CredConfig struct {
	Brokers       []string       // List of Kafka broker addresses.
	SASLMechanism sasl.Mechanism // SASL mechanism for authentication.
	TLSConfig     *tls.Config    // TLS configuration for secure connections.
}

// GetDefaultCredConfig returns a default CredConfig with values from environment variables or default values.
func GetDefaultCredConfig() *CredConfig {
	return &CredConfig{
		Brokers: utils.GetEnvAsSlice(envvariables.KafkaBroker, []string{"0.0.0.0:9092"}, ","),
	}
}

// ProducerConfig holds the configuration for a Kafka producer.
/*
	Environment Variables
	- KAFKA__PRODUCER__ACKNOWLEDGE: Sets [RequiredAcks]
	- KAFKA__PRODUCER__MAX_BUFFER: Sets [MaxBuffer]
	- KAFKA__PRODUCER__AUTO_FLUSH_INTERVAL: Sets [AutoFlushInterval]
	- KAFKA__PRODUCER__ASYNC: Sets [Async]
	- KAFKA__PRODUCER__BATCH: Sets [Batch]
*/
type ProducerConfig struct {
	*CredConfig                     // Embeds CredConfig for credential and connection details.
	RequiredAcks      int           // Number of acknowledgments required from Kafka.
	MaxBuffer         int           // Maximum buffer size for the producer.
	AutoFlushInterval uint64        // Interval in milliseconds to auto flush messages.
	Async             bool          // Flag to indicate if the producer should work asynchronously.
	Batch             bool          // Flag to indicate if messages should be batched.
	Topic             string        // Kafka topic to produce messages to.
	ModuleName        string        // Name of the module for logging.
	Log               log.Log       // Logger instance.
	Trace             ProduceTracer // Tracer for producing messages.
	Writer            *kafka.Writer // Writer for producing messages.
}

func ValidateProducerConfig(config *ProducerConfig) error {
	if config.Batch && config.Async {
		return fmt.Errorf("ValidateProducerConfig: `Batch` and `Async` are mutually exclusive")
	}
	if !config.Batch && !config.Async {
		return fmt.Errorf("ValidateProducerConfig: set either `Batch` or `Async`")
	}
	if config.Batch {
		if config.MaxBuffer <= 0 {
			config.MaxBuffer = 100
		}
		if config.AutoFlushInterval <= 0 {
			config.AutoFlushInterval = 1000
		}
	}
	return nil
}

// GetDefaultProducerConfig creates a new ProducerConfig with the provided options applied.
func GetDefaultProducerConfig() *ProducerConfig {
	config := &ProducerConfig{
		CredConfig:        GetDefaultCredConfig(),
		RequiredAcks:      utils.GetEnvInt(envvariables.KafkaProducerAcknowledge, 1),
		MaxBuffer:         utils.GetEnvInt(envvariables.KafkaProducerMaxBuffer, 0),
		AutoFlushInterval: uint64(utils.GetEnvInt(envvariables.KafkaProducerAutoFlushInterval, 1000)),
		Async:             utils.GetEnvBool(envvariables.KafkaProducerAsync, true),
		Batch:             utils.GetEnvBool(envvariables.KafkaProducerBatch, false),
		Log:               log.New(log.WithModuleName(ModuleProducer)),
		ModuleName:        ModuleProducer,
	}
	return config
}

// ProducerOption defines a function signature for applying options for kafka producer.
type ProducerOption func(*ProducerConfig)

// WithProducerCredConfig sets the Kafka credentials config for kafka producer.
func WithProducerCredConfig(credConfig *CredConfig) ProducerOption {
	return func(c *ProducerConfig) {
		c.CredConfig = credConfig
	}
}

// WithAcknowledge sets the acknowledge value for kafka producer.
func WithAcknowledge(ack int) ProducerOption {
	return func(c *ProducerConfig) {
		c.RequiredAcks = ack
	}
}

// WithProducerBuffer sets the batch max buffer value for kafka producer.
func WithProducerBuffer(buffer int) ProducerOption {
	return func(c *ProducerConfig) {
		c.MaxBuffer = buffer
	}
}

// WithAutoFlushInterval sets the auto flush interval in milliseconds for kafka producer.
func WithAutoFlushInterval(interval uint64) ProducerOption {
	return func(c *ProducerConfig) {
		c.AutoFlushInterval = interval
	}
}

// WithAsync sets the async flag for kafka producer.
func WithAsync(async bool) ProducerOption {
	return func(c *ProducerConfig) {
		c.Async = async
	}
}

// WithBatch sets the batch flag for kafka producer.
func WithBatch(batch bool) ProducerOption {
	return func(c *ProducerConfig) {
		c.Batch = batch
	}
}

// WithProducerTopic sets the topic for kafka producer.
func WithProducerTopic(topic string) ProducerOption {
	return func(c *ProducerConfig) {
		c.Topic = topic
	}
}

// WithProducerModuleName sets the module name for kafka producer.
func WithProducerModuleName(name string) ProducerOption {
	return func(c *ProducerConfig) {
		c.ModuleName = name
	}
}

// WithPoducerLogger sets the logger for kafka producer.
func WithPoducerLogger(logger log.Log) ProducerOption {
	return func(c *ProducerConfig) {
		c.Log = logger
	}
}

// WithProducerTracer sets the tracer for kafka producer.
func WithProducerTracer(tracer ProduceTracer) ProducerOption {
	return func(c *ProducerConfig) {
		c.Trace = tracer
	}
}

// WithWriter sets the [kafka.Writer] for kafka producer.
func WithWriter(writer *kafka.Writer) ProducerOption {
	return func(c *ProducerConfig) {
		c.Writer = writer
	}
}

// ConsumerConfig represents the configuration for a Kafka consumer.
/*
	Environment Variables
	- SERVICE_NAME: Sets [ServiceName]
	- KAFKA__CONSUMER__GROUP_ID: Sets [GroupID]
	- KAFKA__CONSUMER__TOPICS: Sets [Topics]
	- KAFKA__CONSUMER__AUTO_COMMIT: Sets [AutoCommit]
	- KAFKA__CONSUMER__MAX_BUFFER: Sets [MaxBuffer]
	- KAFKA__CONSUMER__AUTO_COMMIT_INTERVAL: Sets [AutoCommitInterval]
*/
type ConsumerConfig struct {
	*CredConfig                       // Embeds CredConfig for credential and connection details.
	GroupID            string         // Consumer group id
	AutoCommit         bool           // Flag to enable auto commit for consumed messages
	MaxBuffer          uint           // Count of message for batch commit
	AutoCommitInterval uint64         // Interval in milliseconds to auto commit messages.
	Log                log.Log        // Logger instance
	Trace              ConsumerTracer // Tracer for consuming messages
	Reader             *kafka.Reader  // Reader for consuming messages
	Topics             []string       // Topics to consume
	ModuleName         string         // Name of the module for logging.
	ServiceName        string         // Name of the service for client id
}

func ValidateConsumerConfig(config *ConsumerConfig) error {
	if config.MaxBuffer <= 0 {
		config.MaxBuffer = 100
	}
	if config.AutoCommitInterval <= 0 {
		config.AutoCommitInterval = 1000
	}
	return nil
}

// GetDefaultConsumerConfig creates a new ConsumerConfig with the provided options.
func GetDefaultConsumerConfig() *ConsumerConfig {
	// Default configuration
	config := &ConsumerConfig{
		CredConfig:         GetDefaultCredConfig(),
		GroupID:            utils.GetEnv(envvariables.KafkaConsumerGroupID, "cg-kafka-consumer"),
		AutoCommit:         utils.GetEnvBool(envvariables.KafkaConsumerAutoCommit, true),
		MaxBuffer:          uint(utils.GetEnvInt(envvariables.KafkaConsumerMaxBuffer, 100)),
		AutoCommitInterval: uint64(utils.GetEnvInt(envvariables.KafkaConsumerAutoCommitInterval, 1000)),
		Log:                log.New(log.WithModuleName(ModuleConsumer)),
		Topics:             utils.GetEnvAsSlice(envvariables.KafkaConsumerTopics, []string{}, ","),
		ModuleName:         ModuleProducer,
		ServiceName:        utils.GetEnv(envvariables.ServiceName, "default"),
	}

	return config
}

// ConsumerOption defines a function type that modifies the ConsumerConfig.
type ConsumerOption func(*ConsumerConfig)

// WithConsumerCredConfig sets the Kafka credentials configuration.
func WithConsumerCredConfig(creds *CredConfig) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.CredConfig = creds
	}
}

// WithGroupID sets the group ID for the Kafka consumer.
func WithGroupID(groupID string) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.GroupID = groupID
	}
}

// WithAutoCommit sets the auto-commit option for the Kafka consumer.
func WithAutoCommit(autoCommit bool) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.AutoCommit = autoCommit
	}
}

// WithConsumerBuffer sets the maximum buffer size for the Kafka consumer.
func WithConsumerBuffer(maxBuffer uint) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.MaxBuffer = maxBuffer
	}
}

// WithAutoCommitInterval sets the auto-commit interval for the Kafka consumer.
func WithAutoCommitInterval(intervalInMs uint64) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.AutoCommitInterval = intervalInMs
	}
}

// WithConsumerLogger sets the logger for the Kafka consumer.
func WithConsumerLogger(logger log.Log) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.Log = logger
	}
}

// WithConsumerTracer sets the tracer for the Kafka consumer.
func WithConsumerTracer(tracer ConsumerTracer) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.Trace = tracer
	}
}

// WithReader sets the Kafka reader for the consumer.
func WithReader(reader *kafka.Reader) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.Reader = reader
	}
}

// WithConsumerTopic sets the topics for the Kafka consumer.
func WithConsumerTopic(topics []string) ConsumerOption {
	return func(config *ConsumerConfig) {
		config.Topics = topics
	}
}

// WithConsumerModuleName sets the module name for kafka consumer.
func WithConsumerModuleName(name string) ConsumerOption {
	return func(c *ConsumerConfig) {
		c.ModuleName = name
	}
}

// WithConsumerServiceName sets the service name for kafka consumer.
func WithConsumerServiceName(name string) ConsumerOption {
	return func(c *ConsumerConfig) {
		c.ServiceName = name
	}
}

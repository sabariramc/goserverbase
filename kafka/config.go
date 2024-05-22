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
)

// KafkaCredConfig holds the configuration for Kafka credentials and connection details.
type KafkaCredConfig struct {
	Brokers       []string       // List of Kafka broker addresses.
	SASLType      string         // SASL authentication type.
	SASLMechanism sasl.Mechanism // SASL mechanism for authentication.
	TLSConfig     *tls.Config    // TLS configuration for secure connections.
}

// GetDefaultKafkaCredConfig returns a default KafkaCredConfig with values from environment variables or default values.
func GetDefaultKafkaCredConfig() *KafkaCredConfig {
	return &KafkaCredConfig{
		Brokers:  utils.GetEnvAsSlice(envvariables.KafkaBroker, []string{"0.0.0.0:9092"}, ","),
		SASLType: utils.GetEnv(envvariables.KafkaSALSMechanism, "NONE"),
	}
}

// ProducerConfig holds the configuration for a Kafka producer.
type ProducerConfig struct {
	*KafkaCredConfig                // Embeds KafkaCredConfig for credential and connection details.
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
		KafkaCredConfig:   GetDefaultKafkaCredConfig(),
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

// ProducerOption defines a function signature for applying options to ProducerConfig.
type ProducerOption func(*ProducerConfig)

// WithKafkaCredConfig sets the Kafka credentials config in the ProducerConfig.
func WithKafkaCredConfig(credConfig *KafkaCredConfig) ProducerOption {
	return func(c *ProducerConfig) {
		c.KafkaCredConfig = credConfig
	}
}

// WithAcknowledge sets the acknowledge value in the ProducerConfig.
func WithAcknowledge(ack int) ProducerOption {
	return func(c *ProducerConfig) {
		c.RequiredAcks = ack
	}
}

// WithBatchMaxBuffer sets the batch max buffer value in the ProducerConfig.
func WithBatchMaxBuffer(buffer int) ProducerOption {
	return func(c *ProducerConfig) {
		c.MaxBuffer = buffer
	}
}

// WithAutoFlushInterval sets the auto flush interval in milliseconds in the ProducerConfig.
func WithAutoFlushInterval(interval uint64) ProducerOption {
	return func(c *ProducerConfig) {
		c.AutoFlushInterval = interval
	}
}

// WithAsync sets the async flag in the ProducerConfig.
func WithAsync(async bool) ProducerOption {
	return func(c *ProducerConfig) {
		c.Async = async
	}
}

// WithBatch sets the batch flag in the ProducerConfig.
func WithBatch(batch bool) ProducerOption {
	return func(c *ProducerConfig) {
		c.Batch = batch
	}
}

// WithTopic sets the topic in the ProducerConfig.
func WithTopic(topic string) ProducerOption {
	return func(c *ProducerConfig) {
		c.Topic = topic
	}
}

// WithModuleName sets the module name in the ProducerConfig.
func WithModuleName(name string) ProducerOption {
	return func(c *ProducerConfig) {
		c.ModuleName = name
	}
}

// WithLogger sets the logger in the ProducerConfig.
func WithLogger(logger log.Log) ProducerOption {
	return func(c *ProducerConfig) {
		c.Log = logger
	}
}

// WithTrace sets the produce tracer in the ProducerConfig.
func WithTrace(tracer ProduceTracer) ProducerOption {
	return func(c *ProducerConfig) {
		c.Trace = tracer
	}
}

// WithWriter sets the [kafka.Writer] in the ProducerConfig.
func WithWriter(writer *kafka.Writer) ProducerOption {
	return func(c *ProducerConfig) {
		c.Writer = writer
	}
}

type KafkaConsumerConfig struct {
	*KafkaCredConfig
	GroupID                string
	AutoCommit             bool
	MaxBuffer              uint
	AutoCommitIntervalInMs uint64
	EnableLog              bool
}

package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
)

var defaultConfig = Config{
	healthCheckInSec: 30,
	healthFilePath:   "/tmp/healthCheck",
	ServerConfig: baseapp.ServerConfig{
		ServiceName: "KafkaConsumer",
	},
	log: log.New(log.WithModuleName("KafkaConsumer")),
	KafkaConsumerConfig: kafka.KafkaConsumerConfig{
		KafkaCredConfig: &kafka.KafkaCredConfig{
			Brokers:  []string{"0.0.0.0:9092"},
			SASLType: "NONE",
		},
		GroupID:                "cg-kafka-base",
		AutoCommit:             true,
		MaxBuffer:              100,
		AutoCommitIntervalInMs: 1000,
		EnableLog:              false,
	},
}

// Options represents options for configuring a KafkaConsumerServer instance.
type Options func(*Config)

// WithLog sets the log instance for KafkaConsumerServer.
func WithLog(log log.Log) Options {
	return func(c *Config) {
		c.log = log
	}
}

// WithNotifier sets the notifier instance for KafkaConsumerServer.
func WithNotifier(notifier notifier.Notifier) Options {
	return func(c *Config) {
		c.notifier = notifier
	}
}

// WithServerConfig sets the server configuration for KafkaConsumerServer.
func WithServerConfig(config baseapp.ServerConfig) Options {
	return func(c *Config) {
		c.ServerConfig = config
	}
}

// WithKafkaConsumerConfig sets the Kafka consumer configuration for KafkaConsumerServer.
func WithKafkaConsumerConfig(config kafka.KafkaConsumerConfig) Options {
	return func(c *Config) {
		c.KafkaConsumerConfig = config
	}
}

// WithTracer sets the tracer instance for KafkaConsumerServer.
func WithTracer(t Tracer) Options {
	return func(c *Config) {
		c.t = t
	}
}

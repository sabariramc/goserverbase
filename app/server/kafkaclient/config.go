package kafkaclient

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/env"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// Config holds the configuration for the application.
type Config struct {
	*baseapp.Config               // Embeds for base config
	*kafka.ConsumerConfig         // Embeds for kafka consumer config
	HealthCheckInterval   uint    // Interval in seconds to do health check of various modules
	HealthCheckResultPath string  // Local disk file path for writing health check results
	Log                   log.Log // Logger instance.
	Tracer                Tracer  // Tracer instance.
}

// GetDefaultConfig creates a new Config with values from environment variables or default values.
/*
	Environment Variables
	- KAFKACS__HEALTH_CHECK_INTERVAL: Sets [HealthCheckInterval]
	- KAFKACS__HEALTH_CHECK_RESULT_PATH: Sets [HealthCheckResultPath]
*/
func GetDefaultConfig() *Config {
	return &Config{
		HealthCheckInterval:   uint(utils.GetEnvInt(env.KafkaClientHealthCheckInterval, 30)),
		HealthCheckResultPath: utils.GetEnv(env.KafkaClientHealthCheckResultPath, "/tmp/healthCheck"),
		Config:                baseapp.GetDefaultConfig(),
		Log:                   log.New(log.WithModuleName("KafkaConsumerServer")),
		ConsumerConfig:        kafka.GetDefaultConsumerConfig(),
	}
}

// Options represents options for configuring a KafkaConsumerServer instance.
type Options func(*Config)

// WithLog sets the log instance for KafkaConsumerServer.
func WithLog(log log.Log) Options {
	return func(c *Config) {
		c.Log = log
	}
}

// WithNotifier sets the notifier instance for KafkaConsumerServer.
func WithNotifier(notifier notifier.Notifier) Options {
	return func(c *Config) {
		c.Notifier = notifier
	}
}

// WithServerConfig sets the server configuration for KafkaConsumerServer.
func WithServerConfig(config *baseapp.Config) Options {
	return func(c *Config) {
		c.Config = config
	}
}

// WithKafkaConsumerConfig sets the Kafka consumer configuration for KafkaConsumerServer.
func WithKafkaConsumerConfig(config *kafka.ConsumerConfig) Options {
	return func(c *Config) {
		c.ConsumerConfig = config
	}
}

// WithTracer sets the tracer instance for KafkaConsumerServer.
func WithTracer(t Tracer) Options {
	return func(c *Config) {
		c.Tracer = t
	}
}

package kafka

import (
	"github.com/sabariramc/goserverbase/v6/envvariables"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
)

type Config struct {
	Log         log.Log
	ServiceName string
	Topic       string
	Producer    *kafka.Producer
}

// GetDefaultConfig returns the default configuration for the Notifier.
func GetDefaultConfig() Config {
	return Config{
		ServiceName: utils.GetEnv(envvariables.ServiceName, "default"),
		Log:         log.New().NewResourceLogger("Notifier"),
		Topic:       utils.GetEnv(envvariables.NotifierTopic, ""),
	}
}

// Option defines a function signature for applying options to Config.
type Option func(*Config)

// WithLogger sets the logger in the Config.
func WithLogger(logger log.Log) Option {
	return func(c *Config) {
		c.Log = logger
	}
}

// WithServiceName sets the service name in the Config.
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithTopic sets the topic in the Config.
func WithTopic(topic string) Option {
	return func(c *Config) {
		c.Topic = topic
	}
}

// WithProducer sets the Kafka producer in the Config.
func WithProducer(producer *kafka.Producer) Option {
	return func(c *Config) {
		c.Producer = producer
	}
}

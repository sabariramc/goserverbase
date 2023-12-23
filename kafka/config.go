package kafka

import (
	"crypto/tls"

	"github.com/segmentio/kafka-go/sasl"
)

type KafkaCredConfig struct {
	Brokers       []string
	ServiceName   string
	SASLType      string
	SASLMechanism sasl.Mechanism
	TLSConfig     *tls.Config
}

type KafkaConsumerConfig struct {
	*KafkaCredConfig
	GroupID                string
	AutoCommit             bool
	MaxBuffer              int
	AutoCommitIntervalInMs uint64
}

type KafkaProducerConfig struct {
	*KafkaCredConfig
	Acknowledge           int
	MaxBuffer             int
	AutoFlushIntervalInMs uint64
	Async                 bool
	Topic                 string
}

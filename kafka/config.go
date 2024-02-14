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
	EnableLog              bool
}

type KafkaProducerConfig struct {
	*KafkaCredConfig
	Acknowledge            int
	BatchMaxBuffer         int
	BatchFlushIntervalInMs uint64
	Async                  bool
	Batch                  bool
	Topic                  string
	EnableLog              bool
	Name                   string
}

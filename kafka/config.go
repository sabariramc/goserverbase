package kafka

import (
	"crypto/tls"

	"github.com/segmentio/kafka-go/sasl"
)

type SASLCredential struct {
	SASLMechanism  string
	SASLCredential interface{}
}

type KafkaCredConfig struct {
	Brokers       []string
	ServiceName   string
	SASLMechanism sasl.Mechanism
	TLSConfig     *tls.Config
}

type KafkaConsumerConfig struct {
	*KafkaCredConfig
	GroupID                string
	OffsetReset            bool
	AutoCommit             bool
	MaxBuffer              int
	AutoCommitIntervalInMs uint64
}

type KafkaProducerConfig struct {
	*KafkaCredConfig
	Acknowledge           int
	MaxBuffer             int
	AutoFlushIntervalInMs uint64
	Channeled             bool //Experimental
	Topic                 string
}

package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/kafka"
)

type KafkaConsumerServerConfig struct {
	baseapp.ServerConfig
	kafka.KafkaConsumerConfig
}

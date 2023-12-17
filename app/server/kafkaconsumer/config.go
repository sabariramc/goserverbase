package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v4/app"
	"github.com/sabariramc/goserverbase/v4/kafka"
)

type KafkaConsumerServerConfig struct {
	baseapp.ServerConfig
	kafka.KafkaConsumerConfig
}

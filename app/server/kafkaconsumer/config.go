package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v3/app"
	"github.com/sabariramc/goserverbase/v3/kafka"
)

type KafkaConsumerServerConfig struct {
	*baseapp.ServerConfig
	*kafka.KafkaConsumerConfig
}

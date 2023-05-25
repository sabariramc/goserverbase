package kafkaclient

import (
	baseapp "github.com/sabariramc/goserverbase/v3/app"
	"github.com/sabariramc/goserverbase/v3/kafka"
)

type KafkaServerConfig struct {
	*baseapp.ServerConfig
	*kafka.KafkaConsumerConfig
}

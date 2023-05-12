package kafkaclient

import (
	baseapp "github.com/sabariramc/goserverbase/v2/app"
	"github.com/sabariramc/goserverbase/v2/kafka"
)

type KafkaServerConfig struct {
	*baseapp.ServerConfig
	*kafka.KafkaConsumerConfig
	Host              string
	AuthHeaderKeyList []string
}

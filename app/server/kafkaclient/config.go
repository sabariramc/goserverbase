package kafkaclient

import (
	"github.com/sabariramc/goserverbase/v2/baseapp"
	"github.com/sabariramc/goserverbase/v2/kafka"
)

type KafkaServerConfig struct {
	*baseapp.ServerConfig
	*kafka.KafkaConsumerConfig
	Host              string
	AuthHeaderKeyList []string
}

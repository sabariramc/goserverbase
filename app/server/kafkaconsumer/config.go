package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/kafka"
)

type KafkaConsumerServerConfig struct {
	baseapp.ServerConfig
	kafka.KafkaConsumerConfig
	HealthCheckInSec int
	HealthFilePath   string
}

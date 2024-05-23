package kafkaconsumer

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
)

type Config struct {
	baseapp.ServerConfig
	kafka.ConsumerConfig
	healthCheckInSec uint
	healthFilePath   string
	log              log.Log
	notifier         notifier.Notifier
	t                Tracer
}

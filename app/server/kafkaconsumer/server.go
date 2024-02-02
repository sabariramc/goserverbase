package kafkaconsumer

import (
	"context"
	"sync"

	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/errors"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	ckafka "github.com/segmentio/kafka-go"
)

type KafkaEventProcessor func(context.Context, *kafka.Message) error

type KafkaConsumerServer struct {
	*baseapp.BaseApp
	client                 *kafka.Consumer
	handler                map[string]KafkaEventProcessor
	log                    log.Log
	ch                     chan *ckafka.Message
	c                      *KafkaConsumerServerConfig
	shutdown, shutdownPoll context.CancelFunc
	requestWG, shutdownWG  sync.WaitGroup
}

func New(appConfig KafkaConsumerServerConfig, logger log.Log, errorNotifier errors.ErrorNotifier) *KafkaConsumerServer {
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	h := &KafkaConsumerServer{
		BaseApp: b,
		log:     logger.NewResourceLogger("KafkaConsumerServer"),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
	}
	h.RegisterOnShutdown(h)
	return h
}

func (k *KafkaConsumerServer) Name(ctx context.Context) string {
	return "KafkaConsumerServer"
}

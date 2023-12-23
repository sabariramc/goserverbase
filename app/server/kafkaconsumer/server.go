package kafkaconsumer

import (
	"context"
	"sync"

	baseapp "github.com/sabariramc/goserverbase/v4/app"
	"github.com/sabariramc/goserverbase/v4/errors"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/log"
	ckafka "github.com/segmentio/kafka-go"
)

type KafkaEventProcessor func(context.Context, *kafka.Message) error

type KafkaConsumerServer struct {
	*baseapp.BaseApp
	client       *kafka.Consumer
	handler      map[string]KafkaEventProcessor
	log          *log.Logger
	ch           chan *ckafka.Message
	c            *KafkaConsumerServerConfig
	shutdown     context.CancelFunc
	shutdownPoll context.CancelFunc
	wg           sync.WaitGroup
}

func New(appConfig KafkaConsumerServerConfig, logger *log.Logger, errorNotifier errors.ErrorNotifier) *KafkaConsumerServer {
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	h := &KafkaConsumerServer{
		BaseApp: b,
		log:     logger.NewResourceLogger("KafkaConsumerServer"),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
	}
	h.AddShutdownHook(h)
	return h
}

func (k *KafkaConsumerServer) Name(ctx context.Context) string {
	return "KafkaConsumerServer"
}

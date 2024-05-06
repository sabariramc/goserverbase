package kafkaconsumer

import (
	"context"
	"io/fs"
	"os"
	"sync"

	baseapp "github.com/sabariramc/goserverbase/v5/app"
	"github.com/sabariramc/goserverbase/v5/instrumentation/span"
	"github.com/sabariramc/goserverbase/v5/kafka"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/notifier"
	ckafka "github.com/segmentio/kafka-go"
)

type KafkaEventProcessor func(context.Context, *kafka.Message) error

type Tracer interface {
	StartKafkaSpanFromMessage(ctx context.Context, msg *ckafka.Message) (context.Context, span.Span)
	span.SpanOp
}

type KafkaConsumerServer struct {
	*baseapp.BaseApp
	client                 *kafka.Poller
	handler                map[string]KafkaEventProcessor
	log                    log.Log
	ch                     chan *ckafka.Message
	c                      *KafkaConsumerServerConfig
	shutdown, shutdownPoll context.CancelFunc
	requestWG, shutdownWG  sync.WaitGroup
	tracer                 Tracer
}

func New(appConfig KafkaConsumerServerConfig, logger log.Log, t Tracer, errorNotifier notifier.Notifier) *KafkaConsumerServer {
	if appConfig.HealthCheckInSec <= 0 {
		appConfig.HealthCheckInSec = 30
	}
	if appConfig.HealthFilePath == "" {
		appConfig.HealthFilePath = "/tmp/healthCheck"
	}
	os.WriteFile(appConfig.HealthFilePath, []byte("Hello"), fs.ModeAppend)
	b := baseapp.New(appConfig.ServerConfig, logger, errorNotifier)
	h := &KafkaConsumerServer{
		BaseApp: b,
		log:     logger.NewResourceLogger("KafkaConsumerServer"),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
		tracer:  t,
	}
	h.RegisterOnShutdownHook(h)
	return h
}

func (k *KafkaConsumerServer) Name(ctx context.Context) string {
	return "KafkaConsumerServer"
}

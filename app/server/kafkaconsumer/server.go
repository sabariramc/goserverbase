// Package kafkaconsumer extends the BaseApp with a kafka consumer server
package kafkaconsumer

import (
	"context"
	"io/fs"
	"os"
	"sync"

	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/instrumentation/span"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	ckafka "github.com/segmentio/kafka-go"
)

// KafkaEventProcessor defines the function signature for processing Kafka events handlers.
type KafkaEventProcessor func(context.Context, *kafka.Message) error

// Tracer defines the interface for tracing functionality.
type Tracer interface {
	StartKafkaSpanFromMessage(ctx context.Context, msg *ckafka.Message) (context.Context, span.Span)
	span.SpanOp
}

// KafkaConsumerServer represents a Kafka consumer server.
// Implements ShutdownHook, HealthCheckHook and StatusCheckHook
type KafkaConsumerServer struct {
	*baseapp.BaseApp
	client                 *kafka.Poller
	handler                map[string]KafkaEventProcessor
	log                    log.Log
	ch                     chan *ckafka.Message
	c                      *Config
	shutdown, shutdownPoll context.CancelFunc
	requestWG, shutdownWG  sync.WaitGroup
	tracer                 Tracer
}

// New creates a new instance of KafkaConsumerServer.
func New(option ...Options) *KafkaConsumerServer {
	config := defaultConfig
	for _, fn := range option {
		fn(&config)
	}
	os.WriteFile(config.healthFilePath, []byte("Hello"), fs.ModeAppend)
	b := baseapp.New(config.Config, config.log, config.notifier)
	h := &KafkaConsumerServer{
		BaseApp: b,
		log:     config.log.NewResourceLogger("KafkaConsumerServer"),
		c:       &config,
		handler: make(map[string]KafkaEventProcessor),
		tracer:  config.t,
	}
	h.RegisterHealthCheckHook(h)
	h.RegisterOnShutdownHook(h)
	h.RegisterStatusCheckHook(h)
	return h
}

// Name returns the name of the KafkaConsumerServer.
// Implementation of the hook interface defined in the BaseApp
func (k *KafkaConsumerServer) Name(ctx context.Context) string {
	return "KafkaConsumerServer"
}

// Shutdown gracefully shuts down the Kafka consumer server.
// Implementation for shutdown hook
func (k *KafkaConsumerServer) Shutdown(ctx context.Context) error {
	defer k.shutdownWG.Done()
	k.shutdownPoll()
	k.requestWG.Wait()
	k.shutdown()
	k.client.Close(ctx)
	return nil
}

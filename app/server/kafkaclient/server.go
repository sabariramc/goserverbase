// Package kafkaconsumer extends the BaseApp with a kafka consumer server
package kafkaclient

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

// KafkaClient represents a Kafka consumer server.
// Implements ShutdownHook, HealthCheckHook and StatusCheckHook
type KafkaClient struct {
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

// New creates a new instance of KafkaClient.
func New(option ...Options) *KafkaClient {
	config := GetDefaultConfig()
	for _, opt := range option {
		opt(config)
	}
	os.WriteFile(config.HealthCheckResultPath, []byte("Hello"), fs.ModeAppend)
	h := &KafkaClient{
		BaseApp: baseapp.NewWithConfig(config.Config),
		log:     config.Log,
		c:       config,
		handler: make(map[string]KafkaEventProcessor),
		tracer:  config.Tracer,
	}
	h.RegisterHealthCheckHook(h)
	h.RegisterOnShutdownHook(h)
	h.RegisterStatusCheckHook(h)
	return h
}

// Name returns the name of the KafkaClient.
// Implementation of the hook interface defined in the BaseApp
func (k *KafkaClient) Name(ctx context.Context) string {
	return "KafkaClient"
}

// Shutdown gracefully shuts down the Kafka consumer server.
// Implementation for shutdown hook
func (k *KafkaClient) Shutdown(ctx context.Context) error {
	defer k.shutdownWG.Done()
	k.shutdownPoll()
	k.requestWG.Wait()
	k.shutdown()
	k.client.Close(ctx)
	return nil
}

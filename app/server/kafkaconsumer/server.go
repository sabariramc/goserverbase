package kafkaconsumer

import (
	"context"
	"fmt"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	baseapp "github.com/sabariramc/goserverbase/v3/app"
	"github.com/sabariramc/goserverbase/v3/errors"
	"github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type KafkaEventProcessor func(context.Context, *kafka.Message) error

type KafkaConsumerServer struct {
	*baseapp.BaseApp
	client  *kafka.Consumer
	handler map[string]KafkaEventProcessor
	Log     *log.Logger
	ch      chan *ckafka.Message
	c       *KafkaConsumerServerConfig
}

func New(appConfig KafkaConsumerServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *KafkaConsumerServer {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	h := &KafkaConsumerServer{
		BaseApp: b,
		Log:     b.GetLogger(),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
	}
	return h
}

func (k *KafkaConsumerServer) Subscribe(ctx context.Context) {
	topicList := make([]string, 0, len(k.handler))
	for h := range k.handler {
		topicList = append(topicList, h)
	}
	ch := make(chan *ckafka.Message)
	k.ch = ch
	client, err := kafka.NewConsumer(ctx, k.c.ServiceName, k.Log, k.c.KafkaConsumerConfig, k.GetErrorNotifier(), topicList...)
	if err != nil {
		k.Log.Emergency(ctx, "Error occurred during client creation", map[string]any{
			"topicList": topicList,
			"config":    k.c.KafkaConsumerConfig,
		}, err)
	}
	k.client = client
}

func (k *KafkaConsumerServer) StartConsumer(ctx context.Context) {
	tracer.Start()
	defer tracer.Stop()
	corr := &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-KAFKA-CONSUMER", k.c.ServiceName)}
	ctx = log.GetContextWithCorrelation(ctx, corr)
	pollCtx, cancelPoll := context.WithCancel(log.GetContextWithCorrelation(context.Background(), corr))
	k.Log.Notice(pollCtx, "Starting kafka consumer", nil)
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(pollCtx, rec, nil)
		}
	}()
	var wg sync.WaitGroup
	k.Subscribe(pollCtx)
	defer k.client.Close(ctx)
	defer wg.Wait()
	defer cancelPoll()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := k.client.Poll(pollCtx, 1, k.ch)
		if err != nil {
			k.Log.Emergency(pollCtx, "Kafka consumer exited", nil, err)
		}
	}()
	k.Log.Notice(pollCtx, "Kafka consumer started", nil)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-k.ch:
			if !ok {
				return
			}
			topicName := *msg.TopicPartition.Topic
			handler := k.handler[topicName]
			if handler == nil {
				k.Log.Emergency(pollCtx, "missing handler for topic - "+topicName, nil, fmt.Errorf("missing handler for topic - %v", topicName))
			}
			emMsg := &kafka.Message{Message: msg}
			msgCtx := context.Background()
			msgCtx = k.GetContextWithCorrelation(msgCtx, k.GetCorrelationParams(emMsg.GetHeaders()))
			msgCtx = k.GetContextWithCustomerId(msgCtx, k.GetCustomerId(emMsg.GetHeaders()))
			k.ProcessEvent(msgCtx, emMsg, handler)
		}
	}
}

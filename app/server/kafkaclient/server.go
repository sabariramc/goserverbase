package kafkaclient

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	baseapp "github.com/sabariramc/goserverbase/v2/app"
	"github.com/sabariramc/goserverbase/v2/errors"
	"github.com/sabariramc/goserverbase/v2/kafka"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

type KafkaEventProcessor func(context.Context, *kafka.Message) error

type KafkaClient struct {
	*baseapp.BaseApp
	client  *kafka.Consumer
	handler map[string]KafkaEventProcessor
	log     *log.Logger
	ch      chan *ckafka.Message
	c       *KafkaServerConfig
}

func New(appConfig KafkaServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *KafkaClient {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	h := &KafkaClient{
		BaseApp: b,
		log:     b.GetLogger(),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
	}
	return h
}

func (k *KafkaClient) AddHandler(topicName string, handler KafkaEventProcessor) {
	k.handler[topicName] = handler
}

func (k *KafkaClient) Subscribe(ctx context.Context) {
	topicList := make([]string, 0, len(k.handler))
	for h := range k.handler {
		topicList = append(topicList, h)
	}
	ch := make(chan *ckafka.Message)
	k.ch = ch
	client, err := kafka.NewConsumer(ctx, k.log, k.c.KafkaConsumerConfig, topicList...)
	if err != nil {
		k.log.Emergency(ctx, "Error occurred during client creation", map[string]any{
			"topicList": topicList,
			"config":    k.c.KafkaConsumerConfig,
		}, err)
	}
	k.client = client
}

func (k *KafkaClient) StartKafkaConsumer() {
	ctx, cancel := context.WithCancel(k.GetContextWithCorrelation(context.Background(), &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-KAFKA-CONSUMER", k.c.ServiceName)}))
	k.log.Notice(ctx, "Starting kafka consumer", nil)
	defer func() {
		if rec := recover(); rec != nil {
			stackTrace := string(debug.Stack())
			k.log.Error(ctx, "Recovered - Panic", rec)
			k.log.Error(ctx, "Recovered - StackTrace", stackTrace)
			err, ok := rec.(error)
			if !ok {
				blob, _ := json.Marshal(rec)
				err = fmt.Errorf("non error panic: %v", string(blob))
			}
			k.ProcessError(ctx, stackTrace, err)
		}
	}()
	defer cancel()
	k.Subscribe(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(k.ch)
		defer wg.Done()
		k.client.Poll(ctx, 1, k.ch)
	}()
	k.log.Notice(ctx, "Kafka consumer started", nil)
	for msg := range k.ch {
		topicName := *msg.TopicPartition.Topic
		handler := k.handler[topicName]
		if handler == nil {
			panic(fmt.Errorf("missing handler for topic - %v", topicName))
		}
		emMsg := &kafka.Message{Message: msg}
		ctx := context.Background()
		ctx = k.GetContextWithCorrelation(ctx, k.GetCorrelationParams(emMsg.GetHeaders()))
		ctx = k.GetContextWithCustomerId(ctx, k.GetCustomerId(emMsg.GetHeaders()))
		k.ProcessEvent(ctx, emMsg, handler)
	}
	wg.Wait()
}

func (k *KafkaClient) ProcessEvent(ctx context.Context, msg *kafka.Message, handler KafkaEventProcessor) {
	defer func() {
		if rec := recover(); rec != nil {
			stackTrace := string(debug.Stack())
			k.log.Error(ctx, "Recovered - Panic", rec)
			k.log.Error(ctx, "Recovered - StackTrace", stackTrace)
			err, ok := rec.(error)
			if !ok {
				blob, _ := json.Marshal(rec)
				err = fmt.Errorf("non error panic: %v", string(blob))
			}
			k.ProcessError(ctx, stackTrace, err)
		}
	}()
	err := handler(ctx, msg)
	if err != nil {
		k.ProcessError(ctx, "", err)
	}
}

func (k *KafkaClient) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	correlation := log.GetDefaultCorrelationParams(k.c.ServiceName)
	err := utils.LenientJsonTransformer(headers, correlation)
	if err != nil {
		return log.GetDefaultCorrelationParams(k.c.ServiceName)
	}
	return correlation
}

func (k *KafkaClient) GetCustomerId(headers map[string]string) *log.CustomerIdentifier {
	customerId := &log.CustomerIdentifier{}
	err := utils.LenientJsonTransformer(headers, customerId)
	if err != nil {
		return &log.CustomerIdentifier{}
	}
	return customerId
}

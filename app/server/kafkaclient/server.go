package kafkaclient

import (
	"context"
	"fmt"
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
	Log     *log.Logger
	ch      chan *ckafka.Message
	c       *KafkaServerConfig
}

func New(appConfig KafkaServerConfig, loggerConfig log.Config, lMux log.LogMux, errorNotifier errors.ErrorNotifier, auditLogger log.AuditLogWriter) *KafkaClient {
	b := baseapp.New(*appConfig.ServerConfig, loggerConfig, lMux, errorNotifier, auditLogger)
	h := &KafkaClient{
		BaseApp: b,
		Log:     b.GetLogger(),
		c:       &appConfig,
		handler: make(map[string]KafkaEventProcessor),
	}
	return h
}

func (k *KafkaClient) AddHandler(ctx context.Context, topicName string, handler KafkaEventProcessor) {
	if handler == nil {
		k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("missing handler for topic - %v", topicName))
	}
	k.handler[topicName] = handler
}

func (k *KafkaClient) Subscribe(ctx context.Context) {
	topicList := make([]string, 0, len(k.handler))
	for h := range k.handler {
		topicList = append(topicList, h)
	}
	ch := make(chan *ckafka.Message)
	k.ch = ch
	client, err := kafka.NewConsumer(ctx, k.Log, k.c.KafkaConsumerConfig, topicList...)
	if err != nil {
		k.Log.Emergency(ctx, "Error occurred during client creation", map[string]any{
			"topicList": topicList,
			"config":    k.c.KafkaConsumerConfig,
		}, err)
	}
	k.client = client
}

func (k *KafkaClient) StartKafkaConsumer() {
	ctx, cancel := context.WithCancel(k.GetContextWithCorrelation(context.Background(), &log.CorrelationParam{CorrelationId: fmt.Sprintf("%v-KAFKA-CONSUMER", k.c.ServiceName)}))
	k.Log.Notice(ctx, "Starting kafka consumer", nil)
	defer func() {
		if rec := recover(); rec != nil {
			k.PanicRecovery(ctx, rec, nil)
		}
	}()
	defer cancel()
	k.Subscribe(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(k.ch)
		err := k.client.Poll(ctx, 1, k.ch)
		if err != nil {
			k.Log.Emergency(ctx, "Kafka consumer exited", nil, err)
		}
	}()
	k.Log.Notice(ctx, "Kafka consumer started", nil)
	for msg := range k.ch {
		topicName := *msg.TopicPartition.Topic
		handler := k.handler[topicName]
		if handler == nil {
			k.Log.Emergency(ctx, "missing handler for topic - "+topicName, nil, fmt.Errorf("missing handler for topic - %v", topicName))
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
			k.PanicRecovery(ctx, rec, msg)
		}
	}()
	err := handler(ctx, msg)
	if err != nil {
		k.ProcessError(ctx, "", err, msg)
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

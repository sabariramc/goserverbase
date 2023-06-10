package kafka_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	cKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	eKafka "github.com/sabariramc/goserverbase/v3/errors/notifier/kafka"
	"github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/utils"
	"gotest.tools/assert"
)

func newProducer(ctx context.Context) (*kafka.Producer, error) {
	log := KafkaTestLogger
	hProducer := kafka.NewHTTPProducer(ctx, log, KafkaTestConfig.KafkaHTTPProxyURL, KafkaTestConfig.KafkaTestTopic, time.Second)
	notifier := eKafka.New(ctx, log, KafkaTestConfig.App.ServiceName, hProducer)
	return kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, notifier)
}

func TestKafkaMessage(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewConsumer(ctx, KafkaTestConfig.App.ServiceName, KafkaTestLogger, KafkaTestConfig.KafkaConsumerConfig, nil, KafkaTestConfig.KafkaTestTopic)
	defer co.Close(ctx)
	assert.NilError(t, err)
	var s sync.WaitGroup
	s.Add(1)
	go func() {
		kMsg, err := co.ReadMessage(ctx, time.Second*10)
		KafkaTestLogger.Info(ctx, "Kafka message", kMsg)
		assert.NilError(t, err)
		s.Done()
	}()
	time.Sleep(time.Second * 5)
	pr, err := kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, nil)
	defer pr.Close(ctx)
	assert.NilError(t, err)
	err = pr.ProduceMessage(ctx, "test", &utils.Message{
		Event: "random event",
	}, nil)
	assert.NilError(t, err)
	s.Wait()
}

func TestKafkaPoll(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewConsumer(ctx, KafkaTestConfig.App.ServiceName, KafkaTestLogger, KafkaTestConfig.KafkaConsumerConfig, nil, KafkaTestConfig.KafkaTestTopic)
	defer co.Close(ctx)
	assert.NilError(t, err)
	pr, err := kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, nil)
	defer pr.Close(ctx)
	assert.NilError(t, err)
	ch := make(chan *cKafka.Message, 100)
	var s sync.WaitGroup
	s.Add(1)
	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 5)
	go func() {
		for i := 0; i < 50; i++ {
			err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
		s.Done()
	}()
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	defer cancel()
	go co.Poll(tCtx, 2000, ch)
	count := 0
	msgCount := 0
	for i := range ch {
		m, err := kafka.LoadMessage(i)
		msgCount++
		if m.Event == uuidVal {
			count++
		}
		if err != nil {
			KafkaTestLogger.Error(ctx, "parse error", err)
		}
		KafkaTestLogger.Info(ctx, "Kafka message", i)
	}
	KafkaTestLogger.Info(ctx, "Total matched", count)
	KafkaTestLogger.Info(ctx, "Total received", msgCount)
	s.Wait()
	assert.Equal(t, 50, count)

}

func TestKafkaPollWithDelay(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewConsumer(ctx, KafkaTestConfig.App.ServiceName, KafkaTestLogger, KafkaTestConfig.KafkaConsumerConfig, nil, KafkaTestConfig.KafkaTestTopic)
	defer co.Close(ctx)
	assert.NilError(t, err)
	pr, err := kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, nil)
	defer pr.Close(ctx)
	assert.NilError(t, err)
	ch := make(chan *cKafka.Message)
	tCtx, cancel := context.WithCancel(ctx)
	go co.Poll(tCtx, 2000, ch)
	time.Sleep(2 * time.Second)
	cancel()
	var s sync.WaitGroup

	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 3)
	for i := 0; i < 10; i++ {
		err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
			Event: uuidVal,
		}, nil)
		assert.NilError(t, err)
	}
	tCtx, cancel = context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	ch = make(chan *cKafka.Message, 100)
	go co.Poll(tCtx, 2000, ch)
	s.Add(1)
	go func() {
		for i := 0; i < 40; i++ {
			err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
		s.Done()
	}()
	count := 0
	msgCount := 0
	for i := range ch {
		m, err := kafka.LoadMessage(i)
		msgCount++
		if m.Event == uuidVal {
			count++
		}
		if err != nil {
			KafkaTestLogger.Error(ctx, "parse error", err)
		}
		KafkaTestLogger.Info(ctx, "Kafka message", i)
	}
	KafkaTestLogger.Info(ctx, "Total matched", count)
	KafkaTestLogger.Info(ctx, "Total received", msgCount)
	s.Wait()
	assert.Equal(t, 50, count)

}

func TestKafkaPollWithDelayExtended(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewConsumer(ctx, KafkaTestConfig.App.ServiceName, KafkaTestLogger, KafkaTestConfig.KafkaConsumerConfig, nil, KafkaTestConfig.KafkaTestTopic)
	defer co.Close(ctx)
	assert.NilError(t, err)
	pr, err := kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducerConfig, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic, nil)
	defer pr.Close(ctx)
	assert.NilError(t, err)
	tCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	ch := make(chan *cKafka.Message, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		co.Poll(tCtx, 2000, ch)
	}()
	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 3)
	for i := 0; i < 10; i++ {
		err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
			Event: uuidVal,
		}, nil)
		assert.NilError(t, err)
	}
	wg.Wait()
}

func TestKafkaPollHTTPProducer(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewConsumer(ctx, "TEST", KafkaTestLogger, KafkaTestConfig.KafkaConsumerConfig, nil, KafkaTestConfig.KafkaTestTopic)
	defer co.Close(ctx)
	assert.NilError(t, err)
	pr := kafka.NewHTTPProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaHTTPProxyURL, KafkaTestConfig.KafkaTestTopic, time.Minute)
	assert.NilError(t, err)
	ch := make(chan *cKafka.Message, 100)
	var s sync.WaitGroup
	s.Add(1)
	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 5)
	go func() {
		tCtx, _ := context.WithDeadline(ctx, time.Now().Add(time.Minute))
		for i := 0; i < 50; i++ {
			err = pr.ProduceMessage(tCtx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
		s.Done()
	}()
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	defer cancel()
	go co.Poll(tCtx, 2000, ch)
	count := 0
	msgCount := 0
	for i := range ch {
		m, err := kafka.LoadMessage(i)
		msgCount++
		if m.Event == uuidVal {
			count++
		}
		if err != nil {
			KafkaTestLogger.Error(ctx, "parse error", err)
		}
		KafkaTestLogger.Info(ctx, "Kafka message", i)
	}
	KafkaTestLogger.Info(ctx, "Total matched", count)
	KafkaTestLogger.Info(ctx, "Total received", msgCount)
	s.Wait()
	assert.Equal(t, 50, count)
}

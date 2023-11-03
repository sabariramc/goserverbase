package kafka_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v3/kafka"
	"github.com/sabariramc/goserverbase/v3/utils"
	cKafka "github.com/segmentio/kafka-go"
	"gotest.tools/assert"
)

func newProducer(ctx context.Context) (*kafka.Producer, error) {
	return kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducer, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic)
}

func newConsumer(ctx context.Context) (*kafka.Consumer, error) {
	return kafka.NewConsumer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaConsumer, KafkaTestConfig.App.ServiceName, KafkaTestConfig.KafkaTestTopic)
}

func TestKafkaConsumer(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	ch := make(chan *cKafka.Message, 100)
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	defer cancel()
	assert.NilError(t, err)
	var s sync.WaitGroup
	s.Add(1)
	go func() {
		co.Poll(tCtx, 1000, ch)
		s.Done()
	}()
	for msg := range ch {
		kMsg := kafka.Message{
			Message: msg,
		}
		KafkaTestLogger.Info(ctx, "Kafka message", kMsg.GetBody())
	}
	s.Wait()
	assert.NilError(t, err)
}

func TestKafkaProducer(t *testing.T) {
	ctx := GetCorrelationContext()
	uuidVal := uuid.NewString()
	totalNoOfMessage := 1000
	connFac := 100
	var wg sync.WaitGroup
	for i := 0; i < connFac; i++ {
		wg.Add(1)
		go func() {
			pr, err := newProducer(ctx)
			assert.NilError(t, err)
			defer pr.Close(ctx)
			for i := 0; i < totalNoOfMessage/connFac; i++ {
				ctx := GetCorrelationContext()
				err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
					Event: uuidVal,
				}, nil)
				assert.NilError(t, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestKafkaPoll(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := newProducer(ctx)
	assert.NilError(t, err)
	defer pr.Close(ctx)
	ch := make(chan *cKafka.Message, 100)
	var s sync.WaitGroup
	s.Add(1)
	totalCount := 1000
	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 5)
	go func() {
		for i := 0; i < totalCount; i++ {
			ctx := GetCorrelationContext()
			err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
		s.Done()
	}()
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	defer cancel()
	st := time.Now()
	go co.Poll(tCtx, 2000, ch)
	count := 0
	msgCount := 0
	for i := range ch {
		m, err := kafka.LoadMessage(i)
		msgCount++
		if m.Event == uuidVal {
			count++
		}
		if totalCount == count {
			cancel()
		}
		if err != nil {
			KafkaTestLogger.Error(ctx, "parse error", err)
		}
		KafkaTestLogger.Info(ctx, "Kafka message", m)
	}
	KafkaTestLogger.Info(ctx, "Total matched", count)
	KafkaTestLogger.Info(ctx, "Total received", msgCount)
	s.Wait()
	assert.Equal(t, totalCount, count)
	et := time.Now()
	KafkaTestLogger.Info(ctx, "Time taken in ms", et.Sub(st)/1000000)
}

func TestKafkaPollWithDelay(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := newProducer(ctx)
	assert.NilError(t, err)
	defer pr.Close(ctx)
	ch := make(chan *cKafka.Message)
	tCtx, cancel := context.WithCancel(ctx)
	go co.Poll(tCtx, 2000, ch)
	time.Sleep(2 * time.Second)
	cancel()
	var s sync.WaitGroup

	uuidVal := uuid.NewString()
	time.Sleep(time.Second * 3)
	for i := 0; i < 10; i++ {
		ctx := GetCorrelationContext()
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
			ctx := GetCorrelationContext()
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
func TestKafkaPollHTTPProducer(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
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
	}
	KafkaTestLogger.Info(ctx, "Total matched", count)
	KafkaTestLogger.Info(ctx, "Total received", msgCount)
	s.Wait()
	assert.Equal(t, 50, count)
}

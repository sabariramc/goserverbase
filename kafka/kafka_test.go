package kafka_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v4/kafka"
	"github.com/sabariramc/goserverbase/v4/utils"
	cKafka "github.com/segmentio/kafka-go"
	"gotest.tools/assert"
)

func newProducer(ctx context.Context) (*kafka.Producer, error) {
	return kafka.NewProducer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaProducer, KafkaTestConfig.KafkaTestTopic)
}

func newChanneledProducer(ctx context.Context) (*kafka.Producer, error) {
	config := *KafkaTestConfig.KafkaProducer
	config.Channeled = true
	return kafka.NewProducer(ctx, KafkaTestLogger, &config, KafkaTestConfig.KafkaTestTopic)
}

func newConsumer(ctx context.Context) (*kafka.Consumer, error) {
	return kafka.NewConsumer(ctx, KafkaTestLogger, KafkaTestConfig.KafkaConsumer, KafkaTestConfig.KafkaTestTopic)
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
	maxCount := 10000
	go func() {
		co.Poll(tCtx, ch)
		s.Done()
	}()
	i := 0
	for msg := range ch {
		i++
		kMsg := kafka.Message{
			Message: msg,
		}
		KafkaTestLogger.Debug(ctx, "Kafka message", kMsg.GetBody())
		if i == maxCount {
			break
		}
	}
	cancel()
	s.Wait()
	assert.NilError(t, err)
}

func TestKafkaProducer(t *testing.T) {
	ctx := GetCorrelationContext()
	uuidVal := uuid.NewString()
	totalNoOfMessage := 10000
	connFac := 10
	var wg sync.WaitGroup
	for i := 0; i < connFac; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pr, err := newProducer(ctx)
			assert.NilError(t, err)
			defer pr.Close(ctx)
			for i := 0; i < totalNoOfMessage/connFac; i++ {
				ctx := GetCorrelationContext()
				err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
					Event: "TestKafkaProducer-" + uuidVal,
				}, nil)
				assert.NilError(t, err)
			}
			err = pr.Flush(ctx)
			assert.NilError(t, err)
		}()
	}
	wg.Wait()
}

func testKafkaPoll(ctx context.Context, t *testing.T, co *kafka.Consumer, pr *kafka.Producer, totalCount int) {
	ch := make(chan *cKafka.Message, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	uuidVal := "TestKafkaPoll" + uuid.NewString()
	time.Sleep(time.Second * 5)
	go func() {
		defer wg.Done()
		for i := 0; i < totalCount; i++ {
			ctx := GetCorrelationContext()
			err := pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
		err := pr.Flush(ctx)
		assert.NilError(t, err)
	}()
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	defer cancel()
	st := time.Now()
	go co.Poll(tCtx, ch)
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
	wg.Wait()
	assert.Equal(t, totalCount, count)
	et := time.Now()
	KafkaTestLogger.Info(ctx, "Time taken in ms", et.Sub(st)/1000000)
}

func TestKafkaPoll(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := newProducer(ctx)
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(ctx, t, co, pr, 1000)
}

func TestKafkaPollChanneledWriter(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := newChanneledProducer(ctx)
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(ctx, t, co, pr, 10)
}

func TestKafkaPollWithDelay(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := newConsumer(ctx)
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := newProducer(ctx)
	assert.NilError(t, err)
	defer pr.Close(ctx)
	ch := make(chan *cKafka.Message, 100)
	tCtx, cancel := context.WithCancel(ctx)
	go co.Poll(tCtx, ch)
	time.Sleep(2 * time.Second)
	cancel()
	var s sync.WaitGroup
	totalCount := 50
	uuidVal := "TestKafkaPollWithDelay" + uuid.NewString()
	time.Sleep(time.Second * 3)
	for i := 0; i < 10; i++ {
		ctx := GetCorrelationContext()
		err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
			Event: uuidVal,
		}, nil)
		assert.NilError(t, err)
	}
	pr.Flush(ctx)
	tCtx, cancel = context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	ch = make(chan *cKafka.Message, 100)
	go co.Poll(tCtx, ch)
	s.Add(1)
	go func() {
		defer s.Done()
		defer pr.Flush(ctx)
		for i := 0; i < (totalCount - 10); i++ {
			ctx := GetCorrelationContext()
			err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
			assert.NilError(t, err)
		}
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
	assert.Equal(t, totalCount, count)
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
	go co.Poll(tCtx, ch)
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

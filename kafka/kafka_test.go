package kafka_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/utils"
	cKafka "github.com/segmentio/kafka-go"
	"gotest.tools/assert"
)

func Example() {
	ctx := GetCorrelationContext()
	co, _ := kafka.NewPoller(kafka.WithConsumerTopic([]string{"test.topic1"}))
	defer co.Close(ctx)
	pr, _ := kafka.NewProducer(kafka.WithProducerTopic("test.topic1"))
	defer pr.Close(ctx)
	ch := make(chan *cKafka.Message, 100)
	var wg sync.WaitGroup
	totalCount := 100000
	wg.Add(1)
	uuidVal := "TestKafkaPoll" + uuid.NewString()
	go func() {
		defer wg.Done()
		for i := 0; i < totalCount; i++ {
			ctx := GetCorrelationContext()
			pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
				Event: uuidVal,
			}, nil)
		}
		pr.Flush(ctx)
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
	KafkaTestLogger.Notice(ctx, "Time taken in ms", time.Now().Sub(st)/1000000)
}

func Example_confluentKafka() {
	ctx := GetCorrelationContext()
	cred := kafka.GetDefaultCredConfig()
	cred.TLSConfig = &tls.Config{ //For Confluent kafka
		MinVersion: tls.VersionTLS12,
	}
	co, _ := kafka.NewPoller(kafka.WithConsumerCredConfig(cred), kafka.WithConsumerTopic([]string{"test.topic1"}))
	defer co.Close(ctx)
	pr, _ := kafka.NewProducer(kafka.WithProducerCredConfig(cred))
	defer pr.Close(ctx)
}

func TestKafkaProducerSync(t *testing.T) {
	ctx := GetCorrelationContext()
	uuidVal := uuid.NewString()
	totalNoOfMessage := 1
	pr, err := kafka.NewProducer(kafka.WithProducerTopic(KafkaTestConfig.KafkaTestTopic2))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	for i := 0; i < totalNoOfMessage; i++ {
		ctx := GetCorrelationContext()
		err = pr.ProduceMessage(ctx, strconv.Itoa(i), &utils.Message{
			Event: "TestKafkaProducer-" + uuidVal,
		}, nil)
		assert.NilError(t, err)
	}
}

func TestKafkaProducer(t *testing.T) {
	ctx := GetCorrelationContext()
	uuidVal := uuid.NewString()
	totalNoOfMessage := 100000
	connFac := 10
	var wg sync.WaitGroup
	for i := 0; i < connFac; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pr, err := kafka.NewProducer(kafka.WithProducerTopic(KafkaTestConfig.KafkaTestTopic))
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

func TestKafkaPoller(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewPoller(kafka.WithAutoCommit(false), kafka.WithConsumerTopic([]string{KafkaTestConfig.KafkaTestTopic, KafkaTestConfig.KafkaTestTopic2}))
	assert.NilError(t, err)
	defer co.Close(ctx)
	ch := make(chan *cKafka.Message, 100)
	tCtx, cancel := context.WithTimeout(ctx, time.Second*45)
	assert.NilError(t, err)
	maxCount := 100000
	st := time.Now()
	go func() {
		co.Poll(tCtx, ch)
	}()
	i := 0
	for msg := range ch {
		i++
		co.StoreOffset(ctx, msg)
		kMsg := kafka.Message{
			Message: msg,
		}
		KafkaTestLogger.Debug(ctx, "Kafka message", kMsg.GetBody())
		if i == maxCount {
			break
		}
	}
	cancel()
	co.Commit(ctx)
	assert.Equal(t, i, maxCount)
	KafkaTestLogger.Notice(ctx, "Time taken in ms", time.Now().Sub(st)/1000000)
}

func testKafkaPoll(ctx context.Context, t *testing.T, co *kafka.Poller, pr *kafka.Producer, totalCount int) {
	ch := make(chan *cKafka.Message, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	uuidVal := "TestKafkaPoll" + uuid.NewString()
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
	KafkaTestLogger.Notice(ctx, "Time taken in ms", time.Now().Sub(st)/1000000)
}

func TestKafkaPoll(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewPoller(kafka.WithConsumerTopic([]string{KafkaTestConfig.KafkaTestTopic}), kafka.WithConsumerBuffer(1000))
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := kafka.NewProducer(kafka.WithBatch(true), kafka.WithAsync(false), kafka.WithProducerBuffer(1000), kafka.WithProducerTopic(KafkaTestConfig.KafkaTestTopic))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(ctx, t, co, pr, 100000)
}

func TestKafkaPollAsyncWriter(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewPoller(kafka.WithConsumerTopic([]string{KafkaTestConfig.KafkaTestTopic}), kafka.WithConsumerBuffer(1000))
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := kafka.NewProducer(kafka.WithProducerTopic(KafkaTestConfig.KafkaTestTopic))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(ctx, t, co, pr, 100000)
}

func TestKafkaPollWithDelay(t *testing.T) {
	ctx := GetCorrelationContext()
	co, err := kafka.NewPoller(kafka.WithConsumerTopic([]string{KafkaTestConfig.KafkaTestTopic}), kafka.WithConsumerBuffer(1000))
	assert.NilError(t, err)
	defer co.Close(ctx)
	pr, err := kafka.NewProducer(kafka.WithProducerTopic(KafkaTestConfig.KafkaTestTopic))
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

package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
	kafkatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	*kafkatrace.Producer
	config *KafkaProducerConfig
	log    *log.Logger
	topic  string
}

func NewProducer(ctx context.Context, log *log.Logger, config *KafkaProducerConfig, topic string) (*Producer, error) {
	parsedConfig := &kafka.ConfigMap{}
	utils.StrictJsonTransformer(config, parsedConfig)
	p, err := kafkatrace.NewProducer(parsedConfig)

	if err != nil {
		log.Error(ctx, "Failed to create kafka producer", err)
		return nil, fmt.Errorf("kafka.NewKafkaProducer.CreateProducer: %w", err)
	}
	k := &Producer{
		config:   config,
		log:      log,
		Producer: p,
		topic:    topic,
	}
	return k, nil
}

func (k *Producer) Produce(ctx context.Context, key string, message *utils.Message) (m *kafka.Message, err error) {
	var buf bytes.Buffer
	deliveryChannel := make(chan kafka.Event)
	defer close(deliveryChannel)
	err = json.NewEncoder(&buf).Encode(message)
	if err != nil {
		k.log.Error(ctx, "Failed to encode message", err)
		k.log.Error(ctx, "Message", message)
		return nil, fmt.Errorf("KafkaProducer.Send.EncodeMessage: %w", err)
	}
	correlationParam := log.GetCorrelationParam(ctx)
	headers := make(map[string]string, 0)
	customerIdentity := log.GetCustomerIdentifier(ctx)
	utils.StrictJsonTransformer(correlationParam, &headers)
	utils.StrictJsonTransformer(customerIdentity, &headers)
	messageHeader := make([]kafka.Header, 0)
	for i, v := range headers {
		messageHeader = append(messageHeader, kafka.Header{
			Key:   i,
			Value: []byte(v),
		})
	}
	k.log.Debug(ctx, "Message payload", map[string]any{"body": message, "key": key, "headers": messageHeader})

	k.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          buf.Bytes(),
		Headers:        messageHeader,
		Timestamp:      time.Now(),
	}, deliveryChannel)
	e := <-deliveryChannel
	m = e.(*kafka.Message)
	err = m.TopicPartition.Error
	if err != nil {
		k.log.Error(ctx, "Send failed for topic: "+k.topic, err)
		return nil, fmt.Errorf("KafkaProducer.Send.ProduceMessage: %w", err)
	}
	k.log.Info(ctx, "Send success for topic: "+k.topic, m)
	return m, nil
}

type HTTPProducer struct {
	baseUrl    string
	log        *log.Logger
	topicName  string
	httpClient *http.Client
}

func NewHTTPProducer(ctx context.Context, log *log.Logger, baseURL, topicName string, timeout time.Duration) *HTTPProducer {
	return &HTTPProducer{baseUrl: baseURL, topicName: topicName, log: log, httpClient: &http.Client{Timeout: timeout}}
}

func (k HTTPProducer) Produce(ctx context.Context, key string, message *utils.Message) (*kafka.Message, error) {
	url := k.baseUrl + "/" + k.topicName
	data := map[string]any{
		"records": []map[string]any{{
			"value": message,
			"key":   key,
		},
		},
	}

	var reqBodyBlob bytes.Buffer
	err := json.NewEncoder(&reqBodyBlob).Encode(&data)
	if err != nil {
		k.log.Error(ctx, "KafkaHTTPProducer.Send.PayloadEncoding", err)
		return nil, fmt.Errorf("KafkaHTTPProducer.Send.PayloadEncoding: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &reqBodyBlob)
	if err != nil {
		k.log.Error(ctx, "KafkaHTTPProducer.Send.RequestCreation", err)
		return nil, fmt.Errorf("KafkaHTTPProducer.Send.RequestCreation: %w", err)
	}
	log.SetCorrelationHeader(ctx, req)
	req.Header.Add("Content-Type", "application/vnd.kafka.json.v2+json")
	k.log.Debug(ctx, "Request payload", data)
	k.log.Debug(ctx, "Request header", req.Header)
	k.log.Debug(ctx, "Request url", req.URL)
	res, err := k.httpClient.Do(req)
	if err != nil {
		k.log.Error(ctx, "Error in sending kafka message", err)
		return nil, fmt.Errorf("KafkaHTTPProducer.Send.HTTPCall: %w", err)
	}
	defer res.Body.Close()
	blobBody, _ := ioutil.ReadAll(res.Body)
	var resBody any
	resBody = make(map[string]any)
	err = json.Unmarshal(blobBody, &data)
	if err != nil {
		k.log.Error(ctx, "KafkaHTTPProducer : Error in JSON Marshal", err)
		resBody = string(blobBody)
	}
	if res.StatusCode > 299 {
		err = fmt.Errorf("KafkaHTTPProducer.Send.HTTPCall.statusCode: %v", res.StatusCode)
		k.log.Error(ctx, fmt.Sprintf("KAFKA HTTP response -%v", res.StatusCode), resBody)
	} else {
		k.log.Debug(ctx, fmt.Sprintf("KAFKA HTTP response -%v", res.StatusCode), resBody)
	}
	return nil, err
}

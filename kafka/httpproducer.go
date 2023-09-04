package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

type HTTPProducer struct {
	baseUrl      string
	log          *log.Logger
	topicName    string
	httpClient   *http.Client
	resourceName string
}

func NewHTTPProducer(ctx context.Context, log *log.Logger, baseURL, topicName string, timeout time.Duration) *HTTPProducer {
	return NewHTTPProducerResource(ctx, log, "KafkaProducerHttp", baseURL, topicName, timeout)
}

func NewHTTPProducerResource(ctx context.Context, log *log.Logger, resourceName, baseURL, topicName string, timeout time.Duration) *HTTPProducer {
	p := &HTTPProducer{baseUrl: baseURL, topicName: topicName, httpClient: &http.Client{Timeout: timeout}, resourceName: resourceName, log: log.NewResourceLogger(resourceName)}

	return p
}

func (k HTTPProducer) ProduceMessage(ctx context.Context, key string, message *utils.Message, headers map[string]string) error {
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
		return fmt.Errorf("KafkaHTTPProducer.Send.PayloadEncoding: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &reqBodyBlob)
	if err != nil {
		k.log.Error(ctx, "KafkaHTTPProducer.Send.RequestCreation", err)
		return fmt.Errorf("KafkaHTTPProducer.Send.RequestCreation: %w", err)
	}
	log.SetCorrelationHeader(ctx, req)
	req.Header.Add("Content-Type", "application/vnd.kafka.json.v2+json")
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	k.log.Debug(ctx, "Request payload", data)
	k.log.Debug(ctx, "Request header", req.Header)
	k.log.Debug(ctx, "Request url", req.URL)
	res, err := k.httpClient.Do(req)
	if err != nil {
		k.log.Error(ctx, "Error in sending kafka message", err)
		return fmt.Errorf("KafkaHTTPProducer.Send.HTTPCall: %w", err)
	}
	defer res.Body.Close()
	blobBody, _ := io.ReadAll(res.Body)
	var resBody any
	resBody = make(map[string]any)
	err = json.Unmarshal(blobBody, &resBody)
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
	return err
}

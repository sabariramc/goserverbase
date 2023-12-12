package kafka

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
	"github.com/sabariramc/goserverbase/v4/utils/httputil"
)

type HTTPProducer struct {
	baseUrl      string
	log          *log.Logger
	topicName    string
	httpClient   *httputil.HttpClient
	resourceName string
}

func NewHTTPProducer(ctx context.Context, log *log.Logger, baseURL, topicName string, timeout time.Duration) *HTTPProducer {
	return NewHTTPProducerResource(ctx, log, "KafkaProducerHttp", baseURL, topicName, timeout)
}

func NewHTTPProducerResource(ctx context.Context, log *log.Logger, resourceName, baseURL, topicName string, timeout time.Duration) *HTTPProducer {
	p := &HTTPProducer{baseUrl: baseURL, topicName: topicName, httpClient: httputil.NewDefaultHttpClient(log), resourceName: resourceName, log: log.NewResourceLogger(resourceName)}
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
	resBody := make(map[string]any)
	res, err := k.httpClient.Post(ctx, url, data, &resBody, map[string]string{
		"Content-Type": "application/vnd.kafka.v2+json",
	})
	if err != nil || (res != nil && res.StatusCode > 299) {
		if res != nil {
			resBlob, _ := io.ReadAll(res.Body)
			err = fmt.Errorf("kafka.HTTPProducer.ProduceMessage: http error with statusCode %v", res.StatusCode)
			k.log.Error(ctx, fmt.Sprintf("KAFKA HTTP response : %v", res.StatusCode), string(resBlob))
		} else {
			return fmt.Errorf("kafka.HTTPProducer.ProduceMessage: error in network call: %w", err)
		}

	} else {
		k.log.Debug(ctx, fmt.Sprintf("KAFKA HTTP response -%v", res.StatusCode), resBody)
	}
	return err
}

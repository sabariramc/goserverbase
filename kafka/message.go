package kafka

import (
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sabariramc/goserverbase/v2/log"
)

type Message struct {
	*kafka.Message
	headers map[string]string
}

func (m *Message) GetKey() string {
	return string(m.Message.Key)
}

func (m *Message) GetHeaders() map[string]string {
	if m.headers != nil {
		return m.headers
	}
	m.headers = make(map[string]string, len(m.Message.Headers))
	for _, v := range m.Message.Headers {
		m.headers[v.Key] = string(v.Value)
	}
	return m.headers
}

func (m *Message) LoadBody(v any) {
	json.Unmarshal(m.Message.Value, v)
}

func (m *Message) Print(ctx context.Context, log *log.Logger) {
	log.Info(ctx, "Kafka Message", map[string]any{
		"Key":            m.GetKey(),
		"Headers":        m.GetHeaders(),
		"TopicPartition": m.TopicPartition,
		"Timestamp":      m.Timestamp,
	})
	log.Debug(ctx, "Kafka Message-Body", string(m.Message.Value))
}

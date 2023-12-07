package kafka

import (
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	*kafka.Message
	headers    map[string]string
	stringBody string
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

func (m *Message) LoadBody(v any) error {
	return json.Unmarshal(m.Message.Value, v)
}

func (m *Message) GetBody() string {
	if m.stringBody == "" {
		m.stringBody = string(m.Message.Value)
	}
	return m.stringBody
}

func (m *Message) GetMeta() map[string]any {
	return map[string]any{
		"Key":       m.GetKey(),
		"Headers":   m.GetHeaders(),
		"Partition": m.Partition,
		"Topic":     m.Topic,
		"Time":      m.Time,
	}
}

package kafka

import (
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

// Message wraps a kafka.Message and provides additional functionalities.
type Message struct {
	*kafka.Message
	headers    map[string]string
	stringBody string
}

// GetKey returns the key of the Kafka message as a string.
func (m *Message) GetKey() string {
	return string(m.Message.Key)
}

// GetHeaders returns the headers of the Kafka message as a map.
// It lazily initializes and caches the headers map on first access.
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

// LoadBody unmarshals the message value into the provided interface.
func (m *Message) LoadBody(v any) error {
	return json.Unmarshal(m.Message.Value, v)
}

// GetBody returns the message value as a string. It caches the result for subsequent calls.
func (m *Message) GetBody() string {
	if m.stringBody == "" {
		m.stringBody = string(m.Message.Value)
	}
	return m.stringBody
}

// GetMeta returns a map containing metadata of the Kafka message, including key, headers, partition, offset, topic, and time.
func (m *Message) GetMeta() map[string]any {
	return map[string]any{
		"Key":       m.GetKey(),
		"Headers":   m.GetHeaders(),
		"Partition": m.Partition,
		"Offset":    m.Offset,
		"Topic":     m.Topic,
		"Time":      m.Time,
	}
}

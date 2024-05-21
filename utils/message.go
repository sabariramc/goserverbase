package utils

import (
	"fmt"
)

type Payload map[string]interface{}

// Message is a generic structure to pass messages between services.
// It includes an entity, event, a list of contained payload names, and a map of payloads.
// This structure is not mature yet and may undergo changes.
type Message struct {
	Entity   string             `json:"entity"`
	Event    string             `json:"event"`
	Contains []string           `json:"contains"`
	Payload  map[string]Payload `json:"payload"`
}

// NewMessage creates a new Message instance with the specified entity and event.
// It initializes the Contains slice and Payload map.
func NewMessage(entity string, event string) *Message {
	return &Message{
		Entity:   entity,
		Event:    event,
		Contains: make([]string, 0),
		Payload:  make(map[string]Payload, 0),
	}
}

// AddPayload adds a new payload to the Message.
// It appends the payload name to the Contains slice and stores the payload in the Payload map.
// If a payload with the same name already exists, it returns an error.
func (m *Message) AddPayload(name string, payload Payload) error {
	for _, v := range m.Contains {
		if v == name {
			return fmt.Errorf("Message.AddPayload: Duplicate payload for key: `%s`", name)
		}
	}
	m.Contains = append(m.Contains, name)
	m.Payload[name] = payload
	return nil
}

// GetPayload retrieves a payload by name from the Message.
// If the payload is found, it returns the payload and no error.
// If the payload is not found, it returns an error.
func (m *Message) GetPayload(name string) (p Payload, err error) {
	for _, v := range m.Contains {
		if v == name {
			var ok bool
			p, ok = m.Payload[name]
			if !ok {
				p = nil
				err = fmt.Errorf("Message.GetPayload: payload %v not found", name)
			}
			return
		}
	}
	err = fmt.Errorf("Message.GetPayload: payload %v not found in contains param", name)
	return
}

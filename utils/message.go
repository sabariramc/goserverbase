package utils

import (
	"fmt"

	"github.com/sabariramc/goserverbase/errors"
)

type Entity string

const (
	StateEntity = "state"
	EventEntity = "event"
)

type Payload struct {
	Entity map[string]interface{} `json:"entity"`
}

type Message struct {
	Entity   Entity              `json:"entity"`
	Event    string              `json:"event"`
	Contains []string            `json:"contains"`
	Payload  map[string]*Payload `json:"payload"`
}

func NewMessage(entity Entity, event string) *Message {
	return &Message{
		Entity:   entity,
		Event:    event,
		Contains: make([]string, 0),
		Payload:  make(map[string]*Payload, 0),
	}
}

func (m *Message) AddPayload(name string, payload *Payload) error {
	for _, v := range m.Contains {
		if v == name {
			return fmt.Errorf("Message.AddPayload : %w", errors.NewCustomError("DUPLICATE_PAYLOAD", fmt.Sprintf("Payload `%v` already exist", name), nil))
		}
	}
	m.Contains = append(m.Contains, name)
	m.Payload[name] = payload
	return nil
}

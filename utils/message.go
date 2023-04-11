package utils

import (
	"fmt"

	"github.com/sabariramc/goserverbase/v2/errors"
)

type Payload map[string]interface{}

type Message struct {
	Entity   string             `json:"entity"`
	Event    string             `json:"event"`
	Contains []string           `json:"contains"`
	Payload  map[string]Payload `json:"payload"`
}

func NewMessage(entity string, event string) *Message {
	return &Message{
		Entity:   entity,
		Event:    event,
		Contains: make([]string, 0),
		Payload:  make(map[string]Payload, 0),
	}
}

func (m *Message) AddPayload(name string, payload Payload) error {
	for _, v := range m.Contains {
		if v == name {
			return errors.NewCustomError("DUPLICATE_PAYLOAD", "Duplicate payload for key :`"+name+"`", nil, nil, true)
		}
	}
	m.Contains = append(m.Contains, name)
	m.Payload[name] = payload
	return nil
}

func (m *Message) GetPayload(name string) (p Payload, err error) {
	for _, v := range m.Contains {
		if v == name {
			var ok bool
			p, ok = m.Payload[name]
			if !ok {
				p = nil
				err = fmt.Errorf("Payload %v not found", name)
			}
			return
		}
	}
	err = fmt.Errorf("Payload %v not found in contains param", name)
	return
}

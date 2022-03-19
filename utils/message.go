package utils

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
		Contains: make([]string, 1),
		Payload:  make(map[string]*Payload, 1),
	}
}

func (m *Message) AddPayload(name string, payload *Payload) {
	m.Contains = append(m.Contains, name)
	m.Payload[name] = payload
}

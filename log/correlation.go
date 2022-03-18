package log

type CorrelationParmas struct {
	CorrelationId string `json:"x-correlation-id"`
	ScenarioId    string `json:"x-scenario-id,omitempty"`
	SessionId     string `json:"x-session-id,omitempty"`
	ScenarioName  string `json:"x-scenario-name,omitempty"`
	ServiceName   string `json:"service-name,omitempty"`
}

type HostParams struct {
	Version     string `json:"version"`
	Host        string `json:"host"`
	ServiceName string `json:"service-name,omitempty"`
}

package log

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CorrelationParam struct {
	CorrelationId string `json:"x-correlation-id"`
	ScenarioId    string `json:"x-scenario-id,omitempty"`
	SessionId     string `json:"x-session-id,omitempty"`
	ScenarioName  string `json:"x-scenario-name,omitempty"`
}

type HostParams struct {
	Version     string `json:"version"`
	Host        string `json:"host"`
	ServiceName string `json:"service-name,omitempty"`
}

func GetDefaultCorrelationParams(serviceName string) *CorrelationParam {
	return &CorrelationParam{
		CorrelationId: fmt.Sprintf("%v-%v", serviceName, uuid.New().String()),
	}
}

func GetCorrelationParam(ctx context.Context) *CorrelationParam {
	val, ok := ctx.Value(CorrelationContextKey).(*CorrelationParam)
	if !ok {
		return &CorrelationParam{}
	}
	return val
}

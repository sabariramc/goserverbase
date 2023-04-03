package log

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type CorrelationParam struct {
	CorrelationId string `json:"x-correlation-id"`
	ScenarioId    string `json:"x-scenario-id,omitempty"`
	SessionId     string `json:"x-session-id,omitempty"`
	ScenarioName  string `json:"x-scenario-name,omitempty"`
}

type CustomerIdentifier struct {
	CustomerId string `json:"customerId"`
	AppUserId  string `json:"appUserId"`
	Id         string `json:"id"`
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
	iVal := ctx.Value(ContextKeyCorrelation)
	if iVal == nil {
		return &CorrelationParam{}
	}
	val, ok := iVal.(*CorrelationParam)
	if !ok {
		return &CorrelationParam{}
	}
	return val
}

func GetCustomerIdentifier(ctx context.Context) *CustomerIdentifier {
	iVal := ctx.Value(ContextKeyCustomerIdentifier)
	if iVal == nil {
		return &CustomerIdentifier{}
	}
	val, ok := iVal.(*CustomerIdentifier)
	if !ok {
		return &CustomerIdentifier{}
	}
	return val
}

func SetCorrelationHeader(ctx context.Context, req *http.Request) {
	correlation := GetCorrelationParam(ctx)
	req.Header.Add("x-correlation-id", correlation.CorrelationId)
	req.Header.Add("x-scenario-id", correlation.ScenarioId)
	req.Header.Add("x-session-id", correlation.SessionId)
	req.Header.Add("x-scenario-id", correlation.ScenarioName)
}

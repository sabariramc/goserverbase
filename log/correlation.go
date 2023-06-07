package log

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v3/utils"
)

type HostParams struct {
	Version     string `json:"version"`
	Host        string `json:"host"`
	ServiceName string `json:"service-name,omitempty"`
}

type CorrelationParam struct {
	CorrelationId string `header:"x-correlation-id" body:"correlationId"`
	ScenarioId    string `header:"x-scenario-id,omitempty" body:"scenarioId,omitempty"`
	SessionId     string `header:"x-session-id,omitempty" body:"sessionId,omitempty"`
	ScenarioName  string `header:"x-scenario-name,omitempty" body:"scenarioName,omitempty"`
}

func (c *CorrelationParam) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJson.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CorrelationParam) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJson.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

type CustomerIdentifier struct {
	CustomerId string `header:"x-customer-id,omitempty" body:"customerId,omitempty"`
	AppUserId  string `header:"x-app-user-id,omitempty" body:"appUserId,omitempty"`
	Id         string `header:"x-entity-id,omitempty" body:"Id,omitempty"`
}

func (c *CustomerIdentifier) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJson.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CustomerIdentifier) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJson.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func GetDefaultCorrelationParam(serviceName string) *CorrelationParam {
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
	headers := GetCorrelationHeader(ctx)
	for i, v := range headers {
		req.Header.Add(i, v)
	}
}

func GetCorrelationHeader(ctx context.Context) map[string]string {
	headers := make(map[string]string, 10)
	corr := GetCorrelationParam(ctx).GetHeader()
	identity := GetCustomerIdentifier(ctx).GetHeader()
	for k, v := range corr {
		if v != "" {
			headers[k] = v
		}
	}
	for k, v := range identity {
		if v != "" {
			headers[k] = v
		}
	}
	return headers
}

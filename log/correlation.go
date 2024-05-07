package log

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/utils"
)

type CorrelationParam struct {
	CorrelationID string  `header:"x-correlation-id" body:"correlationId"`
	ScenarioID    *string `header:"x-scenario-id,omitempty" body:"scenarioId,omitempty"`
	SessionID     *string `header:"x-session-id,omitempty" body:"sessionId,omitempty"`
	ScenarioName  *string `header:"x-scenario-name,omitempty" body:"scenarioName,omitempty"`
}

func (c *CorrelationParam) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CorrelationParam) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CorrelationParam) LoadFromHeader(header map[string]string) error {
	data, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("CorrelationParam.LoadFromHeader: error marshalling header: %w", err)
	}
	err = utils.HeaderJSON.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("CorrelationParam.LoadFromHeader: error unmarshalling header: %w", err)
	}
	return nil
}

func GetDefaultCorrelationParam(serviceName string) *CorrelationParam {
	return &CorrelationParam{
		CorrelationID: fmt.Sprintf("%v-%v", serviceName, uuid.New().String()),
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

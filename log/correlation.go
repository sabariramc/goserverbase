package log

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// CorrelationParam defines context object for a correlation
type CorrelationParam struct {
	CorrelationID string  `header:"x-correlation-id" body:"correlationId"`
	ScenarioID    *string `header:"x-scenario-id,omitempty" body:"scenarioId,omitempty"`
	SessionID     *string `header:"x-session-id,omitempty" body:"sessionId,omitempty"`
	ScenarioName  *string `header:"x-scenario-name,omitempty" body:"scenarioName,omitempty"`
}

// GetPayload encodes CorrelationParam into map[string]string with body struct tag
func (c *CorrelationParam) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// GetHeader encodes CorrelationParam into map[string]string with header struct tag
func (c *CorrelationParam) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// ExtractFromHeader extracts CorrelationParam from map[string]string with header struct tag
func (c *CorrelationParam) ExtractFromHeader(header map[string]string) error {
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

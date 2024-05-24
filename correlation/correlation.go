// Package correlation enhances the context of requests with correlation and user identity.
package correlation

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// CorrelationParam defines a context object for correlation.
type CorrelationParam struct {
	CorrelationID string  `header:"x-correlation-id" body:"correlationId"`
	ScenarioID    *string `header:"x-scenario-id,omitempty" body:"scenarioId,omitempty"`
	SessionID     *string `header:"x-session-id,omitempty" body:"sessionId,omitempty"`
	ScenarioName  *string `header:"x-scenario-name,omitempty" body:"scenarioName,omitempty"`
}

// GetPayload encodes CorrelationParam into a map[string]string with body struct tags.
func (c *CorrelationParam) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// GetHeader encodes CorrelationParam into a map[string]string with header struct tags.
func (c *CorrelationParam) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// LoadFromHeader extracts CorrelationParam from a map[string]string with header struct tags.
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

// NewCorrelationParam creates a new CorrelationParam with a given service name.
func NewCorrelationParam(serviceName string) *CorrelationParam {
	return &CorrelationParam{
		CorrelationID: fmt.Sprintf("%v-%v", serviceName, uuid.New().String()),
	}
}

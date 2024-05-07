package log_test

import (
	"encoding/json"
	"testing"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
	"gotest.tools/assert"
)

func TestCustomJsonBody(t *testing.T) {
	x := log.CorrelationParam{}
	bodyData := map[string]string{
		"correlationId": "xyz",
		"scenarioId":    "c.ScenarioId",
		"sessionId":     "c.SessionId",
		"scenarioName":  "c.ScenarioName",
	}
	data, _ := json.Marshal(bodyData)
	utils.BodyJSON.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationID, bodyData["correlationId"])
	assert.Equal(t, *x.ScenarioID, bodyData["scenarioId"])
	assert.Equal(t, *x.SessionID, bodyData["sessionId"])
	assert.Equal(t, *x.ScenarioName, bodyData["scenarioName"])
	encodedData, _ := utils.BodyJSON.Marshal(x)
	newBodyData := map[string]string{}
	json.Unmarshal(encodedData, &newBodyData)
	assert.DeepEqual(t, newBodyData, bodyData)
}

func TestCustomJsonHeader(t *testing.T) {
	x := log.CorrelationParam{}
	headerData := map[string]string{
		"x-correlation-id": "xyz",
		"x-scenario-id":    "c.ScenarioId",
		"x-session-id":     "c.SessionId",
		"x-scenario-name":  "c.ScenarioName",
	}
	data, _ := json.Marshal(headerData)
	utils.HeaderJSON.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationID, headerData["x-correlation-id"])
	assert.Equal(t, *x.ScenarioID, headerData["x-scenario-id"])
	assert.Equal(t, *x.SessionID, headerData["x-session-id"])
	assert.Equal(t, *x.ScenarioName, headerData["x-scenario-name"])
	encodedData, _ := utils.HeaderJSON.Marshal(x)
	newHeaderData := map[string]string{}
	json.Unmarshal(encodedData, &newHeaderData)
	assert.DeepEqual(t, newHeaderData, headerData)
}

func TestCustomJSON(t *testing.T) {
	x := log.CorrelationParam{}
	bodyData := map[string]string{
		"correlationId": "xyz",
		"scenarioId":    "c.ScenarioId",
		"sessionId":     "c.SessionId",
		"scenarioName":  "c.ScenarioName",
	}
	headerData := map[string]string{
		"x-correlation-id": "xyz",
		"x-scenario-id":    "c.ScenarioId",
		"x-session-id":     "c.SessionId",
		"x-scenario-name":  "c.ScenarioName",
	}
	data, _ := json.Marshal(bodyData)
	utils.BodyJSON.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationID, bodyData["correlationId"])
	assert.Equal(t, *x.ScenarioID, bodyData["scenarioId"])
	assert.Equal(t, *x.SessionID, bodyData["sessionId"])
	assert.Equal(t, *x.ScenarioName, bodyData["scenarioName"])
	encodedData, _ := utils.HeaderJSON.Marshal(x)
	newData := map[string]string{}
	json.Unmarshal(encodedData, &newData)
	assert.DeepEqual(t, newData, headerData)
}

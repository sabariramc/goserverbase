package log_test

import (
	"encoding/json"
	"testing"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
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
	utils.BodyJson.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationId, bodyData["correlationId"])
	assert.Equal(t, x.ScenarioId, bodyData["scenarioId"])
	assert.Equal(t, x.SessionId, bodyData["sessionId"])
	assert.Equal(t, x.ScenarioName, bodyData["scenarioName"])
	encodedData, _ := utils.BodyJson.Marshal(x)
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
	utils.HeaderJson.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationId, headerData["x-correlation-id"])
	assert.Equal(t, x.ScenarioId, headerData["x-scenario-id"])
	assert.Equal(t, x.SessionId, headerData["x-session-id"])
	assert.Equal(t, x.ScenarioName, headerData["x-scenario-name"])
	encodedData, _ := utils.HeaderJson.Marshal(x)
	newHeaderData := map[string]string{}
	json.Unmarshal(encodedData, &newHeaderData)
	assert.DeepEqual(t, newHeaderData, headerData)
}

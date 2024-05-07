package utils

import (
	jsoniter "github.com/json-iterator/go"
)

type CustomJSONTagHandler struct {
	jsoniter.API
}

/*
	NewCustomJSONTagHandler creates new JSON handler with custom struct tag

Example:

	var HeaderJSON = NewCustomJSONTagHandler("header")
	var BodyJSON = NewCustomJSONTagHandler("body")

	type CorrelationParam struct {
		CorrelationID string  `header:"x-correlation-id" body:"correlationId"`
		ScenarioID    *string `header:"x-scenario-id,omitempty" body:"scenarioId,omitempty"`
		SessionID     *string `header:"x-session-id,omitempty" body:"sessionId,omitempty"`
		ScenarioName  *string `header:"x-scenario-name,omitempty" body:"scenarioName,omitempty"`
	}

	x := CorrelationParam{}
	bodyData := map[string]string{
		"correlationId": "xyz",
		"scenarioId":    "c.ScenarioId",
		"sessionId":     "c.SessionId",
		"scenarioName":  "c.ScenarioName",
	}
	data, _ := json.Marshal(bodyData)
	BodyJSON.Unmarshal(data, &x)
	assert.Equal(t, x.CorrelationID, bodyData["correlationId"])
	assert.Equal(t, *x.ScenarioID, bodyData["scenarioId"])
	assert.Equal(t, *x.SessionID, bodyData["sessionId"])
	assert.Equal(t, *x.ScenarioName, bodyData["scenarioName"])

	headerData := map[string]string{
		"x-correlation-id": "xyz",
		"x-scenario-id":    "c.ScenarioId",
		"x-session-id":     "c.SessionId",
		"x-scenario-name":  "c.ScenarioName",
	}
	encodedData, _ := HeaderJSON.Marshal(x) //Note: CustomJSON.Marshal will only work on `struct` objects
	newData := map[string]string{}
	json.Unmarshal(encodedData, &newData)
	assert.DeepEqual(t, newData, headerData)
*/
func NewCustomJSONTagHandler(tag string) *CustomJSONTagHandler {
	return &CustomJSONTagHandler{
		API: jsoniter.Config{
			EscapeHTML:             true,
			SortMapKeys:            true,
			ValidateJsonRawMessage: true,
			TagKey:                 tag,
		}.Froze(),
	}
}

var HeaderJSON = NewCustomJSONTagHandler("header")
var BodyJSON = NewCustomJSONTagHandler("body")

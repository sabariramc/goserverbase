package utils_test

import (
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/utils"
)

func ExampleCustomJSONTagHandler() {
	var tag1Json = utils.NewCustomJSONTagHandler("tag1")
	var tag2Json = utils.NewCustomJSONTagHandler("tag2")

	type CorrelationParam struct {
		CorrelationID string `tag1:"x-correlation-id" tag2:"correlationId"`
		ScenarioID    string `tag1:"x-scenario-id,omitempty" tag2:"scenarioId,omitempty"`
		SessionID     string `tag1:"x-session-id,omitempty" tag2:"sessionId,omitempty"`
		ScenarioName  string `tag1:"x-scenario-name,omitempty" tag2:"scenarioName,omitempty"`
	}

	x := CorrelationParam{}
	tag2Data := map[string]string{
		"correlationId": "xyz",
		"scenarioId":    "ScenarioId",
		"sessionId":     "SessionId",
		"scenarioName":  "ScenarioName",
	}
	data, _ := json.Marshal(tag2Data)
	tag2Json.Unmarshal(data, &x)
	fmt.Println(tag2Data["correlationId"])
	fmt.Println(tag2Data["scenarioId"])
	fmt.Println(tag2Data["sessionId"])
	fmt.Println(tag2Data["scenarioName"])
	encodedData, _ := tag1Json.Marshal(x) //Note: CustomJSON.Marshal will only work on `struct` objects
	newData := map[string]string{}
	json.Unmarshal(encodedData, &newData)
	fmt.Printf("%+v", newData)
	//Output:
	//xyz
	//ScenarioId
	//SessionId
	//ScenarioName
	//map[x-correlation-id:xyz x-scenario-id:ScenarioId x-scenario-name:ScenarioName x-session-id:SessionId]

}

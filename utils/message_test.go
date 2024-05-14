package utils_test

import (
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v5/utils"
)

func ExampleMessage() {
	message := utils.NewMessage("event", "aws.test")
	message.AddPayload("payment", utils.Payload{
		"entity": map[string]interface{}{
			"id":     "pay_14341234",
			"amount": 123,
		},
	})
	message.AddPayload("bank", utils.Payload{
		"entity": map[string]interface{}{
			"id": "bank_fadsfas",
		},
	})
	message.AddPayload("customer", utils.Payload{
		"entity": map[string]interface{}{
			"id": "cust_fasdfsa",
		},
		"state": map[string]interface{}{
			"fromState": "created",
			"toState":   "updated",
		},
	})
	blob, _ := json.MarshalIndent(message, "", "    ")
	fmt.Println(string(blob))
	//Output:
	// {
	//     "entity": "event",
	//     "event": "aws.test",
	//     "contains": [
	//         "payment",
	//         "bank",
	//         "customer"
	//     ],
	//     "payload": {
	//         "bank": {
	//             "entity": {
	//                 "id": "bank_fadsfas"
	//             }
	//         },
	//         "customer": {
	//             "entity": {
	//                 "id": "cust_fasdfsa"
	//             },
	//             "state": {
	//                 "fromState":"created",
	//                 "toState":"updated"
	//             }
	//         },
	//         "payment": {
	//             "entity": {
	//                 "amount": 123,
	//                 "id": "pay_14341234"
	//             }
	//         }
	// }
}

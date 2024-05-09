// Package message abstracts the creation of error message in the format of utils.Message
package message

import (
	"context"
	"time"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
)

/*
CreateErrorMessage creates utils.Message with the provided attributes

Example:

	msg := CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), string(debug.Stack()), map[string]string{"test": "test"}, "5XX")
	// the above msg will decodes to the below json
	// {
	// 	"entity": "error",
	// 	"event": "com.test.message",
	// 	"contains": [
	// 		"category",
	// 		"source",
	// 		"stackTrace",
	// 		"version",
	// 		"errorData",
	// 		"identity",
	// 		"correlation",
	// 		"timestamp"
	// 	],
	// 	"payload": {
	// 		"category": {
	// 			"entity": {
	// 				"category": "5XX"
	// 			}
	// 		},
	// 		"correlation": {
	// 			"entity": {
	// 				"correlationId": ""
	// 			}
	// 		},
	// 		"errorData": {
	// 			"entity": {
	// 				"errorData": {
	// 					"test": "test"
	// 				}
	// 			}
	// 		},
	// 		"identity": {
	// 			"entity": {
	// 				"userId": "cust_test_id"
	// 			}
	// 		},
	// 		"source": {
	// 			"entity": {
	// 				"source": "test_service"
	// 			}
	// 		},
	// 		"stackTrace": {
	// 			"entity": {
	// 				"error": "test",
	// 				"stackTrace": "goroutine 19 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x6b\ngithub.com/sabariramc/goserverbase/v5/errors/message_test.TestMessage(0xc0000c71e0)\n\t/<<>>/Library/goserverbase/errors/message/message_test.go:22 +0x18f\ntesting.tRunner(0xc0000c71e0, 0x7699e0)\n\t/usr/local/go/src/testing/testing.go:1689 +0x1da\ncreated by testing.(*T).Run in goroutine 1\n\t/usr/local/go/src/testing/testing.go:1742 +0x7d3\n"
	// 			}
	// 		},
	// 		"timestamp": {
	// 			"entity": {
	// 				"timestamp": 1715239718
	// 			}
	// 		},
	// 		"version": {
	// 			"entity": {
	// 				"version": "v1"
	// 			}
	// 		}
	// 	}
	// }
*/
func CreateErrorMessage(ctx context.Context, serviceName, errorCode string, err error, stackTrace string, errorData interface{}, alertType string) *utils.Message {
	msg := utils.NewMessage("error", errorCode)
	msg.AddPayload("category", utils.Payload{"entity": map[string]string{"category": alertType}})
	msg.AddPayload("source", utils.Payload{"entity": map[string]string{"source": serviceName}})
	stackTracePayload := map[string]string{
		"stackTrace": stackTrace,
	}
	if err != nil {
		stackTracePayload["error"] = err.Error()
	}
	msg.AddPayload("stackTrace", utils.Payload{"entity": stackTracePayload})
	msg.AddPayload("version", utils.Payload{"entity": map[string]string{"version": "v1"}})
	msg.AddPayload("errorData", utils.Payload{"entity": map[string]interface{}{"errorData": errorData}})
	msg.AddPayload("identity", utils.Payload{"entity": log.ExtractUserIdentifier(ctx).GetPayload()})
	msg.AddPayload("correlation", utils.Payload{"entity": log.ExtractCorrelationParam(ctx).GetPayload()})
	msg.AddPayload("timestamp", utils.Payload{"entity": map[string]int64{"timestamp": time.Now().Unix()}})
	return msg
}

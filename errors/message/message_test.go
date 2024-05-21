package message_test

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/sabariramc/goserverbase/v6/errors/message"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

func TestMessage(t *testing.T) {
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{
		CorrelationID: "xyz",
	})
	custID := "cust_test_id"
	ctx = correlation.GetContextWithUserIdentifier(ctx, &correlation.UserIdentifier{UserID: &custID})
	msg := message.CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), string(debug.Stack()), map[string]string{"test": "test"}, "5XX")
	blob, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(blob))
}

func Example() {
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{
		CorrelationID: "xyz",
	})
	custID := "cust_test_id"
	ctx = correlation.GetContextWithUserIdentifier(ctx, &correlation.UserIdentifier{UserID: &custID})
	msg := message.CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), string(debug.Stack()), map[string]string{"test": "test"}, "5XX")
	blob, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(blob))
	//Output:
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
	// 				"correlationId": "xyz"
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
	// 				"stackTrace": "goroutine 1 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x5e\ngithub.com/sabariramc/goserverbase/v6/errors/message_test.Example()\n\t/<<filepath>>/goserverbase/errors/message/message_test.go:27 +0x15f\ntesting.runExample({{0x64d3df, 0x7}, 0x677028, {0x65c213, 0x4c5}, 0x0})\n\t/usr/local/go/src/testing/run_example.go:63 +0x2de\ntesting.runExamples(0x6b4fc0?, {0x827280, 0x1, 0x1?})\n\t/usr/local/go/src/testing/example.go:40 +0x126\ntesting.(*M).Run(0xc0001188c0)\n\t/usr/local/go/src/testing/testing.go:2029 +0x75d\nmain.main()\n\t_testmain.go:51 +0x16c\n"
	// 			}
	// 		},
	// 		"timestamp": {
	// 			"entity": {
	// 				"timestamp": 1715671891
	// 			}
	// 		},
	// 		"version": {
	// 			"entity": {
	// 				"version": "v1"
	// 			}
	// 		}
	// 	}
	// }
}

package message_test

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/sabariramc/goserverbase/v5/errors/message"
	"github.com/sabariramc/goserverbase/v5/log"
)

func TestMessage(t *testing.T) {
	ctx := log.GetContextWithCorrelationParam(context.Background(), log.GetDefaultCorrelationParam("test_service"))
	custID := "cust_test_id"
	ctx = log.GetContextWithUserIdentifier(ctx, &log.UserIdentifier{UserID: &custID})
	msg := message.CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), string(debug.Stack()), map[string]string{"test": "test"}, "5XX")
	blob, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(blob))
}

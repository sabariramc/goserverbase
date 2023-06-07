package message_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v3/errors/message"
	"github.com/sabariramc/goserverbase/v3/log"
)

func TestMessage(t *testing.T) {
	ctx := log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam("test_service"))
	ctx = log.GetContextWithCustomerId(ctx, &log.CustomerIdentifier{CustomerId: "cust_test_id"})
	msg := message.CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), "fasdfafasdfasf", map[string]string{"test": "test"}, "5XX")
	blob, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(blob))
}

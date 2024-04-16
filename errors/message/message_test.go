package message_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v5/errors/message"
	"github.com/sabariramc/goserverbase/v5/log"
)

func TestMessage(t *testing.T) {
	ctx := log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam("test_service"))
	custId := "cust_test_id"
	ctx = log.GetContextWithCustomerID(ctx, &log.CustomerIdentifier{UserID: &custId})
	msg := message.CreateErrorMessage(ctx, "test_service", "com.test.message", fmt.Errorf("test"), "fasdfafasdfasf", map[string]string{"test": "test"}, "5XX")
	blob, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(blob))
}

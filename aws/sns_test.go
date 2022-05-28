package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/aws"
	"github.com/sabariramc/goserverbase/utils"
)

func TestSNSClient(t *testing.T) {
	arn := AWSTestConfig.SNS.Arn
	ctx := GetCorrelationContext()
	snsClient := aws.GetDefaultSNSClient(AWSTestLogger)
	message := utils.NewMessage(utils.EventEntity, "sns.test")
	message.AddPayload("payment", &utils.Payload{
		Entity: map[string]interface{}{
			"id":     "pay_14341234",
			"amount": 123,
		},
	})
	message.AddPayload("bank", &utils.Payload{
		Entity: map[string]interface{}{
			"id":                "bank_fadsfas",
			"bankAccountNumber": "0000021312",
		},
	})
	message.AddPayload("customer", &utils.Payload{
		Entity: map[string]interface{}{
			"id": "cust_fasdfsa",
		},
	})
	err := snsClient.Publish(ctx, &arn, nil, message, nil)
	if err != nil {
		t.Fatal(err)
	}

}

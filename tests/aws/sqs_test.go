package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/aws"
	"github.com/sabariramc/goserverbase/utils"
)

func GetMessage() *utils.Message {
	message := utils.NewMessage(utils.EventEntity, "aws.test")
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
	return message
}

func TestSQSClient(t *testing.T) {
	queueUrl := AWSTestConfig.SQS.QueueURL
	ctx := GetCorrelationContext()

	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueUrl)
	message := GetMessage()
	err := sqsClient.SendMessage(ctx, message, map[string]string{
		"id": uuid.NewString(),
	}, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessage(ctx, messageList[0].ReceiptHandle)
	if err != nil {
		t.Fatal(err)
	}
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)

	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		sqsMessageList[i] = &aws.BatchQueueMessage{
			Id:      &id,
			Message: message,
			Attribute: map[string]string{
				"id": uuid.NewString(),
			},
		}
	}
	_, err = sqsClient.SendMessageBatch(ctx, sqsMessageList, 1)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatch(ctx, deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

func TestSQSFIFOClient(t *testing.T) {
	queueUrl := AWSTestConfig.FIFOSQS.QueueURL
	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueUrl)
	groupId, dedupId := uuid.NewString(), uuid.NewString()
	ctx := GetCorrelationContext()
	message := GetMessage()
	err := sqsClient.SendMessage(ctx, message, map[string]string{
		"id": uuid.NewString(),
	}, 0, &groupId, &dedupId)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessage(ctx, messageList[0].ReceiptHandle)
	if err != nil {
		t.Fatal(err)
	}
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)
	groupId = "data"
	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		sqsMessageList[i] = &aws.BatchQueueMessage{
			Id:      &id,
			Message: GetMessage(),
			Attribute: map[string]string{
				"id": uuid.NewString(),
			},
			MessageDeduplicationId: &id,
			MessageGroupId:         &groupId,
		}
	}
	_, err = sqsClient.SendMessageBatch(ctx, sqsMessageList, 0)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatch(ctx, deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

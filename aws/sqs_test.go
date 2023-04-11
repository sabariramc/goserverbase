package aws_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v2/aws"
	"github.com/sabariramc/goserverbase/v2/utils"
)

func GetMessage() *utils.Message {
	message := utils.NewMessage("event", "aws.test")
	message.AddPayload("payment", &utils.Payload{
		"entity": map[string]interface{}{
			"id":     "pay_14341234",
			"amount": 123,
		},
	})
	message.AddPayload("bank", &utils.Payload{
		"entity": map[string]interface{}{
			"id":                "bank_fadsfas",
			"bankAccountNumber": "0000021312",
		},
	})
	message.AddPayload("customer", &utils.Payload{
		"entity": map[string]interface{}{
			"id": "cust_fasdfsa",
		},
	})
	return message
}

func TestSQSClient(t *testing.T) {
	queueUrl := AWSTestConfig.AWS.SQS_URL
	ctx := GetCorrelationContext()

	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueUrl)
	message := GetMessage()
	err := sqsClient.SendMessageWithContext(ctx, message, map[string]string{
		"id": uuid.NewString(),
	}, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessageWithContext(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessageWithContext(ctx, messageList[0].ReceiptHandle)
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
	out, err := sqsClient.SendMessageBatchWithContext(ctx, sqsMessageList, 1)
	fmt.Println(out)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessageWithContext(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	fmt.Println(deleteMap)
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatchWithContext(ctx, deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

func TestSQSFIFOClient(t *testing.T) {
	queueUrl := AWSTestConfig.AWS.FIFO_SQS_URL
	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueUrl)
	groupId, dedupeId := uuid.NewString(), uuid.NewString()
	ctx := GetCorrelationContext()
	message := GetMessage()
	err := sqsClient.SendMessageWithContext(ctx, message, map[string]string{
		"id": uuid.NewString(),
	}, 0, &groupId, &dedupeId)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessageWithContext(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessageWithContext(ctx, messageList[0].ReceiptHandle)
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
	_, err = sqsClient.SendMessageBatchWithContext(ctx, sqsMessageList, 0)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessageWithContext(ctx, 10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatchWithContext(ctx, deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

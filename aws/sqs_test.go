package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/aws"
	"github.com/sabariramc/goserverbase/v6/utils"
	"gotest.tools/assert"
)

func GetMessage() *utils.Message {
	message := utils.NewMessage("event", "aws.test")
	message.AddPayload("payment", utils.Payload{
		"entity": map[string]interface{}{
			"id":     "pay_14341234",
			"amount": 123,
		},
	})
	message.AddPayload("bank", utils.Payload{
		"entity": map[string]interface{}{
			"id":                "bank_fadsfas",
			"bankAccountNumber": "0000021312",
		},
	})
	message.AddPayload("customer", utils.Payload{
		"entity": map[string]interface{}{
			"id": "cust_fasdfsa",
		},
	})
	return message
}

func TestSQSClient(t *testing.T) {
	queueURL := AWSTestConfig.AWS.SQS
	ctx := GetCorrelationContext()
	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueURL)
	sqsClient.PurgeQueue(context.Background(), &sqs.PurgeQueueInput{
		QueueUrl: &queueURL,
	})
	message := GetMessage()
	id := uuid.NewString()
	_, err := sqsClient.SendMessage(ctx, message, map[string]any{
		"id": id,
	}, 0)
	assert.NilError(t, err)
	messageRes, err := sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	assert.NilError(t, err)
	messageList := messageRes.Messages
	assert.Equal(t, id, *messageList[0].MessageAttributes["id"].StringValue)
	_, err = sqsClient.DeleteMessage(ctx, messageList[0].ReceiptHandle)
	assert.NilError(t, err)
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)
	idMap := map[string]bool{}
	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		idMap[id] = true
		sqsMessageList[i] = &aws.BatchQueueMessage{
			ID:      &id,
			Message: message,
			Attribute: map[string]any{
				"id": id,
			},
		}
	}
	out, err := sqsClient.SendMessageBatch(ctx, sqsMessageList, 1)
	fmt.Println(out)
	assert.NilError(t, err)
	messageRes, err = sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	assert.NilError(t, err)
	messageList = messageRes.Messages
	deleteMap := make(map[string]*string, len(messageList))
	fmt.Println(deleteMap)
	for _, m := range messageList {
		id := *m.MessageAttributes["id"].StringValue
		deleteMap[id] = m.ReceiptHandle
		delete(idMap, id)
	}
	assert.Equal(t, len(idMap), 0)
	_, err = sqsClient.DeleteMessageBatch(ctx, deleteMap)
	assert.NilError(t, err)

}

func TestSQSFIFOClient(t *testing.T) {
	queueURL := AWSTestConfig.AWS.FIFOSQS
	sqsClient := aws.GetDefaultSQSClient(AWSTestLogger, queueURL)
	sqsClient.PurgeQueue(context.Background(), &sqs.PurgeQueueInput{
		QueueUrl: &queueURL,
	})
	groupID, dedupeID := uuid.NewString(), uuid.NewString()
	ctx := GetCorrelationContext()
	message := GetMessage()
	_, err := sqsClient.SendMessageFIFO(ctx, message, map[string]any{
		"id": uuid.NewString(),
	}, 0, &groupID, &dedupeID)
	assert.NilError(t, err)
	messageRes, err := sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	assert.NilError(t, err)
	messageList := messageRes.Messages
	_, err = sqsClient.DeleteMessage(ctx, messageList[0].ReceiptHandle)
	assert.NilError(t, err)
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)
	groupID = "data"
	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		sqsMessageList[i] = &aws.BatchQueueMessage{
			ID:      &id,
			Message: GetMessage(),
			Attribute: map[string]any{
				"id": uuid.NewString(),
			},
			MessageDeduplicationID: &id,
			MessageGroupID:         &groupID,
		}
	}
	_, err = sqsClient.SendMessageBatch(ctx, sqsMessageList, 0)
	assert.NilError(t, err)
	messageRes, err = sqsClient.ReceiveMessage(ctx, 10, 10, 3)
	assert.NilError(t, err)
	messageList = messageRes.Messages
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatch(ctx, deleteMap)
	assert.NilError(t, err)
}

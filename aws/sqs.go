package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"sabariram.com/goserverbase/log"
	"sabariram.com/goserverbase/utils"
)

type SQS struct {
	_        struct{}
	client   *sqs.SQS
	log      *log.Log
	queueURL *string
	ctx      context.Context
}

var defaultSQSClient *sqs.SQS

var DefaultMaxMessages int64 = 10

func GetDefaultSQSClient(ctx context.Context, queueURL string) *SQS {
	if defaultSecretManagerClient == nil {
		defaultSQSClient = GetAWSSQSClient(defaultAWSSession)
	}
	return GetSQSClient(ctx, defaultSQSClient, queueURL)
}

func GetAWSSQSClient(awsSession *session.Session) *sqs.SQS {
	client := sqs.New(awsSession)
	return client
}

func GetSQSClient(ctx context.Context, client *sqs.SQS, queueURL string) *SQS {
	return &SQS{client: client, queueURL: &queueURL, log: log.GetDefaultLogger(), ctx: ctx}
}

func (s *SQS) IsFIFO() bool {
	return strings.HasSuffix(*s.queueURL, ".fifo")
}

func GetQueueURL(queueName string, sqsClient *sqs.SQS, ctx context.Context) (*string, error) {
	log := log.GetDefaultLogger()
	prefix := utils.Getenv("stage", "dev")
	systemPefix := utils.Getenv("queuePrefix", "")
	if systemPefix != "" {
		prefix = fmt.Sprintf("%v_%v", prefix, systemPefix)
	}
	queueName = fmt.Sprintf("%v_%v", prefix, queueName)
	req := &sqs.GetQueueUrlInput{
		QueueName: &queueName}
	log.Debug("SQS get queue url request", req)
	res, err := sqsClient.GetQueueUrlWithContext(ctx, req)
	if err != nil {
		log.Error("Error creating queue URL", err)
		return nil, err
	}
	log.Debug("SQS get queue url response", res)
	log.Debug("Queue URL", res.QueueUrl)
	return res.QueueUrl, nil
}

func (s *SQS) SendMessage(message interface{}, attribute map[string]string, delayInSeconds int64, messageDeduplicationId, messageGroupId *string) error {
	body, err := utils.GetString(message)
	if err != nil {
		return err
	}
	messageAttributes := s.GetAttribure(attribute)
	req := &sqs.SendMessageInput{
		QueueUrl:          s.queueURL,
		DelaySeconds:      &delayInSeconds,
		MessageBody:       body,
		MessageAttributes: messageAttributes,
	}
	if s.IsFIFO() {
		req.MessageDeduplicationId = messageDeduplicationId
		req.MessageGroupId = messageGroupId
	}
	s.log.Debug("Queue send message request", req)
	res, err := s.client.SendMessageWithContext(s.ctx, req)
	s.log.Debug("Queue send message response", res)
	if err != nil {
		s.log.Error("Error in sending message", err)
		return err
	}
	return nil
}

type BatchQueueMessage struct {
	Id                     *string
	Message                interface{}
	Attribute              map[string]string
	MessageDeduplicationId *string
	MessageGroupId         *string
}

func (s *SQS) SendMessageBatch(messageList []*BatchQueueMessage, delayInSeconds int64) (*sqs.SendMessageBatchOutput, error) {
	isFifo := s.IsFIFO()
	messageReq := make([]*sqs.SendMessageBatchRequestEntry, len(messageList))
	i := 0
	for _, message := range messageList {
		body, err := utils.GetString(message.Message)
		if err != nil {
			return nil, err
		}
		m := &sqs.SendMessageBatchRequestEntry{
			Id:                message.Id,
			DelaySeconds:      &delayInSeconds,
			MessageAttributes: s.GetAttribure(message.Attribute), MessageBody: body,
		}
		if isFifo {
			m.MessageDeduplicationId = message.MessageDeduplicationId
			m.MessageGroupId = message.MessageGroupId
		}
		messageReq[i] = m
		i++
	}
	req := &sqs.SendMessageBatchInput{
		Entries:  messageReq,
		QueueUrl: s.queueURL,
	}
	res, err := s.client.SendMessageBatchWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("Error in batch send message", err)
	}
	s.log.Debug("Queue send message batch message", res)
	return res, err
}

func (s *SQS) GetAttribure(attribute map[string]string) map[string]*sqs.MessageAttributeValue {
	if len(attribute) == 0 {
		return nil
	}
	messageAttributes := make(map[string]*sqs.MessageAttributeValue, len(attribute))
	for key, value := range attribute {
		messageAttributes[key] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	return messageAttributes
}

func (s *SQS) ReceiveMessage(timeoutInSeconds int64, maxNumberOfMessages int64, waitTimeInSeconds int64) ([]*sqs.Message, error) {
	req := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            s.queueURL,
		MaxNumberOfMessages: &maxNumberOfMessages,
		VisibilityTimeout:   &timeoutInSeconds,
		WaitTimeSeconds:     &waitTimeInSeconds,
	}
	s.log.Debug("Queue receive request", req)
	msgResult, err := s.client.ReceiveMessageWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("Error in fetching message", err)
		return nil, err
	}
	s.log.Debug("Queue receive response", msgResult)

	return msgResult.Messages, nil
}

func (s *SQS) DeleteMessage(receiptHandler *string) error {
	req := &sqs.DeleteMessageInput{
		QueueUrl:      s.queueURL,
		ReceiptHandle: receiptHandler,
	}
	s.log.Debug("Queue delete request", req)
	res, err := s.client.DeleteMessageWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("Error in delete message", err)
		return err
	}
	s.log.Debug("Queue delete resposne", res)
	return nil
}

func (s *SQS) DeleteMessageBatch(receiptHandlerMap map[string]*string) (*sqs.DeleteMessageBatchOutput, error) {
	if len(receiptHandlerMap) > 10 {
		return nil, fmt.Errorf("too many message to delete")
	}
	entries := make([]*sqs.DeleteMessageBatchRequestEntry, len(receiptHandlerMap))
	i := 0
	for key, value := range receiptHandlerMap {
		v := key
		entries[i] = &sqs.DeleteMessageBatchRequestEntry{
			Id:            &v,
			ReceiptHandle: value,
		}
		i++
	}
	req := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
		Entries:  entries,
	}
	s.log.Debug("Queue delete batch request", req)
	res, err := s.client.DeleteMessageBatchWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("Error in delete batch message", err)
		return nil, err
	}
	s.log.Debug("Queue delete batch resposne", res)
	return res, nil
}

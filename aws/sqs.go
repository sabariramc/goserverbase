package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sabariramc/goserverbase/log"
	"github.com/sabariramc/goserverbase/utils"
)

type SQS struct {
	_ struct{}
	*sqs.SQS
	log      *log.Logger
	queueURL *string
}

var defaultSQSClient *sqs.SQS
var ErrTooManyMessageToDelete = fmt.Errorf("too many message in receiptHandlerMap(should be less that 10)")
var DefaultMaxMessages int64 = 10

func GetDefaultSQSClient(logger *log.Logger, queueURL string) *SQS {
	if defaultSecretManagerClient == nil {
		defaultSQSClient = NewAWSSQSClient(defaultAWSSession)
	}
	return NewSQSClient(logger, defaultSQSClient, queueURL)
}

func NewAWSSQSClient(awsSession *session.Session) *sqs.SQS {
	client := sqs.New(awsSession)
	return client
}

func NewSQSClient(logger *log.Logger, client *sqs.SQS, queueURL string) *SQS {
	return &SQS{SQS: client, queueURL: &queueURL, log: logger}
}

func (s *SQS) IsFIFO() bool {
	return strings.HasSuffix(*s.queueURL, ".fifo")
}

func GetQueueUrlWithContext(ctx context.Context, logger *log.Logger, queueName string, sqsClient *sqs.SQS) (*string, error) {
	req := &sqs.GetQueueUrlInput{
		QueueName: &queueName}
	logger.Debug(ctx, "SQS get queue url request", req)
	res, err := sqsClient.GetQueueUrlWithContext(ctx, req)
	if err != nil {
		logger.Error(ctx, "Error creating queue URL", err)
		return nil, fmt.Errorf("SQS.GetQueueURL: %w", err)
	}
	logger.Debug(ctx, "SQS get queue url response", res)
	logger.Debug(ctx, "Queue URL", res.QueueUrl)
	return res.QueueUrl, nil
}

func (s *SQS) SendMessageWithContext(ctx context.Context, message *utils.Message, attribute map[string]string, delayInSeconds int64, messageDeduplicationId, messageGroupId *string) error {
	body, err := utils.Searialize(message)
	if err != nil {
		return fmt.Errorf("SQS.SendMessage: %w", err)
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
	s.log.Debug(ctx, "Queue send message request", req)
	res, err := s.SQS.SendMessageWithContext(ctx, req)
	s.log.Debug(ctx, "Queue send message response", res)
	if err != nil {
		s.log.Error(ctx, "Error in sending message", err)
		return fmt.Errorf("SQS.SendMessage: %w", err)
	}
	return nil
}

type BatchQueueMessage struct {
	Id                     *string
	Message                *utils.Message
	Attribute              map[string]string
	MessageDeduplicationId *string
	MessageGroupId         *string
}

func (s *SQS) SendMessageBatchWithContext(ctx context.Context, messageList []*BatchQueueMessage, delayInSeconds int64) (*sqs.SendMessageBatchOutput, error) {
	isFifo := s.IsFIFO()
	messageReq := make([]*sqs.SendMessageBatchRequestEntry, len(messageList))
	i := 0
	for _, message := range messageList {
		body, err := utils.Searialize(message.Message)
		if err != nil {
			return nil, fmt.Errorf("SQS.SendMessageBatch: %w", err)
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
	res, err := s.SQS.SendMessageBatchWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in batch send message", err)
		return res, fmt.Errorf("SQS.SendMessageBatch : %w", err)
	}
	s.log.Debug(ctx, "Queue send message batch message", res)
	return res, nil
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

func (s *SQS) ReceiveMessageWithContext(ctx context.Context, timeoutInSeconds int64, maxNumberOfMessages int64, waitTimeInSeconds int64) ([]*sqs.Message, error) {
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
	s.log.Debug(ctx, "Queue receive request", req)
	msgResult, err := s.SQS.ReceiveMessageWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in fetching message", err)
		return nil, fmt.Errorf("SQS.ReceiveMessage: %w", err)
	}
	s.log.Debug(ctx, "Queue receive response", msgResult)

	return msgResult.Messages, nil
}

func (s *SQS) DeleteMessageWithContext(ctx context.Context, receiptHandler *string) error {
	req := &sqs.DeleteMessageInput{
		QueueUrl:      s.queueURL,
		ReceiptHandle: receiptHandler,
	}
	s.log.Debug(ctx, "Queue delete request", req)
	res, err := s.SQS.DeleteMessageWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in delete message", err)
		return fmt.Errorf("SQS.DeleteMessage: %w", err)
	}
	s.log.Debug(ctx, "Queue delete resposne", res)
	return nil
}

func (s *SQS) DeleteMessageBatchWithContext(ctx context.Context, receiptHandlerMap map[string]*string) (*sqs.DeleteMessageBatchOutput, error) {
	if len(receiptHandlerMap) > 10 {
		return nil, fmt.Errorf("SQS.DeleteMessage: %w", ErrTooManyMessageToDelete)
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
	s.log.Debug(ctx, "Queue delete batch request", req)
	res, err := s.SQS.DeleteMessageBatchWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in delete batch message", err)
		return nil, fmt.Errorf("SQS.DeleteMessage: %w", err)
	}
	s.log.Debug(ctx, "Queue delete batch resposne", res)
	return res, nil
}

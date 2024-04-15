package aws

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
)

type SQS struct {
	_ struct{}
	*sqs.Client
	log      log.Log
	queueURL *string
}

var defaultSQSClient *sqs.Client
var ErrTooManyMessageToDelete = fmt.Errorf("too many message in receiptHandlerMap(should be less that 10)")
var DefaultMaxMessages int64 = 10

func GetDefaultSQSClient(logger log.Log, queueURL string) *SQS {
	if defaultSQSClient == nil {
		defaultSQSClient = NewSQSClientWithConfig(*defaultAWSConfig)
	}
	return NewSQSClient(logger, defaultSQSClient, queueURL)
}

func NewSQSClientWithConfig(awsConfig aws.Config) *sqs.Client {
	client := sqs.NewFromConfig(awsConfig)
	return client
}

func NewSQSClient(logger log.Log, client *sqs.Client, queueURL string) *SQS {
	return &SQS{Client: client, queueURL: &queueURL, log: logger}
}

func (s *SQS) IsFIFO() bool {
	return strings.HasSuffix(*s.queueURL, ".fifo")
}

func GetQueueURL(ctx context.Context, logger log.Log, queueName string, sqsClient *sqs.Client) (*string, error) {
	req := &sqs.GetQueueUrlInput{
		QueueName: &queueName}
	res, err := sqsClient.GetQueueUrl(ctx, req)
	if err != nil {
		logger.Error(ctx, "Error creating queue URL", err)
		return nil, fmt.Errorf("SQS.GetQueueURL: %w", err)
	}
	return res.QueueUrl, nil
}

func (s *SQS) SendMessage(ctx context.Context, message *utils.Message, attribute map[string]string, delayInSeconds int32) (*sqs.SendMessageOutput, error) {
	body, err := utils.LoadString(message)
	if err != nil {
		return nil, fmt.Errorf("SQS.SendMessage: %w", err)
	}
	messageAttributes := s.GenerateAttribute(ctx, attribute)
	req := &sqs.SendMessageInput{
		QueueUrl:          s.queueURL,
		DelaySeconds:      delayInSeconds,
		MessageBody:       body,
		MessageAttributes: messageAttributes,
	}
	res, err := s.Client.SendMessage(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error sending message", err)
		return res, fmt.Errorf("SQS.SendMessage: error sending message: %w", err)
	}
	return res, nil
}

func (s *SQS) SendMessageFIFO(ctx context.Context, message *utils.Message, attribute map[string]string, delayInSeconds int32, messageDeduplicationID, messageGroupID *string) (*sqs.SendMessageOutput, error) {
	body, err := utils.LoadString(message)
	if err != nil {
		return nil, fmt.Errorf("SQS.SendMessage: %w", err)
	}
	messageAttributes := s.GenerateAttribute(ctx, attribute)
	req := &sqs.SendMessageInput{
		QueueUrl:               s.queueURL,
		DelaySeconds:           delayInSeconds,
		MessageBody:            body,
		MessageAttributes:      messageAttributes,
		MessageDeduplicationId: messageDeduplicationID,
		MessageGroupId:         messageGroupID,
	}
	res, err := s.Client.SendMessage(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error sending message", err)
		return res, fmt.Errorf("SQS.SendMessage: error sending message: %w", err)
	}
	return res, nil
}

type BatchQueueMessage struct {
	ID                     *string
	Message                *utils.Message
	Attribute              map[string]string
	MessageDeduplicationID *string
	MessageGroupID         *string
}

func (s *SQS) SendMessageBatch(ctx context.Context, messageList []*BatchQueueMessage, delayInSeconds int32) (*sqs.SendMessageBatchOutput, error) {
	isFifo := s.IsFIFO()
	messageReq := make([]types.SendMessageBatchRequestEntry, len(messageList))
	i := 0
	for _, message := range messageList {
		body, err := utils.LoadString(message.Message)
		if err != nil {
			return nil, fmt.Errorf("SQS.SendMessageBatch: %w", err)
		}
		m := types.SendMessageBatchRequestEntry{
			Id:                message.ID,
			DelaySeconds:      delayInSeconds,
			MessageAttributes: s.GenerateAttribute(ctx, message.Attribute), MessageBody: body,
		}
		if isFifo {
			m.MessageDeduplicationId = message.MessageDeduplicationID
			m.MessageGroupId = message.MessageGroupID
		}
		messageReq[i] = m
		i++
	}
	req := &sqs.SendMessageBatchInput{
		Entries:  messageReq,
		QueueUrl: s.queueURL,
	}
	res, err := s.Client.SendMessageBatch(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in batch send message", err)
		return res, fmt.Errorf("SQS.SendMessageBatch: error sending message: %w", err)
	}
	return res, nil
}

func (s *SQS) GenerateAttribute(ctx context.Context, attribute map[string]string) map[string]types.MessageAttributeValue {
	if attribute == nil {
		attribute = map[string]string{}
	}
	correlation := log.GetCorrelationParam(ctx)
	if correlation != nil {
		headers := correlation.GetHeader()
		maps.Copy(headers, attribute)
	}
	messageAttributes := make(map[string]types.MessageAttributeValue, len(attribute))
	for key, value := range attribute {
		messageAttributes[key] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	return messageAttributes
}

func (s *SQS) ParseAttribute(messageAttributes map[string]types.MessageAttributeValue) map[string]string {
	if len(messageAttributes) == 0 {
		return nil
	}
	attributes := make(map[string]string, len(messageAttributes))
	for key, value := range messageAttributes {
		attributes[key] = *value.StringValue
	}
	return attributes
}

func (s *SQS) ReceiveMessage(ctx context.Context, timeoutInSeconds int32, maxNumberOfMessages int32, waitTimeInSeconds int32) (*sqs.ReceiveMessageOutput, error) {
	req := &sqs.ReceiveMessageInput{
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeName(types.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            s.queueURL,
		MaxNumberOfMessages: maxNumberOfMessages,
		VisibilityTimeout:   timeoutInSeconds,
		WaitTimeSeconds:     waitTimeInSeconds,
	}
	msgResult, err := s.Client.ReceiveMessage(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in fetching message", err)
		return msgResult, fmt.Errorf("SQS.ReceiveMessage: error receiving message: %w", err)
	}
	return msgResult, nil
}

func (s *SQS) DeleteMessage(ctx context.Context, receiptHandler *string) (*sqs.DeleteMessageOutput, error) {
	req := &sqs.DeleteMessageInput{
		QueueUrl:      s.queueURL,
		ReceiptHandle: receiptHandler,
	}
	res, err := s.Client.DeleteMessage(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in delete message", err)
		return res, fmt.Errorf("SQS.DeleteMessage: %w", err)
	}
	return res, nil
}

func (s *SQS) DeleteMessageBatch(ctx context.Context, receiptHandlerMap map[string]*string) (*sqs.DeleteMessageBatchOutput, error) {
	if len(receiptHandlerMap) > 10 {
		return nil, fmt.Errorf("SQS.DeleteMessage: %w", ErrTooManyMessageToDelete)
	}
	entries := make([]types.DeleteMessageBatchRequestEntry, len(receiptHandlerMap))
	i := 0
	for key, value := range receiptHandlerMap {
		v := key
		entries[i] = types.DeleteMessageBatchRequestEntry{
			Id:            &v,
			ReceiptHandle: value,
		}
		i++
	}
	req := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
		Entries:  entries,
	}
	res, err := s.Client.DeleteMessageBatch(ctx, req)
	if err != nil {
		s.log.Error(ctx, "Error in delete batch message", err)
		return nil, fmt.Errorf("SQS.DeleteMessage: %w", err)
	}
	return res, nil
}

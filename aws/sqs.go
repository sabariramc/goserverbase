package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// SQS provides methods to interact with AWS Simple Queue Service (SQS).
type SQS struct {
	_ struct{}
	*sqs.Client
	log      log.Log
	queueURL *string
}

// defaultSQSClient is the default AWS SQS client.
var defaultSQSClient *sqs.Client

// ErrTooManyMessageToDelete is an error indicating too many messages to delete.
var ErrTooManyMessageToDelete = fmt.Errorf("too many messages in receiptHandlerMap (should be less than 10)")

// DefaultMaxMessages is the default maximum number of messages.
var DefaultMaxMessages int64 = 10

// GetDefaultSQSClient returns the default SQS client using the provided logger and queue URL.
func GetDefaultSQSClient(logger log.Log, queueURL string) *SQS {
	if defaultSQSClient == nil {
		defaultSQSClient = NewSQSClientWithConfig(*defaultAWSConfig)
	}
	return NewSQSClient(logger, defaultSQSClient, queueURL)
}

// NewSQSClientWithConfig creates a new SQS client with the provided AWS configuration.
func NewSQSClientWithConfig(awsConfig aws.Config) *sqs.Client {
	client := sqs.NewFromConfig(awsConfig)
	return client
}

// NewSQSClient creates a new SQS instance with the provided logger, SQS client, and queue URL.
func NewSQSClient(logger log.Log, client *sqs.Client, queueURL string) *SQS {
	return &SQS{Client: client, queueURL: &queueURL, log: logger}
}

// IsFIFO checks if the SQS queue is FIFO.
func (s *SQS) IsFIFO() bool {
	return strings.HasSuffix(*s.queueURL, ".fifo")
}

// GetQueueURL returns the URL of the queue with the given name.
func GetQueueURL(ctx context.Context, logger log.Log, queueName string, sqsClient *sqs.Client) (*string, error) {
	req := &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}
	res, err := sqsClient.GetQueueUrl(ctx, req)
	if err != nil {
		logger.Error(ctx, "Error creating queue URL", err)
		return nil, fmt.Errorf("SQS.GetQueueURL: %w", err)
	}
	return res.QueueUrl, nil
}

// SendMessage sends a message to the SQS queue with optional attributes and delay.
func (s *SQS) SendMessage(ctx context.Context, message *utils.Message, attribute map[string]interface{}, delayInSeconds int32) (*sqs.SendMessageOutput, error) {
	body, err := marshal(message)
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

// SendMessageFIFO sends a message to the FIFO SQS queue with optional attributes, delay, and deduplication/group ID.
func (s *SQS) SendMessageFIFO(ctx context.Context, message *utils.Message, attribute map[string]interface{}, delayInSeconds int32, messageDeduplicationID, messageGroupID *string) (*sqs.SendMessageOutput, error) {
	body, err := marshal(message)
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

// BatchQueueMessage represents a message to be sent in a batch to SQS.
type BatchQueueMessage struct {
	ID                     *string
	Message                *utils.Message
	Attribute              map[string]interface{}
	MessageDeduplicationID *string
	MessageGroupID         *string
}

// SendMessageBatch sends multiple messages in a batch to the SQS queue.
func (s *SQS) SendMessageBatch(ctx context.Context, messageList []*BatchQueueMessage, delayInSeconds int32) (*sqs.SendMessageBatchOutput, error) {
	isFifo := s.IsFIFO()
	messageReq := make([]types.SendMessageBatchRequestEntry, len(messageList))
	i := 0
	for _, message := range messageList {
		body, err := marshal(message.Message)
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

// GenerateAttribute generates message attributes from the given attribute map.
func (s *SQS) GenerateAttribute(ctx context.Context, attribute map[string]interface{}) map[string]types.MessageAttributeValue {
	if attribute == nil {
		attribute = map[string]interface{}{}
	}
	correlation := correlation.ExtractCorrelationParam(ctx)
	if correlation != nil {
		headers := correlation.GetHeader()
		for key, val := range headers {
			attribute[key] = val
		}
	}
	messageAttributes := make(map[string]types.MessageAttributeValue, len(attribute))
	for key, value := range attribute {
		switch v := value.(type) {
		case string:
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v),
			}
		case []byte:
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("Binary"),
				BinaryValue: v,
			}
		default:
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(fmt.Sprintf("%v", v)),
			}
		}
	}
	return messageAttributes
}

// ReceiveMessage receives messages from the SQS queue.
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

// DeleteMessage deletes a message from the SQS queue.
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

// DeleteMessageBatch deletes multiple messages from the SQS queue.
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

// marshal converts the provided interface value to JSON string.
func marshal(val interface{}) (*string, error) {
	blob, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("aws.LoadString: %w", err)
	}
	str := string(blob)
	return &str, nil
}

package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// SNS provides methods to interact with AWS Simple Notification Service (SNS).
type SNS struct {
	_ struct{}
	*sns.Client
	log log.Log
}

// defaultSNSClient is the default AWS SNS client.
var defaultSNSClient *sns.Client

// GetDefaultSNSClient returns the default SNS client using the provided logger.
func GetDefaultSNSClient(logger log.Log) *SNS {
	if defaultSNSClient == nil {
		defaultSNSClient = NewSNSClientWithConfig(defaultAWSConfig)
	}
	return NewSNSClient(logger, defaultSNSClient)
}

// NewSNSClientWithConfig creates a new SNS client with the provided AWS configuration.
func NewSNSClientWithConfig(awsConfig *aws.Config) *sns.Client {
	client := sns.NewFromConfig(*awsConfig)
	return client
}

// NewSNSClient creates a new SNS instance with the provided logger and SNS client.
func NewSNSClient(logger log.Log, client *sns.Client) *SNS {
	return &SNS{Client: client, log: logger.NewResourceLogger("SNS")}
}

// Publish publishes a message to the specified SNS topic.
// It returns the publish output and an error if the operation fails.
func (s *SNS) Publish(ctx context.Context, topicArn, subject *string, payload *utils.Message, attributes map[string]interface{}) (*sns.PublishOutput, error) {
	blob, _ := json.Marshal(payload)
	message := string(blob)
	req := &sns.PublishInput{
		TopicArn:          topicArn,
		Subject:           subject,
		Message:           &message,
		MessageAttributes: s.GenerateAttribute(ctx, attributes),
	}
	res, err := s.Client.Publish(ctx, req)
	if err != nil {
		s.log.Error(ctx, "SNS publish error", err)
		return res, fmt.Errorf("SNS.Publish: error publishing event: %w", err)
	}
	return res, nil
}

// GenerateAttribute generates SNS message attributes based on the provided map.
// It extracts correlation parameters from the context and includes them in the message attributes.
func (s *SNS) GenerateAttribute(ctx context.Context, attribute map[string]interface{}) map[string]types.MessageAttributeValue {
	if attribute == nil {
		attribute = make(map[string]interface{})
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

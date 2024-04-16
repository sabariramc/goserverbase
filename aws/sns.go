package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
)

type SNS struct {
	_ struct{}
	*sns.Client
	log log.Log
}

var defaultSNSClient *sns.Client

func GetDefaultSNSClient(logger log.Log) *SNS {
	if defaultSNSClient == nil {
		defaultSNSClient = NewSNSClientWithConfig(defaultAWSConfig)
	}
	return NewSNSClient(logger, defaultSNSClient)
}

func NewSNSClientWithConfig(awsConfig *aws.Config) *sns.Client {
	client := sns.NewFromConfig(*awsConfig)
	return client
}

func NewSNSClient(logger log.Log, client *sns.Client) *SNS {
	return &SNS{Client: client, log: logger.NewResourceLogger("SNS")}
}

func (s *SNS) Publish(ctx context.Context, topicArn, subject *string, payload *utils.Message, attributes map[string]any) (*sns.PublishOutput, error) {
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

func (s *SNS) GenerateAttribute(ctx context.Context, attribute map[string]any) map[string]types.MessageAttributeValue {
	if attribute == nil {
		attribute = map[string]any{}
	}
	correlation := log.GetCorrelationParam(ctx)
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

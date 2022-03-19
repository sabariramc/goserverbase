package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"sabariram.com/goserverbase/log"
	"sabariram.com/goserverbase/utils"
)

type SNS struct {
	_      struct{}
	client *sns.SNS
	log    *log.Logger
}

var defaultSNSClient *sns.SNS

func GetDefaultSNSClient(logger *log.Logger) *SNS {
	if defaultSecretManagerClient == nil {
		defaultSNSClient = GetAWSSNSClient(defaultAWSSession)
	}
	return GetSNSClient(logger, defaultSNSClient)
}

func GetAWSSNSClient(awsSession *session.Session) *sns.SNS {
	client := sns.New(awsSession)
	return client
}

func GetSNSClient(logger *log.Logger, client *sns.SNS) *SNS {
	return &SNS{client: client, log: logger}
}

func GetSNSARN(ctx context.Context, logger *log.Logger, topicName string) (*string, error) {
	prefix := utils.Getenv("stage", "dev")
	systemPefix := utils.Getenv("snsTopicPrefix", "BEDROCK")
	if systemPefix != "" {
		prefix = fmt.Sprintf("%v_%v", prefix, systemPefix)
	}
	region := utils.GetenvMust("region")
	accountId := utils.GetenvMust("account_id")
	arn := fmt.Sprintf("arn:aws:sns:%v:%v:%v_%v", region, accountId, prefix, topicName)
	logger.Debug(ctx, "Topic Arn", arn)
	return &arn, nil
}

func (s *SNS) Publish(ctx context.Context, topicArn, subject *string, payload *utils.Message, attributes map[string]string) error {
	blob, _ := json.Marshal(payload)
	message := string(blob)
	req := &sns.PublishInput{
		TopicArn:          topicArn,
		Subject:           subject,
		Message:           &message,
		MessageAttributes: s.GetAttribure(attributes),
	}
	s.log.Debug(ctx, "SNS publish request", req)
	res, err := s.client.PublishWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "SNS publish error", err)
		return err
	}
	s.log.Debug(ctx, "SNS publish response", res)
	return nil
}

func (s *SNS) GetAttribure(attribute map[string]string) map[string]*sns.MessageAttributeValue {
	if len(attribute) == 0 {
		return nil
	}
	messageAttributes := make(map[string]*sns.MessageAttributeValue, len(attribute))
	for key, value := range attribute {
		messageAttributes[key] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	return messageAttributes
}

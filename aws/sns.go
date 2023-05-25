package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

type SNS struct {
	_ struct{}
	*sns.SNS
	log *log.Logger
}

var defaultSNSClient *sns.SNS

func GetDefaultSNSClient(logger *log.Logger) *SNS {
	if defaultSecretManagerClient == nil {
		defaultSNSClient = NewSNSClientWithSession(defaultAWSSession)
	}
	return NewSNSClient(logger, defaultSNSClient)
}

func NewSNSClientWithSession(awsSession *session.Session) *sns.SNS {
	client := sns.New(awsSession)
	return client
}

func NewSNSClient(logger *log.Logger, client *sns.SNS) *SNS {
	return &SNS{SNS: client, log: logger}
}

func (s *SNS) PublishWithContext(ctx context.Context, topicArn, subject *string, payload *utils.Message, attributes map[string]string) error {
	blob, _ := json.Marshal(payload)
	message := string(blob)
	req := &sns.PublishInput{
		TopicArn:          topicArn,
		Subject:           subject,
		Message:           &message,
		MessageAttributes: s.GetAttribute(attributes),
	}
	s.log.Debug(ctx, "SNS publish request", req)
	res, err := s.SNS.PublishWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "SNS publish error", err)
		return fmt.Errorf("SNS.Publish: %w", err)
	}
	s.log.Debug(ctx, "SNS publish response", res)
	return nil
}

func (s *SNS) GetAttribute(attribute map[string]string) map[string]*sns.MessageAttributeValue {
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

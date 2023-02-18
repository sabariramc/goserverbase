package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/sabariramc/goserverbase/log"
)

type StepFunction struct {
	*sfn.SFN
	log *log.Logger
}

var defaultSFNClient *sfn.SFN

func GetDefaultSFNClient(logger *log.Logger) *StepFunction {
	if defaultSFNClient == nil {
		defaultSFNClient = NewAWSSFNClient(defaultAWSSession)
	}
	return NewSFNClient(logger, defaultSFNClient)
}

func NewAWSSFNClient(awsSession *session.Session) *sfn.SFN {
	client := sfn.New(awsSession)
	return client
}

func NewSFNClient(logger *log.Logger, sfnClient *sfn.SFN) *StepFunction {
	return &StepFunction{SFN: sfnClient, log: logger}
}

func (s *StepFunction) StartExecutionWithContext(ctx context.Context, stateMachineArn, executionName string, payload interface{}) (err error) {
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "State machine payload marshal error", err)
		return
	}
	stringifiedMarshalledPayload := string(marshalledPayload)
	s.log.Info(ctx, "Starting execution of state machine", map[string]string{
		"arn":     stateMachineArn,
		"payload": stringifiedMarshalledPayload,
		"name":    executionName,
	})
	res, err := s.SFN.StartExecutionWithContext(ctx, &sfn.StartExecutionInput{
		Input:           &stringifiedMarshalledPayload,
		Name:            &executionName,
		StateMachineArn: &stateMachineArn,
	})
	if err != nil {
		s.log.Error(ctx, "State machine start execution error", err)
		err = fmt.Errorf("StepFunction.StartExecution: %w", err)
		return
	}
	s.log.Debug(ctx, "State machine start execution response", res)
	return
}

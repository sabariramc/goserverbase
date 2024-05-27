package kafka_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/sabariramc/goserverbase/v6/correlation"
	pKafka "github.com/sabariramc/goserverbase/v6/kafka"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier/kafka"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"gotest.tools/assert"
)

var TestLogger log.Log

const ServiceName = "KafkaNotifierTest"

func init() {
	testutils.Initialize()
	testutils.LoadEnv("../../.env")
	TestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

func TestKafkaNotifier(t *testing.T) {
	p, _ := pKafka.NewProducer()
	notifier := kafka.New(kafka.WithProducer(p))
	ctx := GetCorrelationContext()
	custID := "cust_test_id"
	errorCode := "com.testing.error"
	stackTraceBuff := []byte{}
	runtime.Stack(stackTraceBuff, false)
	stackTrace := string(stackTraceBuff)
	errorData := map[string]any{"check": "Testing error"}
	err := notifier.Notify4XX(correlation.GetContextWithUserIdentifier(ctx, &correlation.UserIdentifier{UserID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
	custID = "app_user_id"
	err = notifier.Notify4XX(correlation.GetContextWithUserIdentifier(ctx, &correlation.UserIdentifier{AppUserID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
	custID = "entity_id"
	err = notifier.Notify4XX(correlation.GetContextWithUserIdentifier(ctx, &correlation.UserIdentifier{EntityID: &custID}), errorCode, nil, stackTrace, errorData)
	assert.NilError(t, err)
}

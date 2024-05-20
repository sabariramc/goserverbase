package logwriter_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/trace"
)

func TestConsoleLogWriter(t *testing.T) {
	ctx := trace.GetContextWithCorrelationParam(context.Background(), &trace.CorrelationParam{
		CorrelationID: "test console log",
	})
	log := log.New()
	log.AddLogWriter(ctx, logwriter.NewConsoleWriter())
	for i := 0; i < 1000000; i++ {
		log.Error(ctx, uuid.NewString())
	}
}

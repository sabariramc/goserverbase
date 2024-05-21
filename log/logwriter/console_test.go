package logwriter_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/correlation"
)

func TestConsoleLogWriter(t *testing.T) {
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{
		CorrelationID: "test console log",
	})
	log := log.New()
	log.AddLogWriter(ctx, logwriter.NewConsoleWriter())
	for i := 0; i < 1000000; i++ {
		log.Error(ctx, uuid.NewString())
	}
}

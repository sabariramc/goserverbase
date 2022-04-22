package logwriter

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
)

func GetCorrelationParam(ctx context.Context) *log.CorrelationParmas {
	val, ok := ctx.Value(constant.CorrelationContextKey).(*log.CorrelationParmas)
	if !ok {
		val = &log.CorrelationParmas{}
	}
	return val
}

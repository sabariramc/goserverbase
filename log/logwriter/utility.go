package logwriter

import (
	"context"

	"github.com/sabariramc/goserverbase/constant"
	"github.com/sabariramc/goserverbase/log"
)

func GetCorrelationParam(ctx context.Context) *log.CorrelationParam {
	val, ok := ctx.Value(constant.CorrelationContextKey).(*log.CorrelationParam)
	if !ok {
		val = &log.CorrelationParam{}
	}
	return val
}

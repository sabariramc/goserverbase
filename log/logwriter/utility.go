package logwriter

import (
	"context"

	"sabariram.com/goserverbase/constant"
	"sabariram.com/goserverbase/log"
)

const ParseErrorMsg = "******************ERROR DURING MARSHAL OF FULLMESSAGE*******************"

func GetCorrelationParam(ctx context.Context) *log.CorrelationParmas {
	val, ok := ctx.Value(constant.CorrelationContextKey).(*log.CorrelationParmas)
	if !ok {
		val = &log.CorrelationParmas{}
	}
	return val
}

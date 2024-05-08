package log

import (
	"context"
	"net/http"
)

type ContextKey struct{}

var (
	ContextKeyCorrelation        ContextKey = ContextKey{}
	ContextKeyCustomerIdentifier ContextKey = ContextKey{}
)

// GetContextWithCorrelationParam returns a context.Context with CorrelationParam
func GetContextWithCorrelationParam(ctx context.Context, c *CorrelationParam) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCorrelation, c)
	return ctx
}

// GetContextWithUserIdentifier returns a context.Context with UserIdentifier
func GetContextWithUserIdentifier(ctx context.Context, c *UserIdentifier) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCustomerIdentifier, c)
	return ctx
}

// ExtractUserIdentifier returns UserIdentifier from within context.Context
func ExtractUserIdentifier(ctx context.Context) *UserIdentifier {
	iVal := ctx.Value(ContextKeyCustomerIdentifier)
	if iVal == nil {
		return &UserIdentifier{}
	}
	val, ok := iVal.(*UserIdentifier)
	if !ok {
		return &UserIdentifier{}
	}
	return val
}

// ExtractCorrelationParam returns CorrelationParam from within context.Context
func ExtractCorrelationParam(ctx context.Context) *CorrelationParam {
	iVal := ctx.Value(ContextKeyCorrelation)
	if iVal == nil {
		return &CorrelationParam{}
	}
	val, ok := iVal.(*CorrelationParam)
	if !ok {
		return &CorrelationParam{}
	}
	return val
}

// SetCorrelationHeader adds CorrelationParam and UserIdentifier available in the context.Context in http.Request.Header marshalled with header struct tag
func SetCorrelationHeader(ctx context.Context, req *http.Request) {
	headers := GetHeader(ctx)
	for i, v := range headers {
		req.Header.Add(i, v)
	}
}

// GetHeader returns CorrelationParam and UserIdentifier available in the context.Context marshalled with header struct tag
func GetHeader(ctx context.Context) map[string]string {
	headers := make(map[string]string, 10)
	corr := ExtractCorrelationParam(ctx).GetHeader()
	identity := ExtractUserIdentifier(ctx).GetHeader()
	for k, v := range corr {
		if v != "" {
			headers[k] = v
		}
	}
	for k, v := range identity {
		if v != "" {
			headers[k] = v
		}
	}
	return headers
}

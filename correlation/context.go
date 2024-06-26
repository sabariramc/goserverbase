package correlation

import (
	"context"
	"net/http"
)

type contextKey string

var (
	ContextKeyCorrelation        contextKey = contextKey("ContextKeyCorrelation")
	ContextKeyCustomerIdentifier contextKey = contextKey("ContextKeyCustomerIdentifier")
)

// GetContextWithCorrelationParam returns a context.Context with the provided CorrelationParam.
func GetContextWithCorrelationParam(ctx context.Context, c *CorrelationParam) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCorrelation, c)
	return ctx
}

// GetContextWithUserIdentifier returns a context.Context with the provided UserIdentifier.
func GetContextWithUserIdentifier(ctx context.Context, c *UserIdentifier) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCustomerIdentifier, c)
	return ctx
}

// ExtractUserIdentifier retrieves the UserIdentifier stored within the context.Context.
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

// ExtractCorrelationParam retrieves the CorrelationParam stored within the context.Context.
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

// SetCorrelationHeader adds the CorrelationParam and UserIdentifier from the context.Context into the http.Request Header.
// These values are marshalled with the header struct tag.
func SetCorrelationHeader(ctx context.Context, req *http.Request) {
	headers := GetHeader(ctx)
	for i, v := range headers {
		req.Header.Add(i, v)
	}
}

// GetHeader retrieves the CorrelationParam and UserIdentifier from the context.Context and marshals them with header struct tags.
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

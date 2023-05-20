package httpserver

type ContextKey string

const (
	HttpContentTypeJSON    = "application/json"
	HttpHeaderContentType  = "Content-Type"
	ContextKeyHandlerError = ContextKey("ContextKeyHandlerError")
	ContextKeyRequestBody  = ContextKey("ContextKeyRequestBody")
)
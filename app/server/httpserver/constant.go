package httpserver

type ContextKey string

const (
	HttpContentTypeJSON              = "application/json"
	HttpHeaderContentType            = "Content-Type"
	ContextKeyRequestBody            = ContextKey("ContextKeyRequestBody")
	ContextKeyHandlerError           = ContextKey("ContextKeyHandlerError")
	ContextKeyHandlerErrorStackTrace = ContextKey("ContextKeyHandlerErrorStackTrace")
)

package httpserver

type ContextKey string

const (
	HttpContentTypeJSON   = "application/json"
	HttpHeaderContentType = "Content-Type"
	ContextKeyRequestBody = ContextKey("ContextKeyRequestBody")
)

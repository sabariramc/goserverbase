package baseapp

import (
	"context"
	"net/http"
)

type ErrorNotifier interface {
	Send(ctx context.Context, r *http.Request, err error)
}

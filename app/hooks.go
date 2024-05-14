package baseapp

import "context"

// ShutdownHook defines interface for a graceful shutdown of different resources used by the app
type ShutdownHook interface {
	Name(ctx context.Context) string
	Shutdown(ctx context.Context) error
}

// HealthCheckHook defines interface for health check of different resources used by the app
type HealthCheckHook interface {
	Name(ctx context.Context) string
	HealthCheck(ctx context.Context) error
}

// StatusCheckHook defines interface to get the current status of different resources used by the app
type StatusCheckHook interface {
	Name(ctx context.Context) string
	StatusCheck(ctx context.Context) (any, error)
}

func (b *BaseApp) RegisterHooks(hook any) {
	if hHook, ok := hook.(HealthCheckHook); ok {
		b.RegisterHealthCheckHook(hHook)
	}
	if sHook, ok := hook.(ShutdownHook); ok {
		b.RegisterOnShutdownHook(sHook)
	}
	if sHook, ok := hook.(StatusCheckHook); ok {
		b.RegisterStatusCheckHook(sHook)
	}
}

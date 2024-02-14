package baseapp

import "context"

type ShutdownHook interface {
	Name(ctx context.Context) string
	Shutdown(ctx context.Context) error
}

type HealthCheckHook interface {
	Name(ctx context.Context) string
	HealthCheck(ctx context.Context) error
}

func (b *BaseApp) RegisterHooks(hook any) {
	if hHook, ok := hook.(HealthCheckHook); ok {
		b.RegisterHealthCheckHook(hHook)
	}
	if sHook, ok := hook.(ShutdownHook); ok {
		b.RegisterOnShutdownHook(sHook)
	}
}

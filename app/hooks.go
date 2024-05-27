package baseapp

import "context"

// Name defines an interface to retrieve the module's identity.
//
// Implementing this interface requires the following method:
//   - Name(ctx context.Context) string
type Name interface {
	Name(ctx context.Context) string
}

// ShutdownHook defines an interface for the graceful shutdown of different resources used by the app.
//
// This interface extends the Name interface and requires the following method:
//   - Shutdown(ctx context.Context) error
type ShutdownHook interface {
	Name
	Shutdown(ctx context.Context) error
}

// HealthCheckHook defines an interface for health checks of different resources used by the app.
//
// This interface extends the Name interface and requires the following method:
//   - HealthCheck(ctx context.Context) error
type HealthCheckHook interface {
	Name
	HealthCheck(ctx context.Context) error
}

// StatusCheckHook defines an interface to get the current status of different resources used by the app.
//
// This interface extends the Name interface and requires the following method:
//   - StatusCheck(ctx context.Context) (any, error)
type StatusCheckHook interface {
	Name
	StatusCheck(ctx context.Context) (any, error)
}

// RegisterHooks registers the provided hooks to the BaseApp.
//
// This function checks the type of the provided hook and registers it as a HealthCheckHook, ShutdownHook, or StatusCheckHook if it implements the respective interface.
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

package baseapp

import (
	"context"
	"fmt"
)

func (b *BaseApp) RegisterHealthCheckHook(handler HealthCheckHook) {
	b.healthHooks = append(b.healthHooks, handler)
}

func (b *BaseApp) RunHealthCheck(ctx context.Context) error {
	b.log.Info(ctx, "Starting health check", nil)
	n := len(b.healthHooks)
	for i, hook := range b.healthHooks {
		b.log.Info(ctx, fmt.Sprintf("Running health check %v of %v : %v", i+1, n, hook.Name(ctx)), nil)
		err := hook.HealthCheck(ctx)
		if err != nil {
			b.log.Error(ctx, "health check failed for hook: "+hook.Name(ctx), err)
			return fmt.Errorf("BaseApp.HealthCheck: %w", err)
		}
		b.log.Info(ctx, fmt.Sprintf("Completed health check %v of %v : %v", i+1, n, hook.Name(ctx)), nil)
	}
	b.log.Info(ctx, "Completed health check", nil)
	return nil
}

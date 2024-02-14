package baseapp

import (
	"context"
	"fmt"
)

func (b *BaseApp) RegisterHealthCheckHook(handler HealthCheckHook) {
	b.healthHooks = append(b.healthHooks, handler)
}

func (b *BaseApp) RunHealthCheck(ctx context.Context) error {
	for _, han := range b.healthHooks {
		err := han.HealthCheck(ctx)
		if err != nil {
			b.log.Error(ctx, "health check failed for hook"+han.Name(ctx), err)
			return fmt.Errorf("BaseApp.HealthCheck: %w", err)
		}
	}
	return nil
}

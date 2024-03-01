package baseapp

import (
	"context"
	"fmt"
	"time"
)

func (b *BaseApp) RegisterHealthCheckHook(handler HealthCheckHook) {
	b.healthHooks = append(b.healthHooks, handler)
}

func (b *BaseApp) RunHealthCheck(ctx context.Context) error {
	b.log.Debug(ctx, "Starting health check", nil)
	n := len(b.healthHooks)
	for i, hook := range b.healthHooks {
		name := hook.Name(ctx)
		b.log.Info(ctx, fmt.Sprintf("Running health check %v of %v : %v", i+1, n, name), nil)
		hookCtx, _ := context.WithTimeout(ctx, time.Second)
		result := make(chan error)
		go func() {
			result <- hook.HealthCheck(hookCtx)
		}()
		var err error
		select {
		case <-hookCtx.Done():
			err = context.DeadlineExceeded
		case err = <-result:
		}
		if err != nil {
			b.log.Error(ctx, "health check failed for hook: "+name, err)
			return fmt.Errorf("BaseApp.HealthCheck: %w", err)
		}
		b.log.Info(ctx, fmt.Sprintf("Completed health check %v of %v : %v", i+1, n, name), nil)
	}
	b.log.Debug(ctx, "Completed health check", nil)
	return nil
}

func (b *BaseApp) RegisterStatusCheckHook(handler StatusCheckHook) {
	b.statusHooks = append(b.statusHooks, handler)
}

func (b *BaseApp) RunStatusCheck(ctx context.Context) map[string]any {
	b.log.Debug(ctx, "Starting status check", nil)
	n := len(b.statusHooks)
	res := map[string]any{}
	for i, hook := range b.statusHooks {
		name := hook.Name(ctx)
		b.log.Info(ctx, fmt.Sprintf("Running status check %v of %v : %v", i+1, n, name), nil)
		hookCtx, _ := context.WithTimeout(ctx, time.Second)
		var err error
		var status any
		result := make(chan bool)
		go func() {
			status, err = hook.StatusCheck(hookCtx)
			result <- true
		}()
		select {
		case <-hookCtx.Done():
			err = context.DeadlineExceeded
		case <-result:
		}
		if err != nil {
			status = map[string]string{
				"status": "failed",
				"error":  err.Error(),
			}
			b.log.Error(ctx, "status check failed for hook: "+name, err)
		} else {
			status = map[string]any{
				"status": "success",
				"data":   status,
			}
		}
		res[name] = status
		b.log.Info(ctx, fmt.Sprintf("Completed status check %v of %v : %v", i+1, n, name), nil)
	}
	b.log.Debug(ctx, "Completed status check", nil)
	return res
}

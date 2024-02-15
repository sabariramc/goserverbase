package kafkaconsumer

import (
	"context"
	"os"
	"time"
)

func (k *KafkaConsumerServer) HealthCheckMonitor(ctx context.Context) {
	go func() {
		timeoutContext, _ := context.WithTimeout(ctx, time.Second*time.Duration(k.c.HealthCheckInSec))
		defer k.log.Warning(ctx, "Health check monitor stopped", nil)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timeoutContext.Done():
				err := k.RunHealthCheck(ctx)
				if err != nil {
					deleteErr := os.Remove(k.c.HealthFilePath)
					if deleteErr != nil {
						k.log.Error(ctx, "error deleting health file", deleteErr)
					}
					k.log.Emergency(ctx, "health check failed", err, nil)
				}
				timeoutContext, _ = context.WithTimeout(ctx, time.Second*time.Duration(k.c.HealthCheckInSec))
			}
		}
	}()
}

func (k *KafkaConsumerServer) HealthCheck(ctx context.Context) error {
	k.client.Stats()
	return nil
}
